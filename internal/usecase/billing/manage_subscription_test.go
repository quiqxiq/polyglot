package billing_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/billing"
)

// ─── ManageSubscriptionUseCase fixture ──────────────────────────────────

type manageFixture struct {
	subs      *mocktest.FakeSubscriptionRepo
	plans     *mocktest.FakeServicePlanRepo
	customers *mocktest.FakeCustomerRepo
	manager   *mocktest.FakeRouterAccountManager
	audit     *mocktest.FakeAuditWriter
	invoices  *mocktest.FakeInvoiceRepo
	usecase   *uc.ManageSubscriptionUseCase
}

func newManageFixture(t *testing.T) *manageFixture {
	t.Helper()
	fx := &manageFixture{
		subs:      mocktest.NewFakeSubscriptionRepo(),
		plans:     mocktest.NewFakeServicePlanRepo(),
		customers: mocktest.NewFakeCustomerRepo(),
		manager:   mocktest.NewFakeRouterAccountManager(),
		audit:     &mocktest.FakeAuditWriter{},
		invoices:  mocktest.NewFakeInvoiceRepo(),
	}
	fx.plans.Seed(newPlan("plan-a", "PLAN-A"))
	require.NoError(t, fx.customers.Save(context.Background(), customerWithPortal("cust-1", "PORTAL1")))
	fx.usecase = uc.NewManageSubscriptionUseCase(
		fx.subs, fx.plans, fx.customers, fx.manager, fx.audit, fx.invoices,
	)
	return fx
}

func baseCreateInput() uc.CreateInput {
	return uc.CreateInput{
		CustomerID:   "cust-1",
		PlanID:       "plan-a",
		ServiceType:  "PPPOE",
		BillingCycle: domainSubscription.CycleMonthly,
		BillingDay:   5,
	}
}

// ─── Create ─────────────────────────────────────────────────────────────

func TestCreateSubscription_AutoGeneratesCredentials(t *testing.T) {
	fx := newManageFixture(t)
	ctx := context.Background()

	sub, err := fx.usecase.Create(ctx, baseCreateInput())
	require.NoError(t, err)

	assert.NotEmpty(t, sub.ID)
	assert.True(t, strings.HasPrefix(sub.RemoteUsername, "pc"), "username starts with initials 'pc', got %s", sub.RemoteUsername)
	assert.NotEmpty(t, sub.RemotePassword)
	assert.True(t, strings.HasSuffix(sub.RemotePassword, "pg"), "password = XXXXXXpg, got %q", sub.RemotePassword)
	assert.NotEqual(t, sub.RemoteUsername, sub.RemotePassword)
	assert.Nil(t, sub.DeviceID)
	assert.Equal(t, domainSubscription.ProvisionNone, sub.ProvisionStatus)
	assert.Equal(t, domainSubscription.StatusPending, sub.Status)
	assert.Equal(t, "PLAN-A", sub.RouterProfile)
	assert.Equal(t, 0, fx.manager.Count("Provision"))
	assert.Equal(t, 1, fx.audit.Count("CREATE_SUBSCRIPTION"))
}

func TestCreateSubscription_WithDeviceProvisionsRouterFirst(t *testing.T) {
	fx := newManageFixture(t)
	ctx := context.Background()
	device := "dev-1"

	in := baseCreateInput()
	in.DeviceID = &device
	sub, err := fx.usecase.Create(ctx, in)
	require.NoError(t, err)

	require.NotNil(t, sub.DeviceID)
	assert.Equal(t, "dev-1", *sub.DeviceID)
	assert.Equal(t, domainSubscription.ProvisionOK, sub.ProvisionStatus)
	assert.Equal(t, domainSubscription.StatusActive, sub.Status)
	assert.Equal(t, "PLAN-A", sub.RouterProfile)
	assert.Equal(t, 1, fx.manager.Count("ProvisionPPPoE:"+sub.RemoteUsername+"@PLAN-A"))
	assert.Equal(t, 1, fx.audit.Count("CREATE_SUBSCRIPTION"))
}

func TestCreateSubscription_ProvisionFail_PendingNotError(t *testing.T) {
	fx := newManageFixture(t)
	ctx := context.Background()
	device := "dev-1"

	in := baseCreateInput()
	in.DeviceID = &device
	fx.manager.Fail = map[string]error{"ProvisionPPPoE:": assert_AnError()}

	sub, err := fx.usecase.Create(ctx, in)
	require.NoError(t, err, "gagal provisi tidak boleh menggagalkan create")
	assert.Equal(t, domainSubscription.ProvisionPending, sub.ProvisionStatus)
	assert.Equal(t, domainSubscription.StatusPending, sub.Status)

	stored, ferr := fx.subs.FindByID(ctx, sub.ID)
	require.NoError(t, ferr)
	assert.Equal(t, domainSubscription.ProvisionPending, stored.ProvisionStatus)
}

