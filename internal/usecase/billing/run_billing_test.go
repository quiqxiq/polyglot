package billing_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/billing"
)

func seedActiveSub(t *testing.T, subs *mocktest.FakeSubscriptionRepo, id string, customPrice *float64) domainSubscription.Subscription {
	t.Helper()
	sub := domainSubscription.Subscription{
		ID:          id,
		TenantID:    "tenant-default",
		CustomerID:  "cust-" + id,
		PlanID:      "plan-1",
		Status:      domainSubscription.StatusActive,
		CustomPrice: customPrice,
	}
	require.NoError(t, subs.Save(context.Background(), sub))
	return sub
}

func TestRunBilling_PeriodValidation(t *testing.T) {
	usecase := uc.NewRunBillingUseCase(mocktest.NewFakeSubscriptionRepo(), mocktest.NewFakeServicePlanRepo(), mocktest.NewFakeInvoiceRepo())
	for _, bad := range []string{"2026-13", "26-08", "abc", "", "2026-8"} {
		_, err := usecase.Run(context.Background(), "tenant-default", bad)
		assert.ErrorIs(t, err, uc.ErrValidation, "period %q harus ditolak", bad)
	}
}

func TestRunBilling_HappyPath_And_Idempotent(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	plans := mocktest.NewFakeServicePlanRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	plans.Seed(domainPlan.ServicePlan{
		ID: "plan-1", Name: "100-RB-100", ServiceType: "PPPOE",
		Price: 100000, TaxPercent: 10, IsActive: true,
	})
	seedActiveSub(t, subs, "sub-a", nil)
	seedActiveSub(t, subs, "sub-b", nil)

	usecase := uc.NewRunBillingUseCase(subs, plans, invoices)
	res, err := usecase.Run(context.Background(), "tenant-default", "2026-08")
	require.NoError(t, err)
	assert.Equal(t, 2, res.Created)
	assert.Equal(t, 0, res.Skipped)

	// Idempoten: run kedua tidak membuat apa pun.
	res2, err := usecase.Run(context.Background(), "tenant-default", "2026-08")
	require.NoError(t, err)
	assert.Equal(t, 0, res2.Created)
	assert.Equal(t, 2, res2.Skipped)

	all, _ := invoices.FindAll(context.Background())
	require.Len(t, all, 2)
	for _, inv := range all {
		assert.Equal(t, "2026-08", inv.Period)
		assert.InDelta(t, 100000, inv.Subtotal, 0.01)
		assert.InDelta(t, 110000, inv.Total, 0.01) // + pajak 10%
		assert.NotEmpty(t, inv.ManualPaymentCode)
		assert.Contains(t, inv.QRPayload, "polyglot://invoice/")
		items := invoices.ItemsOf(inv.ID)
		require.Len(t, items, 1)
		assert.Equal(t, domainBilling.ItemTypeSubscriptionFee, items[0].ItemType)
	}
}

func TestRunBilling_CustomPriceOverride(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	plans := mocktest.NewFakeServicePlanRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	plans.Seed(domainPlan.ServicePlan{ID: "plan-1", Name: "P", Price: 100000, IsActive: true})
	custom := 75000.0
	seedActiveSub(t, subs, "sub-promo", &custom)

	usecase := uc.NewRunBillingUseCase(subs, plans, invoices)
	_, err := usecase.Run(context.Background(), "tenant-default", "2026-08")
	require.NoError(t, err)

	all, _ := invoices.FindAll(context.Background())
	require.Len(t, all, 1)
	assert.InDelta(t, 75000, all[0].Subtotal, 0.01)
}

func TestRunBilling_MissingPlanSkipped(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	seedActiveSub(t, subs, "sub-orphan", nil) // tanpa plan terdaftar

	usecase := uc.NewRunBillingUseCase(subs, mocktest.NewFakeServicePlanRepo(), invoices)
	res, err := usecase.Run(context.Background(), "tenant-default", "2026-08")
	require.NoError(t, err)
	assert.Equal(t, 0, res.Created)
	assert.Equal(t, 1, res.Skipped)
}
