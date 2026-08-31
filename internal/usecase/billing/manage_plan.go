package billing

import (
	"context"
	"fmt"
	"strings"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// PlanUseCase orchestrates ISP service plan CRUD
// (tabel service_plans, DATABASE-SCHEMA-ISP.md §2.2) and synchronization
// to MikroTik BRAS router profiles.
type PlanUseCase struct {
	plans  port.ServicePlanRepository
	subs   port.SubscriptionRepository
	router port.RouterAccountManager
}

// NewPlanUseCase wires dependencies.
func NewPlanUseCase(plans port.ServicePlanRepository, subs port.SubscriptionRepository, router port.RouterAccountManager) *PlanUseCase {
	return &PlanUseCase{plans: plans, subs: subs, router: router}
}

// Create validates, persists a new service plan, and optionally syncs it to a target router.
func (u *PlanUseCase) Create(ctx context.Context, p domainPlan.ServicePlan, deviceID string) (domainPlan.ServicePlan, error) {
	if err := validatePlan(p); err != nil {
		return domainPlan.ServicePlan{}, err
	}
	if p.ID == "" {
		p.ID = idgen.New("plan")
	}
	if p.TenantID == "" {
		p.TenantID = "tenant-default"
	}
	if p.ParentQueue == "" {
		p.ParentQueue = "none"
	}
	if p.ServiceType == domainPlan.TypeHotspot {
		if p.SharedUsers <= 0 {
			if p.Hotspot != nil && p.Hotspot.SharedUsers > 0 {
				p.SharedUsers = p.Hotspot.SharedUsers
			} else {
				p.SharedUsers = 1
			}
		}
	}
	p.IsActive = true

	if _, err := u.plans.FindByName(ctx, p.TenantID, p.Name); err == nil {
		return domainPlan.ServicePlan{}, fmt.Errorf("%w: plan name %q already exists", domainBilling.ErrInvalidInput, p.Name)
	}
	if err := u.plans.Save(ctx, p); err != nil {
		return domainPlan.ServicePlan{}, err
	}

	// Sync profile to MikroTik router if deviceID is provided
	if deviceID != "" && u.router != nil {
		if err := u.router.SyncPlanProfile(ctx, deviceID, p); err != nil {
			return p, fmt.Errorf("plan saved to database but failed to sync to router: %w", err)
		}
	}

	return p, nil
}

// Update replaces the stored plan and optionally syncs changes to the target router.
func (u *PlanUseCase) Update(ctx context.Context, p domainPlan.ServicePlan, deviceID string) (domainPlan.ServicePlan, error) {
	if p.ID == "" {
		return domainPlan.ServicePlan{}, fmt.Errorf("%w: plan id is required", domainBilling.ErrInvalidInput)
	}
	old, err := u.plans.FindByID(ctx, p.ID)
	if err != nil {
		return domainPlan.ServicePlan{}, fmt.Errorf("plan %s not found: %w", p.ID, domainBilling.ErrNotFound)
	}
	if err := validatePlan(p); err != nil {
		return domainPlan.ServicePlan{}, err
	}
	p.CreatedAt = old.CreatedAt
	if p.ServiceType == domainPlan.TypeHotspot && p.SharedUsers <= 0 {
		if p.Hotspot != nil && p.Hotspot.SharedUsers > 0 {
			p.SharedUsers = p.Hotspot.SharedUsers
		} else {
			p.SharedUsers = 1
		}
	}
	if err := u.plans.Save(ctx, p); err != nil {
		return domainPlan.ServicePlan{}, err
	}

	// Sync profile to MikroTik router if deviceID is provided
	if deviceID != "" && u.router != nil {
		if err := u.router.SyncPlanProfile(ctx, deviceID, p); err != nil {
			return p, fmt.Errorf("plan updated in database but failed to sync to router: %w", err)
		}
	}

	return p, nil
}

// Delete removes a plan unless active subscriptions still reference it, and optionally deletes from router.
func (u *PlanUseCase) Delete(ctx context.Context, id string, deviceID string) error {
	inUse, err := u.subs.HasActiveForPlan(ctx, id)
	if err != nil {
		return err
	}
	if inUse {
		return fmt.Errorf("%w: %s; deactivate instead of delete", domainBilling.ErrPlanInUse, id)
	}
	old, err := u.plans.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("plan %s not found: %w", id, domainBilling.ErrNotFound)
	}

	if err := u.plans.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete plan: %w", err)
	}

	// Remove profile from MikroTik router if deviceID is provided
	if deviceID != "" && u.router != nil {
		_ = u.router.DeletePlanProfile(ctx, deviceID, old.ServiceType, old.Name)
	}

	return nil
}

// Get returns one plan by ID.
func (u *PlanUseCase) Get(ctx context.Context, id string) (domainPlan.ServicePlan, error) {
	return u.plans.FindByID(ctx, id)
}

// List returns plans optionally filtered to active ones.
func (u *PlanUseCase) List(ctx context.Context, activeOnly bool) ([]domainPlan.ServicePlan, error) {
	return u.plans.List(ctx, activeOnly)
}

func validatePlan(p domainPlan.ServicePlan) error {
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("%w: name is required", domainBilling.ErrInvalidInput)
	}
	switch strings.ToUpper(p.ServiceType) {
	case domainPlan.TypePPPoE, domainPlan.TypeHotspot, domainPlan.TypeDedicated:
	default:
		return fmt.Errorf("%w: service_type must be PPPOE|HOTSPOT|DEDICATED", domainBilling.ErrInvalidInput)
	}
	if p.BandwidthDownloadKbps <= 0 || p.BandwidthUploadKbps <= 0 {
		return fmt.Errorf("%w: bandwidth must be positive", domainBilling.ErrInvalidInput)
	}
	if p.Price < 0 {
		return fmt.Errorf("%w: price cannot be negative", domainBilling.ErrInvalidInput)
	}
	return nil
}
