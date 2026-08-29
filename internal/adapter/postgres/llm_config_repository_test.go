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
	"github.com/quixiq/polyglot/internal/domain/llm"
)

func setupLLMTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&model.LLMConfigModel{},
		&model.MessageModel{},
	)
	require.NoError(t, err)

	return db
}

func TestLLMConfigRepository_CRUD(t *testing.T) {
	db := setupLLMTestDB(t)
	repo := postgres.NewLLMConfigRepository(db)
	ctx := context.Background()

	// 1. Create
	cfg := &llm.Config{
		Provider:        "openai",
		Model:           "gpt-4o-mini",
		CostPer1MInput:  0.15,
		CostPer1MOutput: 0.60,
		IsActive:        true,
	}
	err := repo.Create(ctx, cfg)
	require.NoError(t, err)
	assert.NotZero(t, cfg.ID)

	// 2. FindByID
	found, err := repo.FindByID(ctx, cfg.ID)
	require.NoError(t, err)
	assert.Equal(t, "openai", found.Provider)
	assert.Equal(t, "gpt-4o-mini", found.Model)

	// 3. FindActive
	active, err := repo.FindActive(ctx)
	require.NoError(t, err)
	assert.Equal(t, cfg.ID, active.ID)

	// 4. FindAll
	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	// 5. Update
	cfg.Model = "gpt-4o"
	err = repo.Update(ctx, cfg)
	require.NoError(t, err)

	updated, err := repo.FindByID(ctx, cfg.ID)
	require.NoError(t, err)
	assert.Equal(t, "gpt-4o", updated.Model)

	// 6. SetActive
	cfg2 := &llm.Config{
		Provider: "groq",
		Model:    "llama-3.3-70b-versatile",
		IsActive: false,
	}
	err = repo.Create(ctx, cfg2)
	require.NoError(t, err)

	err = repo.SetActive(ctx, cfg2.ID)
	require.NoError(t, err)

	activeAfter, err := repo.FindActive(ctx)
	require.NoError(t, err)
	assert.Equal(t, cfg2.ID, activeAfter.ID)

	// 7. Delete
	err = repo.Delete(ctx, cfg.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, cfg.ID)
	assert.Error(t, err)
}
