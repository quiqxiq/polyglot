package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/testutil"
)

func TestUserRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&userModel{}))

	repo := NewUserRepository(db)

	hash, err := auth.HashPassword("secret123")
	require.NoError(t, err)

	u := port.User{
		ID:           "usr-001",
		Username:     "admin",
		PasswordHash: hash,
		FullName:     "Administrator",
		Email:        "admin@example.com",
		Phone:        "0812",
		Role:         "superadmin",
		IsActive:     true,
	}

	created, err := repo.Create(ctx, u)
	require.NoError(t, err)
	assert.Equal(t, u.Username, created.Username)

	found, err := repo.FindByUsername(ctx, "admin")
	require.NoError(t, err)
	require.NoError(t, auth.CheckPassword("secret123", found.PasswordHash))

	require.NoError(t, repo.UpdateLastLogin(ctx, created.ID))

	newHash, err := auth.HashPassword("newsecret")
	require.NoError(t, err)
	require.NoError(t, repo.UpdatePassword(ctx, created.ID, newHash))

	found, err = repo.FindByUsername(ctx, "admin")
	require.NoError(t, err)
	require.NoError(t, auth.CheckPassword("newsecret", found.PasswordHash))
	assert.NotNil(t, found.LastLoginAt)
}

func TestUserRepository_FindByUsername_NotFound(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&userModel{}))

	repo := NewUserRepository(db)

	_, err = repo.FindByUsername(ctx, "missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
