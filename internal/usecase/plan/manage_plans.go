package plan

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	"github.com/quixiq/polyglot/pkg/logger"
)

// planRepository narrows the persistence surface the usecase needs.
type planRepository interface {
	Save(ctx context.Context, p domainPlan.Plan) error
	FindByID(ctx context.Context, id string) (domainPlan.Plan, error)
	List(ctx context.Context, serviceType string, activeOnly bool, limit int) ([]domainPlan.Plan, error)
	Delete(ctx context.Context, id string) error
}

// ManagePlansUseCase handles CRUD for master paket layanan. Plans are pure
// DB entities — no device I/O happens here; provisioning reads them later.
type ManagePlansUseCase struct {
	repo planRepository
}

func NewManagePlansUseCase(repo planRepository) *ManagePlansUseCase {
	return &ManagePlansUseCase{repo: repo}
}

// Create validates and persists a new plan.
func (uc *ManagePlansUseCase) Create(ctx context.Context, p domainPlan.Plan) (domainPlan.Plan, error) {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	if p.TenantID == "" {
		p.TenantID = "tenant-default"
	}
	if p.SharedUsers <= 0 {
		p.SharedUsers = 1
	}
	if err := p.Validate(); err != nil {
		return domainPlan.Plan{}, err
	}

	existing, err := uc.repo.List(ctx, p.ServiceType, false, 0)
	if err != nil {
		return domainPlan.Plan{}, fmt.Errorf("list plans: %w", err)
	}
	for _, e := range existing {
		if e.Name == p.Name {
			return domainPlan.Plan{}, domainPlan.ErrAlreadyExists
		}
	}

	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	if err := uc.repo.Save(ctx, p); err != nil {
		return domainPlan.Plan{}, fmt.Errorf("save plan: %w", err)
	}
	logger.WithComponent("ManagePlans").WithFields(map[string]any{
		"plan_id": p.ID, "name": p.Name, "service_type": p.ServiceType,
	}).Info("plan created")
	return p, nil
}

func (uc *ManagePlansUseCase) Get(ctx context.Context, id string) (domainPlan.Plan, error) {
	if id == "" {
		return domainPlan.Plan{}, domainPlan.ErrNotFound
	}
	return uc.repo.FindByID(ctx, id)
}

// List returns plans ordered by name; empty serviceType means all types.
func (uc *ManagePlansUseCase) List(ctx context.Context, serviceType string, activeOnly bool, limit int) ([]domainPlan.Plan, error) {
	if serviceType != "" && serviceType != domainPlan.ServiceTypePPPoE && serviceType != domainPlan.ServiceTypeHotspot {
		return nil, domainPlan.ErrInvalidServiceType
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return uc.repo.List(ctx, serviceType, activeOnly, limit)
}

// Update overwrites the stored plan identified by p.ID.
func (uc *ManagePlansUseCase) Update(ctx context.Context, p domainPlan.Plan) (domainPlan.Plan, error) {
	if p.ID == "" {
		return domainPlan.Plan{}, domainPlan.ErrNotFound
	}
	current, err := uc.repo.FindByID(ctx, p.ID)
	if err != nil {
		return domainPlan.Plan{}, err
	}
	if err := p.Validate(); err != nil {
		return domainPlan.Plan{}, err
	}
	if current.Name != p.Name {
		existing, err := uc.repo.List(ctx, p.ServiceType, false, 0)
		if err != nil {
			return domainPlan.Plan{}, fmt.Errorf("list plans: %w", err)
		}
		for _, e := range existing {
			if e.Name == p.Name && e.ID != p.ID {
				return domainPlan.Plan{}, domainPlan.ErrAlreadyExists
			}
		}
	}
	// Preserve audit timestamps; updated_at is trigger-managed in prod.
	p.CreatedAt = current.CreatedAt
	p.TenantID = current.TenantID
	if err := uc.repo.Save(ctx, p); err != nil {
		return domainPlan.Plan{}, fmt.Errorf("update plan: %w", err)
	}
	logger.WithComponent("ManagePlans").WithField("plan_id", p.ID).Info("plan updated")
	return p, nil
}

// Delete removes a plan row. Referencing subscriptions block the delete via
// FK RESTRICT — surface that as a clear precondition failure.
func (uc *ManagePlansUseCase) Delete(ctx context.Context, id string) error {
	if id == "" {
		return domainPlan.ErrNotFound
	}
	if _, err := uc.repo.FindByID(ctx, id); err != nil {
		return err
	}
	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete plan: %w", err)
	}
	logger.WithComponent("ManagePlans").WithField("plan_id", id).Info("plan deleted")
	return nil
}
