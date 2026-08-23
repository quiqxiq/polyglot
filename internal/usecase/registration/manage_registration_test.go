package registration_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/registration"
)

func validRegistration() domainRegistration.Registration {
	return domainRegistration.Registration{
		FullName: "Budi Santoso",
		Phone:    "085606846141",
		Address:  "KATAPANG, SAMPANG",
		PlanID:   "plan-1",
	}
}

func newUC(t *testing.T) (*uc.ManageRegistrationUseCase, *mocktest.FakeRegistrationRepo, *mocktest.FakeNotificationRepo, *mocktest.FakeAuditWriter) {
	t.Helper()
	repo := mocktest.NewFakeRegistrationRepo()
	notif := mocktest.NewFakeNotificationRepo()
	auditW := &mocktest.FakeAuditWriter{}
	return uc.NewManageRegistrationUseCase(repo, notif, auditW), repo, notif, auditW
}

func TestSubmit_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		input   domainRegistration.Registration
		wantErr bool
	}{
		{"valid", validRegistration(), false},
		{"missing name", func() domainRegistration.Registration { r := validRegistration(); r.FullName = ""; return r }(), true},
		{"missing address", func() domainRegistration.Registration { r := validRegistration(); r.Address = ""; return r }(), true},
		{"missing plan", func() domainRegistration.Registration { r := validRegistration(); r.PlanID = ""; return r }(), true},
		{"invalid phone", func() domainRegistration.Registration { r := validRegistration(); r.Phone = "12345"; return r }(), true},
		{"landline phone", func() domainRegistration.Registration { r := validRegistration(); r.Phone = "0321123456"; return r }(), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase, _, _, auditW := newUC(t)
			got, err := usecase.Submit(context.Background(), tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, domainRegistration.StatusPending, got.Status)
			assert.NotEmpty(t, got.ID)
			assert.NotEmpty(t, got.RegistrationNo)
			assert.Equal(t, 1, auditW.Count("SUBMIT_REGISTRATION"))
		})
	}
}

func TestApprove_HappyPath(t *testing.T) {
	usecase, repo, notif, _ := newUC(t)
	ctx := context.Background()

	submitted, err := usecase.Submit(ctx, validRegistration())
	require.NoError(t, err)

	approved, err := usecase.Approve(ctx, submitted.ID, 42, "OK")
	require.NoError(t, err)
	assert.Equal(t, domainRegistration.StatusApproved, approved.Status)
	require.NotNil(t, approved.ReviewedBy)
	assert.Equal(t, uint(42), *approved.ReviewedBy)

	stored, err := repo.FindByID(ctx, submitted.ID)
	require.NoError(t, err)
	assert.Equal(t, domainRegistration.StatusApproved, stored.Status)
	require.Len(t, notif.Queued(), 1)
	assert.Equal(t, "REGISTRATION_APPROVED", notif.Queued()[0].MessageType)
}

func TestTransitions_InvalidPaths(t *testing.T) {
	tests := []struct {
		name string
		run  func(u *uc.ManageRegistrationUseCase, id string) error
	}{
		{"approve twice", func(u *uc.ManageRegistrationUseCase, id string) error {
			_, err := u.Approve(context.Background(), id, 1, "")
			if err != nil {
				return err
			}
			_, err = u.Approve(context.Background(), id, 1, "")
			return err
		}},
		{"install before approve", func(u *uc.ManageRegistrationUseCase, id string) error {
			_, err := u.MarkInstalled(context.Background(), id, nil, "")
			return err
		}},
		{"reject after approve", func(u *uc.ManageRegistrationUseCase, id string) error {
			_, err := u.Approve(context.Background(), id, 1, "")
			if err != nil {
				return err
			}
			_, err = u.Reject(context.Background(), id, "late", 1)
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase, _, _, _ := newUC(t)
			sub, err := usecase.Submit(context.Background(), validRegistration())
			require.NoError(t, err)
			err = tt.run(usecase, sub.ID)
			assert.ErrorIs(t, err, uc.ErrInvalidTransition)
		})
	}
}

func TestScheduleInstall_ThenMarkInstalled(t *testing.T) {
	usecase, _, notif, _ := newUC(t)
	ctx := context.Background()
	sub, err := usecase.Submit(ctx, validRegistration())
	require.NoError(t, err)
	_, err = usecase.Approve(ctx, sub.ID, 1, "")
	require.NoError(t, err)

	_, err = usecase.ScheduleInstall(ctx, sub.ID, mustDate("2026-09-01"), nil, nil)
	require.NoError(t, err)
	queued := notif.Queued()
	require.Len(t, queued, 2) // approve + jadwal pasang
	assert.Equal(t, "INSTALLATION_SCHEDULED", queued[1].MessageType)

	techID := uint(7)
	installed, err := usecase.MarkInstalled(ctx, sub.ID, &techID, "ONT terpasang")
	require.NoError(t, err)
	assert.Equal(t, domainRegistration.StatusInstalled, installed.Status)
	require.NotNil(t, installed.InstalledAt)
}

