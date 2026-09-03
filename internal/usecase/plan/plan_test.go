package plan_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	planUC "github.com/quixiq/polyglot/internal/usecase/plan"
)

func TestPlanUseCase_CRUD(t *testing.T) {
	planRepo := mocktest.NewFakeServicePlanRepo()
	subRepo := mocktest.NewFakeSubscriptionRepo()
	manager := mocktest.NewFakeRouterAccountManager()

	uc := planUC.NewManagePlanUseCase(planRepo, subRepo, manager)
	ctx := context.Background()

	// 1. Create
	created, err := uc.Create(ctx, domainPlan.ServicePlan{
		Name:                  "PAKET-10M",
		ServiceType:           domainPlan.TypePPPoE,
		BandwidthDownloadKbps: 10000,
		BandwidthUploadKbps:   10000,
		Price:                 150000,
	}, "")
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "PAKET-10M", created.Name)
	assert.True(t, created.IsActive)

	// 2. Get
	fetched, err := uc.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)

	// 3. List
	plans, err := uc.List(ctx, true)
	require.NoError(t, err)
	assert.Len(t, plans, 1)

	// 4. Update
	created.Name = "PAKET-10M-HEMAT"
	created.Price = 135000.0
	updated, err := uc.Update(ctx, created, "")
	require.NoError(t, err)
	assert.Equal(t, "PAKET-10M-HEMAT", updated.Name)
	assert.Equal(t, 135000.0, updated.Price)

	// 5. Delete
	err = uc.Delete(ctx, created.ID, "")
	require.NoError(t, err)

	_, err = uc.Get(ctx, created.ID)
	assert.Error(t, err)
}
