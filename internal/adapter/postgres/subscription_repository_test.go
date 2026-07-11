package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/testutil"
)

func TestSubscriptionRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&subscriptionModel{}))

	repo := NewSubscriptionRepository(db)

	s := subscription.Subscription{
		ID:            "sub-001",
		CustomerID:    "cust-001",
		PlanID:        "plan-001",
		ServiceType:   "pppoe",
		Status:        "active",
		DeviceID:      "dev-001",
		PPPoEUsername: "budinet",
	}

	created, err := repo.Create(ctx, s)
	require.NoError(t, err)
	assert.Equal(t, s.PPPoEUsername, created.PPPoEUsername)

	found, err := repo.FindByID(ctx, s.ID)
	require.NoError(t, err)
	assert.Equal(t, s.ServiceType, found.ServiceType)

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	byCustomer, err := repo.FindByCustomer(ctx, s.CustomerID)
	require.NoError(t, err)
	assert.Len(t, byCustomer, 1)

	byDevice, err := repo.FindByDevice(ctx, s.DeviceID)
	require.NoError(t, err)
	assert.Len(t, byDevice, 1)

	updated := s
	updated.Status = "suspended"
	_, err = repo.Update(ctx, updated)
	require.NoError(t, err)

	found, err = repo.FindByID(ctx, s.ID)
	require.NoError(t, err)
	assert.Equal(t, "suspended", found.Status)

	require.NoError(t, repo.Delete(ctx, s.ID))
	_, err = repo.FindByID(ctx, s.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, subscription.ErrNotFound)
}

func TestSubscriptionRepository_FindByID_NotFound(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&subscriptionModel{}))

	repo := NewSubscriptionRepository(db)

	_, err = repo.FindByID(ctx, "missing")
	require.Error(t, err)
	assert.ErrorIs(t, err, subscription.ErrNotFound)
}
