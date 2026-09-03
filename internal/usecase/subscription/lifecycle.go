package subscription

import (
	"context"
	"fmt"
	"time"

	domainAudit "github.com/quixiq/polyglot/internal/domain/audit"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	planUC "github.com/quixiq/polyglot/internal/usecase/plan"
	"github.com/quixiq/polyglot/pkg/logger"
)

// LifecycleUseCase manages subscription status transitions that interact with router.
type LifecycleUseCase struct {
	subs    port.SubscriptionRepository
	plans   port.ServicePlanRepository
	manager port.RouterAccountManager
	audit   port.AuditLogWriter

	now func() time.Time
}

// NewLifecycleUseCase wires dependencies.
func NewLifecycleUseCase(
	subs port.SubscriptionRepository,
	plans port.ServicePlanRepository,
	manager port.RouterAccountManager,
	auditW port.AuditLogWriter,
) *LifecycleUseCase {
	return &LifecycleUseCase{
		subs:    subs,
		plans:   plans,
		manager: manager,
		audit:   auditW,
		now:     time.Now,
	}
}

// Activate assigns a device & provisions the account on the router.
func (u *LifecycleUseCase) Activate(ctx context.Context, subID, deviceID string) (domainSub.Subscription, error) {
	sub, err := u.mustGet(ctx, subID)
	if err != nil {
		return sub, err
	}
	if deviceID == "" && sub.DeviceID != nil && *sub.DeviceID != "" {
		deviceID = *sub.DeviceID
	}
	if deviceID == "" {
		return domainSub.Subscription{}, fmt.Errorf("%w: device id is required", domainSub.ErrInvalidInput)
	}
	pl, err := u.plans.FindByID(ctx, sub.PlanID)
	if err != nil {
		return sub, fmt.Errorf("plan %s: %w", sub.PlanID, err)
	}

	sub.DeviceID = &deviceID
	if u.manager != nil {
		var perr error
		if isHotspot(sub.ServiceType) {
			hotSpec := planUC.BuildHotspotProvisionSpec(sub, pl)
			perr = u.manager.ProvisionHotspot(ctx, deviceID, hotSpec)
		} else if isDedicated(sub.ServiceType) {
			dedSpec := planUC.BuildDedicatedProvisionSpec(sub, pl)
			perr = u.manager.ProvisionDedicated(ctx, deviceID, dedSpec)
		} else {
			pppSpec := planUC.BuildPPPoEProvisionSpec(sub, pl)
			perr = u.manager.ProvisionPPPoE(ctx, deviceID, pppSpec)
		}
		if perr != nil {
			sub.ProvisionStatus = domainSub.ProvisionPending
			if serr := u.subs.Save(ctx, sub); serr != nil {
				return sub, fmt.Errorf("save subscription: %w", serr)
			}
			return sub, fmt.Errorf("provision ke router: %w", perr)
		}
	}
	sub.ProvisionStatus = domainSub.ProvisionOK
	sub.Status = domainSub.StatusActive
	sub.RouterProfile = pl.Name
	if err := u.subs.Save(ctx, sub); err != nil {
		return sub, fmt.Errorf("save subscription: %w", err)
	}
	u.writeAudit(ctx, "", "ACTIVATE_SUBSCRIPTION", "subscription", sub.ID)
	return sub, nil
}

// ChangePlan moves a subscription to a new service plan.
func (u *LifecycleUseCase) ChangePlan(ctx context.Context, subID, newPlanID string) (domainSub.Subscription, error) {
	sub, err := u.mustGet(ctx, subID)
	if err != nil {
		return sub, err
	}
	if newPlanID == sub.PlanID {
		return sub, nil // no-op
	}
	pl, err := u.plans.FindByID(ctx, newPlanID)
	if err != nil {
		return sub, fmt.Errorf("target plan %s: %w", newPlanID, err)
	}
	if !pl.IsActive {
		return sub, fmt.Errorf("%w: target plan is inactive", domainSub.ErrInvalidInput)
	}

	provisioned := sub.ProvisionStatus == domainSub.ProvisionOK && sub.DeviceID != nil && *sub.DeviceID != ""
	if provisioned {
		if err := u.manager.EnsureProfile(ctx, *sub.DeviceID, sub.ServiceType, pl.Name, pl.RateLimit()); err != nil {
			return sub, fmt.Errorf("ensure profil router: %w", err)
		}
		if err := u.manager.UpdateAccount(ctx, *sub.DeviceID, sub.ServiceType, sub.RemoteUsername, pl.Name); err != nil {
			return sub, fmt.Errorf("update profil router: %w", err)
		}
	}

	sub.PlanID = newPlanID
	sub.CustomPrice = nil
	sub.RouterProfile = pl.Name
	if err := u.subs.Save(ctx, sub); err != nil {
		return sub, fmt.Errorf("save subscription: %w", err)
	}
	u.writeAudit(ctx, "", "CHANGE_PLAN", "subscription", sub.ID)
	return sub, nil
}

// Suspend disables the subscriber temporarily.
func (u *LifecycleUseCase) Suspend(ctx context.Context, subID, reason string) (domainSub.Subscription, error) {
	sub, err := u.mustGetAny(ctx, subID, domainSub.StatusActive, domainSub.StatusIsolated)
	if err != nil {
		return sub, err
	}
	if provisioned(sub) {
		if err := u.manager.Suspend(ctx, derefDevice(sub.DeviceID), sub.ServiceType, sub.RemoteUsername); err != nil {
			return sub, fmt.Errorf("suspend akun router: %w", err)
		}
	}
	sub.Status = domainSub.StatusSuspended
	sub.Notes = appendNote(sub.Notes, "SUSPENDED: "+reason)
	if err := u.subs.Save(ctx, sub); err != nil {
		return sub, fmt.Errorf("save subscription: %w", err)
	}
	u.writeAudit(ctx, "", "SUSPEND_SUBSCRIPTION", "subscription", sub.ID)
	return sub, nil
}

