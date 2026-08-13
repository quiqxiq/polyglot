package models

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/llm"
)

// LLMConfigModel is the GORM database model for LLM configurations.
// TableName eksplisit ke `llm_configs` (migrasi 000002) — tanpa ini GORM
// memakai `llm_config_models`, divergen dari migrasi (prod).
// CostPer1MInput/Output memakai tag `column:` karena GORM memetakan angka 1M
// menjadi `cost_per1_m_input`, sementara migrasi memakai `cost_per_1m_input`.
type LLMConfigModel struct {
	ID              uint    `gorm:"primaryKey"`
	Provider        string  `gorm:"not null"`
	Model           string  `gorm:"not null"`
	APIKeyEncrypted string  `gorm:"not null"`
	Params          string  `gorm:"type:text"`
	MaxOutputTokens int     `gorm:"default:512"`
	IsActive        bool    `gorm:"default:false"`
	CostPer1MInput  float64 `gorm:"column:cost_per_1m_input;default:0"`
	CostPer1MOutput float64 `gorm:"column:cost_per_1m_output;default:0"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TableName maps LLMConfigModel ke tabel migrasi `llm_configs`.
func (LLMConfigModel) TableName() string { return "llm_configs" }

func (m *LLMConfigModel) ToDomain() *llm.LLMConfig {
	if m == nil {
		return nil
	}
	return &llm.LLMConfig{
		ID:              m.ID,
		Provider:        m.Provider,
		Model:           m.Model,
		APIKeyEncrypted: m.APIKeyEncrypted,
		Params:          m.Params,
		MaxOutputTokens: m.MaxOutputTokens,
		IsActive:        m.IsActive,
		CostPer1MInput:  m.CostPer1MInput,
		CostPer1MOutput: m.CostPer1MOutput,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func LLMConfigModelFromDomain(c *llm.LLMConfig) *LLMConfigModel {
	if c == nil {
		return nil
	}
	return &LLMConfigModel{
		ID:              c.ID,
		Provider:        c.Provider,
		Model:           c.Model,
		APIKeyEncrypted: c.APIKeyEncrypted,
		Params:          c.Params,
		MaxOutputTokens: c.MaxOutputTokens,
		IsActive:        c.IsActive,
		CostPer1MInput:  c.CostPer1MInput,
		CostPer1MOutput: c.CostPer1MOutput,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}
