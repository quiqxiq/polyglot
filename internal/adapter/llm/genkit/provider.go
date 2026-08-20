package genkit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/compat_oai/anthropic"
	"github.com/firebase/genkit/go/plugins/compat_oai/openai"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/firebase/genkit/go/plugins/ollama"
	"github.com/openai/openai-go/option"
	"github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/port"
)

// sanitizingRoundTripper membersihkan atribut yang tidak didukung upstream (seperti reasoning_content pada Groq)
type sanitizingRoundTripper struct {
	base http.RoundTripper
}

func (s *sanitizingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil && req.Method == http.MethodPost && strings.Contains(req.URL.Path, "/chat/completions") {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil {
			_ = req.Body.Close()

			var payload map[string]any
			if err := json.Unmarshal(bodyBytes, &payload); err == nil {
				if msgs, ok := payload["messages"].([]any); ok {
					changed := false
					for _, m := range msgs {
						if msgMap, ok := m.(map[string]any); ok {
							if _, has := msgMap["reasoning_content"]; has {
								delete(msgMap, "reasoning_content")
								changed = true
							}
						}
					}
					if changed {
						if modifiedBytes, err := json.Marshal(payload); err == nil {
							bodyBytes = modifiedBytes
							req.ContentLength = int64(len(bodyBytes))
						}
					}
				}
			}
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
	}

	base := s.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

// Provider implements port.LLMProvider using Google Genkit Go SDK.
type Provider struct {
	g         *genkit.Genkit
	modelName string
	maxTokens int
}

type customParams struct {
	BaseURL     string  `json:"base_url,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// NewProvider creates a new Genkit-backed LLMProvider.
func NewProvider(ctx context.Context, cfg *llm.Config, apiKey string) (*Provider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("llm config is nil")
	}

	var params customParams
	if strings.TrimSpace(cfg.Params) != "" {
		_ = json.Unmarshal([]byte(cfg.Params), &params)
	}

	rawModel := strings.TrimSpace(cfg.Model)
	var modelName string
	var g *genkit.Genkit

	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "gemini":
		if apiKey == "" {
			return nil, fmt.Errorf("gemini api key is required")
		}
		if rawModel == "" {
			rawModel = "gemini-2.0-flash"
		}
		modelName = "googlegenai/" + strings.TrimPrefix(strings.TrimPrefix(rawModel, "googlegenai/"), "googleai/")
		g = genkit.Init(ctx, genkit.WithPlugins(&googlegenai.GoogleAI{
			APIKey: apiKey,
		}))

	case "openai":
		if apiKey == "" {
			return nil, fmt.Errorf("openai api key is required")
		}
		if rawModel == "" {
			rawModel = "gpt-4o-mini"
		}
		var opts []option.RequestOption
		if params.BaseURL != "" {
			opts = append(opts, option.WithBaseURL(params.BaseURL))
		}
		modelName = "openai/" + rawModel
		g = genkit.Init(ctx, genkit.WithPlugins(&openai.OpenAI{
			APIKey: apiKey,
			Opts:   opts,
		}))

	case "groq":
		if apiKey == "" {
			return nil, fmt.Errorf("groq api key is required")
		}
		baseURL := params.BaseURL
		if baseURL == "" {
			baseURL = "https://api.groq.com/openai/v1"
		}
		if rawModel == "" {
			rawModel = "qwen/qwen3.6-27b"
		}
		modelName = "openai/" + rawModel
		httpClient := &http.Client{
			Transport: &sanitizingRoundTripper{base: http.DefaultTransport},
		}
		g = genkit.Init(ctx, genkit.WithPlugins(&openai.OpenAI{
			APIKey: apiKey,
			Opts: []option.RequestOption{
				option.WithBaseURL(baseURL),
				option.WithHTTPClient(httpClient),
			},
		}))

	case "deepseek":
		if apiKey == "" {
			return nil, fmt.Errorf("deepseek api key is required")
		}
		baseURL := params.BaseURL
		if baseURL == "" {
			baseURL = "https://api.deepseek.com/v1"
		}
		if rawModel == "" {
			rawModel = "deepseek-chat"
		}
		modelName = "openai/" + rawModel
		g = genkit.Init(ctx, genkit.WithPlugins(&openai.OpenAI{
			APIKey: apiKey,
			Opts:   []option.RequestOption{option.WithBaseURL(baseURL)},
		}))

	case "claude":
		if apiKey == "" {
			return nil, fmt.Errorf("claude api key is required")
		}
		if rawModel == "" {
			rawModel = "claude-3-5-sonnet-latest"
		}
		var opts []option.RequestOption
		if params.BaseURL != "" {
			opts = append(opts, option.WithBaseURL(params.BaseURL))
		}
		modelName = "anthropic/" + strings.TrimPrefix(rawModel, "anthropic/")
		g = genkit.Init(ctx, genkit.WithPlugins(&anthropic.Anthropic{
			APIKey: apiKey,
			Opts:   opts,
		}))

	case "ollama":
		serverAddr := params.BaseURL
		if serverAddr == "" {
			serverAddr = "http://localhost:11434"
		}
		if rawModel == "" {
			rawModel = "llama3.2"
		}
		modelName = "ollama/" + strings.TrimPrefix(rawModel, "ollama/")
		g = genkit.Init(ctx, genkit.WithPlugins(&ollama.Ollama{
			ServerAddress: serverAddr,
		}))

	case "custom":
		baseURL := params.BaseURL
		if baseURL == "" {
			return nil, fmt.Errorf("base_url is required for custom openai-compatible provider")
		}
		if rawModel == "" {
			return nil, fmt.Errorf("model name is required for custom provider")
		}
		modelName = "openai/" + rawModel
		g = genkit.Init(ctx, genkit.WithPlugins(&openai.OpenAI{
			APIKey: apiKey,
			Opts:   []option.RequestOption{option.WithBaseURL(baseURL)},
		}))

	default:
		return nil, fmt.Errorf("unsupported llm provider: %s", cfg.Provider)
	}

	return &Provider{
		g:         g,
		modelName: modelName,
		maxTokens: cfg.MaxOutputTokens,
	}, nil
}

// Chat executes a conversational completion through Genkit unified interface.
func (p *Provider) Chat(ctx context.Context, systemPrompt string, messages []llm.ChatMessage, maxTokens int) (*llm.ChatResponse, error) {
	return p.ChatWithTools(ctx, systemPrompt, messages, nil, maxTokens)
}

// ChatWithTools executes a conversational completion with native Genkit tool calling support.
func (p *Provider) ChatWithTools(ctx context.Context, systemPrompt string, messages []llm.ChatMessage, tools []llm.Tool, maxTokens int) (*llm.ChatResponse, error) {
	if p.g == nil {
		return nil, fmt.Errorf("genkit instance is not initialized")
	}

	if maxTokens <= 0 {
		maxTokens = p.maxTokens
	}
	if maxTokens <= 0 {
		maxTokens = 1024
	}

	// Susun history percakapan ke tipe ai.Message
	var genkitMessages []*ai.Message
	for _, m := range messages {
		var role ai.Role
		switch strings.ToLower(m.Role) {
		case "user":
			role = ai.RoleUser
		case "assistant", "model", "bot":
			role = ai.RoleModel
		case "system":
			role = ai.RoleSystem
		default:
			role = ai.RoleUser
		}

		genkitMessages = append(genkitMessages, &ai.Message{
			Role: role,
			Content: []*ai.Part{
				ai.NewTextPart(m.Content),
			},
		})
	}

	opts := []ai.GenerateOption{
		ai.WithModelName(p.modelName),
		ai.WithMessages(genkitMessages...),
	}

	if strings.TrimSpace(systemPrompt) != "" {
		opts = append(opts, ai.WithSystem(systemPrompt))
	}

	if maxTokens > 0 && (strings.HasPrefix(p.modelName, "googlegenai/") || strings.HasPrefix(p.modelName, "googleai/")) {
		opts = append(opts, ai.WithConfig(&ai.GenerationCommonConfig{
			MaxOutputTokens: maxTokens,
		}))
	}

	// Daftarkan tools jika disediakan
	if len(tools) > 0 {
		var genkitTools []ai.ToolRef
		for _, t := range tools {
			toolDef := t
			gt := ai.NewTool(
				toolDef.Name,
				toolDef.Description,
				func(toolCtx *ai.ToolContext, input map[string]any) (string, error) {
					var argsJSON string
					if input != nil {
						b, err := json.Marshal(input)
						if err == nil {
							argsJSON = string(b)
						}
					}
					if toolDef.Handler != nil {
						return toolDef.Handler(toolCtx.Context, argsJSON)
					}
					return "tool executed", nil
				},
			)
			genkitTools = append(genkitTools, gt)
		}
		opts = append(opts, ai.WithTools(genkitTools...), ai.WithMaxTurns(5))
	}

	resp, err := genkit.Generate(ctx, p.g, opts...)
	if err != nil {
		return nil, fmt.Errorf("genkit generate failed: %w", err)
	}

	content := resp.Text()
	tokenIn := 0
	tokenOut := 0

	if resp.Usage != nil {
		tokenIn = resp.Usage.InputTokens
		tokenOut = resp.Usage.OutputTokens
	}

	return &llm.ChatResponse{
		Content:  content,
		TokenIn:  tokenIn,
		TokenOut: tokenOut,
	}, nil
}

// TestConnection verifies connectivity to the AI model with a minimal prompt.
func (p *Provider) TestConnection(ctx context.Context) error {
	if p.g == nil {
		return fmt.Errorf("genkit instance is not initialized")
	}

	opts := []ai.GenerateOption{
		ai.WithModelName(p.modelName),
		ai.WithPrompt("ping"),
	}
	if strings.HasPrefix(p.modelName, "googlegenai/") || strings.HasPrefix(p.modelName, "googleai/") {
		opts = append(opts, ai.WithConfig(&ai.GenerationCommonConfig{
			MaxOutputTokens: 5,
		}))
	}

	resp, err := genkit.GenerateText(ctx, p.g, opts...)
	if err != nil {
		return fmt.Errorf("test connection failed for %s: %w", p.modelName, err)
	}

	if strings.TrimSpace(resp) == "" {
		return fmt.Errorf("received empty response from %s", p.modelName)
	}

	return nil
}

var _ port.LLMProvider = (*Provider)(nil)
