package models

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/knowledge"
)

// KnowledgeEntryModel is the GORM database model for Knowledge Base entries.
type KnowledgeEntryModel struct {
	ID        uint   `gorm:"primaryKey"`
	Title     string `gorm:"not null"`
	Content   string `gorm:"type:text"`
	Tags      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *KnowledgeEntryModel) ToDomain() *knowledge.KnowledgeEntry {
	if m == nil {
		return nil
	}
	return &knowledge.KnowledgeEntry{
		ID:        m.ID,
		Title:     m.Title,
		Content:   m.Content,
		Tags:      m.Tags,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func KnowledgeEntryModelFromDomain(k *knowledge.KnowledgeEntry) *KnowledgeEntryModel {
	if k == nil {
		return nil
	}
	return &KnowledgeEntryModel{
		ID:        k.ID,
		Title:     k.Title,
		Content:   k.Content,
		Tags:      k.Tags,
		CreatedAt: k.CreatedAt,
		UpdatedAt: k.UpdatedAt,
	}
}
