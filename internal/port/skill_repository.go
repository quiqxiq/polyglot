package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/skill"
)

// SkillRepository mengelola persistensi metadata skill di database relasional (PostgreSQL).
type SkillRepository interface {
	ListSkills(ctx context.Context, userID string) ([]skill.SkillMetadataRecord, error)
	GetSkill(ctx context.Context, userID, name string) (*skill.SkillMetadataRecord, error)
	SaveSkillMetadata(ctx context.Context, rec *skill.SkillMetadataRecord) error
	DeleteSkillMetadata(ctx context.Context, userID, name string) error
	ListGitSkills(ctx context.Context) ([]skill.SkillMetadataRecord, error)
	ToggleSkillEnabled(ctx context.Context, userID, name string, enabled bool) error

	// Global System Prompt
	GetGlobalSystemPrompt(ctx context.Context) (string, error)
	SaveGlobalSystemPrompt(ctx context.Context, content string) error
}
