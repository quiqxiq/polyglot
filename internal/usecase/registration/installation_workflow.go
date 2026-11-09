package registration

import (
	"context"
	"fmt"
	"time"

	"github.com/quixiq/polyglot/internal/domain/audit"
	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
)

// ScheduleInstall sets the installation appointment on an APPROVED signup.
func (u *ManageRegistrationUseCase) ScheduleInstall(ctx context.Context, id string, date time.Time, installTime *time.Time, techID *uint) (domainRegistration.Registration, error) {
	reg, err := u.mustGet(ctx, id, domainRegistration.StatusApproved)
	if err != nil {
		return domainRegistration.Registration{}, err
	}
	if date.IsZero() {
		return domainRegistration.Registration{}, fmt.Errorf("%w: install date is required", domainRegistration.ErrInvalidInput)
	}
	reg.ScheduledInstallDate = &date
	reg.ScheduledInstallTime = installTime
	if techID != nil {
		reg.AssignedTechnicianID = techID
	}
	if err := u.repo.Save(ctx, reg); err != nil {
		return domainRegistration.Registration{}, fmt.Errorf("save registration: %w", err)
	}
	u.appendEvent(ctx, reg.ID, domainRegistration.EventInstallationScheduled, "Jadwal pemasangan ditetapkan",
		date.Format("02 Jan 2006"), nil,
		map[string]any{"install_date": date.Format("2006-01-02")})
	u.queueTemplate(ctx, reg.TenantID, "INSTALLATION_SCHEDULED",
		"Halo {{full_name}}, teknisi kami akan datang pada {{install_schedule}} ke alamat {{address}}.",
		map[string]string{
			"full_name":        reg.FullName,
			"install_schedule": date.Format("02 Jan 2006"),
			"address":          reg.Address,
		}, reg.Phone)
	return reg, nil
}

// SurveyInput adalah input hasil survey lokasi oleh teknisi (PSB v2).
type SurveyInput struct {
	SurveyorID                *uint
	SurveyDate                *time.Time
	SurveyTime                *time.Time
	SurveyStatus              string // ELIGIBLE / NOT_ELIGIBLE (kosong = ELIGIBLE)
	SurveyNotes               string
	EstimatedCableLengthMeter int
}

// MarkInstalledInput adalah input penyelesaian instalasi oleh teknisi (PSB v2).
type MarkInstalledInput struct {
	InstallerID          *uint
	DeviceID             string
	TechnicianNotes      string
	InstallationPhotoURL string // bukti foto instalasi selesai
	ONTMACAddress        string // identitas fisik modem (bukti, bukan manajemen ONT)
	ONTSerialNumber      string
	OLTDeviceID          string // OLT tempat ONT terdaftar (integrasi OLT/ACS)
	PONIndex             *int
	ONUIndex             *int
}

// RecordSurvey mencatat hasil survey lokasi: PENDING/APPROVED → SURVEY.
func (u *ManageRegistrationUseCase) RecordSurvey(ctx context.Context, id string, in SurveyInput) (domainRegistration.Registration, error) {
	reg, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return domainRegistration.Registration{}, domainRegistration.ErrNotFound
	}
	switch reg.Status {
	case domainRegistration.StatusPending, domainRegistration.StatusApproved:
	default:
		return domainRegistration.Registration{}, fmt.Errorf("%w: survey requires PENDING/APPROVED, got %s",
			domainRegistration.ErrInvalidTransition, reg.Status)
	}
	status := in.SurveyStatus
	if status == "" {
		status = domainRegistration.SurveyEligible
	}
	if status != domainRegistration.SurveyEligible && status != domainRegistration.SurveyNotEligible {
		return domainRegistration.Registration{}, fmt.Errorf("%w: survey_status must be ELIGIBLE or NOT_ELIGIBLE",
			domainRegistration.ErrInvalidInput)
	}
	now := u.now()
	reg.Status = domainRegistration.StatusSurvey
	reg.SurveyStatus = status
	reg.SurveyNotes = in.SurveyNotes
	reg.EstimatedCableLengthMeter = in.EstimatedCableLengthMeter
	surveyDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if in.SurveyDate != nil {
		surveyDate = *in.SurveyDate
	}
	reg.SurveyDate = &surveyDate
	reg.SurveyTime = in.SurveyTime
	if in.SurveyorID != nil {
		reg.SurveyorID = in.SurveyorID
		reg.AssignedTechnicianID = in.SurveyorID
	}
	reg.UpdatedAt = now
	if err := u.repo.Save(ctx, reg); err != nil {
		return domainRegistration.Registration{}, fmt.Errorf("save registration: %w", err)
	}
	u.writeAudit(ctx, audit.ActorUser, actorStr(in.SurveyorID), "RECORD_SURVEY", "registration", reg.ID)
	u.appendEvent(ctx, reg.ID, domainRegistration.EventSurveyCompleted, "Survey lokasi selesai",
		in.SurveyNotes, in.SurveyorID, map[string]any{
			"survey_status":                status,
			"estimated_cable_length_meter": in.EstimatedCableLengthMeter,
		})
	return reg, nil
}