func TestCancel_FromActiveRejected(t *testing.T) {
	usecase, _, _, _ := newUC(t)
	sub, err := usecase.Submit(context.Background(), validRegistration())
	require.NoError(t, err)
	_, err = usecase.Cancel(context.Background(), sub.ID, "pelanggan mundur")
	require.NoError(t, err)

	err = func() error {
		_, err := usecase.Cancel(context.Background(), sub.ID, "again")
		return err
	}()
	assert.ErrorIs(t, err, uc.ErrInvalidTransition)
}

func TestConvert_FullFlow_CreatesArtifactsAndLinksBack(t *testing.T) {
	repo := mocktest.NewFakeRegistrationRepo()
	plans := mocktest.NewFakeServicePlanRepo()
	customers := mocktest.NewFakeCustomerRepo()
	subs := mocktest.NewFakeSubscriptionRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	notif := mocktest.NewFakeNotificationRepo()
	auditW := &mocktest.FakeAuditWriter{}

	mgr := uc.NewManageRegistrationUseCase(repo, notif, auditW)
	plans.Seed(domainPlan.ServicePlan{
		ID: "plan-1", Name: "100-RB-100", ServiceType: "PPPOE",
		BandwidthDownloadKbps: 5120, BandwidthUploadKbps: 5120,
		Price: 100000, InstallationFee: 150000, TaxPercent: 10, IsActive: true,
	})
	conv := uc.NewConvertUseCase(uc.ConvertDeps{
		Repo: repo, Plans: plans, Customers: customers,
		Subs: subs, Invoices: invoices, Audit: auditW,
	})
	var codeSeq int
	conv.WithGenerators(
		func() string { codeSeq++; return "CUST-0000" + string(rune('0'+codeSeq)) },
		func() string { return "12345678" },
		func() string { return "secret99" },
	)

	ctx := context.Background()
	sub, err := mgr.Submit(ctx, validRegistration())
	require.NoError(t, err)
	_, err = mgr.Approve(ctx, sub.ID, 1, "")
	require.NoError(t, err)
	tech := uint(3)
	_, err = mgr.MarkInstalled(ctx, sub.ID, &tech, "")
	require.NoError(t, err)

	converted, err := conv.Convert(ctx, sub.ID, "9")
	require.NoError(t, err)
	assert.Equal(t, domainRegistration.StatusActive, converted.Status)

	// Semua artefak tercipta dan tertaut.
	assert.NotEmpty(t, converted.CustomerID)
	assert.NotEmpty(t, converted.SubscriptionID)
	assert.NotEmpty(t, converted.InvoiceID)

	storedCust, err := customers.FindByID(ctx, converted.CustomerID)
	require.NoError(t, err)
	assert.NotEmpty(t, storedCust.CustomerCode)
	assert.Len(t, storedCust.PortalAccessCode, 8)

	storedSubRow, err := subs.FindByID(ctx, converted.SubscriptionID)
	require.NoError(t, err)
	assert.Equal(t, "BUDI-SANTOSO", storedSubRow.RemoteUsername)
	assert.NotEmpty(t, storedSubRow.RemotePassword)

	inv, err := invoices.FindByID(ctx, converted.InvoiceID)
	require.NoError(t, err)
	assert.InDelta(t, 100000, inv.Subtotal, 0.01) // subtotal = harga paket saja
	assert.Len(t, invoices.ItemsOf(inv.ID), 2)    // fee langganan + biaya pasang
	for _, it := range invoices.ItemsOf(inv.ID) {
		assert.Equal(t, inv.ID, it.InvoiceID)
	}
	assert.InDelta(t, 110000, inv.Total, 0.01) // 100000 + pajak 10%

	// Audit trail lengkap.
	assert.Equal(t, 1, auditW.Count("CREATE_CUSTOMER"))
	assert.Equal(t, 1, auditW.Count("CREATE_SUBSCRIPTION"))
	assert.Equal(t, 1, auditW.Count("CONVERT_REGISTRATION"))
}

func TestConvert_Guards(t *testing.T) {
	repo := mocktest.NewFakeRegistrationRepo()
	plans := mocktest.NewFakeServicePlanRepo()
	conv := uc.NewConvertUseCase(uc.ConvertDeps{
		Repo: repo, Plans: plans,
		Customers: mocktest.NewFakeCustomerRepo(),
		Subs:      mocktest.NewFakeSubscriptionRepo(),
		Invoices:  mocktest.NewFakeInvoiceRepo(),
	})
	ctx := context.Background()

	// Belum INSTALLED → ditolak.
	sub, err := uc.NewManageRegistrationUseCase(repo, nil, nil).Submit(ctx, validRegistration())
	require.NoError(t, err)
	_, err = conv.Convert(ctx, sub.ID, "1")
	assert.ErrorIs(t, err, uc.ErrInvalidTransition)

	// Plan hilang → error jelas.
	reg, err := repo.FindByID(ctx, sub.ID)
	require.NoError(t, err)
	reg.Status = domainRegistration.StatusInstalled
	require.NoError(t, repo.Save(ctx, reg))
	_, err = conv.Convert(ctx, sub.ID, "1")
	assert.ErrorContains(t, err, "plan plan-1")
}

// ─── helpers ────────────────────────────────────────────────────────────

func mustDate(s string) time.Time {
	parsed, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return parsed
}
