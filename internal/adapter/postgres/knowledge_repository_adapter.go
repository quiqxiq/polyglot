package postgres

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/knowledge"
	"github.com/quixiq/polyglot/internal/port"
)

// Compile-time proof bahwa *KnowledgeRepositoryAdapter memenuhi
// port.KnowledgeRepository.
var _ port.KnowledgeRepository = (*KnowledgeRepositoryAdapter)(nil)

// KnowledgeRepositoryAdapter memenuhi port.KnowledgeRepository dengan nama
// method yang tidak bentrok dengan method domain lain di Store.
type KnowledgeRepositoryAdapter struct {
	store *Store
}

// NewKnowledgeRepository membungkus Store sebagai port.KnowledgeRepository.
func NewKnowledgeRepository(store *Store) *KnowledgeRepositoryAdapter {
	return &KnowledgeRepositoryAdapter{store: store}
}

func (a *KnowledgeRepositoryAdapter) Create(ctx context.Context, entry *knowledge.Entry) error {
	return a.store.CreateKnowledgeEntry(ctx, entry)
}

func (a *KnowledgeRepositoryAdapter) FindByID(ctx context.Context, id uint) (*knowledge.Entry, error) {
	return a.store.FindKnowledgeEntryByID(ctx, id)
}

func (a *KnowledgeRepositoryAdapter) FindAll(ctx context.Context) ([]knowledge.Entry, error) {
	return a.store.FindAllKnowledgeEntries(ctx)
}

func (a *KnowledgeRepositoryAdapter) Update(ctx context.Context, entry *knowledge.Entry) error {
	return a.store.UpdateKnowledgeEntry(ctx, entry)
}

func (a *KnowledgeRepositoryAdapter) Delete(ctx context.Context, id uint) error {
	return a.store.DeleteKnowledgeEntry(ctx, id)
}

func (a *KnowledgeRepositoryAdapter) SearchByTags(ctx context.Context, tags []string) ([]knowledge.Entry, error) {
	return a.store.SearchKnowledgeByTags(ctx, tags)
}

// Retrieve memenuhi port.KnowledgeRetriever untuk keyword retriever (dokumen
// lokal Postgres). Pencarian keyword sebenarnya dilakukan filter di caller —
// di sini semua entry dikembalikan, konsisten dengan perilaku lama Store.
func (a *KnowledgeRepositoryAdapter) Retrieve(ctx context.Context, _ string) ([]knowledge.Entry, error) {
	return a.store.FindAllKnowledgeEntries(ctx)
}
