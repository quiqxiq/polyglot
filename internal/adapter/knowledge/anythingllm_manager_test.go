package knowledge

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type recordedCall struct {
	method string
	path   string
	auth   string
	body   map[string]any
}

func recordCalls(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *[]recordedCall) {
	t.Helper()
	var calls []recordedCall
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call := recordedCall{method: r.Method, path: r.URL.Path, auth: r.Header.Get("Authorization")}
		if r.Body != nil {
			if raw, err := io.ReadAll(r.Body); err == nil && len(raw) > 0 {
				_ = json.Unmarshal(raw, &call.body)
			}
		}
		calls = append(calls, call)
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

func writeJSON(w http.ResponseWriter, status int, payload string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(payload))
}

func TestNewManager(t *testing.T) {
	tests := []struct {
		name          string
		baseURL       string
		apiKey        string
		workspaceSlug string
		wantErr       bool
		wantBaseURL   string
	}{
		{"empty api key", "", "", "netops", true, ""},
		{"empty workspace slug", "", "secret-key", "", true, ""},
		{"default base url", "", "secret-key", "netops", false, DefaultBaseURL},
		{"trailing slash trimmed", "http://localhost:3001/", "secret-key", "netops", false, "http://localhost:3001"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewManager(tt.baseURL, tt.apiKey, tt.workspaceSlug)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantBaseURL, m.baseURL)
			require.Equal(t, DefaultDocumentFolder, m.folder)
		})
	}
}

func TestManagerUpsertDocument(t *testing.T) {
	srv, calls := recordCalls(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, `{"success":true,"error":null,"documents":[{"location":"custom-documents/raw-harga-paket-abc123.json","title":"Harga Paket"}]}`)
	})

	m, err := NewManager(srv.URL, "secret-key", "netops")
	require.NoError(t, err)

	loc, err := m.UpsertDocument(context.Background(), "", "Harga Paket", "# Harga Paket\nPaket 20 Mbps Rp250.000/bulan")
	require.NoError(t, err)
	require.Equal(t, "custom-documents/raw-harga-paket-abc123.json", loc)

	require.Len(t, *calls, 1)
	call := (*calls)[0]
	require.Equal(t, http.MethodPost, call.method)
	require.Equal(t, "/api/v1/document/raw-text", call.path)
	require.Equal(t, "Bearer secret-key", call.auth)
	require.Equal(t, "# Harga Paket\nPaket 20 Mbps Rp250.000/bulan", call.body["textContent"])
	require.Equal(t, "netops", call.body["addToWorkspaces"])
	metadata, ok := call.body["metadata"].(map[string]any)
	require.True(t, ok, "metadata harus object JSON")
	require.Equal(t, "Harga Paket", metadata["title"])
}

func TestManagerUpsertDocumentReplacesExisting(t *testing.T) {
	srv, calls := recordCalls(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/system/remove-documents":
			writeJSON(w, http.StatusOK, `{}`)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/document/raw-text":
			writeJSON(w, http.StatusOK, `{"success":true,"documents":[{"location":"custom-documents/raw-baru-456.json"}]}`)
		default:
			writeJSON(w, http.StatusNotFound, `{"success":false}`)
		}
	})

	m, err := NewManager(srv.URL, "secret-key", "netops")
	require.NoError(t, err)

	loc, err := m.UpsertDocument(context.Background(), "custom-documents/raw-lama-789.json", "Judul", "isi baru")
	require.NoError(t, err)
	require.Equal(t, "custom-documents/raw-baru-456.json", loc)

	require.Len(t, *calls, 2)
	del := (*calls)[0]
	require.Equal(t, http.MethodDelete, del.method)
	require.Equal(t, "/api/v1/system/remove-documents", del.path)
	names, ok := del.body["names"].([]any)
	require.True(t, ok)
	require.Equal(t, []any{"custom-documents/raw-lama-789.json"}, names)

	up := (*calls)[1]
	require.Equal(t, http.MethodPost, up.method)
	require.Equal(t, "/api/v1/document/raw-text", up.path)
}

func TestManagerUpsertDocumentDeleteFailureStillUploads(t *testing.T) {
	// Dokumen lama yang sudah dihapus manual dari UI AnythingLLM membuat
	// remove-documents error — upload tetap harus dilanjutkan (best-effort).
	srv, calls := recordCalls(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete:
			writeJSON(w, http.StatusInternalServerError, `{"success":false,"error":"document not found"}`)
		default:
			writeJSON(w, http.StatusOK, `{"success":true,"documents":[{"location":"custom-documents/raw-baru-456.json"}]}`)
		}
	})

	m, err := NewManager(srv.URL, "secret-key", "netops")
	require.NoError(t, err)

	loc, err := m.UpsertDocument(context.Background(), "custom-documents/raw-lama-789.json", "Judul", "isi")
	require.NoError(t, err)
	require.Equal(t, "custom-documents/raw-baru-456.json", loc)
	require.Len(t, *calls, 2, "delete gagal tapi upload tetap dijalankan")
}

