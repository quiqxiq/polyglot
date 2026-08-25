package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/plan"
)

// PlanRepository persists master paket layanan (service plans).
type PlanRepository interface {
	Save(ctx context.Context, p plan.Plan) error
	FindByID(ctx context.Context, id string) (plan.Plan, error)
	List(ctx context.Context, serviceType string, activeOnly bool, limit int) ([]plan.Plan, error)
	Delete(ctx context.Context, id string) error
}
