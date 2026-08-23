package registration_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/registration"
)

// TestConvert_DefaultGenerators menjalankan konversi tanpa override
// generator — mengeksekusi jalur idgen default termasuk retry-loop kode unik.
func TestConvert_DefaultGenerators(t *testing.T) {
	repo := mocktest.NewFakeRegistrationRepo()
	plans := mocktest.NewFakeServicePlanRepo()
	customers := mocktest.NewFakeCustomerRepo()
	subs := mocktest.NewFakeSubscriptionRepo()
	invoices := mocktest.NewFakeInvoiceRepo()

	plans.Seed(domainPlan.ServicePlan{
		ID: "plan-d", Name: "HOME-20M", ServiceType: domainPlan.TypePPPoE,
		Price: 150000, TaxPercent: 0, IsActive: true,
	})

	mgr := uc.NewManageRegistrationUseCase(repo, nil, nil)
	conv := uc.NewConvertUseCase(uc.ConvertDeps{
		Repo: repo, Plans: plans, Customers: customers,
		Subs: subs, Invoices: invoices,
	})

	ctx := context.Background()
	input := validRegistration()
	input.PlanID = "plan-d"
	sub, err := mgr.Submit(ctx, input)
	require.NoError(t, err)
	reg, err := repo.FindByID(ctx, sub.ID)
	require.NoError(t, err)
	reg.Status = domainRegistration.StatusInstalled
	require.NoError(t, repo.Save(ctx, reg))

	converted, err := conv.Convert(ctx, sub.ID, "1")
	require.NoError(t, err)
	assert.NotEmpty(t, converted.CustomerID)

	cust, err := customers.FindByID(ctx, converted.CustomerID)
	require.NoError(t, err)
	assert.Len(t, cust.CustomerCode, 10) // CUST-XXXXX
	assert.Len(t, cust.PortalAccessCode, 8)

	srow, err := subs.FindByID(ctx, converted.SubscriptionID)
	require.NoError(t, err)
	assert.Equal(t, "BUDI-SANTOSO", srow.RemoteUsername)
	assert.NotEmpty(t, srow.RemotePassword)
}

func TestReject_HappyPath(t *testing.T) {
	repo := mocktest.NewFakeRegistrationRepo()
	auditW := &mocktest.FakeAuditWriter{}
	mgr := uc.NewManageRegistrationUseCase(repo, nil, auditW)

	sub, err := mgr.Submit(context.Background(), validRegistration())
	require.NoError(t, err)

	rejected, err := mgr.Reject(context.Background(), sub.ID, "dokumen tidak lengkap", 77)
	require.NoError(t, err)
	assert.Equal(t, domainRegistration.StatusRejected, rejected.Status)
	assert.Equal(t, "dokumen tidak lengkap", rejected.RejectedReason)
	assert.Equal(t, 1, auditW.Count("REJECT_REGISTRATION"))

	_, err = repo.FindByID(context.Background(), sub.ID)
	require.NoError(t, err)
}
