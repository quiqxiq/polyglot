package billing_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/billing"
)

func defaultSettings(extra map[string]string) *mocktest.FakeSettingReader {
	base := map[string]string{
		"isp.auto_isolate":         "true",
		"isp.isolate_grace_days":   "3",
		"isp.suspend_after_days":   "90",
		"isp.pppoe_isolir_profile": "isolir",
		"isp.isolir_address_list":  "ISOLIR_USERS",
		"isp.payment_redirect_url": "bayar.example.com:8080",
	}
	for k, v := range extra {
		base[k] = v
	}
	return mocktest.NewFakeSettingReader(base)
}

func newWorker(subs *mocktest.FakeSubscriptionRepo, inv *mocktest.FakeInvoiceRepo,
	cust *mocktest.FakeCustomerRepo, plans *mocktest.FakeServicePlanRepo,
	mgr *mocktest.FakeRouterAccountManager, notif *mocktest.FakeNotificationRepo,
	settings *mocktest.FakeSettingReader) *uc.IsolateWorker {
	return uc.NewIsolateWorker(subs, inv, cust, plans, mgr, notif, settings)
}

func TestIsolateWorker_SettingsDriven_Isolate(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	customers := mocktest.NewFakeCustomerRepo()
	plans := mocktest.NewFakeServicePlanRepo()
	isolator := mocktest.NewFakeRouterAccountManager()
	notif := mocktest.NewFakeNotificationRepo()

	deviceID := "dev-router-1"
	sub := seedActiveSub(t, subs, "sub-late", nil)
	sub.DeviceID = &deviceID
	sub.RemoteUsername = "BUDI"
	sub.ServiceType = "PPPOE"
	require.NoError(t, subs.Save(context.Background(), sub))
	// Jatuh tempo 10 hari lalu — lewat grace 3 hari dari settings.
	require.NoError(t, invoices.Save(context.Background(), unpaidInvoice("inv-late", sub.CustomerID, sub.ID, -10)))
	require.NoError(t, customers.Save(context.Background(), customerWithPortal(sub.CustomerID, "87654321")))

	worker := newWorker(subs, invoices, customers, plans, isolator, notif, defaultSettings(nil))
	res, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, res.Isolated)

	updated, _ := subs.FindByID(context.Background(), sub.ID)
	assert.Equal(t, domainSubscription.StatusIsolated, updated.Status)
	assert.Equal(t, 1, isolator.Count("Isolate:BUDI->isolir"))
	require.Len(t, notif.Queued(), 1)
}

func TestIsolateWorker_AutoIsolateOff(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	plans := mocktest.NewFakeServicePlanRepo()
	isolator := mocktest.NewFakeRouterAccountManager()

	deviceID := "dev-r"
	sub := seedActiveSub(t, subs, "sub-x", nil)
	sub.DeviceID = &deviceID
	sub.RemoteUsername = "XUSER"
	sub.ServiceType = "PPPOE"
	require.NoError(t, subs.Save(context.Background(), sub))
	require.NoError(t, invoices.Save(context.Background(), unpaidInvoice("inv-x", sub.CustomerID, sub.ID, -30)))

	settings := defaultSettings(map[string]string{"isp.auto_isolate": "false"})
	worker := newWorker(subs, invoices, mocktest.NewFakeCustomerRepo(), plans, isolator,
		mocktest.NewFakeNotificationRepo(), settings)
	res, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, res.Isolated)
	assert.Equal(t, 0, isolator.Count("Isolate"))
}

func TestIsolateWorker_GraceFromSettings(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	plans := mocktest.NewFakeServicePlanRepo()
	isolator := mocktest.NewFakeRouterAccountManager()

	deviceID := "dev-g"
	sub := seedActiveSub(t, subs, "sub-grace", nil)
	sub.DeviceID = &deviceID
	sub.RemoteUsername = "GUSER"
	sub.ServiceType = "PPPOE"
	require.NoError(t, subs.Save(context.Background(), sub))

	// Settings grace=14 → invoice -5 hari belum boleh diisolir.
	require.NoError(t, invoices.Save(context.Background(), unpaidInvoice("inv-g", sub.CustomerID, sub.ID, -5)))
	settings := defaultSettings(map[string]string{"isp.isolate_grace_days": "14"})
	worker := newWorker(subs, invoices, mocktest.NewFakeCustomerRepo(), plans, isolator,
		mocktest.NewFakeNotificationRepo(), settings)
	res, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, res.Isolated)
}

