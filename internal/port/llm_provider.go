package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/llm"
)

// LLMConfigRepository defines the persistence interface for LLM configurations.
type LLMConfigRepository interface {
	Create(config *llm.LLMConfig) error
	FindByID(id uint) (*llm.LLMConfig, error)
	FindActive() (*llm.LLMConfig, error)
	FindAll() ([]llm.LLMConfig, error)
	Update(config *llm.LLMConfig) error
	SetActive(id uint) error
	Delete(id uint) error
}

// LLMProvider defines the interface for interacting with LLM services.
type LLMProvider interface {
	Chat(ctx context.Context, systemPrompt string, messages []llm.ChatMessage, maxTokens int) (*llm.ChatResponse, error)
	TestConnection(ctx context.Context) error
}
