package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/plan"
	"github.com/quixiq/polyglot/internal/testutil"
)

func TestPlanRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&planModel{}))

	repo := NewPlanRepository(db)

	p := plan.Plan{
		ID:                  "plan-001",
		Name:                "Home 10M",
		ServiceType:         "pppoe",
		Description:         "10 Mbps home plan",
		Price:               150000,
		BillingPeriodMonths: 1,
		BandwidthDownKbps:   10240,
		BandwidthUpKbps:     2048,
		IsActive:            true,
	}

	created, err := repo.Create(ctx, p)
	require.NoError(t, err)
	assert.Equal(t, p.Name, created.Name)
	assert.Equal(t, float64(150000), created.Price)

	found, err := repo.FindByID(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, p.Name, found.Name)
	assert.Equal(t, "pppoe", found.ServiceType)
	assert.Equal(t, 10240, found.BandwidthDownKbps)

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	updated := p
	updated.Name = "Home 20M"
	updated.Price = 200000
	updated.BandwidthDownKbps = 20480
	_, err = repo.Update(ctx, updated)
	require.NoError(t, err)

	found, err = repo.FindByID(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, "Home 20M", found.Name)
	assert.Equal(t, float64(200000), found.Price)

	require.NoError(t, repo.Delete(ctx, p.ID))
	_, err = repo.FindByID(ctx, p.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, plan.ErrNotFound)
}

func TestPlanRepository_FindActive(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&planModel{}))

	repo := NewPlanRepository(db)

	active := plan.Plan{
		ID:                  "plan-active",
		Name:                "Active Plan",
		ServiceType:         "pppoe",
		Price:               100000,
		BillingPeriodMonths: 1,
		BandwidthDownKbps:   5120,
		BandwidthUpKbps:     1024,
		IsActive:            true,
	}
	inactive := plan.Plan{
		ID:                  "plan-inactive",
		Name:                "Inactive Plan",
		ServiceType:         "hotspot",
		Price:               50000,
		BillingPeriodMonths: 1,
		BandwidthDownKbps:   2048,
		BandwidthUpKbps:     512,
		IsActive:            false,
	}

	_, err = repo.Create(ctx, active)
	require.NoError(t, err)
	_, err = repo.Create(ctx, inactive)
	require.NoError(t, err)

	activePlans, err := repo.FindActive(ctx)
	require.NoError(t, err)
	assert.Len(t, activePlans, 1)
	assert.Equal(t, "Active Plan", activePlans[0].Name)

	allPlans, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, allPlans, 2)
}

func TestPlanRepository_FindByID_NotFound(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&planModel{}))

	repo := NewPlanRepository(db)

	_, err = repo.FindByID(ctx, "non-existent")
	require.Error(t, err)
	assert.ErrorIs(t, err, plan.ErrNotFound)
}
