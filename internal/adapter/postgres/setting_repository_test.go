package postgres_test

import (
	"context"
	"strings"
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

// TestSettingRepository_EncryptsSecretSettings membuktikan key sensitif
// gateway disimpan terenkripsi (prefix enc:v1:) dan didekripsi transparan
// saat dibaca (F4-7).
func TestSettingRepository_EncryptsSecretSettings(t *testing.T) {
	db := setupSettingTestDB(t)
	repo := postgres.NewSettingRepository(db).WithVault(fakeVaultForTest{})
	ctx := context.Background()

	require.NoError(t, repo.Set(ctx, "gw.tripay.private_key", "RAHASIA-123", "gateway", "Tripay private key"))
	require.NoError(t, repo.Set(ctx, "company_name", "Polyglot", "general", "nama"))

	// Nilai mentah di DB terenkripsi.
	var raw model.SystemSettingModel
	require.NoError(t, db.First(&raw, "key = ?", "gw.tripay.private_key").Error)
	assert.True(t, strings.HasPrefix(raw.Value, "enc:v1:"), raw.Value)
	assert.NotEqual(t, "RAHASIA-123", raw.Value)

	// GetValue & Get mengembalikan plaintext.
	assert.Equal(t, "RAHASIA-123", repo.GetValue(ctx, "gw.tripay.private_key", ""))
	s, err := repo.Get(ctx, "gw.tripay.private_key")
	require.NoError(t, err)
	assert.Equal(t, "RAHASIA-123", s.Value)

	// Key non-sensitif tidak dienkripsi.
	var rawCompany model.SystemSettingModel
	require.NoError(t, db.First(&rawCompany, "key = ?", "company_name").Error)
	assert.Equal(t, "Polyglot", rawCompany.Value)

	// BatchSet juga mengenkripsi.
	require.NoError(t, repo.BatchSet(ctx, []setting.Setting{
		{Key: "gw.xendit.secret_key", Value: "SK-1", Category: "gateway"},
	}))
	var rawX model.SystemSettingModel
	require.NoError(t, db.First(&rawX, "key = ?", "gw.xendit.secret_key").Error)
	assert.True(t, strings.HasPrefix(rawX.Value, "enc:v1:"))
	assert.Equal(t, "SK-1", repo.GetValue(ctx, "gw.xendit.secret_key", ""))

	// Tanpa vault → nilai apa adanya (kompatibel repo lama).
	plain := postgres.NewSettingRepository(db)
	assert.True(t, strings.HasPrefix(plain.GetValue(ctx, "gw.tripay.private_key", ""), "enc:v1:"))
}
