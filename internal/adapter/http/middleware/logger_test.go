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