func TestCreateSubscription_ValidatesPlanAndCustomer(t *testing.T) {
	fx := newManageFixture(t)
	ctx := context.Background()

	// Customer tak ada.
	in := baseCreateInput()
	in.CustomerID = "missing"
	_, err := fx.usecase.Create(ctx, in)
	require.Error(t, err)
	assert.ErrorContains(t, err, "customer")

	// Plan tak ada.
	in = baseCreateInput()
	in.PlanID = "missing"
	_, err = fx.usecase.Create(ctx, in)
	require.Error(t, err)
	assert.ErrorContains(t, err, "plan")

	// Plan non-aktif.
	p := newPlan("plan-off", "PLAN-OFF")
	p.IsActive = false
	fx.plans.Seed(p)
	in = baseCreateInput()
	in.PlanID = "plan-off"
	_, err = fx.usecase.Create(ctx, in)
	require.Error(t, err)

	// ServiceType mismatch dengan plan (case-insensitive match wajib).
	in = baseCreateInput()
	in.ServiceType = "HOTSPOT"
	_, err = fx.usecase.Create(ctx, in)
	require.Error(t, err)
	assert.ErrorContains(t, err, "service")
}

func TestCreateSubscription_ServiceTypeFallbackFromPlan(t *testing.T) {
	fx := newManageFixture(t)
	ctx := context.Background()

	in := baseCreateInput()
	in.ServiceType = ""
	sub, err := fx.usecase.Create(ctx, in)
	require.NoError(t, err)
	assert.Equal(t, "PPPOE", sub.ServiceType)
}

// ─── Update ─────────────────────────────────────────────────────────────

func TestUpdateSubscription_AppliesOnlyProvidedFields(t *testing.T) {
	fx := newManageFixture(t)
	ctx := context.Background()
	device := "dev-1"

	in := baseCreateInput()
	in.DeviceID = &device
	in.Notes = "awal"
	price := 150000.0
	cycle := domainSubscription.CycleMonthly
	day := 10
	newPass := "rahasia99pg"
	sub, err := fx.usecase.Create(ctx, in)
	require.NoError(t, err)

	upd := uc.UpdateInput{
		RemotePassword: &newPass,
		CustomPrice:    &price,
		BillingCycle:   &cycle,
		BillingDay:     &day,
	}
	updated, err := fx.usecase.Update(ctx, sub.ID, upd)
	require.NoError(t, err)
	assert.Equal(t, newPass, updated.RemotePassword)
	assert.InDelta(t, 150000, *updated.CustomPrice, 0.01)
	assert.Equal(t, day, updated.BillingDay)
	assert.Equal(t, cycle, updated.BillingCycle)
	assert.Equal(t, "awal", updated.Notes, "field kosong tidak boleh berubah")

	// Password "" = skip (tidak diubah).
	empty := ""
	upd2 := uc.UpdateInput{RemotePassword: &empty, Notes: &empty}
	updated2, err := fx.usecase.Update(ctx, sub.ID, upd2)
	require.NoError(t, err)
	assert.Equal(t, newPass, updated2.RemotePassword, "password kosong tidak boleh menimpa")
	assert.Equal(t, "", updated2.Notes)
	assert.Equal(t, 2, fx.audit.Count("UPDATE_SUBSCRIPTION"), "dua panggilan update")

	stored, _ := fx.subs.FindByID(ctx, sub.ID)
	assert.Equal(t, newPass, stored.RemotePassword)
}

func TestUpdateSubscription_NotFound(t *testing.T) {
	fx := newManageFixture(t)
	_, err := fx.usecase.Update(context.Background(), "missing", uc.UpdateInput{})
	assert.ErrorIs(t, err, domainBilling.ErrNotFound)
}

// ─── Delete ─────────────────────────────────────────────────────────────

func TestDeleteSubscription_BlockedWhenInvoiced(t *testing.T) {
	fx := newManageFixture(t)
	ctx := context.Background()
	device := "dev-1"

	in := baseCreateInput()
	in.DeviceID = &device
	sub, err := fx.usecase.Create(ctx, in)
	require.NoError(t, err)

	subID := sub.ID
	if err := fx.invoices.Save(ctx, domainBilling.Invoice{
		ID: "inv-1", SubscriptionID: &subID, Status: "UNPAID",
	}); err != nil {
		t.Fatalf("save invoice: %v", err)
	}

	err = fx.usecase.Delete(ctx, sub.ID)
	require.Error(t, err)
	assert.ErrorContains(t, err, "still has invoices")
	assert.ErrorIs(t, err, domainBilling.ErrInvalidInput)

	// Sub masih ada.
	_, ferr := fx.subs.FindByID(ctx, sub.ID)
	assert.NoError(t, ferr)
	assert.Equal(t, 0, fx.manager.Count("Terminate"))
}

func TestDeleteSubscription_TerminatesRouterThenDeletes(t *testing.T) {
	fx := newManageFixture(t)
	ctx := context.Background()
	device := "dev-1"

	in := baseCreateInput()
	in.DeviceID = &device
	sub, err := fx.usecase.Create(ctx, in)
	require.NoError(t, err)

	require.NoError(t, fx.usecase.Delete(ctx, sub.ID))
	assert.Equal(t, 1, fx.manager.Count("Terminate:"+sub.RemoteUsername))
	assert.Equal(t, 1, fx.audit.Count("DELETE_SUBSCRIPTION"))

	_, ferr := fx.subs.FindByID(ctx, sub.ID)
	assert.ErrorIs(t, ferr, mocktest.ErrFakeNotFound)
}
