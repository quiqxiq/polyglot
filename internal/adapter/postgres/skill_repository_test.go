package postgres_test

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/skill"
)

func setupSkillTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&model.SkillMetadataModel{},
		&model.GlobalPromptModel{},
	)
	require.NoError(t, err)

	return db
}

func TestSkillRepository_CRUD(t *testing.T) {
	db := setupSkillTestDB(t)
	repo := postgres.NewSkillRepository(db)
	ctx := context.Background()

	// 1. SaveSkillMetadata
	rec := &skill.SkillMetadataRecord{
		Name:       "mikrotik-doctor",
		Definition: "Skill script instructions...",
		Enabled:    true,
		SourceType: "git",
	}
	err := repo.SaveSkillMetadata(ctx, rec)
	require.NoError(t, err)
	assert.NotEmpty(t, rec.ID)

	// 2. GetSkill
	found, err := repo.GetSkill(ctx, "", "mikrotik-doctor")
	require.NoError(t, err)
	assert.Equal(t, "mikrotik-doctor", found.Name)
	assert.True(t, found.Enabled)

	// 3. ListSkills & ListGitSkills
	skills, err := repo.ListSkills(ctx, "")
	require.NoError(t, err)
	assert.Len(t, skills, 1)

	gitSkills, err := repo.ListGitSkills(ctx)
	require.NoError(t, err)
	assert.Len(t, gitSkills, 1)

	// 4. ToggleSkillEnabled
	err = repo.ToggleSkillEnabled(ctx, "", "mikrotik-doctor", false)
	require.NoError(t, err)

	afterToggle, err := repo.GetSkill(ctx, "", "mikrotik-doctor")
	require.NoError(t, err)
	assert.False(t, afterToggle.Enabled)

	// 5. Global System Prompt
	err = repo.SaveGlobalSystemPrompt(ctx, "You are a network engineer assistant.")
	require.NoError(t, err)

	prompt, err := repo.GetGlobalSystemPrompt(ctx)
	require.NoError(t, err)
	assert.Equal(t, "You are a network engineer assistant.", prompt)

	// 6. DeleteSkillMetadata
	err = repo.DeleteSkillMetadata(ctx, "", "mikrotik-doctor")
	require.NoError(t, err)

	_, err = repo.GetSkill(ctx, "", "mikrotik-doctor")
	assert.Error(t, err)
}
