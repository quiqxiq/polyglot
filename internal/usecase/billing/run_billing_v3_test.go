package billing_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/billing"
)

func seedSubWithAnchor(t *testing.T, subs *mocktest.FakeSubscriptionRepo, id string, anchor int) {
	t.Helper()
	require.NoError(t, subs.Save(context.Background(), domainSubscription.Subscription{
		ID: id, TenantID: "tenant-default", CustomerID: "cust-" + id, PlanID: "plan-1",
		Status: domainSubscription.StatusActive, AutoIsolate: true,
		BillingDay: anchor, BillingAnchorDay: anchor,
	}))
}

func TestRunBilling_PrabayarIssuesNextPeriodWithLeadDays(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	plans := mocktest.NewFakeServicePlanRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	settings := mocktest.NewFakeSettingReader(map[string]string{
		"isp.billing_mode":             "PRABAYAR",
		"isp.prepay_invoice_lead_days": "7",
		"isp.closing_date":             "1",
	})
	plans.Seed(domainPlan.ServicePlan{ID: "plan-1", Name: "P", Price: 100000, IsActive: true})
	seedSubWithAnchor(t, subs, "sub-pre", 1)

	// 25 September: periode Oktober (mulai 1 Okt) sudah dalam jendela 7 hari.
	clock := time.Date(2026, 9, 25, 6, 0, 0, 0, time.UTC)
	usecase := uc.NewRunBillingUseCase(subs, plans, invoices).
		WithSettings(settings).
		WithClock(func() time.Time { return clock })

	res, err := usecase.Run(context.Background(), "tenant-default", "2026-10")
	require.NoError(t, err)
	assert.Equal(t, 1, res.Created)

	all, _ := invoices.FindAll(context.Background())
	require.Len(t, all, 1)
	inv := all[0]
	assert.Equal(t, "2026-10", inv.Period)
	// Prabayar: jatuh tempo = hari mulai siklus (1 Okt 23:59).
	assert.Equal(t, 1, inv.DueDate.Day())
	assert.Equal(t, time.October, inv.DueDate.Month())
	assert.InDelta(t, 100000, inv.Total, 0.01, "tanpa PPN default")
}

func TestRunBilling_PascabayarSkipsNextPeriod(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	plans := mocktest.NewFakeServicePlanRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	settings := mocktest.NewFakeSettingReader(map[string]string{
		"isp.billing_mode": "PASCABAYAR",
		"isp.closing_date": "1",
	})
	plans.Seed(domainPlan.ServicePlan{ID: "plan-1", Name: "P", Price: 100000, IsActive: true})
	seedSubWithAnchor(t, subs, "sub-post", 1)

	clock := time.Date(2026, 9, 25, 6, 0, 0, 0, time.UTC)
	usecase := uc.NewRunBillingUseCase(subs, plans, invoices).
		WithSettings(settings).
		WithClock(func() time.Time { return clock })

	res, err := usecase.Run(context.Background(), "tenant-default", "2026-10")
	require.NoError(t, err)
	assert.Equal(t, 0, res.Created)
	assert.Equal(t, 1, res.Skipped)
}

func TestRunBilling_TaxInclusiveMode(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	plans := mocktest.NewFakeServicePlanRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	plans.Seed(domainPlan.ServicePlan{
		ID: "plan-1", Name: "P", Price: 111000, IsActive: true,
		TaxPercent: 11, TaxMode: "INCLUSIVE",
	})
	seedSubWithAnchor(t, subs, "sub-inc", 1)

	usecase := uc.NewRunBillingUseCase(subs, plans, invoices)
	_, err := usecase.Run(context.Background(), "tenant-default", "2026-08")
	require.NoError(t, err)

	all, _ := invoices.FindAll(context.Background())
	require.Len(t, all, 1)
	inv := all[0]
	assert.InDelta(t, 100000, inv.Subtotal, 0.01)
	assert.InDelta(t, 11000, inv.TaxAmount, 0.01)
	assert.InDelta(t, 111000, inv.Total, 0.01)
	assert.InDelta(t, 11, inv.TaxPercent, 0.01)
}

func TestRunBilling_TaxDefaultFallbackFromSettings(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	plans := mocktest.NewFakeServicePlanRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	settings := mocktest.NewFakeSettingReader(map[string]string{
		"isp.tax_percent": "11",
	})
	plans.Seed(domainPlan.ServicePlan{ID: "plan-1", Name: "P", Price: 100000, IsActive: true})
	seedSubWithAnchor(t, subs, "sub-tax", 1)

	usecase := uc.NewRunBillingUseCase(subs, plans, invoices).WithSettings(settings)
	_, err := usecase.Run(context.Background(), "tenant-default", "2026-08")
	require.NoError(t, err)

	all, _ := invoices.FindAll(context.Background())
	require.Len(t, all, 1)
	assert.InDelta(t, 11000, all[0].TaxAmount, 0.01)
	assert.InDelta(t, 111000, all[0].Total, 0.01)
}

func TestRunBilling_ClosingDateAsDefaultAnchor(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	plans := mocktest.NewFakeServicePlanRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	settings := mocktest.NewFakeSettingReader(map[string]string{
		"isp.closing_date": "15",
	})
	plans.Seed(domainPlan.ServicePlan{ID: "plan-1", Name: "P", Price: 100000, IsActive: true})
	// Billing day 0 → memakai closing_date 15.
	seedSubWithAnchor(t, subs, "sub-anchor", 0)

	clock := time.Date(2026, 9, 20, 6, 0, 0, 0, time.UTC)
	usecase := uc.NewRunBillingUseCase(subs, plans, invoices).
		WithSettings(settings).
		WithClock(func() time.Time { return clock })
	_, err := usecase.Run(context.Background(), "tenant-default", "2026-09")
	require.NoError(t, err)

	all, _ := invoices.FindAll(context.Background())
	require.Len(t, all, 1)
	require.NotNil(t, all[0].PeriodStart)
	assert.Equal(t, 15, all[0].PeriodStart.Day())
}
