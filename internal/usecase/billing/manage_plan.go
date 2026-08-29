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
// (tabel service_plans, DATABASE-SCHEMA-ISP.md §2.2).
type PlanUseCase struct {
	plans port.ServicePlanRepository
	subs  port.SubscriptionRepository
}

// NewPlanUseCase wires dependencies.
func NewPlanUseCase(plans port.ServicePlanRepository, subs port.SubscriptionRepository) *PlanUseCase {
	return &PlanUseCase{plans: plans, subs: subs}
}

// Create validates and persists a new service plan with safe defaults.
func (u *PlanUseCase) Create(ctx context.Context, p domainPlan.ServicePlan) (domainPlan.ServicePlan, error) {
	if err := validatePlan(p); err != nil {
		return domainPlan.ServicePlan{}, err
	}
	if p.ID == "" {
		p.ID = idgen.New("plan")
	}
	if p.TenantID == "" {
		p.TenantID = "tenant-default"
	}
	if p.Validity == "" {
		p.Validity = "30d"
	}
	if p.ValidityMode == "" {
		p.ValidityMode = domainPlan.ValidityCalendar
	}
	if p.ExpireMode == "" {
		p.ExpireMode = domainPlan.ExpireNotFiltered
	}
	if p.ParentQueue == "" {
		p.ParentQueue = "none"
	}
	if p.SimultaneousUse <= 0 {
		p.SimultaneousUse = 1
	}
	if p.SharedUsers <= 0 {
		p.SharedUsers = 1
	}
	p.IsActive = true

	if _, err := u.plans.FindByName(ctx, p.TenantID, p.Name); err == nil {
		return domainPlan.ServicePlan{}, fmt.Errorf("%w: plan name %q already exists", domainBilling.ErrInvalidInput, p.Name)
	}
	if err := u.plans.Save(ctx, p); err != nil {
		return domainPlan.ServicePlan{}, err
	}
	return p, nil
}

// Update replaces the stored plan after validation.
func (u *PlanUseCase) Update(ctx context.Context, p domainPlan.ServicePlan) (domainPlan.ServicePlan, error) {
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
	if err := u.plans.Save(ctx, p); err != nil {
		return domainPlan.ServicePlan{}, err
	}
	return p, nil
}

// Delete removes a plan unless active subscriptions still reference it.
func (u *PlanUseCase) Delete(ctx context.Context, id string) error {
	inUse, err := u.subs.HasActiveForPlan(ctx, id)
	if err != nil {
		return err
	}
	if inUse {
		return fmt.Errorf("%w: %s; deactivate instead of delete", domainBilling.ErrPlanInUse, id)
	}
	return u.plans.Delete(ctx, id)
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
