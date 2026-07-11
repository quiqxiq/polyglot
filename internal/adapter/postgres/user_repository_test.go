package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/testutil"
)

func TestUserRepository_CreateAndFind(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&userModel{}))

	repo := NewUserRepository(db)

	u := port.User{
		ID:           "user-001",
		Username:     "admin",
		PasswordHash: "$2a$10$fakehash",
		FullName:     "Admin User",
		Email:        "admin@example.com",
		Phone:        "08123456789",
		Role:         "superadmin",
		IsActive:     true,
	}

	created, err := repo.Create(ctx, u)
	require.NoError(t, err)
	assert.Equal(t, u.Username, created.Username)
	assert.Equal(t, u.Role, created.Role)

	found, err := repo.FindByUsername(ctx, "admin")
	require.NoError(t, err)
	assert.Equal(t, u.FullName, found.FullName)
	assert.Equal(t, u.Role, found.Role)
}

func TestUserRepository_FindByUsername_NotFound(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&userModel{}))

	repo := NewUserRepository(db)

	_, err = repo.FindByUsername(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestUserRepository_UpdatePassword(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&userModel{}))

	repo := NewUserRepository(db)

	u := port.User{
		ID:           "user-002",
		Username:     "staff",
		PasswordHash: "$2a$10$oldhash",
		FullName:     "Staff User",
		Role:         "staff",
		IsActive:     true,
	}
	_, err = repo.Create(ctx, u)
	require.NoError(t, err)

	require.NoError(t, repo.UpdatePassword(ctx, "user-002", "$2a$10$newhash"))

	found, err := repo.FindByUsername(ctx, "staff")
	require.NoError(t, err)
	assert.Equal(t, "$2a$10$newhash", found.PasswordHash)
}

func TestUserRepository_UpdateLastLogin(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&userModel{}))

	repo := NewUserRepository(db)

	u := port.User{
		ID:           "user-003",
		Username:     "teknisi",
		PasswordHash: "$2a$10$hash",
		FullName:     "Teknisi User",
		Role:         "teknisi",
		IsActive:     true,
	}
	_, err = repo.Create(ctx, u)
	require.NoError(t, err)

	require.NoError(t, repo.UpdateLastLogin(ctx, "user-003"))

	found, err := repo.FindByUsername(ctx, "teknisi")
	require.NoError(t, err)
	assert.NotNil(t, found.LastLoginAt)
}

func TestUserRepository_UpdatePassword_NotFound(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&userModel{}))

	repo := NewUserRepository(db)

	err = repo.UpdatePassword(ctx, "no-such-user", "$2a$10$hash")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
