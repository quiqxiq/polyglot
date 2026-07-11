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
		Status:        "pending_install",
		DeviceID:      "dev-001",
		PPPoEUsername: "pppoe-user1",
	}

	created, err := repo.Create(ctx, s)
	require.NoError(t, err)
	assert.Equal(t, s.PPPoEUsername, created.PPPoEUsername)
	assert.Equal(t, "pending_install", created.Status)

	found, err := repo.FindByID(ctx, s.ID)
	require.NoError(t, err)
	assert.Equal(t, s.CustomerID, found.CustomerID)
	assert.Equal(t, "pppoe", found.ServiceType)

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	updated := s
	updated.Status = "active"
	_, err = repo.Update(ctx, updated)
	require.NoError(t, err)

	found, err = repo.FindByID(ctx, s.ID)
	require.NoError(t, err)
	assert.Equal(t, "active", found.Status)

	require.NoError(t, repo.Delete(ctx, s.ID))
	_, err = repo.FindByID(ctx, s.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, subscription.ErrNotFound)
}

func TestSubscriptionRepository_FindByCustomer(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&subscriptionModel{}))

	repo := NewSubscriptionRepository(db)

	s1 := subscription.Subscription{
		ID:          "sub-101",
		CustomerID:  "cust-A",
		PlanID:      "plan-001",
		ServiceType: "pppoe",
		Status:      "active",
		DeviceID:    "dev-001",
	}
	s2 := subscription.Subscription{
		ID:          "sub-102",
		CustomerID:  "cust-A",
		PlanID:      "plan-002",
		ServiceType: "hotspot",
		Status:      "active",
		DeviceID:    "dev-002",
	}
	s3 := subscription.Subscription{
		ID:          "sub-103",
		CustomerID:  "cust-B",
		PlanID:      "plan-001",
		ServiceType: "pppoe",
		Status:      "active",
		DeviceID:    "dev-001",
	}

	_, err = repo.Create(ctx, s1)
	require.NoError(t, err)
	_, err = repo.Create(ctx, s2)
	require.NoError(t, err)
	_, err = repo.Create(ctx, s3)
	require.NoError(t, err)

	custA, err := repo.FindByCustomer(ctx, "cust-A")
	require.NoError(t, err)
	assert.Len(t, custA, 2)

	custB, err := repo.FindByCustomer(ctx, "cust-B")
	require.NoError(t, err)
	assert.Len(t, custB, 1)
}

func TestSubscriptionRepository_FindByDevice(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&subscriptionModel{}))

	repo := NewSubscriptionRepository(db)

	s1 := subscription.Subscription{
		ID:          "sub-201",
		CustomerID:  "cust-A",
		PlanID:      "plan-001",
		ServiceType: "pppoe",
		Status:      "active",
		DeviceID:    "dev-X",
	}
	s2 := subscription.Subscription{
		ID:          "sub-202",
		CustomerID:  "cust-B",
		PlanID:      "plan-001",
		ServiceType: "pppoe",
		Status:      "active",
		DeviceID:    "dev-Y",
	}

	_, err = repo.Create(ctx, s1)
	require.NoError(t, err)
	_, err = repo.Create(ctx, s2)
	require.NoError(t, err)

	devX, err := repo.FindByDevice(ctx, "dev-X")
	require.NoError(t, err)
	assert.Len(t, devX, 1)
	assert.Equal(t, "sub-201", devX[0].ID)
}

func TestSubscriptionRepository_FindByID_NotFound(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&subscriptionModel{}))

	repo := NewSubscriptionRepository(db)

	_, err = repo.FindByID(ctx, "non-existent")
	require.Error(t, err)
	assert.ErrorIs(t, err, subscription.ErrNotFound)
}
