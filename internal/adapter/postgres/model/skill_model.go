package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/skill"
)

// SkillMetadataModel represents the skills_metadata table in PostgreSQL.
type SkillMetadataModel struct {
	ID         string    `gorm:"primaryKey;size:36" json:"id"`
	UserID     string    `gorm:"index;size:36" json:"user_id,omitempty"`
	Name       string    `gorm:"index;size:255;not null" json:"name"`
	Definition string    `gorm:"type:text" json:"definition,omitempty"`
	SourceType string    `gorm:"size:32;default:'inline'" json:"source_type"`
	SourceURL  string    `gorm:"size:512" json:"source_url,omitempty"`
	Version    string    `gorm:"size:64" json:"version,omitempty"`
	Enabled    bool      `gorm:"default:true;index" json:"enabled"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SkillMetadataModel) TableName() string {
	return "skills_metadata"
}

func (m *SkillMetadataModel) ToDomain() *skill.SkillMetadataRecord {
	if m == nil {
		return nil
	}
	return &skill.SkillMetadataRecord{
		ID:         m.ID,
		UserID:     m.UserID,
		Name:       m.Name,
		Definition: m.Definition,
		SourceType: m.SourceType,
		SourceURL:  m.SourceURL,
		Version:    m.Version,
		Enabled:    m.Enabled,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

func SkillMetadataModelFromDomain(r *skill.SkillMetadataRecord) *SkillMetadataModel {
	if r == nil {
		return nil
	}
	return &SkillMetadataModel{
		ID:         r.ID,
		UserID:     r.UserID,
		Name:       r.Name,
		Definition: r.Definition,
		SourceType: r.SourceType,
		SourceURL:  r.SourceURL,
		Version:    r.Version,
		Enabled:    r.Enabled,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

// GlobalPromptModel represents the bot_global_prompts table in PostgreSQL.
type GlobalPromptModel struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Key       string    `gorm:"type:varchar(50);unique;not null;default:'default'"`
	Content   string    `gorm:"type:text"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (GlobalPromptModel) TableName() string {
	return "bot_global_prompts"
}
