package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/quixiq/polyglot/internal/domain/llm"
)

const apiURL = "https://api.anthropic.com/v1/messages"

type Provider struct {
	apiKey     string
	model      string
	maxTokens  int
	httpClient *http.Client
}

func NewProvider(apiKey, model string, maxTokens int) (*Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Claude API key is required")
	}
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	return &Provider{
		apiKey:     apiKey,
		model:      model,
		maxTokens:  maxTokens,
		httpClient: &http.Client{},
	}, nil
}

type claudeRequest struct {
	Model     string          `json:"model"`
	System    string          `json:"system,omitempty"`
	Messages  []claudeMessage `json:"messages"`
	MaxTokens int             `json:"max_tokens"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func (p *Provider) Chat(ctx context.Context, systemPrompt string, messages []llm.ChatMessage, maxTokens int) (*llm.ChatResponse, error) {
	if maxTokens <= 0 {
		maxTokens = p.maxTokens
	}

	apiMessages := make([]claudeMessage, 0, len(messages))
	for _, msg := range messages {
		apiMessages = append(apiMessages, claudeMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	reqBody := claudeRequest{
		Model:     p.model,
		System:    systemPrompt,
		Messages:  apiMessages,
		MaxTokens: maxTokens,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Claude API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if claudeResp.Error != nil {
		return nil, fmt.Errorf("Claude API error: %s - %s", claudeResp.Error.Type, claudeResp.Error.Message)
	}

	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("Claude returned no content")
	}

	var content string
	for _, block := range claudeResp.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	return &llm.ChatResponse{
		Content:  content,
		TokenIn:  claudeResp.Usage.InputTokens,
		TokenOut: claudeResp.Usage.OutputTokens,
	}, nil
}

func (p *Provider) TestConnection(ctx context.Context) error {
	_, err := p.Chat(ctx, "You are a test.", []llm.ChatMessage{
		{Role: "user", Content: "Say OK"},
	}, 10)
	return err
}
