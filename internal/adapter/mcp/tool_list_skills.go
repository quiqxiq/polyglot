package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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

func (s *Server) listSkills(ctx context.Context, _ *mcp.CallToolRequest, _ listSkillsArgs) (*mcp.CallToolResult, listSkillsOutput, error) {
	if s.skillRepo == nil {
		return toolOK(listSkillsOutput{
			Status:  "success",
			Summary: "Skill repository not configured",
		})
	}

	skills, err := s.skillRepo.ListSkills(ctx, "")
	if err != nil {
		return toolError(listSkillsOutput{Status: "error", Summary: err.Error()})
	}

	if len(skills) == 0 {
		return toolOK(listSkillsOutput{
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

	return toolOK(listSkillsOutput{
		Status:  "success",
		Summary: fmt.Sprintf("Found %d skills", len(skills)),
		Skills:  items,
	})
}
