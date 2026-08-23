package registration

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/audit"
	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/logger"
)

// Kesalahan alur pendaftaran.
var (
	ErrInvalidTransition = errors.New("invalid registration status transition")
	ErrNotFound          = errors.New("registration not found")
	ErrValidation        = errors.New("validation failed")
)

var phoneRe = regexp.MustCompile(`^(\+?62|0)8\d{7,12}$`)

// ManageRegistrationUseCase orchestrates the signup flow:
// PENDING → APPROVED → INSTALLED → ACTIVE (atau REJECTED/CANCELLED).
type ManageRegistrationUseCase struct {
	repo  port.RegistrationRepository
	notif port.NotificationRepository
	audit port.AuditLogWriter

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

// Submit validates and stores a new signup (ringkas tanpa NIK/KTP).
func (u *ManageRegistrationUseCase) Submit(ctx context.Context, reg domainRegistration.Registration) (domainRegistration.Registration, error) {
	if strings.TrimSpace(reg.FullName) == "" || strings.TrimSpace(reg.Address) == "" {
		return domainRegistration.Registration{}, fmt.Errorf("%w: full_name and address are required", ErrValidation)
	}
	if !phoneRe.MatchString(reg.Phone) {
		return domainRegistration.Registration{}, fmt.Errorf("%w: phone must be an Indonesian mobile number", ErrValidation)
	}
	if reg.PlanID == "" {
		return domainRegistration.Registration{}, fmt.Errorf("%w: plan_id is required", ErrValidation)
	}

	now := u.now()
	reg.ID = idgen.New("reg")
	reg.Status = domainRegistration.StatusPending
	reg.RegistrationNo = fmt.Sprintf("REG-%s-%04d", now.Format("200601"), now.UnixNano()%10000)
	reg.CreatedAt, reg.UpdatedAt = now, now

	if err := u.repo.Save(ctx, reg); err != nil {
		return domainRegistration.Registration{}, err
	}
	u.writeAudit(ctx, audit.ActorPortal, "", "SUBMIT_REGISTRATION", "registration", reg.ID)
	return reg, nil
}

// Approve moves PENDING → APPROVED and queues the approval WA notice.
func (u *ManageRegistrationUseCase) Approve(ctx context.Context, id string, reviewerID uint, adminNotes string) (domainRegistration.Registration, error) {
	reg, err := u.mustGet(ctx, id, domainRegistration.StatusPending)
	if err != nil {
		return domainRegistration.Registration{}, err
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
	u.queueTemplate(ctx, reg.TenantID, "REGISTRATION_APPROVED", map[string]string{
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
		return domainRegistration.Registration{}, fmt.Errorf("%w: install date is required", ErrValidation)
	}
	reg.ScheduledInstallDate = &date
	reg.ScheduledInstallTime = installTime
	if techID != nil {
		reg.AssignedTechnicianID = techID
	}
	if err := u.repo.Save(ctx, reg); err != nil {
		return domainRegistration.Registration{}, err
	}
	u.queueTemplate(ctx, reg.TenantID, "INSTALLATION_SCHEDULED", map[string]string{
		"full_name":        reg.FullName,
		"install_schedule": date.Format("02 Jan 2006"),
		"address":          reg.Address,
	}, reg.Phone)
	return reg, nil
}

// MarkInstalled records technician completion: APPROVED → INSTALLED.
func (u *ManageRegistrationUseCase) MarkInstalled(ctx context.Context, id string, installerID *uint, techNotes string) (domainRegistration.Registration, error) {
	reg, err := u.mustGet(ctx, id, domainRegistration.StatusApproved)
	if err != nil {
		return domainRegistration.Registration{}, err
	}
	now := u.now()
	reg.Status = domainRegistration.StatusInstalled
	reg.InstalledAt = &now
	reg.TechnicianNotes = techNotes
	if installerID != nil {
		reg.AssignedTechnicianID = installerID
	}
	if err := u.repo.Save(ctx, reg); err != nil {
		return domainRegistration.Registration{}, err
	}
	u.writeAudit(ctx, audit.ActorUser, actorStr(installerID), "MARK_INSTALLED", "registration", reg.ID)
	return reg, nil
}

// Reject closes a pending signup with a reason.
func (u *ManageRegistrationUseCase) Reject(ctx context.Context, id, reason string, reviewerID uint) (domainRegistration.Registration, error) {
	reg, err := u.mustGet(ctx, id, domainRegistration.StatusPending)
	if err != nil {
		return domainRegistration.Registration{}, err
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
	return reg, nil
}

// Cancel withdraws a signup before activation.
func (u *ManageRegistrationUseCase) Cancel(ctx context.Context, id, reason string) (domainRegistration.Registration, error) {
	reg, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return domainRegistration.Registration{}, ErrNotFound
	}
	switch reg.Status {
	case domainRegistration.StatusActive, domainRegistration.StatusCancelled, domainRegistration.StatusRejected:
		return domainRegistration.Registration{}, fmt.Errorf("%w: cannot cancel from %s", ErrInvalidTransition, reg.Status)
	}
	now := u.now()
	reg.Status = domainRegistration.StatusCancelled
	reg.CancelledAt = &now
	reg.CancelReason = reason
	if err := u.repo.Save(ctx, reg); err != nil {
		return domainRegistration.Registration{}, err
	}
	u.writeAudit(ctx, audit.ActorUser, "", "CANCEL_REGISTRATION", "registration", reg.ID)
	return reg, nil
}

func (u *ManageRegistrationUseCase) mustGet(ctx context.Context, id, wantStatus string) (domainRegistration.Registration, error) {
	reg, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return domainRegistration.Registration{}, ErrNotFound
	}
	if reg.Status != wantStatus {
		return domainRegistration.Registration{}, fmt.Errorf("%w: want %s, got %s", ErrInvalidTransition, wantStatus, reg.Status)
	}
	return reg, nil
}

func (u *ManageRegistrationUseCase) writeAudit(ctx context.Context, actorType, actorID, action, entityType, entityID string) {
	if u.audit == nil {
		return
	}
	err := u.audit.Write(ctx, audit.AuditLog{
		TenantID:   "tenant-default",
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

func (u *ManageRegistrationUseCase) queueTemplate(ctx context.Context, tenantID, key string, vars map[string]string, phone string) {
	if u.notif == nil || phone == "" {
		return
	}
	tpl, err := u.notif.FindTemplateByKey(ctx, tenantID, key)
	content := tpl.Content
	if err != nil {
		content = key // fallback teks polos bila template hilang
	} else {
		rep := make([]string, 0, len(vars)*2)
		for k, v := range vars {
			rep = append(rep, "{{"+k+"}}", v)
		}
		content = strings.NewReplacer(rep...).Replace(tpl.Content)
	}
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

func actorStr(id *uint) string {
	if id == nil {
		return ""
	}
	return fmt.Sprint(*id)
}
