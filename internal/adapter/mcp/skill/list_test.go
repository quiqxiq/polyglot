package skill

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/skill"
	"github.com/quixiq/polyglot/internal/port"
)

type memSkillRepo struct {
	skills []skill.SkillMetadataRecord
}

func (m *memSkillRepo) ListSkills(_ context.Context, _ string) ([]skill.SkillMetadataRecord, error) {
	return m.skills, nil
}

func (m *memSkillRepo) GetSkill(_ context.Context, _, _ string) (*skill.SkillMetadataRecord, error) {
	return nil, nil
}

func (m *memSkillRepo) SaveSkillMetadata(_ context.Context, _ *skill.SkillMetadataRecord) error {
	return nil
}

func (m *memSkillRepo) DeleteSkillMetadata(_ context.Context, _, _ string) error {
	return nil
}

func (m *memSkillRepo) ListGitSkills(_ context.Context) ([]skill.SkillMetadataRecord, error) {
	return nil, nil
}

func (m *memSkillRepo) ToggleSkillEnabled(_ context.Context, _, _ string, _ bool) error {
	return nil
}

func (m *memSkillRepo) GetGlobalSystemPrompt(_ context.Context) (string, error) {
	return "", nil
}

func (m *memSkillRepo) SaveGlobalSystemPrompt(_ context.Context, _ string) error {
	return nil
}

func TestListSkills_NoRepo(t *testing.T) {
	h := NewHandler(nil)
	res, out, err := h.ListSkills(context.Background(), nil, listSkillsArgs{})
	require.NoError(t, err)
	assert.False(t, res.IsError)
	assert.Equal(t, "success", out.Status)
	assert.Contains(t, out.Summary, "not configured")
}

func TestListSkills_Empty(t *testing.T) {
	repo := &memSkillRepo{skills: nil}
	h := NewHandler(func() port.SkillRepository { return repo })
	res, out, err := h.ListSkills(context.Background(), nil, listSkillsArgs{})
	require.NoError(t, err)
	assert.False(t, res.IsError)
	assert.Equal(t, "success", out.Status)
	assert.Contains(t, out.Summary, "No skills configured")
}

func TestListSkills_Success(t *testing.T) {
	repo := &memSkillRepo{
		skills: []skill.SkillMetadataRecord{
			{ID: "sk-1", Name: "Mikrotik Reboot", Definition: "Reboot router SOP", SourceType: "builtin", Enabled: true},
		},
	}
	h := NewHandler(func() port.SkillRepository { return repo })
	res, out, err := h.ListSkills(context.Background(), nil, listSkillsArgs{})
	require.NoError(t, err)
	assert.False(t, res.IsError)
	assert.Equal(t, "success", out.Status)
	assert.Len(t, out.Skills, 1)
	assert.Equal(t, "Mikrotik Reboot", out.Skills[0].Name)
}
