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

// ─── PlanUseCase ────────────────────────────────────────────────────────

func TestPlanCreate_Defaults_AndDuplicate(t *testing.T) {
	plans := mocktest.NewFakeServicePlanRepo()
	usecase := uc.NewPlanUseCase(plans, mocktest.NewFakeSubscriptionRepo())
	ctx := context.Background()

	created, err := usecase.Create(ctx, domainPlan.ServicePlan{
		Name: "100-RB-100", ServiceType: "PPPOE",
		BandwidthDownloadKbps: 5120, BandwidthUploadKbps: 5120,
		Price: 100000,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "30d", created.Validity)
	assert.Equal(t, domainPlan.ValidityCalendar, created.ValidityMode)
	assert.True(t, created.IsActive)

	_, err = usecase.Create(ctx, domainPlan.ServicePlan{
		Name: "100-RB-100", ServiceType: "PPPOE",
		BandwidthDownloadKbps: 5120, BandwidthUploadKbps: 5120, Price: 1,
	})
	assert.ErrorIs(t, err, domainBilling.ErrInvalidInput)

	_, err = usecase.Create(ctx, domainPlan.ServicePlan{
		Name: "BAD", ServiceType: "WIFI",
		BandwidthDownloadKbps: 1, BandwidthUploadKbps: 1,
	})
	assert.ErrorContains(t, err, "service_type")

	_, err = usecase.Create(ctx, domainPlan.ServicePlan{
		Name: "NEG", ServiceType: "PPPOE",
		BandwidthDownloadKbps: -5, BandwidthUploadKbps: 5,
	})
	assert.ErrorContains(t, err, "bandwidth")
}

func TestPlanDelete_GuardAktif(t *testing.T) {
	plans := mocktest.NewFakeServicePlanRepo()
	subs := mocktest.NewFakeSubscriptionRepo()
	usecase := uc.NewPlanUseCase(plans, subs)
	ctx := context.Background()

	p := newPlan("plan-guard", "GUARD")
	require.NoError(t, plans.Save(ctx, p))

	// Ada langganan aktif → hapus ditolak.
	deviceID := "dev-g"
	sub := domainSubscription.Subscription{
		ID: "s1", CustomerID: "c1", PlanID: p.ID,
		Status: domainSubscription.StatusActive,
	}
	require.NoError(t, subs.Save(ctx, sub))
	err := usecase.Delete(ctx, p.ID)
	assert.ErrorIs(t, err, domainBilling.ErrPlanInUse)

	// Tidak ada langganan aktif lagi → terhapus.
	_ = deviceID
	sub.Status = domainSubscription.StatusTerminated
	require.NoError(t, subs.Save(ctx, sub))
	require.NoError(t, usecase.Delete(ctx, p.ID))
	_, err = plans.FindByID(ctx, p.ID)
	assert.ErrorIs(t, err, mocktest.ErrFakeNotFound)
}

// ─── SubscriptionLifecycleUseCase ───────────────────────────────────────

type lifecycleFixture struct {
	subs    *mocktest.FakeSubscriptionRepo
	plans   *mocktest.FakeServicePlanRepo
	manager *mocktest.FakeRouterAccountManager
	audit   *mocktest.FakeAuditWriter
	usecase *uc.SubscriptionLifecycleUseCase
}

func newLifecycleFixture(t *testing.T, initialStatus string, provisioned bool) *lifecycleFixture {
	t.Helper()
	subs := mocktest.NewFakeSubscriptionRepo()
	plans := mocktest.NewFakeServicePlanRepo()
	manager := mocktest.NewFakeRouterAccountManager()
	audit := &mocktest.FakeAuditWriter{}

	plans.Seed(newPlan("plan-a", "PLAN-A"))
	plans.Seed(domainPlan.ServicePlan{
		ID: "plan-b", TenantID: "tenant-default", Name: "PLAN-B",
		ServiceType:           domainPlan.TypePPPoE,
		BandwidthDownloadKbps: 10240, BandwidthUploadKbps: 10240,
		Price: 200000, IsActive: true,
	})

	sub := seedActiveSub(t, subs, "sub-lc", nil)
	sub.PlanID = "plan-a"
	sub.RemoteUsername = "LCUSER"
	sub.RemotePassword = "pw"
	sub.ServiceType = "PPPOE"
	if provisioned {
		d := "dev-lc"
		sub.DeviceID = &d
		sub.ProvisionStatus = domainSubscription.ProvisionOK
		sub.RouterProfile = "PLAN-A"
	}
	sub.Status = initialStatus
	require.NoError(t, subs.Save(context.Background(), sub))

	return &lifecycleFixture{
		subs:    subs,
		plans:   plans,
		manager: manager,
		audit:   audit,
		usecase: uc.NewSubscriptionLifecycleUseCase(subs, plans, manager, audit),
	}
}

func TestChangePlan_Provisioned_RouterFirstThenDB(t *testing.T) {
	fx := newLifecycleFixture(t, domainSubscription.StatusActive, true)

	updated, err := fx.usecase.ChangePlan(context.Background(), "sub-lc", "plan-b")
	require.NoError(t, err)
	assert.Equal(t, "plan-b", updated.PlanID)
	assert.Equal(t, "PLAN-B", updated.RouterProfile)
	assert.Nil(t, updated.CustomPrice)

	assert.Equal(t, 1, fx.manager.Count("UpdateAccount:LCUSER->PLAN-B"))
	assert.Equal(t, 1, fx.audit.Count("CHANGE_PLAN"))

	stored, _ := fx.subs.FindByID(context.Background(), "sub-lc")
	assert.Equal(t, "plan-b", stored.PlanID)
}

func TestChangePlan_RouterFail_KeepsDB(t *testing.T) {
	fx := newLifecycleFixture(t, domainSubscription.StatusActive, true)
	fx.manager.Fail = map[string]error{"UpdateAccount:LCUSER->PLAN-B": assert_AnError()}

	_, err := fx.usecase.ChangePlan(context.Background(), "sub-lc", "plan-b")
	require.Error(t, err)
	stored, _ := fx.subs.FindByID(context.Background(), "sub-lc")
	assert.Equal(t, "plan-a", stored.PlanID, "DB tidak boleh berubah saat router gagal")
}

func TestChangePlan_SamePlanNoop(t *testing.T) {
	fx := newLifecycleFixture(t, domainSubscription.StatusActive, true)
	updated, err := fx.usecase.ChangePlan(context.Background(), "sub-lc", "plan-a")
	require.NoError(t, err)
	assert.Equal(t, "plan-a", updated.PlanID)
	assert.Equal(t, 0, fx.manager.Count("UpdateAccount"))
}

func TestChangePlan_EnsuresTargetProfileBeforeSwitch(t *testing.T) {
	fx := newLifecycleFixture(t, domainSubscription.StatusActive, true)

	_, err := fx.usecase.ChangePlan(context.Background(), "sub-lc", "plan-b")
	require.NoError(t, err)
	assert.Equal(t, 1, fx.manager.Count("EnsureProfile:PLAN-B"))
	assert.Equal(t, 1, fx.manager.Count("UpdateAccount:LCUSER->PLAN-B"))
}

func TestChangePlan_ProfileFail_KeepsDB(t *testing.T) {
	fx := newLifecycleFixture(t, domainSubscription.StatusActive, true)
	fx.manager.Fail = map[string]error{"EnsureProfile:PLAN-B": assert_AnError()}

	_, err := fx.usecase.ChangePlan(context.Background(), "sub-lc", "plan-b")
	require.Error(t, err)
	stored, _ := fx.subs.FindByID(context.Background(), "sub-lc")
	assert.Equal(t, "plan-a", stored.PlanID, "DB tak boleh berubah bila profil gagal")
}

func TestSuspendResumeTerminate_Flow(t *testing.T) {
	fx := newLifecycleFixture(t, domainSubscription.StatusActive, true)
	ctx := context.Background()

	suspended, err := fx.usecase.Suspend(ctx, "sub-lc", "cuti 2 bulan")
	require.NoError(t, err)
	assert.Equal(t, domainSubscription.StatusSuspended, suspended.Status)
	assert.Equal(t, 1, fx.manager.Count("Suspend:LCUSER"))

	resumed, err := fx.usecase.Resume(ctx, "sub-lc")
	require.NoError(t, err)
	assert.Equal(t, domainSubscription.StatusActive, resumed.Status)
	assert.Equal(t, 1, fx.manager.Count("Restore:LCUSER->PLAN-A"))

	terminated, err := fx.usecase.Terminate(ctx, "sub-lc", "berhenti")
	require.NoError(t, err)
	assert.Equal(t, domainSubscription.StatusTerminated, terminated.Status)
	require.NotNil(t, terminated.EndDate)
	assert.Equal(t, 1, fx.manager.Count("Terminate:LCUSER"))

	// Transisi ilegal dari TERMINATED.
	_, err = fx.usecase.Suspend(ctx, "sub-lc", "again")
	assert.ErrorIs(t, err, domainBilling.ErrInvalidTransition)
}

func TestLifecycle_NotProvisioned_SkipsRouterCalls(t *testing.T) {
	fx := newLifecycleFixture(t, domainSubscription.StatusActive, false)
	ctx := context.Background()

	_, err := fx.usecase.Suspend(ctx, "sub-lc", "manual")
	require.NoError(t, err)
	assert.Equal(t, 0, fx.manager.Count("Suspend")) // belum provisioned → tanpa call router

	updated, _ := fx.subs.FindByID(ctx, "sub-lc")
	assert.Equal(t, domainSubscription.StatusSuspended, updated.Status)
}

func TestActivate_ProvisionSuccessAndFailure(t *testing.T) {
	fx := newLifecycleFixture(t, domainSubscription.StatusActive, false)
	ctx := context.Background()

	// Sukses: device ditugaskan & akun dibuat di router.
	sub, err := fx.usecase.Activate(ctx, "sub-lc", "dev-act")
	require.NoError(t, err)
	assert.Equal(t, domainSubscription.ProvisionOK, sub.ProvisionStatus)
	assert.Equal(t, "PLAN-A", sub.RouterProfile)
	assert.Equal(t, 1, fx.manager.Count("Provision:LCUSER@PLAN-A"))

	// Gagal router → PENDING (worker retry), error jelas.
	fx2 := newLifecycleFixture(t, domainSubscription.StatusActive, false)
	fx2.manager.Fail = map[string]error{"Provision:LCUSER@PLAN-A": assert_AnError()}
	sub2, err := fx2.usecase.Activate(ctx, "sub-lc", "dev-act2")
	require.Error(t, err)
	assert.Equal(t, domainSubscription.ProvisionPending, sub2.ProvisionStatus)
}

func TestPlanUpdateAndGet(t *testing.T) {
	plans := mocktest.NewFakeServicePlanRepo()
	usecase := uc.NewPlanUseCase(plans, mocktest.NewFakeSubscriptionRepo())
	ctx := context.Background()

	created, err := usecase.Create(ctx, newPlan("plan-u", "OLD"))
	require.NoError(t, err)

	created.Name = "NEW"
	created.Price = 150000
	updated, err := usecase.Update(ctx, created)
	require.NoError(t, err)
	assert.Equal(t, "NEW", updated.Name)
	assert.InDelta(t, 150000, updated.Price, 0.01)

	got, err := usecase.Get(ctx, "plan-u")
	require.NoError(t, err)
	assert.Equal(t, "NEW", got.Name)

	list, err := usecase.List(ctx, true)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// ID kosong / tidak ada.
	_, err = usecase.Update(ctx, domainPlan.ServicePlan{})
	assert.ErrorIs(t, err, domainBilling.ErrInvalidInput)
	_, err = usecase.Get(ctx, "missing")
	assert.ErrorIs(t, err, mocktest.ErrFakeNotFound)
}
