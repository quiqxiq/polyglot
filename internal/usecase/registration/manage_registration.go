package registration

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/audit"
	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
	domainTenant "github.com/quixiq/polyglot/internal/domain/tenant"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/logger"
)

// Error sentinels untuk alur pendaftaran tinggal di
// internal/domain/registration/errors.go per DEVELOPMENT-GUIDELINES.md §6.

var phoneRe = regexp.MustCompile(`^(\+?62|0)8\d{7,12}$`)

// ManageRegistrationUseCase orchestrates the signup flow:
// PENDING → APPROVED → INSTALLED → ACTIVE (atau REJECTED/CANCELLED).
type ManageRegistrationUseCase struct {
	repo   port.RegistrationRepository
	notif  port.NotificationRepository
	audit  port.AuditLogWriter
	events port.RegistrationEventRepository
	plans  port.ServicePlanRepository

	now func() time.Time // dapat dioverride di test
}

// NewManageRegistrationUseCase wires the use case to its ports.
func NewManageRegistrationUseCase(
	repo port.RegistrationRepository,
	notif port.NotificationRepository,
	auditW port.AuditLogWriter,
) *ManageRegistrationUseCase {
	return &ManageRegistrationUseCase{repo: repo, notif: notif, audit: auditW, now: time.Now}
}

// WithClock mengganti sumber waktu (untuk test deterministik).
func (u *ManageRegistrationUseCase) WithClock(now func() time.Time) *ManageRegistrationUseCase {
	u.now = now
	return u
}

// WithEvents menautkan timeline/audit PSB (registration_events).
func (u *ManageRegistrationUseCase) WithEvents(events port.RegistrationEventRepository) *ManageRegistrationUseCase {
	u.events = events
	return u
}

// WithPlans menautkan sumber paket aktif untuk formulir pendaftaran publik.
func (u *ManageRegistrationUseCase) WithPlans(plans port.ServicePlanRepository) *ManageRegistrationUseCase {
	u.plans = plans
	return u
}

// PublicPlans mengembalikan paket aktif yang boleh dipilih calon pelanggan.
func (u *ManageRegistrationUseCase) PublicPlans(ctx context.Context) ([]domainPlan.ServicePlan, error) {
	if u.plans == nil {
		return nil, nil
	}
	list, err := u.plans.List(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("list active plans: %w", err)
	}
	return list, nil
}

// Submit validates and stores a new signup (ringkas tanpa NIK/KTP).
func (u *ManageRegistrationUseCase) Submit(ctx context.Context, reg domainRegistration.Registration) (domainRegistration.Registration, error) {
	if strings.TrimSpace(reg.FullName) == "" || strings.TrimSpace(reg.Address) == "" {
		return domainRegistration.Registration{}, fmt.Errorf("%w: full_name and address are required", domainRegistration.ErrInvalidInput)
	}
	if !phoneRe.MatchString(reg.Phone) {
		return domainRegistration.Registration{}, fmt.Errorf("%w: phone must be an Indonesian mobile number", domainRegistration.ErrInvalidInput)
	}
	if reg.PlanID == "" {
		return domainRegistration.Registration{}, fmt.Errorf("%w: plan_id is required", domainRegistration.ErrInvalidInput)
	}

	now := u.now()
	reg.ID = idgen.New("reg")
	reg.Status = domainRegistration.StatusPending
	reg.RegistrationNo = fmt.Sprintf("REG-%s-%04d", now.Format("200601"), now.UnixNano()%10000)
	if reg.PublicToken == "" {
		reg.PublicToken = newPublicToken()
	}
	reg.CreatedAt, reg.UpdatedAt = now, now

	if err := u.repo.Save(ctx, reg); err != nil {
		return domainRegistration.Registration{}, err
	}
	u.writeAudit(ctx, audit.ActorPortal, "", "SUBMIT_REGISTRATION", "registration", reg.ID)
	u.appendEvent(ctx, reg.ID, domainRegistration.EventCreated, "Pendaftaran diterima",
		"Calon pelanggan mengirim formulir pendaftaran.", nil,
		map[string]any{"plan_id": reg.PlanID})
	return reg, nil
}

