// Fase 2 worker lifecycle tests: opt-out auto_isolate per langganan,
// grace per langganan, status OVERDUE untuk tagihan parsial, dan retry
// restore pasca-bayar yang gagal (F2-11, F3-2, F2-13).
package billing_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port/mocktest"
)

func TestIsolateWorker_PerSubscriptionOptOut(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	isolator := mocktest.NewFakeRouterAccountManager()

	deviceID := "dev-opt"
	sub := domainSubscription.Subscription{
		ID: "sub-opt", TenantID: "tenant-default", CustomerID: "cust-opt",
		PlanID: "plan-1", DeviceID: &deviceID, ServiceType: "PPPOE",
		RemoteUsername: "OPTUSER", Status: domainSubscription.StatusActive,
		AutoIsolate: false, // opt-out per langganan
	}
	require.NoError(t, subs.Save(context.Background(), sub))
	require.NoError(t, invoices.Save(context.Background(), unpaidInvoice("inv-opt", sub.CustomerID, sub.ID, -30)))

	worker := newWorker(subs, invoices, mocktest.NewFakeCustomerRepo(), mocktest.NewFakeServicePlanRepo(),
		isolator, mocktest.NewFakeNotificationRepo(), defaultSettings(nil))
	res, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, res.Isolated)
	assert.Equal(t, 0, isolator.Count("Isolate"))

	got, _ := subs.FindByID(context.Background(), sub.ID)
	assert.Equal(t, domainSubscription.StatusActive, got.Status)

	// Tagihan tetap ditandai OVERDUE walau layanan tidak diisolir.
	inv, err := invoices.FindByID(context.Background(), "inv-opt")
	require.NoError(t, err)
	assert.Equal(t, domainBilling.StatusOverdue, inv.Status)
}

func TestIsolateWorker_PerSubscriptionGraceDays(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	isolator := mocktest.NewFakeRouterAccountManager()

	deviceID := "dev-grace2"
	sub := domainSubscription.Subscription{
		ID: "sub-grace2", TenantID: "tenant-default", CustomerID: "cust-grace2",
		PlanID: "plan-1", DeviceID: &deviceID, ServiceType: "PPPOE",
		RemoteUsername: "GRACEUSER", Status: domainSubscription.StatusActive,
		AutoIsolate: true, IsolationGraceDays: 14, // menang atas global 3
	}
	require.NoError(t, subs.Save(context.Background(), sub))
	require.NoError(t, invoices.Save(context.Background(), unpaidInvoice("inv-grace2", sub.CustomerID, sub.ID, -5)))

	worker := newWorker(subs, invoices, mocktest.NewFakeCustomerRepo(), mocktest.NewFakeServicePlanRepo(),
		isolator, mocktest.NewFakeNotificationRepo(), defaultSettings(nil))
	res, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, res.Isolated, "grace 14 hari belum boleh isolir tagihan -5 hari")

	// Grace per-langganan diubah menjadi 1 hari → isolir di siklus berikutnya.
	got, err := subs.FindByID(context.Background(), sub.ID)
	require.NoError(t, err)
	got.IsolationGraceDays = 1
	require.NoError(t, subs.Save(context.Background(), got))

	res2, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, res2.Isolated)
	assert.Equal(t, 1, isolator.Count("Isolate:GRACEUSER->isolir"))
}

func TestIsolateWorker_PartialInvoiceMarkedOverdueAndIsolated(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	isolator := mocktest.NewFakeRouterAccountManager()

	deviceID := "dev-part"
	sub := domainSubscription.Subscription{
		ID: "sub-part", TenantID: "tenant-default", CustomerID: "cust-part",
		PlanID: "plan-1", DeviceID: &deviceID, ServiceType: "PPPOE",
		RemoteUsername: "PARTUSER", Status: domainSubscription.StatusActive,
		AutoIsolate: true,
	}
	require.NoError(t, subs.Save(context.Background(), sub))

	inv := unpaidInvoice("inv-part", sub.CustomerID, sub.ID, -10)
	inv.PaidAmount = 50000
	inv.Status = domainBilling.StatusPartial
	require.NoError(t, invoices.Save(context.Background(), inv))

	worker := newWorker(subs, invoices, mocktest.NewFakeCustomerRepo(), mocktest.NewFakeServicePlanRepo(),
		isolator, mocktest.NewFakeNotificationRepo(), defaultSettings(nil))
	res, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, res.Isolated, "tagihan PARTIAL menunggak harus tetap diisolir")

	stored, err := invoices.FindByID(context.Background(), "inv-part")
	require.NoError(t, err)
	assert.Equal(t, domainBilling.StatusPartial, stored.Status, "status PARTIAL dipertahankan (bukan dikonversi OVERDUE)")
}

func TestIsolateWorker_RestoreRetryAfterFailedOnPaid(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	plans := mocktest.NewFakeServicePlanRepo()
	isolator := mocktest.NewFakeRouterAccountManager()
	plans.Seed(newPlan("plan-p", "HOME-20M"))

	deviceID := "dev-restore"
	sub := domainSubscription.Subscription{
		ID: "sub-restore", TenantID: "tenant-default", CustomerID: "cust-restore",
		PlanID: "plan-p", DeviceID: &deviceID, ServiceType: "PPPOE",
		RemoteUsername: "RESTOREUSER", RemotePassword: "secret99",
		Status:          domainSubscription.StatusIsolated,
		ProvisionStatus: domainSubscription.ProvisionFailed, // restore pasca-bayar gagal
		AutoIsolate:     true,
	}
	require.NoError(t, subs.Save(context.Background(), sub))

	worker := newWorker(subs, invoices, mocktest.NewFakeCustomerRepo(), plans,
		isolator, mocktest.NewFakeNotificationRepo(), defaultSettings(nil))
	res, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, res.Provisioned)
	assert.Equal(t, 1, res.Restored, "provision ulang langganan ISOLATED = restore yang tertunda")

	got, err := subs.FindByID(context.Background(), sub.ID)
	require.NoError(t, err)
	assert.Equal(t, domainSubscription.StatusActive, got.Status)
	assert.Equal(t, domainSubscription.ProvisionOK, got.ProvisionStatus)
}