// StartInstallation memulai pekerjaan lapangan: APPROVED → ON_PROGRESS.
func (u *ManageRegistrationUseCase) StartInstallation(ctx context.Context, id, actorID string) (domainRegistration.Registration, error) {
	reg, err := u.mustGet(ctx, id, domainRegistration.StatusApproved)
	if err != nil {
		return domainRegistration.Registration{}, err
	}
	now := u.now()
	reg.Status = domainRegistration.StatusOnProgress
	reg.InstallationStartedAt = &now
	reg.UpdatedAt = now
	if err := u.repo.Save(ctx, reg); err != nil {
		return domainRegistration.Registration{}, fmt.Errorf("save registration: %w", err)
	}
	u.writeAudit(ctx, audit.ActorUser, actorID, "START_INSTALLATION", "registration", reg.ID)
	u.appendEvent(ctx, reg.ID, domainRegistration.EventInstallationStarted, "Pemasangan dimulai",
		"Teknisi mulai pemasangan di lokasi.", nil, nil)
	return reg, nil
}

// MarkInstalledWithEvidence mencatat penyelesaian instalasi beserta bukti
// (foto, MAC/SN modem): APPROVED/ON_PROGRESS → INSTALLED.
func (u *ManageRegistrationUseCase) MarkInstalledWithEvidence(ctx context.Context, id string, in MarkInstalledInput) (domainRegistration.Registration, error) {
	reg, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return domainRegistration.Registration{}, domainRegistration.ErrNotFound
	}
	switch reg.Status {
	case domainRegistration.StatusApproved, domainRegistration.StatusOnProgress:
	default:
		return domainRegistration.Registration{}, fmt.Errorf("%w: mark installed requires APPROVED/ON_PROGRESS, got %s",
			domainRegistration.ErrInvalidTransition, reg.Status)
	}
	now := u.now()
	reg.Status = domainRegistration.StatusInstalled
	reg.InstalledAt = &now
	reg.TechnicianNotes = in.TechnicianNotes
	if in.DeviceID != "" {
		reg.TargetDeviceID = in.DeviceID
	}
	if in.InstallerID != nil {
		reg.AssignedTechnicianID = in.InstallerID
	}
	reg.InstallationPhotoURL = in.InstallationPhotoURL
	reg.ONTMACAddress = in.ONTMACAddress
	reg.ONTSerialNumber = in.ONTSerialNumber
	reg.OLTDeviceID = in.OLTDeviceID
	reg.PONIndex = in.PONIndex
	reg.ONUIndex = in.ONUIndex
	reg.UpdatedAt = now
	if err := u.repo.Save(ctx, reg); err != nil {
		return domainRegistration.Registration{}, fmt.Errorf("save registration: %w", err)
	}
	u.writeAudit(ctx, audit.ActorUser, actorStr(in.InstallerID), "MARK_INSTALLED", "registration", reg.ID)
	u.appendEvent(ctx, reg.ID, domainRegistration.EventInstallationCompleted, "Pemasangan selesai",
		in.TechnicianNotes, in.InstallerID, map[string]any{
			"device_id":  in.DeviceID,
			"ont_mac":    in.ONTMACAddress,
			"ont_serial": in.ONTSerialNumber,
		})
	u.queueTemplate(ctx, reg.TenantID, "INSTALLATION_COMPLETED",
		"Halo {{full_name}}, pemasangan layanan di {{address}} telah selesai. Akun Anda akan segera aktif.",
		map[string]string{
			"full_name": reg.FullName,
			"address":   reg.Address,
		}, reg.Phone)
	return reg, nil
}

// MarkInstalled records technician completion: APPROVED/ON_PROGRESS → INSTALLED.
// deviceID adalah router BRAS yang dipilih teknisi di lokasi pasang (F2-7).
func (u *ManageRegistrationUseCase) MarkInstalled(ctx context.Context, id string, installerID *uint, deviceID, techNotes string) (domainRegistration.Registration, error) {
	return u.MarkInstalledWithEvidence(ctx, id, MarkInstalledInput{
		InstallerID:     installerID,
		DeviceID:        deviceID,
		TechnicianNotes: techNotes,
	})
}
