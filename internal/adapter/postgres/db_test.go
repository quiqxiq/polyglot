package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDB_InvalidDSN(t *testing.T) {
	_, err := NewDB(context.Background(), "postgres://invalid::/not-a-url")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "postgres")
}

func TestNewDB_Integration(t *testing.T) {
	if os.Getenv("POLYGLOT_RUN_DB_INTEGRATION") != "1" {
		t.Skip("set POLYGLOT_RUN_DB_INTEGRATION=1 to run database integration test")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	db, err := NewDB(ctx, dsn)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	require.NoError(t, db.Ping(ctx))
	assert.NotNil(t, db.GORM())
}
