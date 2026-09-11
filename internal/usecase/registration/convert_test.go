// Convert-use-case tests (Fase 0/Fase 1 — PLAN-ISP-CORE-HARDENING.md).
//
//   - Happy path: INSTALLED → ACTIVE membuat customer + subscription + invoice
//     pertama dan memprovisikan akun bila device diberikan.
//   - F1-13: biaya pemasangan masuk Subtotal/Total invoice.
//   - F1-12: kegagalan persistensi tidak boleh meninggalkan artefak yatim
//     (konversi dipersistensikan atomik lewat port.ConversionWriter).
package registration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/registration"
)

type convertFixture struct {
	repo      *mocktest.FakeRegistrationRepo
	plans     *mocktest.FakeServicePlanRepo
	customers *mocktest.FakeCustomerRepo
	subs      *mocktest.FakeSubscriptionRepo
	invoices  *mocktest.FakeInvoiceRepo
	mgr       *mocktest.FakeRouterAccountManager
	writer    *mocktest.FakeConversionWriter
	conv      *uc.ConvertUseCase
}

func newConvertFixture(t *testing.T) *convertFixture {
	t.Helper()
	f := &convertFixture{
		repo:      mocktest.NewFakeRegistrationRepo(),
		plans:     mocktest.NewFakeServicePlanRepo(),
		customers: mocktest.NewFakeCustomerRepo(),
		subs:      mocktest.NewFakeSubscriptionRepo(),
		invoices:  mocktest.NewFakeInvoiceRepo(),
		mgr:       mocktest.NewFakeRouterAccountManager(),
	}
	f.writer = mocktest.NewFakeConversionWriter(f.customers, f.subs, f.invoices, f.repo)
	f.plans.Seed(domainPlan.ServicePlan{
		ID: "plan-1", Name: "100-RB-100", ServiceType: domainPlan.TypePPPoE,
		BandwidthDownloadKbps: 5120, BandwidthUploadKbps: 5120,
		Price: 100000, InstallationFee: 150000, TaxPercent: 10, IsActive: true,
	})
	f.conv = uc.NewConvertUseCase(uc.ConvertDeps{
		Repo: f.repo, Plans: f.plans, Customers: f.customers,
		Subs: f.subs, Audit: &mocktest.FakeAuditWriter{},
		Writer: f.writer, Manager: f.mgr,
	})
	f.conv.WithGenerators(
		func() string { return "CUST-00001" },
		func() string { return "12345678" },
		func() string { return "secret99" },
	)
	return f
}

func seedInstalledRegistration(t *testing.T, repo *mocktest.FakeRegistrationRepo) {
	t.Helper()
	require.NoError(t, repo.Save(context.Background(), domainRegistration.Registration{
		ID: "reg-1", RegistrationNo: "REG-202609-0001",
		PlanID: "plan-1", FullName: "Budi Santoso", Phone: "085606846141",
		Address: "KATAPANG, SAMPANG", Status: domainRegistration.StatusInstalled,
	}))
}

// ─── Happy path ─────────────────────────────────────────────────────────

func TestConvertWithDevice_ProvisionsAccountWhenDeviceGiven(t *testing.T) {
	f := newConvertFixture(t)
	ctx := context.Background()
	seedInstalledRegistration(t, f.repo)

	reg, err := f.conv.ConvertWithDevice(ctx, "reg-1", "dev-9", "9")
	require.NoError(t, err)
	assert.Equal(t, domainRegistration.StatusActive, reg.Status)
	assert.Equal(t, 1, f.mgr.Count("ProvisionPPPoE:"))

	sub, err := f.subs.FindByID(ctx, reg.SubscriptionID)
	require.NoError(t, err)
	require.NotNil(t, sub.DeviceID)
	assert.Equal(t, "dev-9", *sub.DeviceID)
	assert.Equal(t, domainSubscription.ProvisionOK, sub.ProvisionStatus)
	assert.Equal(t, "100-RB-100", sub.RouterProfile)
}

// ─── F1-13: biaya pemasangan masuk total invoice ────────────────────────

func TestConvertWithDevice_InstallationFeeIncludedInTotal(t *testing.T) {
	f := newConvertFixture(t)
	ctx := context.Background()
	seedInstalledRegistration(t, f.repo)

	reg, err := f.conv.Convert(ctx, "reg-1", "9")
	require.NoError(t, err)

	inv, err := f.invoices.FindByID(ctx, reg.InvoiceID)
	require.NoError(t, err)
	assert.InDelta(t, 250000, inv.Subtotal, 0.01) // 100.000 paket + 150.000 pasang
	assert.InDelta(t, 275000, inv.Total, 0.01)    // subtotal + pajak 10% dari 250.000
	assert.Len(t, f.invoices.ItemsOf(inv.ID), 2)
}

// ─── F1-12: konversi atomik ─────────────────────────────────────────────

func TestConvertWithDevice_PartialFailureRollsBackArtifacts(t *testing.T) {
	f := newConvertFixture(t)
	ctx := context.Background()
	seedInstalledRegistration(t, f.repo)

	f.writer.Fail = errors.New("simulated db failure")

	_, err := f.conv.ConvertWithDevice(ctx, "reg-1", "dev-9", "9")
	require.Error(t, err)

	// Tidak boleh ada artefak yatim yang tersisa.
	custs, err := f.customers.FindAll(ctx)
	require.NoError(t, err)
	assert.Empty(t, custs)

	allSubs, err := f.subs.FindAll(ctx)
	require.NoError(t, err)
	assert.Empty(t, allSubs)

	invs, err := f.invoices.FindAll(ctx)
	require.NoError(t, err)
	assert.Empty(t, invs)

	stored, err := f.repo.FindByID(ctx, "reg-1")
	require.NoError(t, err)
	assert.Equal(t, domainRegistration.StatusInstalled, stored.Status)
	assert.Empty(t, stored.CustomerID)
}

// ─── F2-7: fallback router dari pilihan teknisi ─────────────────────────

func TestConvertWithDevice_FallsBackToTechnicianTargetDevice(t *testing.T) {
	f := newConvertFixture(t)
	ctx := context.Background()
	require.NoError(t, f.repo.Save(ctx, domainRegistration.Registration{
		ID: "reg-1", RegistrationNo: "REG-202609-0001",
		PlanID: "plan-1", FullName: "Budi Santoso", Phone: "085606846141",
		Address: "KATAPANG, SAMPANG", Status: domainRegistration.StatusInstalled,
		TargetDeviceID: "dev-target",
	}))

	reg, err := f.conv.Convert(ctx, "reg-1", "9") // deviceID kosong
	require.NoError(t, err)

	sub, err := f.subs.FindByID(ctx, reg.SubscriptionID)
	require.NoError(t, err)
	require.NotNil(t, sub.DeviceID)
	assert.Equal(t, "dev-target", *sub.DeviceID)
	assert.Equal(t, 1, f.mgr.Count("ProvisionPPPoE:"))
}
