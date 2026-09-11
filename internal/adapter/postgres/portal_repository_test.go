// Portal OTP repository tests (Fase 0 — PLAN-ISP-CORE-HARDENING.md).
//
// F5-1 (RED): ConsumeOTP mengembalikan ErrOTPLocked dari dalam Transaction,
// sehingga GORM me-rollback attempts++ dan consumed_at — lockout brute-force
// tidak pernah benar-benar tersimpan.
package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
)

const (
	testOTPPhone = "6281200000000"
	testOTPWrong = "hash-wrong"
	testOTPRight = "hash-right"
)

func setupPortalOTPDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.PortalOTPModel{}))
	return db
}

func seedOTP(t *testing.T, db *gorm.DB, codeHash string, attempts int) *postgres.PortalRepository {
	t.Helper()
	repo := postgres.NewPortalRepository(db)
	require.NoError(t, repo.SaveOTP(context.Background(), domainCustomer.PortalOTP{
		ID: "otp-1", TenantID: "tenant-default", Phone: testOTPPhone,
		CodeHash: codeHash, Purpose: "PORTAL_LOGIN", Attempts: attempts,
		ExpiresAt: time.Now().Add(5 * time.Minute), CreatedAt: time.Now(),
	}))
	return repo
}

func loadOTP(t *testing.T, db *gorm.DB) model.PortalOTPModel {
	t.Helper()
	var m model.PortalOTPModel
	require.NoError(t, db.First(&m, "id = ?", "otp-1").Error)
	return m
}

func TestConsumeOTP_CorrectCodeConsumesOnce(t *testing.T) {
	db := setupPortalOTPDB(t)
	repo := seedOTP(t, db, testOTPRight, 0)
	ctx := context.Background()

	matched, err := repo.ConsumeOTP(ctx, testOTPPhone, testOTPRight, 3)
	require.NoError(t, err)
	assert.True(t, matched)
	assert.NotNil(t, loadOTP(t, db).ConsumedAt)

	// OTP yang sudah dikonsumsi tidak bisa dipakai lagi.
	_, err = repo.ConsumeOTP(ctx, testOTPPhone, testOTPRight, 3)
	assert.ErrorIs(t, err, domainCustomer.ErrOTPNotFound)
}

func TestConsumeOTP_WrongCodeIncrementsAttempts(t *testing.T) {
	db := setupPortalOTPDB(t)
	repo := seedOTP(t, db, testOTPRight, 0)
	ctx := context.Background()

	matched, err := repo.ConsumeOTP(ctx, testOTPPhone, testOTPWrong, 3)
	require.NoError(t, err)
	assert.False(t, matched)
	assert.Equal(t, 1, loadOTP(t, db).Attempts)
}

func TestConsumeOTP_LocksAfterMaxAttempts(t *testing.T) {
	db := setupPortalOTPDB(t)
	repo := seedOTP(t, db, testOTPRight, 0)
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		_, err := repo.ConsumeOTP(ctx, testOTPPhone, testOTPWrong, 3)
		if i < 3 {
			require.NoError(t, err)
			continue
		}
		assert.ErrorIs(t, err, domainCustomer.ErrOTPLocked)
	}

	locked := loadOTP(t, db)
	assert.GreaterOrEqual(t, locked.Attempts, 3)
	assert.NotNil(t, locked.ConsumedAt, "OTP harus dikonsumsi saat terkunci")

	// Bahkan kode yang benar harus ditolak setelah terkunci.
	matched, err := repo.ConsumeOTP(ctx, testOTPPhone, testOTPRight, 3)
	assert.False(t, matched)
	assert.Error(t, err)
}

func TestConsumeOTP_AlreadyAtMaxIsLockedAndConsumed(t *testing.T) {
	db := setupPortalOTPDB(t)
	repo := seedOTP(t, db, testOTPRight, 3)
	ctx := context.Background()

	matched, err := repo.ConsumeOTP(ctx, testOTPPhone, testOTPRight, 3)
	assert.False(t, matched)
	assert.ErrorIs(t, err, domainCustomer.ErrOTPLocked)
	assert.NotNil(t, loadOTP(t, db).ConsumedAt)
}
