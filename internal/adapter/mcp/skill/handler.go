package skill

import "github.com/quixiq/polyglot/internal/port"

// Handler holds dependencies for Skill MCP tools.
type Handler struct {
	GetSkillRepo func() port.SkillRepository
}

// NewHandler creates a new Skill tool handler.
func NewHandler(repoGetter func() port.SkillRepository) *Handler {
	return &Handler{
		GetSkillRepo: repoGetter,
	}
}

func (h *Handler) repo() port.SkillRepository {
	if h.GetSkillRepo != nil {
		return h.GetSkillRepo()
	}
	return nil
}