// Approve moves PENDING/SURVEY → APPROVED and queues the approval WA notice.
func (u *ManageRegistrationUseCase) Approve(ctx context.Context, id string, reviewerID uint, adminNotes string) (domainRegistration.Registration, error) {
	reg, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return domainRegistration.Registration{}, domainRegistration.ErrNotFound
	}
	if reg.Status != domainRegistration.StatusPending && reg.Status != domainRegistration.StatusSurvey {
		return domainRegistration.Registration{}, fmt.Errorf("%w: approve requires PENDING/SURVEY, got %s",
			domainRegistration.ErrInvalidTransition, reg.Status)
	}
	now := u.now()
	reg.Status = domainRegistration.StatusApproved
	reg.ReviewedBy = &reviewerID
	reg.ReviewedAt = &now
	reg.AdminNotes = adminNotes

	if err := u.repo.Save(ctx, reg); err != nil {
		return domainRegistration.Registration{}, err
	}
	u.writeAudit(ctx, audit.ActorUser, fmt.Sprint(reviewerID), "APPROVE_REGISTRATION", "registration", reg.ID)
	actorID := reviewerID
	u.appendEvent(ctx, reg.ID, domainRegistration.EventApproved, "Pendaftaran disetujui",
		adminNotes, &actorID, nil)
	u.queueTemplate(ctx, reg.TenantID, "REGISTRATION_APPROVED",
		"Halo {{full_name}}, pendaftaran paket {{plan_name}} Anda disetujui. Jadwal pemasangan: {{install_schedule}}.",
		map[string]string{
			"full_name":        reg.FullName,
			"plan_name":        reg.PlanID,
			"install_schedule": "akan diinformasikan",
		}, reg.Phone)
	return reg, nil
}

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
		return domainRegistration.Registration{}, err
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
		return domainRegistration.Registration{}, err
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
		return domainRegistration.Registration{}, err
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
		return domainRegistration.Registration{}, err
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

// Reject closes a signup with a reason (PENDING/SURVEY → REJECTED, sesuai
// state machine plan v2: APPROVED/ON_PROGRESS dibatalkan lewat Cancel).
func (u *ManageRegistrationUseCase) Reject(ctx context.Context, id, reason string, reviewerID uint) (domainRegistration.Registration, error) {
	reg, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return domainRegistration.Registration{}, domainRegistration.ErrNotFound
	}
	switch reg.Status {
	case domainRegistration.StatusPending, domainRegistration.StatusSurvey:
	default:
		return domainRegistration.Registration{}, fmt.Errorf("%w: reject requires PENDING/SURVEY, got %s",
			domainRegistration.ErrInvalidTransition, reg.Status)
	}
	now := u.now()
	reg.Status = domainRegistration.StatusRejected
	reg.RejectedAt = &now
	reg.RejectedReason = reason
	reg.ReviewedBy = &reviewerID
	reg.ReviewedAt = &now
	if err := u.repo.Save(ctx, reg); err != nil {
		return domainRegistration.Registration{}, err
	}
	u.writeAudit(ctx, audit.ActorUser, fmt.Sprint(reviewerID), "REJECT_REGISTRATION", "registration", reg.ID)
	actorID := reviewerID
	u.appendEvent(ctx, reg.ID, domainRegistration.EventRejected, "Pendaftaran ditolak",
		reason, &actorID, nil)
	u.queueTemplate(ctx, reg.TenantID, "REGISTRATION_REJECTED",
		"Halo {{full_name}}, pendaftaran Anda tidak dapat diproses. Alasan: {{reason}}. Hubungi kami untuk informasi lebih lanjut.",
		map[string]string{
			"full_name": reg.FullName,
			"reason":    reason,
		}, reg.Phone)
	return reg, nil
}

// Cancel withdraws a signup before activation.
func (u *ManageRegistrationUseCase) Cancel(ctx context.Context, id, reason string) (domainRegistration.Registration, error) {
	reg, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return domainRegistration.Registration{}, domainRegistration.ErrNotFound
	}
	switch reg.Status {
	case domainRegistration.StatusActive, domainRegistration.StatusCancelled, domainRegistration.StatusRejected:
		return domainRegistration.Registration{}, fmt.Errorf("%w: cannot cancel from %s", domainRegistration.ErrInvalidTransition, reg.Status)
	}
	now := u.now()
	reg.Status = domainRegistration.StatusCancelled
	reg.CancelledAt = &now
	reg.CancelReason = reason
	if err := u.repo.Save(ctx, reg); err != nil {
		return domainRegistration.Registration{}, err
	}
	u.writeAudit(ctx, audit.ActorUser, "", "CANCEL_REGISTRATION", "registration", reg.ID)
	u.appendEvent(ctx, reg.ID, domainRegistration.EventCancelled, "Pendaftaran dibatalkan",
		reason, nil, nil)
	u.queueTemplate(ctx, reg.TenantID, "REGISTRATION_CANCELLED",
		"Halo {{full_name}}, pendaftaran Anda telah dibatalkan. Alasan: {{reason}}.",
		map[string]string{
			"full_name": reg.FullName,
			"reason":    reason,
		}, reg.Phone)
	return reg, nil
}

