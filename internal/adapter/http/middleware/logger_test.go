package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/quixiq/polyglot/pkg/logger"
)

// TestRequestLoggerPreservesStreamingInterfaces guards against regressions
// where the logging wrapper drops optional ResponseWriter interfaces. The
// SSE endpoint (internal/adapter/ws/sse_hub.go) asserts w.(http.Flusher)
// directly, and coder/websocket follows Unwrap() to find the Hijacker — both
// must survive the middleware chain or /events and /ws/* return 500.
func TestRequestLoggerPreservesStreamingInterfaces(t *testing.T) {
	var sawFlusher bool
	var unwrapped *httptest.ResponseRecorder

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, sawFlusher = w.(http.Flusher)
		if uw, ok := w.(interface{ Unwrap() http.ResponseWriter }); ok {
			unwrapped, _ = uw.Unwrap().(*httptest.ResponseRecorder)
		}
		w.WriteHeader(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/events", nil)

	Chain(handler, RequestLogger()).ServeHTTP(rec, req)

	if !sawFlusher {
		t.Error("RequestLogger wrapper no longer implements http.Flusher — SSE /events will 500")
	}
	if unwrapped == nil {
		t.Error("RequestLogger wrapper no longer exposes Unwrap() — websocket hijack will fail")
	}
}

func TestRequestIDPreservesIncomingID(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set(requestIDHeader, "client-request-42")

	Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := RequestIDFromContext(r.Context()); got != "client-request-42" {
			t.Fatalf("request ID = %q, want client-request-42", got)
		}
	}), RequestID()).ServeHTTP(rec, req)

	if got := rec.Header().Get(requestIDHeader); got != "client-request-42" {
		t.Fatalf("response request ID = %q, want client-request-42", got)
	}
}

func TestRequestIDGeneratesMissingID(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if RequestIDFromContext(r.Context()) == "" {
			t.Fatal("request ID is empty")
		}
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(rec, req)

	if rec.Header().Get(requestIDHeader) == "" {
		t.Fatal("response request ID is empty")
	}
}

func TestRequestLoggerUnaryOutput(t *testing.T) {
	var buf bytes.Buffer
	logger.Init("info", "production", &buf)
	t.Cleanup(func() { logger.Init("info", "development") })

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set(requestIDHeader, "req-unary-101")

	Chain(handler, RequestID(), RequestLogger()).ServeHTTP(rec, req)

	output := buf.String()
	if !strings.Contains(output, "request completed") {
		t.Errorf("expected 'request completed' in log, got: %s", output)
	}
	if !strings.Contains(output, `"request_id":"req-unary-101"`) {
		t.Errorf("expected request_id in log, got: %s", output)
	}
	if !strings.Contains(output, `"path":"/api/v1/health"`) {
		t.Errorf("expected path in log, got: %s", output)
	}
	if !strings.Contains(output, `"status":200`) {
		t.Errorf("expected status 200 in log, got: %s", output)
	}
}

func TestRequestLoggerStreamingLifecycle(t *testing.T) {
	var buf bytes.Buffer
	logger.Init("info", "production", &buf)
	t.Cleanup(func() { logger.Init("info", "development") })

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("data: ping\n\n"))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set(requestIDHeader, "req-stream-202")

	Chain(handler, RequestID(), RequestLogger()).ServeHTTP(rec, req)

	output := buf.String()
	if !strings.Contains(output, "streaming started") {
		t.Errorf("expected 'streaming started' in log, got: %s", output)
	}
	if !strings.Contains(output, "streaming completed") {
		t.Errorf("expected 'streaming completed' in log, got: %s", output)
	}
	if !strings.Contains(output, `"request_id":"req-stream-202"`) {
		t.Errorf("expected request_id in log, got: %s", output)
	}
}
