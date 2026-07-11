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
		Name:                "Basic 10Mbps",
		ServiceType:         "pppoe",
		Description:         "Basic plan",
		Price:               150000,
		BillingPeriodMonths: 1,
		BandwidthDownKbps: 10000,
		BandwidthUpKbps:   5000,
		IsActive:            true,
	}

	created, err := repo.Create(ctx, p)
	require.NoError(t, err)
	assert.Equal(t, p.Name, created.Name)

	found, err := repo.FindByID(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, p.Price, found.Price)

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	active, err := repo.FindActive(ctx)
	require.NoError(t, err)
	assert.Len(t, active, 1)

	updated := p
	updated.Name = "Basic 10Mbps Updated"
	_, err = repo.Update(ctx, updated)
	require.NoError(t, err)

	found, err = repo.FindByID(ctx, p.ID)
	require.NoError(t, err)
	assert.Equal(t, "Basic 10Mbps Updated", found.Name)

	require.NoError(t, repo.Delete(ctx, p.ID))
	_, err = repo.FindByID(ctx, p.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, plan.ErrNotFound)
}

func TestPlanRepository_FindByID_NotFound(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&planModel{}))

	repo := NewPlanRepository(db)

	_, err = repo.FindByID(ctx, "missing")
	require.Error(t, err)
	assert.ErrorIs(t, err, plan.ErrNotFound)
}
