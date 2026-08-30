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
	"github.com/quixiq/polyglot/internal/domain/customer"
)

func setupUserTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.UserModel{}, &model.UserDeviceModel{})
	require.NoError(t, err)

	return db
}

func TestUserRepository_CRUD(t *testing.T) {
	db := setupUserTestDB(t)
	repo := postgres.NewUserRepository(db)
	ctx := context.Background()

	// 1. Create
	user := &customer.User{
		Username:     "netops_admin",
		Email:        "admin@polyglot.net",
		PasswordHash: "$2a$10$hashedsecret",
		Role:         "admin",
		IsActive:     true,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)
	assert.NotZero(t, user.ID)

	// 2. FindByID
	found, err := repo.FindByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "netops_admin", found.Username)
	assert.Equal(t, "admin@polyglot.net", found.Email)
	assert.Equal(t, "admin", found.Role)

	// 3. FindByUsername
	byUsername, err := repo.FindByUsername(ctx, "netops_admin")
	require.NoError(t, err)
	assert.Equal(t, user.ID, byUsername.ID)

	// 4. FindByEmail
	byEmail, err := repo.FindByEmail(ctx, "admin@polyglot.net")
	require.NoError(t, err)
	assert.Equal(t, user.ID, byEmail.ID)

	// 5. Count & List
	count, err := repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	users, total, err := repo.List(ctx, 1, 10, "netops")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, users, 1)

	// 6. FindAll & FindByRoles
	allUsers, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, allUsers, 1)

	byRoles, err := repo.FindByRoles(ctx, []string{"admin", "technician"}, true)
	require.NoError(t, err)
	assert.Len(t, byRoles, 1)

	// 7. Update Password & Status
	err = repo.UpdatePassword(ctx, user.ID, "$2a$10$newsecret")
	require.NoError(t, err)

	err = repo.UpdateStatus(ctx, user.ID, false)
	require.NoError(t, err)

	afterStatus, err := repo.FindByID(ctx, user.ID)
	require.NoError(t, err)
	assert.False(t, afterStatus.IsActive)
	assert.Equal(t, "$2a$10$newsecret", afterStatus.PasswordHash)

	// 8. Update general info
	user.Email = "updated_admin@polyglot.net"
	user.IsActive = true
	err = repo.Update(ctx, user)
	require.NoError(t, err)

	afterUpdate, err := repo.FindByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "updated_admin@polyglot.net", afterUpdate.Email)

	// 9. Assign Devices & Access Check
	devIDs := []string{"dev-001", "dev-002"}
	err = repo.AssignDevices(ctx, user.ID, devIDs, nil)
	require.NoError(t, err)

	assigned, err := repo.GetAssignedDeviceIDs(ctx, user.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, devIDs, assigned)

	accessible, err := repo.IsDeviceAccessibleByUser(ctx, user.ID, "dev-001")
	require.NoError(t, err)
	assert.True(t, accessible)

	notAccessible, err := repo.IsDeviceAccessibleByUser(ctx, user.ID, "dev-999")
	require.NoError(t, err)
	assert.False(t, notAccessible)

	// 10. Delete
	err = repo.Delete(ctx, user.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, user.ID)
	assert.Error(t, err)
}
