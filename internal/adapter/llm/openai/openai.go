package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/quixiq/polyglot/internal/domain/llm"
)

const apiURL = "https://api.openai.com/v1/chat/completions"

type Provider struct {
	apiKey     string
	model      string
	maxTokens  int
	httpClient *http.Client
}

func NewProvider(apiKey, model string, maxTokens int) (*Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &Provider{
		apiKey:     apiKey,
		model:      model,
		maxTokens:  maxTokens,
		httpClient: &http.Client{},
	}, nil
}

type chatRequest struct {
	Model     string       `json:"model"`
	Messages  []apiMessage `json:"messages"`
	MaxTokens int          `json:"max_tokens,omitempty"`
}

type apiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (p *Provider) Chat(ctx context.Context, systemPrompt string, messages []llm.ChatMessage, maxTokens int) (*llm.ChatResponse, error) {
	if maxTokens <= 0 {
		maxTokens = p.maxTokens
	}

	apiMessages := make([]apiMessage, 0, len(messages)+1)
	apiMessages = append(apiMessages, apiMessage{Role: "system", Content: systemPrompt})
	for _, msg := range messages {
		apiMessages = append(apiMessages, apiMessage{Role: msg.Role, Content: msg.Content})
	}

	reqBody := chatRequest{
		Model:     p.model,
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
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if chatResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("OpenAI returned no choices")
	}

	return &llm.ChatResponse{
		Content:  chatResp.Choices[0].Message.Content,
		TokenIn:  chatResp.Usage.PromptTokens,
		TokenOut: chatResp.Usage.CompletionTokens,
	}, nil
}

func (p *Provider) TestConnection(ctx context.Context) error {
	_, err := p.Chat(ctx, "You are a test.", []llm.ChatMessage{
		{Role: "user", Content: "Say OK"},
	}, 10)
	return err
}
