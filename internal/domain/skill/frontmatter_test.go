package skill_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/skill"
)

func TestBuildAndParseFrontmatter(t *testing.T) {
	name := "troubleshoot-los-onu"
	desc := "Panduan penanganan modem indikator LOS merah"
	license := "Apache-2.0"
	compat := ">=1.0.0"
	tools := "mcp_device, bash"
	meta := map[string]string{
		"category": "network",
		"author":   "noc-polyglot",
	}
	bodyContent := "# SOP Troubleshoot LOS\n\n1. Periksa kabel patch cord.\n2. Cek redaman OLT."

	fm := skill.BuildFrontmatter(name, desc, license, compat, tools, meta)
	fullDoc := fm + bodyContent

	parsedMeta, parsedBody, err := skill.ParseFrontmatter(fullDoc)
	require.NoError(t, err)

	assert.Equal(t, name, parsedMeta.Name)
	assert.Equal(t, desc, parsedMeta.Description)
	assert.Equal(t, license, parsedMeta.License)
	assert.Equal(t, compat, parsedMeta.Compatibility)
	assert.Equal(t, tools, parsedMeta.AllowedTools)
	assert.Equal(t, "network", parsedMeta.Metadata["category"])
	assert.Equal(t, "noc-polyglot", parsedMeta.Metadata["author"])
	assert.Equal(t, bodyContent, parsedBody)
}

func TestParseFrontmatter_NoFrontmatter(t *testing.T) {
	doc := "# Pure Markdown Document\n\nNo frontmatter here."
	meta, body, err := skill.ParseFrontmatter(doc)
	require.NoError(t, err)
	assert.Empty(t, meta.Name)
	assert.Equal(t, doc, body)
}

func TestValidateSkillName(t *testing.T) {
	assert.NoError(t, skill.ValidateSkillName("valid-skill-name-123"))
	assert.NoError(t, skill.ValidateSkillName("simple"))

	assert.ErrorIs(t, skill.ValidateSkillName(""), skill.ErrInvalidSkillName)
	assert.ErrorIs(t, skill.ValidateSkillName("Invalid_Skill"), skill.ErrInvalidSkillName)
	assert.ErrorIs(t, skill.ValidateSkillName("skill name with spaces"), skill.ErrInvalidSkillName)
	assert.ErrorIs(t, skill.ValidateSkillName("-leading-dash"), skill.ErrInvalidSkillName)
}

func TestValidateResourcePath(t *testing.T) {
	assert.NoError(t, skill.ValidateResourcePath("scripts/helper.rsc"))
	assert.NoError(t, skill.ValidateResourcePath("templates/config.json"))

	assert.ErrorIs(t, skill.ValidateResourcePath(""), skill.ErrInvalidResourcePath)
	assert.ErrorIs(t, skill.ValidateResourcePath("../secret.txt"), skill.ErrInvalidResourcePath)
	assert.ErrorIs(t, skill.ValidateResourcePath("/etc/passwd"), skill.ErrInvalidResourcePath)
	assert.ErrorIs(t, skill.ValidateResourcePath("scripts/../../danger"), skill.ErrInvalidResourcePath)
}
