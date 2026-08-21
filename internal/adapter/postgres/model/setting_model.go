package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/setting"
)

// SystemSettingModel is the GORM database model for system_settings.
type SystemSettingModel struct {
	Key         string    `gorm:"type:varchar(100);primaryKey"`
	Value       string    `gorm:"type:text;not null"`
	Category    string    `gorm:"type:varchar(50);not null;default:'general'"`
	Description string    `gorm:"type:text"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// TableName maps SystemSettingModel to the migration table `system_settings`.
func (SystemSettingModel) TableName() string { return "system_settings" }

func (m *SystemSettingModel) ToDomain() *setting.Setting {
	if m == nil {
		return nil
	}
	return &setting.Setting{
		Key:         m.Key,
		Value:       m.Value,
		Category:    m.Category,
		Description: m.Description,
		UpdatedAt:   m.UpdatedAt,
	}
}

func SystemSettingModelFromDomain(s *setting.Setting) *SystemSettingModel {
	if s == nil {
		return nil
	}
	return &SystemSettingModel{
		Key:         s.Key,
		Value:       s.Value,
		Category:    s.Category,
		Description: s.Description,
		UpdatedAt:   s.UpdatedAt,
	}
}
