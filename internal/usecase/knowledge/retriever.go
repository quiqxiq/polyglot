package knowledge

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/knowledge"
	"github.com/quixiq/polyglot/internal/port"
)

type KeywordRetriever struct {
	retriever port.KnowledgeRetriever
}

func NewKeywordRetriever(retriever port.KnowledgeRetriever) *KeywordRetriever {
	return &KeywordRetriever{retriever: retriever}
}

func (r *KeywordRetriever) Retrieve(ctx context.Context, query string) ([]knowledge.Entry, error) {
	if r.retriever != nil {
		return r.retriever.Retrieve(ctx, query)
	}
	return nil, nil
}
