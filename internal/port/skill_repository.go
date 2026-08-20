package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/skill"
)

// SkillRepository mengelola persistensi data skill di database relasional (PostgreSQL).
type SkillRepository interface {
	ListSkills(ctx context.Context) ([]skill.Skill, error)
	GetSkillByID(ctx context.Context, id uint) (*skill.Skill, error)
	GetSkillBySlug(ctx context.Context, slug string) (*skill.Skill, error)
	CreateSkill(ctx context.Context, s *skill.Skill) error
	UpdateSkill(ctx context.Context, s *skill.Skill) error
	DeleteSkill(ctx context.Context, id uint) error
	ToggleSkillEnabled(ctx context.Context, slug string, enabled bool) error

	// File operations
	SaveSkillFile(ctx context.Context, skillID uint, f *skill.SkillFile) error
	DeleteSkillFile(ctx context.Context, fileID uint) error
	DeleteSkillFileByPath(ctx context.Context, skillID uint, filePath string) error

	// Global System Prompt
	GetGlobalSystemPrompt(ctx context.Context) (string, error)
	SaveGlobalSystemPrompt(ctx context.Context, content string) error
}
