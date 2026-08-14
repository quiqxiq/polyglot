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
// method yang tidak bentrok dengan method domain lain di Store. Store sendiri
// tidak bisa mengimplement interface ini langsung karena method generic
// (`Create`, `FindByID`, `Update`, `Delete`) sudah dipakai domain lain
// (mis. `Create(*llm.LLMConfig)`), dan Go tidak punya overloading.
type KnowledgeRepositoryAdapter struct {
	store *Store
}

// NewKnowledgeRepository membungkus Store sebagai port.KnowledgeRepository.
func NewKnowledgeRepository(store *Store) *KnowledgeRepositoryAdapter {
	return &KnowledgeRepositoryAdapter{store: store}
}

func (a *KnowledgeRepositoryAdapter) Create(entry *knowledge.KnowledgeEntry) error {
	return a.store.CreateKnowledgeEntry(entry)
}

func (a *KnowledgeRepositoryAdapter) FindByID(id uint) (*knowledge.KnowledgeEntry, error) {
	return a.store.FindKnowledgeEntryByID(id)
}

func (a *KnowledgeRepositoryAdapter) FindAll() ([]knowledge.KnowledgeEntry, error) {
	return a.store.FindAllKnowledgeEntries()
}

func (a *KnowledgeRepositoryAdapter) Update(entry *knowledge.KnowledgeEntry) error {
	return a.store.UpdateKnowledgeEntry(entry)
}

func (a *KnowledgeRepositoryAdapter) Delete(id uint) error {
	return a.store.DeleteKnowledgeEntry(id)
}

func (a *KnowledgeRepositoryAdapter) SearchByTags(tags []string) ([]knowledge.KnowledgeEntry, error) {
	return a.store.SearchKnowledgeByTags(tags)
}

// Retrieve memenuhi port.KnowledgeRetriever untuk keyword retriever (dokumen
// lokal Postgres). Pencarian keyword sebenarnya dilakukan filter di caller —
// di sini semua entry dikembalikan, konsisten dengan perilaku lama Store.
func (a *KnowledgeRepositoryAdapter) Retrieve(ctx context.Context, _ string) ([]knowledge.KnowledgeEntry, error) {
	return a.store.FindAllKnowledgeEntries()
}
