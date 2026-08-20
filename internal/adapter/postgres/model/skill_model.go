package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/skill"
)

// SkillModel represents the bot_skills table in PostgreSQL.
type SkillModel struct {
	ID          uint              `gorm:"primaryKey;autoIncrement"`
	Slug        string            `gorm:"type:varchar(100);uniqueIndex;not null"`
	Name        string            `gorm:"type:varchar(255);not null"`
	Description string            `gorm:"type:text"`
	IsEnabled   bool              `gorm:"default:true;index"`
	Files       []SkillFileModel  `gorm:"foreignKey:SkillID;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time         `gorm:"autoCreateTime"`
	UpdatedAt   time.Time         `gorm:"autoUpdateTime"`
}

func (SkillModel) TableName() string {
	return "bot_skills"
}

func (m *SkillModel) ToDomain() *skill.Skill {
	if m == nil {
		return nil
	}
	var files []skill.SkillFile
	for _, f := range m.Files {
		if d := f.ToDomain(); d != nil {
			files = append(files, *d)
		}
	}
	return &skill.Skill{
		ID:          m.ID,
		Slug:        m.Slug,
		Name:        m.Name,
		Description: m.Description,
		IsEnabled:   m.IsEnabled,
		Files:       files,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func SkillModelFromDomain(s *skill.Skill) *SkillModel {
	if s == nil {
		return nil
	}
	var files []SkillFileModel
	for _, f := range s.Files {
		if m := SkillFileModelFromDomain(&f); m != nil {
			files = append(files, *m)
		}
	}
	return &SkillModel{
		ID:          s.ID,
		Slug:        s.Slug,
		Name:        s.Name,
		Description: s.Description,
		IsEnabled:   s.IsEnabled,
		Files:       files,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

// SkillFileModel represents the bot_skill_files table in PostgreSQL.
type SkillFileModel struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	SkillID     uint      `gorm:"index;not null"`
	Name        string    `gorm:"type:varchar(255);not null"`
	FilePath    string    `gorm:"type:varchar(500);index;not null"`
	Content     string    `gorm:"type:text"`
	IsReference bool      `gorm:"default:false"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (SkillFileModel) TableName() string {
	return "bot_skill_files"
}

func (m *SkillFileModel) ToDomain() *skill.SkillFile {
	if m == nil {
		return nil
	}
	return &skill.SkillFile{
		ID:          m.ID,
		SkillID:     m.SkillID,
		Name:        m.Name,
		FilePath:    m.FilePath,
		Content:     m.Content,
		IsReference: m.IsReference,
		UpdatedAt:   m.UpdatedAt,
	}
}

func SkillFileModelFromDomain(f *skill.SkillFile) *SkillFileModel {
	if f == nil {
		return nil
	}
	return &SkillFileModel{
		ID:          f.ID,
		SkillID:     f.SkillID,
		Name:        f.Name,
		FilePath:    f.FilePath,
		Content:     f.Content,
		IsReference: f.IsReference,
		UpdatedAt:   f.UpdatedAt,
	}
}

// GlobalPromptModel represents the bot_global_prompts table in PostgreSQL.
type GlobalPromptModel struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Key       string    `gorm:"type:varchar(50);uniqueIndex;not null;default:'default'"`
	Content   string    `gorm:"type:text"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (GlobalPromptModel) TableName() string {
	return "bot_global_prompts"
}