// appendEvent menulis satu entri timeline PSB. Best-effort: event tidak
// tersedia / gagal tidak menggagalkan transisi utama.
func (u *ManageRegistrationUseCase) appendEvent(ctx context.Context, regID, eventType, title, description string, actorID *uint, metadata map[string]any) {
	if u.events == nil {
		return
	}
	var raw json.RawMessage
	if len(metadata) > 0 {
		if b, err := json.Marshal(metadata); err == nil {
			raw = b
		}
	}
	e := domainRegistration.RegistrationEvent{
		ID:             idgen.New("regevt"),
		RegistrationID: regID,
		EventType:      eventType,
		ActorID:        actorID,
		Title:          title,
		Description:    description,
		Metadata:       raw,
		CreatedAt:      u.now().UTC(),
	}
	if err := u.events.Append(ctx, e); err != nil {
		logger.WithComponent("RegistrationUC").WithError(err).Warn("append registration event failed")
	}
}

// ListEvents mengembalikan timeline perjalanan PSB satu pendaftaran.
func (u *ManageRegistrationUseCase) ListEvents(ctx context.Context, id string) ([]domainRegistration.RegistrationEvent, error) {
	if u.events == nil {
		return nil, nil
	}
	return u.events.ListByRegistration(ctx, id)
}

// TrackByPublicToken mencari pendaftaran lewat token publik (lacak status
// tanpa login) — halaman /register/status?token=...
func (u *ManageRegistrationUseCase) TrackByPublicToken(ctx context.Context, token string) (domainRegistration.Registration, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return domainRegistration.Registration{}, fmt.Errorf("%w: token is required", domainRegistration.ErrInvalidInput)
	}
	reg, err := u.repo.FindByPublicToken(ctx, token)
	if err != nil {
		return domainRegistration.Registration{}, domainRegistration.ErrNotFound
	}
	return reg, nil
}

// newPublicToken menghasilkan token lacak 32 hex char (128-bit).
func newPublicToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "tok" + idgen.Digits(16)
	}
	return hex.EncodeToString(b)
}

func (u *ManageRegistrationUseCase) mustGet(ctx context.Context, id, wantStatus string) (domainRegistration.Registration, error) {
	reg, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return domainRegistration.Registration{}, domainRegistration.ErrNotFound
	}
	if reg.Status != wantStatus {
		return domainRegistration.Registration{}, fmt.Errorf("%w: want %s, got %s", domainRegistration.ErrInvalidTransition, wantStatus, reg.Status)
	}
	return reg, nil
}

func (u *ManageRegistrationUseCase) writeAudit(ctx context.Context, actorType, actorID, action, entityType, entityID string) {
	if u.audit == nil {
		return
	}
	err := u.audit.Write(ctx, audit.AuditLog{
		TenantID:   domainTenant.DefaultID,
		ActorType:  actorType,
		ActorID:    actorID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
	})
	if err != nil {
		logger.WithComponent("RegistrationUC").WithError(err).Warn("audit log write failed")
	}
}

// queueTemplate mengantre notifikasi WA dari template DB; fallback dirender
// dengan variabel yang sama bila template belum tersedia.
func (u *ManageRegistrationUseCase) queueTemplate(ctx context.Context, tenantID, key, fallback string, vars map[string]string, phone string) {
	if u.notif == nil || phone == "" {
		return
	}
	content := fallback
	if tpl, err := u.notif.FindTemplateByKey(ctx, tenantID, key); err == nil && strings.TrimSpace(tpl.Content) != "" {
		content = tpl.Content
	}
	rep := make([]string, 0, len(vars)*2)
	for k, v := range vars {
		rep = append(rep, "{{"+k+"}}", v)
	}
	content = strings.NewReplacer(rep...).Replace(content)
	n := domainNotification.WANotification{
		ID:             idgen.New("wa"),
		TenantID:       tenantID,
		RecipientPhone: phone,
		MessageType:    key,
		MessageContent: content,
		Status:         domainNotification.StatusQueued,
		CreatedAt:      u.now(),
	}
	if err := u.notif.Queue(ctx, n); err != nil {
		logger.WithComponent("RegistrationUC").WithError(err).Warn("queue WA notification failed")
	}
}

// List retrieves registrations matching the given filter.
func (u *ManageRegistrationUseCase) List(ctx context.Context, filter port.RegistrationFilter) ([]domainRegistration.Registration, error) {
	if u.repo == nil {
		return nil, domainRegistration.ErrNotFound
	}
	list, err := u.repo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list registrations: %w", err)
	}
	return list, nil
}

// FindByID retrieves a registration by its unique identifier.
func (u *ManageRegistrationUseCase) FindByID(ctx context.Context, id string) (domainRegistration.Registration, error) {
	if u.repo == nil {
		return domainRegistration.Registration{}, domainRegistration.ErrNotFound
	}
	reg, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return domainRegistration.Registration{}, fmt.Errorf("find registration by id: %w", err)
	}
	return reg, nil
}

func actorStr(id *uint) string {
	if id == nil {
		return ""
	}
	return fmt.Sprint(*id)
}
