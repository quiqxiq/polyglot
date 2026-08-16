package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/knowledge"
)

// KnowledgeRetriever defines the interface for retrieving relevant knowledge.
type KnowledgeRetriever interface {
	Retrieve(ctx context.Context, query string) ([]knowledge.Entry, error)
}

