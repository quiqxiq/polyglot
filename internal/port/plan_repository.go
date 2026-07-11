package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/plan"
)

// PlanRepository defines persistence operations for service plans.
// Mapped to the `plans` table per migration 000005.
type PlanRepository interface {
	// FindByID returns the plan for the given id, or an error wrapping
	// plan.ErrNotFound if no such plan exists.
	FindByID(ctx context.Context, id string) (plan.Plan, error)

	// FindAll returns all plans ordered by name.
	FindAll(ctx context.Context) ([]plan.Plan, error)

	// FindActive returns only active plans (is_active = true).
	FindActive(ctx context.Context) ([]plan.Plan, error)

	// Create inserts a new plan.
	Create(ctx context.Context, p plan.Plan) (plan.Plan, error)

	// Update modifies an existing plan.
	Update(ctx context.Context, p plan.Plan) (plan.Plan, error)

	// Delete removes a plan by id.
	Delete(ctx context.Context, id string) error
}
