package llm

import (
	"context"
	"fmt"

	"github.com/quixiq/polyglot/internal/adapter/llm/genkit"
	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/port"
)

// NewProvider creates the appropriate LLMProvider based on the given config using Google Genkit.
func NewProvider(cfg *llm.Config, encryptionKey string) (port.LLMProvider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("llm config is nil")
	}

	var apiKey string
	if cfg.APIKeyEncrypted != "" {
		decrypted, err := config.Decrypt(cfg.APIKeyEncrypted, encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt API key: %w", err)
		}
		apiKey = decrypted
	}

	return genkit.NewProvider(context.Background(), cfg, apiKey)
}