func TestIsolateWorker_ProvisionRetry(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	customers := mocktest.NewFakeCustomerRepo()
	plans := mocktest.NewFakeServicePlanRepo()
	isolator := mocktest.NewFakeRouterAccountManager()

	plans.Seed(newPlan("plan-p", "HOME-20M"))
	deviceID := "dev-prov"
	sub := seedActiveSub(t, subs, "sub-prov", nil)
	sub.PlanID = "plan-p"
	sub.DeviceID = &deviceID
	sub.RemoteUsername = "PROVUSER"
	sub.RemotePassword = "secret99"
	sub.ServiceType = "PPPOE"
	sub.ProvisionStatus = domainSubscription.ProvisionNone
	require.NoError(t, subs.Save(context.Background(), sub))

	// Tanpa tagihan tertunggak — hanya retry provisi.
	worker := newWorker(subs, invoices, customers, plans, isolator,
		mocktest.NewFakeNotificationRepo(), defaultSettings(nil))
	res, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, res.Provisioned)

	got, _ := subs.FindByID(context.Background(), sub.ID)
	assert.Equal(t, domainSubscription.ProvisionOK, got.ProvisionStatus)
	assert.Equal(t, "HOME-20M", got.RouterProfile)
	assert.Equal(t, 1, isolator.Count("Provision:PROVUSER@HOME-20M"))

	// Siklus kedua tidak provision ulang.
	res2, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, res2.Provisioned)
}

func TestIsolateWorker_AutoSuspendAfterDays(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	plans := mocktest.NewFakeServicePlanRepo()
	isolator := mocktest.NewFakeRouterAccountManager()

	deviceID := "dev-s"
	sub := seedActiveSub(t, subs, "sub-old", nil)
	sub.DeviceID = &deviceID
	sub.RemoteUsername = "OLDUSER"
	sub.ServiceType = "PPPOE"
	sub.Status = domainSubscription.StatusIsolated
	require.NoError(t, subs.Save(context.Background(), sub))
	// Tunggakan 200 hari; suspend_after_days=90 + grace 3 → memenuhi.
	require.NoError(t, invoices.Save(context.Background(), unpaidInvoice("inv-old", sub.CustomerID, sub.ID, -200)))

	worker := newWorker(subs, invoices, mocktest.NewFakeCustomerRepo(), plans, isolator,
		mocktest.NewFakeNotificationRepo(), defaultSettings(nil))
	res, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, res.Suspended)
	assert.Equal(t, 1, isolator.Count("Suspend:OLDUSER"))
	updated, _ := subs.FindByID(context.Background(), sub.ID)
	assert.Equal(t, domainSubscription.StatusSuspended, updated.Status)

	// Suspend off via settings → manual saja.
	subs2 := mocktest.NewFakeSubscriptionRepo()
	invoices2 := mocktest.NewFakeInvoiceRepo()
	sub2 := seedActiveSub(t, subs2, "sub-old2", nil)
	sub2.DeviceID = &deviceID
	sub2.RemoteUsername = "OLDUSER2"
	sub2.Status = domainSubscription.StatusIsolated
	require.NoError(t, subs2.Save(context.Background(), sub2))
	require.NoError(t, invoices2.Save(context.Background(), unpaidInvoice("inv-o2", sub2.CustomerID, sub2.ID, -200)))

	iso2 := mocktest.NewFakeRouterAccountManager()
	settingsOff := defaultSettings(map[string]string{"isp.suspend_after_days": "0"})
	_, err = newWorker(subs2, invoices2, mocktest.NewFakeCustomerRepo(), plans, iso2,
		mocktest.NewFakeNotificationRepo(), settingsOff).Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, iso2.Count("Suspend"))
	updated2, _ := subs2.FindByID(context.Background(), sub2.ID)
	assert.Equal(t, domainSubscription.StatusIsolated, updated2.Status)
}

func TestIsolateWorker_RouterFailureLeavesDBUntouched(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	plans := mocktest.NewFakeServicePlanRepo()
	isolator := &mocktest.FakeRouterAccountManager{Fail: map[string]error{
		"Isolate:BUDI-GAGAL->isolir": assert_AnError(),
	}}

	deviceID := "dev-f"
	sub := seedActiveSub(t, subs, "sub-fail", nil)
	sub.DeviceID = &deviceID
	sub.RemoteUsername = "BUDI-GAGAL"
	sub.ServiceType = "PPPOE"
	require.NoError(t, subs.Save(context.Background(), sub))
	require.NoError(t, invoices.Save(context.Background(), unpaidInvoice("inv-fail", sub.CustomerID, sub.ID, -30)))

	worker := newWorker(subs, invoices, mocktest.NewFakeCustomerRepo(), plans, isolator,
		mocktest.NewFakeNotificationRepo(), defaultSettings(nil))
	res, err := worker.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, res.RouterFailures)

	updated, _ := subs.FindByID(context.Background(), sub.ID)
	assert.Equal(t, domainSubscription.StatusActive, updated.Status)
}
