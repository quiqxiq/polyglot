package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/knowledge"
)

// KnowledgeRepository defines the persistence interface for knowledge entries.
type KnowledgeRepository interface {
	Create(entry *knowledge.KnowledgeEntry) error
	FindByID(id uint) (*knowledge.KnowledgeEntry, error)
	FindAll() ([]knowledge.KnowledgeEntry, error)
	Update(entry *knowledge.KnowledgeEntry) error
	Delete(id uint) error
	SearchByTags(tags []string) ([]knowledge.KnowledgeEntry, error)
}

// KnowledgeRetriever defines the interface for retrieving relevant knowledge.
type KnowledgeRetriever interface {
	Retrieve(ctx context.Context, query string) ([]knowledge.KnowledgeEntry, error)
}
