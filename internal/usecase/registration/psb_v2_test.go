// PSB v2 (planning v2 Fase 2): survey → approve → mulai instalasi →
// instalasi selesai dengan bukti, plus timeline registration_events dan
// lacak status via token publik.
package registration_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/registration"
)

func newPSBFixture(t *testing.T) (*uc.ManageRegistrationUseCase, *mocktest.FakeRegistrationRepo, *mocktest.FakeRegistrationEventRepo) {
	t.Helper()
	repo := mocktest.NewFakeRegistrationRepo()
	events := mocktest.NewFakeRegistrationEventRepo()
	mgr := uc.NewManageRegistrationUseCase(repo, mocktest.NewFakeNotificationRepo(), &mocktest.FakeAuditWriter{}).
		WithEvents(events).
		WithClock(func() time.Time { return time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC) })
	return mgr, repo, events
}

func submitPSB(t *testing.T, mgr *uc.ManageRegistrationUseCase) domainRegistration.Registration {
	t.Helper()
	reg, err := mgr.Submit(context.Background(), validRegistration())
	require.NoError(t, err)
	return reg
}

func TestSubmit_GeneratesPublicTokenAndEvent(t *testing.T) {
	mgr, repo, events := newPSBFixture(t)
	reg := submitPSB(t, mgr)

	assert.Len(t, reg.PublicToken, 32, "public token 128-bit hex")
	stored, err := repo.FindByPublicToken(context.Background(), reg.PublicToken)
	require.NoError(t, err)
	assert.Equal(t, reg.ID, stored.ID)
	assert.Equal(t, 1, events.Count(domainRegistration.EventCreated))
}

func TestRecordSurvey_MovesPendingToSurvey(t *testing.T) {
	mgr, repo, events := newPSBFixture(t)
	reg := submitPSB(t, mgr)
	tech := uint(7)
	surveyTime := time.Date(2026, 9, 11, 9, 30, 0, 0, time.UTC)

	updated, err := mgr.RecordSurvey(context.Background(), reg.ID, uc.SurveyInput{
		SurveyorID:                &tech,
		SurveyTime:                &surveyTime,
		SurveyStatus:              domainRegistration.SurveyEligible,
		SurveyNotes:               "ODP 200m, port tersedia",
		EstimatedCableLengthMeter: 180,
	})
	require.NoError(t, err)
	assert.Equal(t, domainRegistration.StatusSurvey, updated.Status)
	assert.Equal(t, domainRegistration.SurveyEligible, updated.SurveyStatus)
	assert.Equal(t, "ODP 200m, port tersedia", updated.SurveyNotes)
	assert.Equal(t, 180, updated.EstimatedCableLengthMeter)
	require.NotNil(t, updated.SurveyDate)
	assert.Equal(t, dateOf(2026, 9, 10), *updated.SurveyDate, "default survey date = hari ini")
	require.NotNil(t, updated.SurveyorID)
	assert.Equal(t, tech, *updated.SurveyorID)
	assert.Equal(t, 1, events.Count(domainRegistration.EventSurveyCompleted))

	stored, err := repo.FindByID(context.Background(), reg.ID)
	require.NoError(t, err)
	assert.Equal(t, domainRegistration.StatusSurvey, stored.Status)
}

func TestRecordSurvey_RejectsInvalidStatus(t *testing.T) {
	mgr, _, _ := newPSBFixture(t)
	reg := submitPSB(t, mgr)
	_, err := mgr.RecordSurvey(context.Background(), reg.ID, uc.SurveyInput{SurveyStatus: "MAYBE"})
	assert.ErrorIs(t, err, domainRegistration.ErrInvalidInput)
}

func TestApprove_FromSurvey(t *testing.T) {
	mgr, _, _ := newPSBFixture(t)
	reg := submitPSB(t, mgr)
	_, err := mgr.RecordSurvey(context.Background(), reg.ID, uc.SurveyInput{
		SurveyStatus: domainRegistration.SurveyEligible,
	})
	require.NoError(t, err)

	approved, err := mgr.Approve(context.Background(), reg.ID, 2, "lanjut")
	require.NoError(t, err)
	assert.Equal(t, domainRegistration.StatusApproved, approved.Status)
}

func TestStartInstallation_ApprovedToOnProgress(t *testing.T) {
	mgr, _, events := newPSBFixture(t)
	reg := submitPSB(t, mgr)
	_, err := mgr.Approve(context.Background(), reg.ID, 2, "")
	require.NoError(t, err)

	started, err := mgr.StartInstallation(context.Background(), reg.ID, "3")
	require.NoError(t, err)
	assert.Equal(t, domainRegistration.StatusOnProgress, started.Status)
	require.NotNil(t, started.InstallationStartedAt)
	assert.Equal(t, 1, events.Count(domainRegistration.EventInstallationStarted))
}

