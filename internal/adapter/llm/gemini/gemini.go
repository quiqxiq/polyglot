package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/quixiq/polyglot/internal/domain/llm"
)

const apiURLTemplate = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s"

type Provider struct {
	apiKey     string
	model      string
	maxTokens  int
	httpClient *http.Client
}

func NewProvider(apiKey, model string, maxTokens int) (*Provider, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}
	if model == "" {
		model = "gemini-2.0-flash"
	}
	return &Provider{
		apiKey:     apiKey,
		model:      model,
		maxTokens:  maxTokens,
		httpClient: &http.Client{},
	}, nil
}

type geminiRequest struct {
	Contents          []geminiContent         `json:"contents"`
	SystemInstruction *geminiContent          `json:"systemInstruction,omitempty"`
	GenerationConfig  *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	MaxOutputTokens int `json:"maxOutputTokens,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
	} `json:"usageMetadata"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (p *Provider) Chat(ctx context.Context, systemPrompt string, messages []llm.ChatMessage, maxTokens int) (*llm.ChatResponse, error) {
	if maxTokens <= 0 {
		maxTokens = p.maxTokens
	}

	contents := make([]geminiContent, 0, len(messages))
	for _, msg := range messages {
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: msg.Content}},
		})
	}

	reqBody := geminiRequest{
		Contents: contents,
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		GenerationConfig: &geminiGenerationConfig{
			MaxOutputTokens: maxTokens,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf(apiURLTemplate, p.model, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Gemini API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if geminiResp.Error != nil {
		return nil, fmt.Errorf("Gemini API error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("Gemini returned no content")
	}

	return &llm.ChatResponse{
		Content:  geminiResp.Candidates[0].Content.Parts[0].Text,
		TokenIn:  geminiResp.UsageMetadata.PromptTokenCount,
		TokenOut: geminiResp.UsageMetadata.CandidatesTokenCount,
	}, nil
}

func (p *Provider) TestConnection(ctx context.Context) error {
	_, err := p.Chat(ctx, "You are a test.", []llm.ChatMessage{
		{Role: "user", Content: "Say OK"},
	}, 10)
	return err
}
