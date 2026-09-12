// F6-1: job locker berbasis PostgreSQL advisory lock.
package postgres_test

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
)

func TestAdvisoryLocker_NonPostgresAlwaysLocks(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	l := postgres.NewAdvisoryLocker(db)

	unlock, ok, err := l.TryLock(context.Background(), "run-billing")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.NotNil(t, unlock)
	unlock()
}

func TestAdvisoryLocker_PostgresMutualExclusion(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	_, _, dsn := startMigratedPostgres(t)
	db1, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	db2, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	l1 := postgres.NewAdvisoryLocker(db1)
	l2 := postgres.NewAdvisoryLocker(db2)
	ctx := context.Background()

	unlock1, ok, err := l1.TryLock(ctx, "run-billing")
	require.NoError(t, err)
	require.True(t, ok, "instance pertama harus mendapat lock")

	_, ok2, err := l2.TryLock(ctx, "run-billing")
	require.NoError(t, err)
	assert.False(t, ok2, "instance kedua tidak boleh mendapat lock yang sama")

	unlock1()
	unlock2, ok3, err := l2.TryLock(ctx, "run-billing")
	require.NoError(t, err)
	assert.True(t, ok3, "setelah unlock, instance kedua boleh mengambil lock")
	unlock2()
}
