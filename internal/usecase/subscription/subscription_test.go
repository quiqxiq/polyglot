package subscription_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	subUC "github.com/quixiq/polyglot/internal/usecase/subscription"
)

func TestSubscriptionUseCase_Lifecycle(t *testing.T) {
	subRepo := mocktest.NewFakeSubscriptionRepo()
	planRepo := mocktest.NewFakeServicePlanRepo()
	custRepo := mocktest.NewFakeCustomerRepo()
	invRepo := mocktest.NewFakeInvoiceRepo()
	manager := mocktest.NewFakeRouterAccountManager()
	audit := &mocktest.FakeAuditWriter{}

	ctx := context.Background()

	require.NoError(t, custRepo.Save(ctx, domainCustomer.Customer{
		ID:   "cust-1",
		Name: "Budi",
	}))

	planRepo.Seed(domainPlan.ServicePlan{
		ID:                    "plan-10m",
		Name:                  "PAKET-10M",
		ServiceType:           domainPlan.TypePPPoE,
		BandwidthDownloadKbps: 10000,
		BandwidthUploadKbps:   10000,
		Price:                 150000,
		IsActive:              true,
	})

	planRepo.Seed(domainPlan.ServicePlan{
		ID:                    "plan-20m",
		Name:                  "PAKET-20M",
		ServiceType:           domainPlan.TypePPPoE,
		BandwidthDownloadKbps: 20000,
		BandwidthUploadKbps:   20000,
		Price:                 250000,
		IsActive:              true,
	})

	manageUC := subUC.NewManageSubscriptionUseCase(subRepo, planRepo, custRepo, nil, manager, audit, invRepo)
	lifecycleUC := subUC.NewLifecycleUseCase(subRepo, planRepo, manager, audit)

	// 1. Create Subscription
	sub, err := manageUC.Create(ctx, subUC.CreateInput{
		CustomerID:     "cust-1",
		PlanID:         "plan-10m",
		ServiceType:    domainPlan.TypePPPoE,
		RemoteUsername: "budi_pppoe",
		RemotePassword: "secretpassword",
	})
	require.NoError(t, err)
	assert.Equal(t, domainSub.StatusPending, sub.Status)

	// 2. Activate
	sub, err = lifecycleUC.Activate(ctx, sub.ID, "dev-r1")
	require.NoError(t, err)
	assert.Equal(t, domainSub.StatusActive, sub.Status)

	// 3. ChangePlan
	sub, err = lifecycleUC.ChangePlan(ctx, sub.ID, "plan-20m")
	require.NoError(t, err)
	assert.Equal(t, "plan-20m", sub.PlanID)

	// 4. Suspend
	sub, err = lifecycleUC.Suspend(ctx, sub.ID, "Cuti liburan")
	require.NoError(t, err)
	assert.Equal(t, domainSub.StatusSuspended, sub.Status)

	// 5. Resume
	sub, err = lifecycleUC.Resume(ctx, sub.ID)
	require.NoError(t, err)
	assert.Equal(t, domainSub.StatusActive, sub.Status)

	// 6. Isolate
	sub, err = lifecycleUC.Isolate(ctx, sub.ID, "Tagihan tertunggak")
	require.NoError(t, err)
	assert.Equal(t, domainSub.StatusIsolated, sub.Status)

	// 7. Restore
	sub, err = lifecycleUC.Restore(ctx, sub.ID)
	require.NoError(t, err)
	assert.Equal(t, domainSub.StatusActive, sub.Status)

	// 8. Terminate
	sub, err = lifecycleUC.Terminate(ctx, sub.ID, "Pindah rumah")
	require.NoError(t, err)
	assert.Equal(t, domainSub.StatusTerminated, sub.Status)
}
