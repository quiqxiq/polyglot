package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/knowledge"
)

// KnowledgeEntryModel is the GORM database model for Knowledge Base entries.
// TableName eksplisit ke `knowledge_entries` (migrasi 000002) — tanpa ini GORM
// memakai `knowledge_entry_models`, divergen dari migrasi (prod).
type KnowledgeEntryModel struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"not null"`
	Content     string `gorm:"type:text"`
	Category    string `gorm:"type:varchar(100);default:umum"`
	Tags        string
	EmbedToLLM  bool   `gorm:"default:false"`
	EmbedStatus string `gorm:"type:varchar(20);default:none;index:idx_knowledge_embed_status"`
	// Tanpa tag `column:` GORM memetakan AnythingLLMDocName →
	// anything_llm_doc_name (snake_case + pemisah acronym), padahal migrasi
	// 000008 membuat kolom `anythingllm_doc_name`. Tanpa tag ini, AutoMigrate
	// (dev) membuat kolom kedua yang tidak pernah dibaca migrasi — dokumen
	// ter-embed tapi doc name hilang di DB.
	AnythingLLMDocName string `gorm:"column:anythingllm_doc_name;type:varchar(255)"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// TableName maps KnowledgeEntryModel ke tabel migrasi `knowledge_entries`.
func (KnowledgeEntryModel) TableName() string { return "knowledge_entries" }

func (m *KnowledgeEntryModel) ToDomain() *knowledge.Entry {
	if m == nil {
		return nil
	}
	return &knowledge.Entry{
		ID:                 m.ID,
		Title:              m.Title,
		Content:            m.Content,
		Category:           m.Category,
		Tags:               m.Tags,
		EmbedToLLM:         m.EmbedToLLM,
		EmbedStatus:        m.EmbedStatus,
		AnythingLLMDocName: m.AnythingLLMDocName,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}

func KnowledgeEntryModelFromDomain(k *knowledge.Entry) *KnowledgeEntryModel {
	if k == nil {
		return nil
	}
	return &KnowledgeEntryModel{
		ID:                 k.ID,
		Title:              k.Title,
		Content:            k.Content,
		Category:           k.Category,
		Tags:               k.Tags,
		EmbedToLLM:         k.EmbedToLLM,
		EmbedStatus:        k.EmbedStatus,
		AnythingLLMDocName: k.AnythingLLMDocName,
		CreatedAt:          k.CreatedAt,
		UpdatedAt:          k.UpdatedAt,
	}
}
