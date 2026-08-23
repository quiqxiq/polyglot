package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/plan"
)

// ServicePlanRepository defines persistence operations for ISP service plans
// (tabel service_plans, DATABASE-SCHEMA-ISP.md §2.2).
type ServicePlanRepository interface {
	Save(ctx context.Context, p plan.ServicePlan) error
	FindByID(ctx context.Context, id string) (plan.ServicePlan, error)
	FindByName(ctx context.Context, tenantID, name string) (plan.ServicePlan, error)
	List(ctx context.Context, activeOnly bool) ([]plan.ServicePlan, error)
	Delete(ctx context.Context, id string) error
}
