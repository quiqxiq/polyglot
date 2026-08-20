package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/llm"
)

// LLMConfigRepository defines the persistence interface for LLM configurations.
type LLMConfigRepository interface {
	Create(ctx context.Context, config *llm.Config) error
	FindByID(ctx context.Context, id uint) (*llm.Config, error)
	FindActive(ctx context.Context) (*llm.Config, error)
	FindAll(ctx context.Context) ([]llm.Config, error)
	Update(ctx context.Context, config *llm.Config) error
	SetActive(ctx context.Context, id uint) error
	Delete(ctx context.Context, id uint) error
}

