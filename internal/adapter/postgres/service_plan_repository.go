package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/plan"
	"github.com/quixiq/polyglot/internal/port"
)

type ServicePlanRepository struct {
	db *gorm.DB
}

var _ port.ServicePlanRepository = (*ServicePlanRepository)(nil)

// NewServicePlanRepository returns a port.ServicePlanRepository backed by GORM/Postgres.
func NewServicePlanRepository(db *gorm.DB) *ServicePlanRepository {
	return &ServicePlanRepository{db: db}
}

func (r *ServicePlanRepository) Save(ctx context.Context, p plan.ServicePlan) error {
	m := model.ServicePlanModelFromDomain(p)
	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		return err
	}
	p.CreatedAt = m.CreatedAt
	p.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *ServicePlanRepository) FindByID(ctx context.Context, id string) (plan.ServicePlan, error) {
	var m model.ServicePlanModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if err != nil {
		return plan.ServicePlan{}, mapNotFound(err)
	}
	return m.ToDomain(), nil
}

func (r *ServicePlanRepository) FindByName(ctx context.Context, tenantID, name string) (plan.ServicePlan, error) {
	var m model.ServicePlanModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND name = ?", tenantID, name).
		First(&m).Error
	if err != nil {
		return plan.ServicePlan{}, mapNotFound(err)
	}
	return m.ToDomain(), nil
}

func (r *ServicePlanRepository) List(ctx context.Context, activeOnly bool) ([]plan.ServicePlan, error) {
	q := r.db.WithContext(ctx)
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	var mList []model.ServicePlanModel
	if err := q.Order("name").Find(&mList).Error; err != nil {
		return nil, err
	}
	out := make([]plan.ServicePlan, len(mList))
	for i := range mList {
		out[i] = mList[i].ToDomain()
	}
	return out, nil
}

func (r *ServicePlanRepository) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&model.ServicePlanModel{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
