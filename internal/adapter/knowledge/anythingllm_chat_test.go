package knowledge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewChatClient(t *testing.T) {
	if _, err := NewChatClient("", "", "netops"); err == nil {
		t.Fatal("expected error for empty API key")
	}
	if _, err := NewChatClient("", "key", ""); err == nil {
		t.Fatal("expected error for empty workspace slug")
	}
	c, err := NewChatClient("http://example.com:3001/", "key", "netops")
	if err != nil {
		t.Fatalf("NewChatClient: %v", err)
	}
	if c.baseURL != "http://example.com:3001" {
		t.Fatalf("expected trailing slash trimmed, got %q", c.baseURL)
	}
}

func TestChatClientHappyPath(t *testing.T) {
	var gotMethod, gotPath, gotAuth string
	var gotBody chatRequest
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "chat-1",
			"type": "textResponse",
			"close": true,
			"error": null,
			"textResponse": "Paket GNET 100 Mbps Rp 525.000/bulan.",
			"sources": [{"title": "paket-internet.md"}, {"title": "promo.md"}],
			"metrics": {"prompt_tokens": 120, "completion_tokens": 45}
		}`))
	}))
	defer ts.Close()

	client, err := NewChatClient(ts.URL, "dev-key", "netops")
	if err != nil {
		t.Fatalf("NewChatClient: %v", err)
	}
	res, err := client.Chat(context.Background(), "berapa harga paket 100?", "conv-7")
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("expected POST, got %s", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/api/v1/workspace/netops/chat") {
		t.Fatalf("unexpected path %q", gotPath)
	}
	if gotAuth != "Bearer dev-key" {
		t.Fatalf("unexpected auth %q", gotAuth)
	}
	if gotBody.Mode != "chat" || gotBody.Message != "berapa harga paket 100?" || gotBody.SessionID != "conv-7" {
		t.Fatalf("unexpected request body: %+v", gotBody)
	}
	if res.Content != "Paket GNET 100 Mbps Rp 525.000/bulan." {
		t.Fatalf("unexpected content %q", res.Content)
	}
	if len(res.Sources) != 2 || res.Sources[0] != "paket-internet.md" {
		t.Fatalf("unexpected sources: %v", res.Sources)
	}
	if res.TokenIn != 120 || res.TokenOut != 45 {
		t.Fatalf("unexpected tokens: %+v", res)
	}
}

func TestChatClientErrorField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "chat-1", "type": "abort", "textResponse": null,
			"sources": [], "close": true, "error": "Something went wrong"
		}`))
	}))
	defer ts.Close()

	client, _ := NewChatClient(ts.URL, "key", "netops")
	_, err := client.Chat(context.Background(), "halo", "")
	if err == nil || !strings.Contains(err.Error(), "Something went wrong") {
		t.Fatalf("expected abort error, got %v", err)
	}
}

func TestChatClientEmptyResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "chat-1", "type": "textResponse", "textResponse": "",
			"sources": [], "close": true, "error": null
		}`))
	}))
	defer ts.Close()

	client, _ := NewChatClient(ts.URL, "key", "netops")
	if _, err := client.Chat(context.Background(), "halo", ""); err == nil {
		t.Fatal("expected error for empty textResponse")
	}
}

func TestChatClientNon200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "invalid workspace"}`))
	}))
	defer ts.Close()

	client, _ := NewChatClient(ts.URL, "key", "netops")
	_, err := client.Chat(context.Background(), "halo", "")
	if err == nil || !strings.Contains(err.Error(), "400") {
		t.Fatalf("expected 400 error, got %v", err)
	}
}

func TestChatClientNetworkError(t *testing.T) {
	client, _ := NewChatClient("http://127.0.0.1:1", "key", "netops")
	if _, err := client.Chat(context.Background(), "halo", ""); err == nil {
		t.Fatal("expected network error")
	}
}

func TestChatClientEmptyMessage(t *testing.T) {
	client, _ := NewChatClient("http://127.0.0.1:1", "key", "netops")
	if _, err := client.Chat(context.Background(), "  ", "conv-1"); err == nil {
		t.Fatal("expected error for empty message")
	}
}
