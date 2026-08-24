package postgres_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
	"github.com/quixiq/polyglot/internal/port"
)

func TestPortalRepository_OTPLifecycle(t *testing.T) {
	db := setupISPDB(t)
	repo := postgres.NewPortalRepository(db)
	ctx := context.Background()

	code := "123456"
	sum := sha256.Sum256([]byte(code))
	hash := hex.EncodeToString(sum[:])
	now := time.Now()

	require.NoError(t, repo.SaveOTP(ctx, domainCustomer.PortalOTP{
		ID: "otp-1", TenantID: "tenant-default", Phone: "0812",
		CodeHash: hash, Purpose: "PORTAL_LOGIN",
		ExpiresAt: now.Add(5 * time.Minute), CreatedAt: now,
	}))

	// OTP salah → attempts naik, belum locked.
	ok, err := repo.ConsumeOTP(ctx, "0812", "salah", 5)
	require.NoError(t, err)
	assert.False(t, ok)

	// OTP benar → consumed.
	ok, err = repo.ConsumeOTP(ctx, "0812", hash, 5)
	require.NoError(t, err)
	assert.True(t, ok)

	// Sudah dipakai → not found.
	_, err = repo.ConsumeOTP(ctx, "0812", hash, 5)
	assert.ErrorIs(t, err, port.ErrOTPNotFound)
}

func TestPortalRepository_OTPLocked(t *testing.T) {
	db := setupISPDB(t)
	repo := postgres.NewPortalRepository(db)
	ctx := context.Background()
	sum := sha256.Sum256([]byte("999999"))
	now := time.Now()
	require.NoError(t, repo.SaveOTP(ctx, domainCustomer.PortalOTP{
		ID: "otp-2", Phone: "0813", CodeHash: hex.EncodeToString(sum[:]),
		Purpose:   "PORTAL_LOGIN",
		ExpiresAt: now.Add(5 * time.Minute), CreatedAt: now,
	}))

	for i := 0; i < 4; i++ {
		ok, err := repo.ConsumeOTP(ctx, "0813", "wrong", 5)
		require.NoError(t, err)
		assert.False(t, ok)
	}
	// Percobaan ke-5 salah → terkunci.
	_, err := repo.ConsumeOTP(ctx, "0813", "wrong", 5)
	assert.ErrorIs(t, err, port.ErrOTPLocked)
}

func TestNotificationRetry_PendingLimit_AndAttempts(t *testing.T) {
	db := setupISPDB(t)
	repo := postgres.NewNotificationRepository(db)
	ctx := context.Background()
	for i, st := range []string{
		domainNotification.StatusQueued, domainNotification.StatusQueued,
		domainNotification.StatusFailed,
	} {
		attempts := 0
		if st == domainNotification.StatusFailed {
			attempts = 3
		}
		require.NoError(t, db.Exec(
			`INSERT INTO wa_notifications (id, recipient_phone, message_type, message_content, status, attempts)
			 VALUES (?, '0811','T','x', ?, ?)`,
			[]string{"wa-a", "wa-b", "wa-c"}[i], st, attempts).Error)
	}

	pending, err := repo.PendingWithRetryLimit(ctx, 10, 3)
	require.NoError(t, err)
	require.Len(t, pending, 2) // yang failed attempts=3 tak diambil

	require.NoError(t, repo.MarkFailedWithAttempt(ctx, pending[0].ID, "gagal kirim", 1))
	var got struct{ Attempts int }
	require.NoError(t, db.Raw(`SELECT attempts FROM wa_notifications WHERE id=?`, pending[0].ID).
		Scan(&got).Error)
	assert.Equal(t, 1, got.Attempts)
}
