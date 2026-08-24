package registration_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	connectReg "github.com/quixiq/polyglot/internal/adapter/connect/registration"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/registration"
)

func newHandlerFixture(t *testing.T) (*connectReg.RegistrationConnectHandler, *mocktest.FakeRegistrationRepo, *mocktest.FakeServicePlanRepo) {
	t.Helper()
	repo := mocktest.NewFakeRegistrationRepo()
	plans := mocktest.NewFakeServicePlanRepo()
	customers := mocktest.NewFakeCustomerRepo()
	subs := mocktest.NewFakeSubscriptionRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	notif := mocktest.NewFakeNotificationRepo()
	audit := &mocktest.FakeAuditWriter{}

	plans.Seed(domainPlan.ServicePlan{
		ID: "plan-1", Name: "HOME-20M", ServiceType: "PPPOE",
		BandwidthDownloadKbps: 20000, BandwidthUploadKbps: 20000,
		Price: 150000, IsActive: true,
	})

	mgr := uc.NewManageRegistrationUseCase(repo, notif, audit)
	conv := uc.NewConvertUseCase(uc.ConvertDeps{
		Repo: repo, Plans: plans, Customers: customers,
		Subs: subs, Invoices: invoices, Audit: audit,
	})

	handler := connectReg.NewRegistrationConnectHandler(mgr, conv, repo)
	return handler, repo, plans
}

func TestRegistrationConnectHandler_Submit_Approve_Convert(t *testing.T) {
	handler, repo, _ := newHandlerFixture(t)
	ctx := context.Background()

	// 1. Submit (Public RPC)
	subResp, err := handler.SubmitRegistration(ctx, connect.NewRequest(&devicepb.SubmitRegistrationRequest{
		FullName: "Ahmad Dahlan",
		Phone:    "085606846141",
		Address:  "Jl. Ahmad Yani No. 1",
		PlanId:   "plan-1",
	}))
	require.NoError(t, err)
	regID := subResp.Msg.Registration.Id
	assert.NotEmpty(t, regID)
	assert.Equal(t, "PENDING", subResp.Msg.Registration.Status)

	// 2. Get
	getResp, err := handler.GetRegistration(ctx, connect.NewRequest(&devicepb.GetRegistrationRequest{Id: regID}))
	require.NoError(t, err)
	assert.Equal(t, "Ahmad Dahlan", getResp.Msg.Registration.FullName)

	// 3. Approve
	appResp, err := handler.ApproveRegistration(ctx, connect.NewRequest(&devicepb.ApproveRegistrationRequest{
		Id:         regID,
		AdminNotes: "Disetujui admin",
	}))
	require.NoError(t, err)
	assert.Equal(t, "APPROVED", appResp.Msg.Registration.Status)

	// 4. Mark Installed
	instResp, err := handler.MarkInstalled(ctx, connect.NewRequest(&devicepb.MarkInstalledRequest{
		Id:              regID,
		TechnicianNotes: "ONT terpasang di ruang tamu",
	}))
	require.NoError(t, err)
	assert.Equal(t, "INSTALLED", instResp.Msg.Registration.Status)

	// 5. Convert
	convResp, err := handler.ConvertRegistration(ctx, connect.NewRequest(&devicepb.ConvertRegistrationRequest{
		Id: regID,
	}))
	require.NoError(t, err)
	assert.Equal(t, "ACTIVE", convResp.Msg.Registration.Status)
	assert.NotEmpty(t, convResp.Msg.CustomerId)
	assert.NotEmpty(t, convResp.Msg.SubscriptionId)
	assert.NotEmpty(t, convResp.Msg.InvoiceId)

	_ = repo
}

func TestRegistrationConnectHandler_ValidationErrors(t *testing.T) {
	handler, _, _ := newHandlerFixture(t)
	ctx := context.Background()

	// Submit dengan nama kosong
	_, err := handler.SubmitRegistration(ctx, connect.NewRequest(&devicepb.SubmitRegistrationRequest{
		FullName: "",
		Phone:    "081234567890",
		Address:  "Jl. Test",
		PlanId:   "plan-1",
	}))
	assert.Error(t, err)

	// Get dengan ID kosong
	_, err = handler.GetRegistration(ctx, connect.NewRequest(&devicepb.GetRegistrationRequest{Id: ""}))
	assert.Error(t, err)

	_ = domainRegistration.StatusActive
}