func TestManagerUpsertDocumentErrors(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		payload    string
		wantErrSub string
	}{
		{"collector offline", http.StatusInternalServerError, `{"success":false,"error":"Document processing API is not online."}`, "raw-text returned 500"},
		{"missing metadata", http.StatusUnprocessableEntity, `{"success":false,"error":"You are missing required metadata key"}`, "raw-text returned 422"},
		{"rejected with success false", http.StatusOK, `{"success":false,"error":"something went wrong"}`, "raw-text rejected"},
		{"empty documents", http.StatusOK, `{"success":true,"documents":[]}`, "no document location"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := recordCalls(t, func(w http.ResponseWriter, r *http.Request) {
				writeJSON(w, tt.status, tt.payload)
			})
			m, err := NewManager(srv.URL, "secret-key", "netops")
			require.NoError(t, err)

			_, err = m.UpsertDocument(context.Background(), "", "Judul", "isi")
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErrSub)
		})
	}
}

func TestManagerUpsertDocumentNetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	srv.Close() // server sudah mati → request gagal

	m, err := NewManager(srv.URL, "secret-key", "netops")
	require.NoError(t, err)

	_, err = m.UpsertDocument(context.Background(), "", "Judul", "isi")
	require.Error(t, err)
	require.Contains(t, err.Error(), "raw-text request failed")
}

func TestManagerDeleteDocument(t *testing.T) {
	srv, calls := recordCalls(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, `{}`)
	})

	m, err := NewManager(srv.URL, "secret-key", "netops")
	require.NoError(t, err)

	err = m.DeleteDocument(context.Background(), "custom-documents/raw-harga-paket-abc123.json")
	require.NoError(t, err)

	require.Len(t, *calls, 1)
	call := (*calls)[0]
	require.Equal(t, http.MethodDelete, call.method)
	require.Equal(t, "/api/v1/system/remove-documents", call.path)
	require.Equal(t, "Bearer secret-key", call.auth)
	names, ok := call.body["names"].([]any)
	require.True(t, ok)
	require.Equal(t, []any{"custom-documents/raw-harga-paket-abc123.json"}, names)
}

func TestManagerDeleteDocumentEmptyNameNoOp(t *testing.T) {
	srv, calls := recordCalls(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, `{}`)
	})

	m, err := NewManager(srv.URL, "secret-key", "netops")
	require.NoError(t, err)

	require.NoError(t, m.DeleteDocument(context.Background(), ""))
	require.NoError(t, m.DeleteDocument(context.Background(), "   "))
	require.Empty(t, *calls, "doc name kosong tidak boleh memicu HTTP call")
}

func TestManagerDeleteDocumentErrors(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		payload    string
		wantErrSub string
	}{
		{"server error", http.StatusInternalServerError, `{"success":false}`, "remove-documents returned 500"},
		{"unauthorized", http.StatusForbidden, `{"success":false,"error":"Invalid API Key"}`, "remove-documents returned 403"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := recordCalls(t, func(w http.ResponseWriter, r *http.Request) {
				writeJSON(w, tt.status, tt.payload)
			})
			m, err := NewManager(srv.URL, "secret-key", "netops")
			require.NoError(t, err)

			err = m.DeleteDocument(context.Background(), "custom-documents/raw-xyz.json")
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErrSub)
		})
	}
}

func TestManagerDeleteDocumentNetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	srv.Close()

	m, err := NewManager(srv.URL, "secret-key", "netops")
	require.NoError(t, err)

	err = m.DeleteDocument(context.Background(), "custom-documents/raw-xyz.json")
	require.Error(t, err)
	require.Contains(t, err.Error(), "remove-documents request failed")
}

// pastikan string kosong di body metadata tetap dikirim sebagai object {}
// (server AnythingLLM 422 kalau metadata bukan object atau title kosong).
func TestManagerUpsertDocumentMetadataAlwaysPresent(t *testing.T) {
	srv, calls := recordCalls(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, `{"success":true,"documents":[{"location":"custom-documents/raw-x.json"}]}`)
	})

	m, err := NewManager(srv.URL, "secret-key", "netops")
	require.NoError(t, err)

	_, err = m.UpsertDocument(context.Background(), "", "Judul", "isi")
	require.NoError(t, err)

	call := (*calls)[0]
	raw, err := json.Marshal(call.body["metadata"])
	require.NoError(t, err)
	require.True(t, strings.Contains(string(raw), `"title":"Judul"`), "metadata harus selalu memuat title: %s", raw)
}
