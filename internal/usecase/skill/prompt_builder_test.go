package skill_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/skill"
	skillUC "github.com/quixiq/polyglot/internal/usecase/skill"
)

func TestRenderSkillsPrompt(t *testing.T) {
	skills := []skill.SkillInfo{
		{
			Name:        "web-search",
			Description: "Search the web for real-time data",
			Content:     "# Web Search SOP\n\nSearch and summarize concisely.",
		},
		{
			Name:        "troubleshoot-los",
			Description: "Handle red LOS light on ONU",
			Content:     "",
		},
	}

	prompt := skillUC.RenderSkillsPrompt(skills, "")
	assert.Contains(t, prompt, "<available_skills>")
	assert.Contains(t, prompt, "<name>web-search</name>")
	assert.Contains(t, prompt, "<content># Web Search SOP\n\nSearch and summarize concisely.</content>")
	assert.Contains(t, prompt, "<name>troubleshoot-los</name>")
	assert.Contains(t, prompt, "<description>Handle red LOS light on ONU</description>")
	assert.Contains(t, prompt, "</available_skills>")
}

func TestRequestSkillTool(t *testing.T) {
	skills := []skill.SkillInfo{
		{
			Name:        "billing-faq",
			Description: "Billing FAQ SOP",
			Content:     "Step 1: Check invoice.\nStep 2: Send payment link.",
		},
	}

	tool := skillUC.RequestSkillTool{Skills: skills}

	// 1. Found skill
	resp, err := tool.Run(skillUC.RequestSkillArgs{SkillName: "billing-faq"})
	require.NoError(t, err)
	assert.Contains(t, resp, "Skill 'billing-faq':")
	assert.Contains(t, resp, "Step 1: Check invoice.")

	// 2. Not found skill
	respNotFound, err := tool.Run(skillUC.RequestSkillArgs{SkillName: "non-existent"})
	require.NoError(t, err)
	assert.Contains(t, respNotFound, "Skill 'non-existent' not found. Available skills: billing-faq")
}

func TestFilterSkills(t *testing.T) {
	all := []skill.SkillInfo{
		{Name: "skill-a"},
		{Name: "skill-b"},
		{Name: "skill-c"},
	}

	// Empty selected returns all
	assert.Len(t, skillUC.FilterSkills(all, nil), 3)

	// Filtered returns only selected
	filtered := skillUC.FilterSkills(all, []string{"skill-a", "skill-c"})
	require.Len(t, filtered, 2)
	assert.Equal(t, "skill-a", filtered[0].Name)
	assert.Equal(t, "skill-c", filtered[1].Name)
}
