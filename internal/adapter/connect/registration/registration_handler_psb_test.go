// Handler ConnectRPC untuk PSB v2 (planning v2 Fase 2): survey, mulai
// instalasi, bukti instalasi, lacak publik, dan daftar paket publik.
package registration_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	connectReg "github.com/quixiq/polyglot/internal/adapter/connect/registration"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/registration"
)

func newPSBHandlerFixture(t *testing.T) *connectReg.RegistrationConnectHandler {
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

	mgr := uc.NewManageRegistrationUseCase(repo, notif, audit).
		WithEvents(mocktest.NewFakeRegistrationEventRepo()).
		WithPlans(plans)
	conv := uc.NewConvertUseCase(uc.ConvertDeps{
		Repo: repo, Plans: plans, Customers: customers,
		Subs: subs, Audit: audit,
		Writer: mocktest.NewFakeConversionWriter(customers, subs, invoices, repo),
	})
	return connectReg.NewRegistrationConnectHandler(mgr, conv)
}

func submitPSBRegistration(t *testing.T, h *connectReg.RegistrationConnectHandler) *polyglotv1.Registration {
	t.Helper()
	resp, err := h.SubmitRegistration(context.Background(), connect.NewRequest(&polyglotv1.SubmitRegistrationRequest{
		FullName: "Siti Aminah",
		Phone:    "085606846141",
		Address:  "Jl. Merdeka No. 2",
		PlanId:   "plan-1",
		Nik:      "3507123456780001",
	}))
	require.NoError(t, err)
	return resp.Msg.Registration
}

func TestPSBHandler_SurveyStartInstallEvidence(t *testing.T) {
	h := newPSBHandlerFixture(t)
	ctx := context.Background()
	reg := submitPSBRegistration(t, h)
	require.NotEmpty(t, reg.PublicToken, "token lacak harus dibuat saat submit")

	// 1. RecordSurvey
	surveyResp, err := h.RecordSurvey(ctx, connect.NewRequest(&polyglotv1.RecordSurveyRequest{
		Id:                        reg.Id,
		SurveyStatus:              domainRegistration.SurveyEligible,
		SurveyNotes:               "ODP 150m, port kosong",
		EstimatedCableLengthMeter: 140,
	}))
	require.NoError(t, err)
	assert.Equal(t, domainRegistration.StatusSurvey, surveyResp.Msg.Registration.Status)
	assert.Equal(t, domainRegistration.SurveyEligible, surveyResp.Msg.Registration.SurveyStatus)
	assert.Equal(t, int32(140), surveyResp.Msg.Registration.EstimatedCableLengthMeter)

	// 2. Approve dari SURVEY
	appResp, err := h.ApproveRegistration(ctx, connect.NewRequest(&polyglotv1.ApproveRegistrationRequest{
		Id: reg.Id, AdminNotes: "lanjut",
	}))
	require.NoError(t, err)
	assert.Equal(t, domainRegistration.StatusApproved, appResp.Msg.Registration.Status)

	// 3. StartInstallation
	startResp, err := h.StartInstallation(ctx, connect.NewRequest(&polyglotv1.StartInstallationRequest{Id: reg.Id}))
	require.NoError(t, err)
	assert.Equal(t, domainRegistration.StatusOnProgress, startResp.Msg.Registration.Status)
	assert.NotZero(t, startResp.Msg.Registration.InstallationStartedAtUnix)

	// 4. MarkInstalled dengan bukti
	instResp, err := h.MarkInstalled(ctx, connect.NewRequest(&polyglotv1.MarkInstalledRequest{
		Id:                   reg.Id,
		DeviceId:             "dev-bras-1",
		TechnicianNotes:      "redaman -19.5 dBm",
		InstallationPhotoUrl: "/uploads/instalasi/reg.jpg",
		OntMacAddress:        "ZTEGC1234567",
		OntSerialNumber:      "ZTEGC12345678",
	}))
	require.NoError(t, err)
	inst := instResp.Msg.Registration
	assert.Equal(t, domainRegistration.StatusInstalled, inst.Status)
	assert.Equal(t, "/uploads/instalasi/reg.jpg", inst.InstallationPhotoUrl)
	assert.Equal(t, "ZTEGC1234567", inst.OntMacAddress)
	assert.Equal(t, "ZTEGC12345678", inst.OntSerialNumber)

	// 5. Timeline via ListRegistrationEvents
	evResp, err := h.ListRegistrationEvents(ctx, connect.NewRequest(&polyglotv1.ListRegistrationEventsRequest{Id: reg.Id}))
	require.NoError(t, err)
	require.NotEmpty(t, evResp.Msg.Events)
	assert.Equal(t, domainRegistration.EventCreated, evResp.Msg.Events[0].EventType)

	// 6. TrackRegistration publik via token
	trackResp, err := h.TrackRegistration(ctx, connect.NewRequest(&polyglotv1.TrackRegistrationRequest{
		PublicToken: reg.PublicToken,
	}))
	require.NoError(t, err)
	assert.Equal(t, reg.Id, trackResp.Msg.Registration.Id)
	assert.NotEmpty(t, trackResp.Msg.Events)
}

func TestPSBHandler_ListPublicPlans(t *testing.T) {
	h := newPSBHandlerFixture(t)
	resp, err := h.ListPublicPlans(context.Background(), connect.NewRequest(&polyglotv1.ListPublicPlansRequest{}))
	require.NoError(t, err)
	require.Len(t, resp.Msg.Plans, 1)
	assert.Equal(t, "plan-1", resp.Msg.Plans[0].Id)
	assert.Equal(t, "HOME-20M", resp.Msg.Plans[0].Name)
	assert.InDelta(t, 150000, resp.Msg.Plans[0].Price, 0.01)
}

func TestPSBHandler_TrackUnknownToken(t *testing.T) {
	h := newPSBHandlerFixture(t)
	_, err := h.TrackRegistration(context.Background(), connect.NewRequest(&polyglotv1.TrackRegistrationRequest{
		PublicToken: "tidak-ada",
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}
