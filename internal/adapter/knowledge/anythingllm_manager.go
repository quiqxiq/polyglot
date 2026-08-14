package knowledge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/quixiq/polyglot/internal/port"
)

const (
	// DefaultDocumentFolder adalah folder storage AnythingLLM tempat dokumen
	// raw-text dari knowledge admin ditaruh ("custom-documents" adalah folder
	// default instance AnythingLLM untuk dokumen API/upload).
	DefaultDocumentFolder = "custom-documents"
)

// Compile-time proof bahwa *Manager benar-benar memenuhi
// port.KnowledgeDocumentManager — tanpa ini, ketidakcocokan signature bisa
// lolos `go build` selama tidak ada kode lain yang menugaskan *Manager ke
// variabel interface.
var _ port.KnowledgeDocumentManager = (*Manager)(nil)

// Manager implements port.KnowledgeDocumentManager terhadap satu instance
// AnythingLLM. Sisi tulis knowledge admin: raw-text upload (replace) dan
// remove-documents (delete). Endpoint diverifikasi dari source server
// AnythingLLM v1.16.0:
//   - POST   /api/v1/document/raw-text        (endpoints/api/document/index.js)
//   - DELETE /api/v1/system/remove-documents  (endpoints/api/system/index.js)
//
// Berbeda dari Retriever (sisi baca untuk bot), Manager dipakai halaman
// admin knowledge untuk sinkronisasi per-dokumen.
type Manager struct {
	baseURL       string
	apiKey        string
	workspaceSlug string
	folder        string
	httpClient    *http.Client
}

// NewManager builds a Manager against a self-hosted AnythingLLM instance.
// baseURL kosong memakai DefaultBaseURL; folder memakai DefaultDocumentFolder.
func NewManager(baseURL, apiKey, workspaceSlug string) (*Manager, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("anythingllm: API key is required")
	}
	if strings.TrimSpace(workspaceSlug) == "" {
		return nil, fmt.Errorf("anythingllm: workspace slug is required")
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = DefaultBaseURL
	}
	return &Manager{
		baseURL:       strings.TrimRight(baseURL, "/"),
		apiKey:        strings.TrimSpace(apiKey),
		workspaceSlug: strings.TrimSpace(workspaceSlug),
		folder:        DefaultDocumentFolder,
		httpClient:    &http.Client{},
	}, nil
}

// rawTextRequest mirrors the body of POST /api/v1/document/raw-text.
// metadata.title wajib (server 422 kalau kosong); key lain bebas ditambah.
type rawTextRequest struct {
	TextContent     string            `json:"textContent"`
	AddToWorkspaces string            `json:"addToWorkspaces,omitempty"`
	Metadata        map[string]string `json:"metadata"`
}

type rawTextDocument struct {
	// Location adalah path relatif file JSON di storage AnythingLLM,
	// contoh "custom-documents/raw-my-doc-text-<uuid>.json".
	Location string `json:"location"`
	Title    string `json:"title"`
}

type rawTextResponse struct {
	Success   bool              `json:"success"`
	Error     string            `json:"error"`
	Documents []rawTextDocument `json:"documents"`
}

type removeDocumentsRequest struct {
	Names []string `json:"names"`
}

// UpsertDocument meng-upload markdown sebagai dokumen raw-text ke workspace.
// Kalau docName lama tidak kosong, dokumen lama dihapus dulu (semantik
// replace). Mengembalikan doc name JSON terbaru.
func (m *Manager) UpsertDocument(ctx context.Context, docName, title, markdown string) (string, error) {
	if strings.TrimSpace(docName) != "" {
		if err := m.DeleteDocument(ctx, docName); err != nil {
			// Best-effort: dokumen lama mungkin sudah dihapus manual dari UI
			// AnythingLLM (purgeDocument error "not found"). Upload tetap
			// dilanjutkan supaya dokumen baru selalu ter-embed.
			_ = err
		}
	}

	reqBody := rawTextRequest{
		TextContent:     markdown,
		AddToWorkspaces: m.workspaceSlug,
		Metadata: map[string]string{
			"title": title,
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("anythingllm: marshal raw-text request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/document/raw-text", m.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("anythingllm: build raw-text request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("anythingllm: raw-text request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("anythingllm: read raw-text response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anythingllm: raw-text returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var rawResp rawTextResponse
	if err := json.Unmarshal(respBody, &rawResp); err != nil {
		return "", fmt.Errorf("anythingllm: parse raw-text response: %w", err)
	}
	if !rawResp.Success {
		return "", fmt.Errorf("anythingllm: raw-text rejected: %s", rawResp.Error)
	}
	if len(rawResp.Documents) == 0 || strings.TrimSpace(rawResp.Documents[0].Location) == "" {
		return "", fmt.Errorf("anythingllm: raw-text returned no document location")
	}
	return rawResp.Documents[0].Location, nil
}

// DeleteDocument menghapus satu dokumen dari AnythingLLM. docName kosong
// dianggap no-op — tidak ada HTTP call.
func (m *Manager) DeleteDocument(ctx context.Context, docName string) error {
	if strings.TrimSpace(docName) == "" {
		return nil
	}

	reqBody := removeDocumentsRequest{Names: []string{docName}}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("anythingllm: marshal remove-documents request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/system/remove-documents", m.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("anythingllm: build remove-documents request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("anythingllm: remove-documents request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("anythingllm: remove-documents returned %d", resp.StatusCode)
		}
		return fmt.Errorf("anythingllm: remove-documents returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return nil
}
