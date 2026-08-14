package knowledge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestServer returns an httptest server that records the last vector-search
// request and serves a canned response.
func newTestServer(t *testing.T, status int, respBody any) (*httptest.Server, *string) {
	t.Helper()
	var lastBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/workspace/netops/vector-search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("expected Bearer auth, got %q", got)
		}
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		lastBody = string(buf)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if respBody != nil {
			_ = json.NewEncoder(w).Encode(respBody)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &lastBody
}

func TestNewRetriever(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		apiKey      string
		workspace   string
		wantErrText string
	}{
		{"valid", "http://localhost:3001", "key", "netops", ""},
		{"empty api key", "http://localhost:3001", "", "netops", "API key is required"},
		{"empty workspace", "http://localhost:3001", "key", "", "workspace slug is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewRetriever(tt.baseURL, tt.apiKey, tt.workspace, 0)
			if tt.wantErrText != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrText) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErrText, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if r == nil {
				t.Fatal("expected non-nil retriever")
			}
			if r.baseURL != "http://localhost:3001" || r.topN != DefaultTopN {
				t.Fatalf("defaults not applied: baseURL=%s topN=%d", r.baseURL, r.topN)
			}
		})
	}
}

func TestRetrieverRetrieve(t *testing.T) {
	respBody := map[string]any{
		"results": []map[string]any{
			{
				"text": "Paket GNET 20 Mbps Rp250.000/bulan.",
				"metadata": map[string]any{
					"title":       "paket-gnet.md",
					"chunkSource": "paket-gnet.md",
				},
				"score": 0.92,
			},
			{
				"text": "Cara cek tagihan via WhatsApp.",
				"metadata": map[string]any{
					"chunkSource": "tagihan.md",
				},
				"score": 0.81,
			},
			{
				"text":     "",
				"metadata": map[string]any{"title": "kosong.md"},
				"score":    0.5,
			},
		},
	}
	srv, lastBody := newTestServer(t, http.StatusOK, respBody)
	r, err := NewRetriever(srv.URL, "test-key", "netops", 6)
	if err != nil {
		t.Fatal(err)
	}

	entries, err := r.Retrieve(context.Background(), "berapa harga paket?")
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}

	// Query terkirim ke endpoint vector-search.
	if *lastBody == "" {
		t.Fatal("expected request body to be sent")
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries (empty text chunk skipped), got %d", len(entries))
	}
	if entries[0].Title != "paket-gnet.md" || entries[0].Content != "Paket GNET 20 Mbps Rp250.000/bulan." {
		t.Fatalf("entry 0 mapping wrong: %+v", entries[0])
	}
	// Tanpa metadata.title, fallback ke chunkSource.
	if entries[1].Title != "tagihan.md" {
		t.Fatalf("expected fallback title from chunkSource, got %q", entries[1].Title)
	}
}

func TestRetrieverRetrieveEmptyQuery(t *testing.T) {
	r, err := NewRetriever("http://localhost:3001", "test-key", "netops", 0)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := r.Retrieve(context.Background(), "   ")
	if err != nil {
		t.Fatalf("expected no error for empty query, got %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no entries for empty query, got %d", len(entries))
	}
}

func TestRetrieverRetrieveNoResults(t *testing.T) {
	srv, _ := newTestServer(t, http.StatusOK, map[string]any{"results": []any{}})
	r, err := NewRetriever(srv.URL, "test-key", "netops", 0)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := r.Retrieve(context.Background(), "tidak ada di dokumen")
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no entries, got %d", len(entries))
	}
}

func TestRetrieverRetrieveHTTPError(t *testing.T) {
	srv, _ := newTestServer(t, http.StatusUnauthorized, map[string]any{"message": "invalid api key"})
	r, err := NewRetriever(srv.URL, "test-key", "netops", 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Retrieve(context.Background(), "harga"); err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

func TestRetrieverRetrieveMalformedBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}))
	t.Cleanup(srv.Close)

	r, err := NewRetriever(srv.URL, "test-key", "netops", 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Retrieve(context.Background(), "harga"); err == nil {
		t.Fatal("expected error for malformed response body")
	}
}
