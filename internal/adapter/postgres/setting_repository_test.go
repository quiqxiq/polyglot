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
	"github.com/quixiq/polyglot/internal/domain/setting"
)

func setupSettingTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.SystemSettingModel{})
	require.NoError(t, err)

	return db
}

func TestSettingRepository_CRUD(t *testing.T) {
	db := setupSettingTestDB(t)
	repo := postgres.NewSettingRepository(db)
	ctx := context.Background()

	// 1. Set
	err := repo.Set(ctx, "company_name", "Polyglot ISP", "general", "Nama perusahaan")
	require.NoError(t, err)

	// 2. Get
	s, err := repo.Get(ctx, "company_name")
	require.NoError(t, err)
	assert.Equal(t, "company_name", s.Key)
	assert.Equal(t, "Polyglot ISP", s.Value)

	// 3. GetValue
	val := repo.GetValue(ctx, "company_name", "Default")
	assert.Equal(t, "Polyglot ISP", val)

	missing := repo.GetValue(ctx, "missing_key", "Fallback")
	assert.Equal(t, "Fallback", missing)

	// 4. BatchSet
	err = repo.BatchSet(ctx, []setting.Setting{
		{Key: "tax_rate", Value: "11", Category: "billing", Description: "PPN rate"},
		{Key: "currency", Value: "IDR", Category: "billing", Description: "Mata uang"},
	})
	require.NoError(t, err)

	// 5. GetByCategory
	billingSettings, err := repo.GetByCategory(ctx, "billing")
	require.NoError(t, err)
	assert.Len(t, billingSettings, 2)

	// 6. GetAll
	all, err := repo.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 3)

	// 7. BotSettings
	botSettings, err := repo.GetBotSettings(ctx)
	require.NoError(t, err)
	assert.NotNil(t, botSettings)

	botSettings.BurstLimit = 10
	err = repo.SaveBotSettings(ctx, botSettings)
	require.NoError(t, err)

	updatedBotSettings, err := repo.GetBotSettings(ctx)
	require.NoError(t, err)
	assert.Equal(t, 10, updatedBotSettings.BurstLimit)
}

