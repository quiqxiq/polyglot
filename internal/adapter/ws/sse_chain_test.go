package ws

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/quixiq/polyglot/internal/adapter/http/middleware"
)

// TestSSEThroughMiddlewareChain replicates the production wiring from
// internal/app/app.go (Recovery + RequestLogger + CORS around the root mux)
// and asserts GET /events still streams. Regression for the 500
// "Streaming unsupported" caused by the logging wrapper dropping
// http.Flusher.
func TestSSEThroughMiddlewareChain(t *testing.T) {
	h := NewSSEHub()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /events", h.ServeHTTP)

	handler := middleware.Chain(mux,
		middleware.Recovery(),
		middleware.RequestLogger(),
		middleware.CORS(nil, "development"),
	)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/events")
	if err != nil {
		t.Fatalf("open SSE connection: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /events status = %d, want 200 (previous bug: 500 Streaming unsupported)", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("content-type = %q, want text/event-stream", ct)
	}

	reader := bufio.NewReader(resp.Body)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read first SSE line: %v", err)
	}
	if !strings.HasPrefix(line, "event: ready") {
		t.Fatalf("first SSE line = %q, want %q", line, "event: ready")
	}
}
