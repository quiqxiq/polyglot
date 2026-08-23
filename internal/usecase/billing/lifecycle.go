package billing

import (
	"context"
	"fmt"
	"strconv"
	"time"

	domainAudit "github.com/quixiq/polyglot/internal/domain/audit"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

// SubscriptionLifecycleUseCase mengelola transisi status langganan yang
// menyentuh router: ChangePlan, Suspend, Resume, Terminate, Activate.
// Router-first untuk operasi perubahan akun: DB hanya berubah setelah
// perintah router sukses (kecuali bila akun memang belum diprovisikan).
type SubscriptionLifecycleUseCase struct {
	subs    port.SubscriptionRepository
	plans   port.ServicePlanRepository
	manager port.RouterAccountManager
	audit   port.AuditLogWriter

	now func() time.Time
}

// NewSubscriptionLifecycleUseCase wires dependencies.
func NewSubscriptionLifecycleUseCase(
	subs port.SubscriptionRepository,
	plans port.ServicePlanRepository,
	manager port.RouterAccountManager,
	auditW port.AuditLogWriter,
) *SubscriptionLifecycleUseCase {
	return &SubscriptionLifecycleUseCase{subs: subs, plans: plans, manager: manager, audit: auditW, now: time.Now}
}

// Activate menugaskan device & memprovisikan akun sekarang (dipakai saat
// teknisi memasang / admin menugaskan router). Gagal router = error, status
// provisi PENDING agar worker mencoba ulang.
func (u *SubscriptionLifecycleUseCase) Activate(ctx context.Context, subID, deviceID string) (domainSubscription.Subscription, error) {
	if deviceID == "" {
		return domainSubscription.Subscription{}, fmt.Errorf("%w: device id is required", ErrValidation)
	}
	sub, err := u.mustGet(ctx, subID)
	if err != nil {
		return sub, err
	}
	pl, err := u.plans.FindByID(ctx, sub.PlanID)
	if err != nil {
		return sub, fmt.Errorf("plan %s: %w", sub.PlanID, err)
	}

	sub.DeviceID = &deviceID
	acct := port.SubscriberAccount{
		Username:  sub.RemoteUsername,
		Password:  sub.RemotePassword,
		Profile:   pl.Name,
		RateLimit: planRate(pl),
		Comment:   "polyglot:" + sub.ID,
	}
	if err := u.manager.Provision(ctx, deviceID, sub.ServiceType, acct); err != nil {
		sub.ProvisionStatus = domainSubscription.ProvisionPending
		if serr := u.subs.Save(ctx, sub); serr != nil {
			return sub, serr
		}
		return sub, fmt.Errorf("provision ke router: %w", err)
	}
	sub.ProvisionStatus = domainSubscription.ProvisionOK
	sub.RouterProfile = pl.Name
	if err := u.subs.Save(ctx, sub); err != nil {
		return sub, err
	}
	u.writeAudit(ctx, "", "ACTIVATE_SUBSCRIPTION", "subscription", sub.ID)
	return sub, nil
}

// ChangePlan memindahkan langganan ke paket baru: profil router di-update
// dulu (bila sudah provisioned), lalu DB.
func (u *SubscriptionLifecycleUseCase) ChangePlan(ctx context.Context, subID, newPlanID string) (domainSubscription.Subscription, error) {
	sub, err := u.mustGet(ctx, subID)
	if err != nil {
		return sub, err
	}
	if newPlanID == sub.PlanID {
		return sub, nil // no-op
	}
	pl, err := u.plans.FindByID(ctx, newPlanID)
	if err != nil {
		return sub, fmt.Errorf("plan tujuan %s: %w", newPlanID, err)
	}
	if !pl.IsActive {
		return sub, fmt.Errorf("%w: plan tujuan tidak aktif", ErrValidation)
	}

	provisioned := sub.ProvisionStatus == domainSubscription.ProvisionOK && sub.DeviceID != nil && *sub.DeviceID != ""
	if provisioned {
		if err := u.manager.UpdateAccount(ctx, *sub.DeviceID, sub.ServiceType, sub.RemoteUsername, pl.Name); err != nil {
			return sub, fmt.Errorf("update profil router: %w", err)
		}
	}

	sub.PlanID = newPlanID
	sub.CustomPrice = nil // kembali ke harga paket baru
	sub.RouterProfile = pl.Name
	if err := u.subs.Save(ctx, sub); err != nil {
		return sub, err
	}
	u.writeAudit(ctx, "", "CHANGE_PLAN", "subscription", sub.ID)
	return sub, nil
}

// Suspend menonaktifkan sementara (cuti/penghentian bertahap): disable akun
// di router + status SUSPENDED. Boleh dari ACTIVE atau ISOLATED.
func (u *SubscriptionLifecycleUseCase) Suspend(ctx context.Context, subID, reason string) (domainSubscription.Subscription, error) {
	sub, err := u.mustGetAny(ctx, subID, domainSubscription.StatusActive, domainSubscription.StatusIsolated)
	if err != nil {
		return sub, err
	}
	if provisioned(sub) {
		if err := u.manager.Suspend(ctx, derefDevice(sub.DeviceID), sub.ServiceType, sub.RemoteUsername); err != nil {
			return sub, fmt.Errorf("suspend akun router: %w", err)
		}
	}
	sub.Status = domainSubscription.StatusSuspended
	sub.Notes = appendNote(sub.Notes, "SUSPENDED: "+reason)
	if err := u.subs.Save(ctx, sub); err != nil {
		return sub, err
	}
	u.writeAudit(ctx, "", "SUSPEND_SUBSCRIPTION", "subscription", sub.ID)
	return sub, nil
}

// Resume mengaktifkan kembali pelanggan SUSPENDED: restore profil paket.
func (u *SubscriptionLifecycleUseCase) Resume(ctx context.Context, subID string) (domainSubscription.Subscription, error) {
	sub, err := u.mustGetAny(ctx, subID, domainSubscription.StatusSuspended)
	if err != nil {
		return sub, err
	}
	if provisioned(sub) {
		if err := u.manager.Restore(ctx, derefDevice(sub.DeviceID), sub.ServiceType,
			sub.RemoteUsername, u.normalProfile(ctx, sub), ""); err != nil {
			return sub, fmt.Errorf("resume akun router: %w", err)
		}
	}
	sub.Status = domainSubscription.StatusActive
	sub.EndDate = nil
	if err := u.subs.Save(ctx, sub); err != nil {
		return sub, err
	}
	u.writeAudit(ctx, "", "RESUME_SUBSCRIPTION", "subscription", sub.ID)
	return sub, nil
}

// Terminate menghapus langganan permanen: hapus akun router, status
// TERMINATED, isi end_date.
func (u *SubscriptionLifecycleUseCase) Terminate(ctx context.Context, subID, reason string) (domainSubscription.Subscription, error) {
	sub, err := u.mustGetAny(ctx, subID,
		domainSubscription.StatusActive, domainSubscription.StatusIsolated, domainSubscription.StatusSuspended)
	if err != nil {
		return sub, err
	}
	if provisioned(sub) {
		if err := u.manager.Terminate(ctx, derefDevice(sub.DeviceID), sub.ServiceType, sub.RemoteUsername); err != nil {
			return sub, fmt.Errorf("terminate akun router: %w", err)
		}
	}
	now := u.now()
	sub.Status = domainSubscription.StatusTerminated
	sub.EndDate = &now
	sub.Notes = appendNote(sub.Notes, "TERMINATED: "+reason)
	if err := u.subs.Save(ctx, sub); err != nil {
		return sub, err
	}
	u.writeAudit(ctx, "", "TERMINATE_SUBSCRIPTION", "subscription", sub.ID)
	return sub, nil
}

// ─── helpers ────────────────────────────────────────────────────────────

func (u *SubscriptionLifecycleUseCase) mustGet(ctx context.Context, id string) (domainSubscription.Subscription, error) {
	sub, err := u.subs.FindByID(ctx, id)
	if err != nil {
		return domainSubscription.Subscription{}, ErrNotFoundBilling
	}
	return sub, nil
}

func (u *SubscriptionLifecycleUseCase) mustGetAny(ctx context.Context, id string, allowed ...string) (domainSubscription.Subscription, error) {
	sub, err := u.mustGet(ctx, id)
	if err != nil {
		return sub, err
	}
	for _, a := range allowed {
		if sub.Status == a {
			return sub, nil
		}
	}
	return domainSubscription.Subscription{}, fmt.Errorf("%w: status %s tidak diizinkan untuk operasi ini",
		ErrInvalidTransitionBilling, sub.Status)
}

func (u *SubscriptionLifecycleUseCase) normalProfile(ctx context.Context, sub domainSubscription.Subscription) string {
	if sub.RouterProfile != "" {
		return sub.RouterProfile
	}
	if pl, err := u.plans.FindByID(ctx, sub.PlanID); err == nil && pl.Name != "" {
		return pl.Name
	}
	return sub.PlanID
}

func (u *SubscriptionLifecycleUseCase) writeAudit(ctx context.Context, actorID, action, entityType, entityID string) {
	if u.audit == nil {
		return
	}
	err := u.audit.Write(ctx, domainAudit.AuditLog{
		TenantID:   "tenant-default",
		ActorType:  domainAudit.ActorUser,
		ActorID:    actorID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
	})
	if err != nil {
		logger.WithComponent("LifecycleUC").WithError(err).Warn("audit log write failed")
	}
}

func provisioned(sub domainSubscription.Subscription) bool {
	return sub.ProvisionStatus == domainSubscription.ProvisionOK &&
		sub.DeviceID != nil && *sub.DeviceID != ""
}

func derefDevice(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func appendNote(existing, addition string) string {
	if existing == "" {
		return addition
	}
	return existing + " | " + addition
}

// planRate formats the MikroTik rate limit from plan bandwidth columns.
func planRate(pl domainPlan.ServicePlan) string {
	dl, ul := planRateSide(pl.BandwidthDownloadKbps), planRateSide(pl.BandwidthUploadKbps)
	if dl == "" && ul == "" {
		return ""
	}
	return dl + "/" + ul
}

func planRateSide(kbps int) string {
	if kbps <= 0 {
		return ""
	}
	if kbps >= 1000 {
		return strconv.Itoa((kbps+500)/1000) + "M"
	}
	return strconv.Itoa(kbps) + "k"
}
