package models

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/llm"
)

// LLMConfigModel is the GORM database model for LLM configurations.
type LLMConfigModel struct {
	ID              uint    `gorm:"primaryKey"`
	Provider        string  `gorm:"not null"`
	Model           string  `gorm:"not null"`
	APIKeyEncrypted string  `gorm:"not null"`
	Params          string  `gorm:"type:text"`
	MaxOutputTokens int     `gorm:"default:512"`
	IsActive        bool    `gorm:"default:false"`
	CostPer1MInput  float64 `gorm:"default:0"`
	CostPer1MOutput float64 `gorm:"default:0"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

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
