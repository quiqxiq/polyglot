package llm

import (
	"fmt"

	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/llm"
	llmclaude "github.com/quixiq/polyglot/internal/adapter/llm/claude"
	llmgemini "github.com/quixiq/polyglot/internal/adapter/llm/gemini"
	llmgroq "github.com/quixiq/polyglot/internal/adapter/llm/groq"
	llmopenai "github.com/quixiq/polyglot/internal/adapter/llm/openai"
	"github.com/quixiq/polyglot/internal/port"
)

// NewProvider creates the appropriate LLMProvider based on the given config.
func NewProvider(cfg *llm.LLMConfig, encryptionKey string) (port.LLMProvider, error) {
	apiKey, err := config.Decrypt(cfg.APIKeyEncrypted, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt API key: %w", err)
	}

	switch cfg.Provider {
	case "openai":
		return llmopenai.NewProvider(apiKey, cfg.Model, cfg.MaxOutputTokens)
	case "gemini":
		return llmgemini.NewProvider(apiKey, cfg.Model, cfg.MaxOutputTokens)
	case "claude":
		return llmclaude.NewProvider(apiKey, cfg.Model, cfg.MaxOutputTokens)
	case "groq":
		return llmgroq.NewProvider(apiKey, cfg.Model, cfg.MaxOutputTokens)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.Provider)
	}
}
