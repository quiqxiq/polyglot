package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
