package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/testutil"
)

func TestCustomerRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&customerModel{}))

	repo := NewCustomerRepository(db)

	c := customer.Customer{
		ID:           "cust-001",
		FullName:     "Budi Santoso",
		Phone:        "08123456789",
		Address:      "Jl. Mawar 1",
		CustomerType: "residential",
		Status:       "active",
	}

	created, err := repo.Create(ctx, c)
	require.NoError(t, err)
	assert.Equal(t, c.FullName, created.FullName)

	found, err := repo.FindByID(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, c.FullName, found.FullName)

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	updated := c
	updated.FullName = "Budi Santoso Updated"
	_, err = repo.Update(ctx, updated)
	require.NoError(t, err)

	found, err = repo.FindByID(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, "Budi Santoso Updated", found.FullName)

	require.NoError(t, repo.Delete(ctx, c.ID))
	_, err = repo.FindByID(ctx, c.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, customer.ErrNotFound)
}

func TestCustomerRepository_FindByID_NotFound(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&customerModel{}))

	repo := NewCustomerRepository(db)

	_, err = repo.FindByID(ctx, "missing")
	require.Error(t, err)
	assert.ErrorIs(t, err, customer.ErrNotFound)
}