func TestStartInstallation_RequiresApproved(t *testing.T) {
	mgr, _, _ := newPSBFixture(t)
	reg := submitPSB(t, mgr)
	_, err := mgr.StartInstallation(context.Background(), reg.ID, "3")
	assert.ErrorIs(t, err, domainRegistration.ErrInvalidTransition)
}

func TestMarkInstalledWithEvidence_FromOnProgress(t *testing.T) {
	mgr, repo, events := newPSBFixture(t)
	reg := submitPSB(t, mgr)
	_, err := mgr.Approve(context.Background(), reg.ID, 2, "")
	require.NoError(t, err)
	_, err = mgr.StartInstallation(context.Background(), reg.ID, "3")
	require.NoError(t, err)

	tech := uint(3)
	installed, err := mgr.MarkInstalledWithEvidence(context.Background(), reg.ID, uc.MarkInstalledInput{
		InstallerID:          &tech,
		DeviceID:             "dev-01",
		TechnicianNotes:      "kabel 180m, redaman -19.5 dBm",
		InstallationPhotoURL: "/uploads/instalasi/reg-1.jpg",
		ONTMACAddress:        "ZTEGC1234567",
		ONTSerialNumber:      "ZTEGC12345678",
	})
	require.NoError(t, err)
	assert.Equal(t, domainRegistration.StatusInstalled, installed.Status)
	assert.Equal(t, "dev-01", installed.TargetDeviceID)
	assert.Equal(t, "/uploads/instalasi/reg-1.jpg", installed.InstallationPhotoURL)
	assert.Equal(t, "ZTEGC1234567", installed.ONTMACAddress)
	assert.Equal(t, "ZTEGC12345678", installed.ONTSerialNumber)
	assert.Equal(t, 1, events.Count(domainRegistration.EventInstallationCompleted))

	stored, err := repo.FindByID(context.Background(), reg.ID)
	require.NoError(t, err)
	assert.Equal(t, domainRegistration.StatusInstalled, stored.Status)
}

func TestMarkInstalled_LegacyFromApprovedStillWorks(t *testing.T) {
	mgr, _, _ := newPSBFixture(t)
	reg := submitPSB(t, mgr)
	_, err := mgr.Approve(context.Background(), reg.ID, 2, "")
	require.NoError(t, err)

	installed, err := mgr.MarkInstalled(context.Background(), reg.ID, nil, "dev-02", "selesai")
	require.NoError(t, err)
	assert.Equal(t, domainRegistration.StatusInstalled, installed.Status)
}

func TestReject_FromSurvey(t *testing.T) {
	mgr, _, events := newPSBFixture(t)
	reg := submitPSB(t, mgr)
	_, err := mgr.RecordSurvey(context.Background(), reg.ID, uc.SurveyInput{
		SurveyStatus: domainRegistration.SurveyNotEligible,
		SurveyNotes:  "di luar jangkauan ODP",
	})
	require.NoError(t, err)

	rejected, err := mgr.Reject(context.Background(), reg.ID, "di luar coverage", 4)
	require.NoError(t, err)
	assert.Equal(t, domainRegistration.StatusRejected, rejected.Status)
	assert.Equal(t, 1, events.Count(domainRegistration.EventRejected))
}

func TestTrackByPublicToken(t *testing.T) {
	mgr, _, _ := newPSBFixture(t)
	reg := submitPSB(t, mgr)

	got, err := mgr.TrackByPublicToken(context.Background(), reg.PublicToken)
	require.NoError(t, err)
	assert.Equal(t, reg.ID, got.ID)

	_, err = mgr.TrackByPublicToken(context.Background(), "token-tidak-ada")
	assert.ErrorIs(t, err, domainRegistration.ErrNotFound)
	_, err = mgr.TrackByPublicToken(context.Background(), " ")
	assert.ErrorIs(t, err, domainRegistration.ErrInvalidInput)
}

func TestListEvents_Timeline(t *testing.T) {
	mgr, _, _ := newPSBFixture(t)
	reg := submitPSB(t, mgr)
	_, err := mgr.RecordSurvey(context.Background(), reg.ID, uc.SurveyInput{
		SurveyStatus: domainRegistration.SurveyEligible,
	})
	require.NoError(t, err)

	events, err := mgr.ListEvents(context.Background(), reg.ID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, domainRegistration.EventCreated, events[0].EventType)
	assert.Equal(t, domainRegistration.EventSurveyCompleted, events[1].EventType)
	assert.Equal(t, reg.ID, events[0].RegistrationID)
}

func dateOf(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
