package genkit

import (
	"context"
	"encoding/json"
	"fmt"
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
		g = genkit.Init(ctx, genkit.WithPlugins(&openai.OpenAI{
			APIKey: apiKey,
			Opts:   []option.RequestOption{option.WithBaseURL(baseURL)},
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
