package knowledge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/knowledge"
)

const (
	// DefaultBaseURL is where the self-hosted AnythingLLM instance listens.
	DefaultBaseURL = "http://localhost:3001"
	// DefaultTopN is how many most-relevant chunks to request per query.
	DefaultTopN = 6
	// DefaultScoreThreshold filters out weak matches (0-1, higher = stricter).
	// 0.25 adalah default AnythingLLM sendiri (workspace.similarityThreshold).
	DefaultScoreThreshold = 0.25
)

// Retriever implements port.KnowledgeRetriever by querying the vector store of
// satu workspace AnythingLLM (POST /api/v1/workspace/:slug/vector-search).
// Dokumen di-upload dan dikelola lewat UI/admin AnythingLLM; adapter ini hanya
// mengambil chunk paling relevan untuk disuntikkan ke prompt bot. LLM call
// tetap dilakukan engine di Go — AnythingLLM di sini murni retrieval.
type Retriever struct {
	baseURL        string
	apiKey         string
	workspaceSlug  string
	topN           int
	scoreThreshold float64
	httpClient     *http.Client
}

// NewRetriever builds a Retriever against a self-hosted AnythingLLM instance.
// baseURL kosong memakai DefaultBaseURL; topN <= 0 memakai DefaultTopN.
func NewRetriever(baseURL, apiKey, workspaceSlug string, topN int) (*Retriever, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("anythingllm: API key is required")
	}
	if strings.TrimSpace(workspaceSlug) == "" {
		return nil, fmt.Errorf("anythingllm: workspace slug is required")
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = DefaultBaseURL
	}
	if topN <= 0 {
		topN = DefaultTopN
	}
	return &Retriever{
		baseURL:        strings.TrimRight(baseURL, "/"),
		apiKey:         strings.TrimSpace(apiKey),
		workspaceSlug:  strings.TrimSpace(workspaceSlug),
		topN:           topN,
		scoreThreshold: DefaultScoreThreshold,
		httpClient:     &http.Client{},
	}, nil
}

// vectorSearchRequest mirrors the body of POST /api/v1/workspace/:slug/vector-search.
type vectorSearchRequest struct {
	Query          string  `json:"query"`
	TopN           int     `json:"topN,omitempty"`
	ScoreThreshold float64 `json:"scoreThreshold,omitempty"`
}

type vectorSearchResponse struct {
	Results []vectorSearchResult `json:"results"`
	Message string               `json:"message"`
}

type vectorSearchResult struct {
	Text     string `json:"text"`
	Metadata struct {
		Title       string `json:"title"`
		ChunkSource string `json:"chunkSource"`
		URL         string `json:"url"`
	} `json:"metadata"`
	Score float64 `json:"score"`
}

// Retrieve queries the AnythingLLM workspace vector store and maps the top
// matching chunks into knowledge entries: Title = nama dokumen sumber,
// Content = isi chunk. Query kosong → tidak ada yang di-retrieve.
func (r *Retriever) Retrieve(ctx context.Context, query string) ([]knowledge.KnowledgeEntry, error) {
	if strings.TrimSpace(query) == "" {
		return nil, nil
	}

	reqBody := vectorSearchRequest{
		Query:          query,
		TopN:           r.topN,
		ScoreThreshold: r.scoreThreshold,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("anythingllm: marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/workspace/%s/vector-search", r.baseURL, r.workspaceSlug)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("anythingllm: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.apiKey)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("anythingllm: vector-search request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("anythingllm: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anythingllm: vector-search returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var searchResp vectorSearchResponse
	if err := json.Unmarshal(respBody, &searchResp); err != nil {
		return nil, fmt.Errorf("anythingllm: parse response: %w", err)
	}

	entries := make([]knowledge.KnowledgeEntry, 0, len(searchResp.Results))
	for _, result := range searchResp.Results {
		if strings.TrimSpace(result.Text) == "" {
			continue
		}
		title := result.Metadata.Title
		if title == "" {
			title = result.Metadata.ChunkSource
		}
		if title == "" {
			title = result.Metadata.URL
		}
		entries = append(entries, knowledge.KnowledgeEntry{
			Title:   title,
			Content: result.Text,
		})
	}
	return entries, nil
}