// Resume re-enables a suspended subscriber.
func (u *LifecycleUseCase) Resume(ctx context.Context, subID string) (domainSub.Subscription, error) {
	sub, err := u.mustGetAny(ctx, subID, domainSub.StatusSuspended)
	if err != nil {
		return sub, err
	}
	if provisioned(sub) {
		if err := u.manager.Restore(ctx, derefDevice(sub.DeviceID), sub.ServiceType,
			sub.RemoteUsername, u.normalProfile(ctx, sub), ""); err != nil {
			return sub, fmt.Errorf("resume akun router: %w", err)
		}
	}
	sub.Status = domainSub.StatusActive
	sub.EndDate = nil
	if err := u.subs.Save(ctx, sub); err != nil {
		return sub, fmt.Errorf("save subscription: %w", err)
	}
	u.writeAudit(ctx, "", "RESUME_SUBSCRIPTION", "subscription", sub.ID)
	return sub, nil
}

// Terminate terminates a subscriber permanently.
func (u *LifecycleUseCase) Terminate(ctx context.Context, subID, reason string) (domainSub.Subscription, error) {
	sub, err := u.mustGetAny(ctx, subID,
		domainSub.StatusActive, domainSub.StatusIsolated, domainSub.StatusSuspended)
	if err != nil {
		return sub, err
	}
	if provisioned(sub) {
		if err := u.manager.Terminate(ctx, derefDevice(sub.DeviceID), sub.ServiceType, sub.RemoteUsername); err != nil {
			return sub, fmt.Errorf("terminate akun router: %w", err)
		}
	}
	now := u.now()
	sub.Status = domainSub.StatusTerminated
	sub.EndDate = &now
	sub.Notes = appendNote(sub.Notes, "TERMINATED: "+reason)
	if err := u.subs.Save(ctx, sub); err != nil {
		return sub, fmt.Errorf("save subscription: %w", err)
	}
	u.writeAudit(ctx, "", "TERMINATE_SUBSCRIPTION", "subscription", sub.ID)
	return sub, nil
}

// Isolate isolates a subscriber manually.
func (u *LifecycleUseCase) Isolate(ctx context.Context, subID, reason string) (domainSub.Subscription, error) {
	sub, err := u.mustGetAny(ctx, subID, domainSub.StatusActive)
	if err != nil {
		return sub, err
	}
	if provisioned(sub) {
		opt := port.IsolationOptions{
			IsolirProfile: "ISOLIR",
			AddressList:   "ISOLIR_USERS",
		}
		if err := u.manager.Isolate(ctx, derefDevice(sub.DeviceID), sub.ServiceType, sub.RemoteUsername, opt); err != nil {
			return sub, fmt.Errorf("isolate akun router: %w", err)
		}
	}
	sub.Status = domainSub.StatusIsolated
	if reason != "" {
		sub.Notes = appendNote(sub.Notes, "MANUAL_ISOLATED: "+reason)
	}
	if err := u.subs.Save(ctx, sub); err != nil {
		return sub, fmt.Errorf("save subscription: %w", err)
	}
	u.writeAudit(ctx, "", "ISOLATE_SUBSCRIPTION", "subscription", sub.ID)
	return sub, nil
}

// Restore restores an isolated subscriber back to active.
func (u *LifecycleUseCase) Restore(ctx context.Context, subID string) (domainSub.Subscription, error) {
	sub, err := u.mustGetAny(ctx, subID, domainSub.StatusIsolated)
	if err != nil {
		return sub, err
	}
	if provisioned(sub) {
		normalProfile := u.normalProfile(ctx, sub)
		if err := u.manager.Restore(ctx, derefDevice(sub.DeviceID), sub.ServiceType,
			sub.RemoteUsername, normalProfile, "ISOLIR_USERS"); err != nil {
			return sub, fmt.Errorf("restore akun router: %w", err)
		}
	}
	sub.Status = domainSub.StatusActive
	if err := u.subs.Save(ctx, sub); err != nil {
		return sub, fmt.Errorf("save subscription: %w", err)
	}
	u.writeAudit(ctx, "", "RESTORE_SUBSCRIPTION", "subscription", sub.ID)
	return sub, nil
}

func (u *LifecycleUseCase) mustGet(ctx context.Context, id string) (domainSub.Subscription, error) {
	sub, err := u.subs.FindByID(ctx, id)
	if err != nil {
		return domainSub.Subscription{}, domainSub.ErrNotFound
	}
	return sub, nil
}

func (u *LifecycleUseCase) mustGetAny(ctx context.Context, id string, allowed ...string) (domainSub.Subscription, error) {
	sub, err := u.mustGet(ctx, id)
	if err != nil {
		return sub, err
	}
	for _, a := range allowed {
		if sub.Status == a {
			return sub, nil
		}
	}
	return domainSub.Subscription{}, fmt.Errorf("%w: status %s is not allowed for this operation",
		domainSub.ErrInvalidTransition, sub.Status)
}

func (u *LifecycleUseCase) normalProfile(ctx context.Context, sub domainSub.Subscription) string {
	if sub.RouterProfile != "" {
		return sub.RouterProfile
	}
	if pl, err := u.plans.FindByID(ctx, sub.PlanID); err == nil && pl.Name != "" {
		return pl.Name
	}
	return sub.PlanID
}

func (u *LifecycleUseCase) writeAudit(ctx context.Context, actorID, action, entityType, entityID string) {
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

func appendNote(existing, addition string) string {
	if existing == "" {
		return addition
	}
	return existing + " | " + addition
}
