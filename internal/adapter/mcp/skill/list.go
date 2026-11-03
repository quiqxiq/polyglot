package skill

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/quixiq/polyglot/internal/adapter/mcp/mcputil"
)

type listSkillsArgs struct{}

type skillItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Definition string `json:"definition"`
	SourceType string `json:"source_type"`
	IsEnabled  bool   `json:"is_enabled"`
}

type listSkillsOutput struct {
	Status  string      `json:"status"`
	Summary string      `json:"summary"`
	Skills  []skillItem `json:"skills,omitempty"`
}

// ListSkills handles the "list_skills" tool.
func (h *Handler) ListSkills(ctx context.Context, _ *mcp.CallToolRequest, _ listSkillsArgs) (*mcp.CallToolResult, listSkillsOutput, error) {
	repo := h.repo()
	if repo == nil {
		return mcputil.ToolOK(listSkillsOutput{
			Status:  "success",
			Summary: "Skill repository not configured",
		})
	}

	skills, err := repo.ListSkills(ctx, "")
	if err != nil {
		return mcputil.ToolError(listSkillsOutput{Status: "error", Summary: err.Error()})
	}

	if len(skills) == 0 {
		return mcputil.ToolOK(listSkillsOutput{
			Status:  "success",
			Summary: "No skills configured in system",
		})
	}

	items := make([]skillItem, len(skills))
	for i, sk := range skills {
		items[i] = skillItem{
			ID:         sk.ID,
			Name:       sk.Name,
			Definition: sk.Definition,
			SourceType: sk.SourceType,
			IsEnabled:  sk.Enabled,
		}
	}

	return mcputil.ToolOK(listSkillsOutput{
		Status:  "success",
		Summary: fmt.Sprintf("Found %d skills", len(skills)),
		Skills:  items,
	})
}
