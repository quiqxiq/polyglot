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

// ChatClient implements port.KnowledgeChat via AnythingLLM's synchronous
// workspace chat endpoint (POST /api/v1/workspace/:slug/chat, mode "chat").
// Satu panggilan ini = retrieval dari vector store workspace + LLM answer,
// plus rolling chat history per sessionId — jadi AnythingLLM menjadi "otak"
// bot, bukan sekadar gudang vector.
type ChatClient struct {
	baseURL       string
	apiKey        string
	workspaceSlug string
	httpClient    *http.Client
}

// Compile-time proof bahwa *ChatClient benar-benar memenuhi port.KnowledgeChat.
var _ port.KnowledgeChat = (*ChatClient)(nil)

// NewChatClient builds a ChatClient against a self-hosted AnythingLLM instance.
// baseURL kosong memakai DefaultBaseURL. API key / workspace slug kosong → error
// (sama seperti NewRetriever).
func NewChatClient(baseURL, apiKey, workspaceSlug string) (*ChatClient, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("anythingllm: API key is required")
	}
	if strings.TrimSpace(workspaceSlug) == "" {
		return nil, fmt.Errorf("anythingllm: workspace slug is required")
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = DefaultBaseURL
	}
	return &ChatClient{
		baseURL:       strings.TrimRight(baseURL, "/"),
		apiKey:        strings.TrimSpace(apiKey),
		workspaceSlug: strings.TrimSpace(workspaceSlug),
		httpClient:    &http.Client{},
	}, nil
}

// chatRequest mirrors the body of POST /api/v1/workspace/:slug/chat.
// Mode "chat" dipilih: pakai LLM + custom embeddings + rolling history
// (berbeda dengan "query" yang tidak mengingat riwayat percakapan).
// DEVIASI: tag JSON "sessionId" (camelCase) sengaja mengikuti wire format
// API AnythingLLM, bukan konvensi snake_case project ini.
type chatRequest struct {
	Message   string `json:"message"`
	Mode      string `json:"mode"`
	SessionID string `json:"sessionId,omitempty"`
}

type chatResponse struct {
	Type         string       `json:"type"`
	TextResponse string       `json:"textResponse"`
	Sources      []chatSource `json:"sources"`
	Error        string       `json:"error"`
	Metrics      chatMetrics  `json:"metrics"`
}

type chatSource struct {
	Title string `json:"title"`
}

// chatMetrics menampung token usage yang dikembalikan provider LLM
// (shape provider-dependent; field boleh absen → zero value).
type chatMetrics struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

// Chat sends a single message to the workspace (mode "chat") and returns the
// LLM answer plus grounding sources. sessionID mempartisi rolling history
// per percakapan (engine bot memakai conversation ID). Error dikembalikan
// untuk: network failure, status != 200, tipe "abort", atau field error —
// caller (engine bot) memakai ini sebagai sinyal untuk fallback ke LLM lokal.
func (c *ChatClient) Chat(ctx context.Context, message string, sessionID string) (port.KnowledgeChatResult, error) {
	if strings.TrimSpace(message) == "" {
		return port.KnowledgeChatResult{}, fmt.Errorf("anythingllm: message is required")
	}

	reqBody := chatRequest{
		Message:   message,
		Mode:      "chat",
		SessionID: sessionID,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return port.KnowledgeChatResult{}, fmt.Errorf("anythingllm: marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/workspace/%s/chat", c.baseURL, c.workspaceSlug)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return port.KnowledgeChatResult{}, fmt.Errorf("anythingllm: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return port.KnowledgeChatResult{}, fmt.Errorf("anythingllm: chat request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return port.KnowledgeChatResult{}, fmt.Errorf("anythingllm: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return port.KnowledgeChatResult{}, fmt.Errorf("anythingllm: chat returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return port.KnowledgeChatResult{}, fmt.Errorf("anythingllm: parse response: %w", err)
	}
	if chatResp.Type == "abort" || strings.TrimSpace(chatResp.Error) != "" {
		return port.KnowledgeChatResult{}, fmt.Errorf("anythingllm: chat aborted: %s", chatResp.Error)
	}
	if strings.TrimSpace(chatResp.TextResponse) == "" {
		return port.KnowledgeChatResult{}, fmt.Errorf("anythingllm: chat returned empty response")
	}

	sources := make([]string, 0, len(chatResp.Sources))
	for _, s := range chatResp.Sources {
		if s.Title != "" {
			sources = append(sources, s.Title)
		}
	}

	return port.KnowledgeChatResult{
		Content:  chatResp.TextResponse,
		Sources:  sources,
		TokenIn:  chatResp.Metrics.PromptTokens,
		TokenOut: chatResp.Metrics.CompletionTokens,
	}, nil
}
