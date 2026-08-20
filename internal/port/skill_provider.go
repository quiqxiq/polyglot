package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/skill"
)

// SkillProvider menyediakan daftar skill aktif untuk runtime agent LLM dan tool calling.
type SkillProvider interface {
	ListSkills(ctx context.Context) ([]skill.SkillInfo, error)
	GetSkillContent(ctx context.Context, name string) (string, error)
}
