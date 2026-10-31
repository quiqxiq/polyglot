package genieacs

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// nbiRequest is a snapshot of one HTTP request received by a test NBI server.
// Assertions run on the test goroutine against snapshots instead of inside the
// server handler, which is safer with t.FailNow-based helpers.
type nbiRequest struct {
	Method string
	Path   string
	Query  url.Values
	Header http.Header
	Body   []byte
}

// bodyJSON decodes a JSON request body.
func (r nbiRequest) bodyJSON(t *testing.T) map[string]any {
	t.Helper()
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(r.Body, &decoded))
	return decoded
}

// requestRecorder records requests handled by a test NBI server. It is safe
// for concurrent use because handlers run on server goroutines.
type requestRecorder struct {
	mu       sync.Mutex
	requests []nbiRequest
}

func (rec *requestRecorder) record(r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	snapshot := nbiRequest{
		Method: r.Method,
		Path:   r.URL.Path,
		Query:  r.URL.Query(),
		Header: r.Header.Clone(),
		Body:   body,
	}
	rec.mu.Lock()
	defer rec.mu.Unlock()
	rec.requests = append(rec.requests, snapshot)
}

func (rec *requestRecorder) count() int {
	rec.mu.Lock()
	defer rec.mu.Unlock()
	return len(rec.requests)
}

func (rec *requestRecorder) last() nbiRequest {
	rec.mu.Lock()
	defer rec.mu.Unlock()
	if len(rec.requests) == 0 {
		return nbiRequest{}
	}
	return rec.requests[len(rec.requests)-1]
}

// newTestNBI starts an httptest server that records every request and delegates
// the response to respond, then returns an NBI client pointed at it.
func newTestNBI(t *testing.T, recorder *requestRecorder, respond func(w http.ResponseWriter, r *http.Request)) *NBI {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder.record(r)
		respond(w, r)
	}))
	t.Cleanup(server.Close)

	client, err := NewNBI(server.URL, "test-api-key", 5*time.Second)
	require.NoError(t, err)
	return client
}

// failResponse answers any request with a server error; used by tests that
// assert no request was made.
func failResponse(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}
