package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/knowledge"
)

// KnowledgeRepository defines the persistence interface for knowledge entries.
type KnowledgeRepository interface {
	Create(ctx context.Context, entry *knowledge.Entry) error
	FindByID(ctx context.Context, id uint) (*knowledge.Entry, error)
	FindAll(ctx context.Context) ([]knowledge.Entry, error)
	Update(ctx context.Context, entry *knowledge.Entry) error
	Delete(ctx context.Context, id uint) error
	SearchByTags(ctx context.Context, tags []string) ([]knowledge.Entry, error)
}

