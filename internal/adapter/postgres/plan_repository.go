package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/plan"
	"github.com/quixiq/polyglot/internal/port"
)

var _ port.PlanRepository = (*PlanRepository)(nil)

// PlanRepository persists master paket layanan (service_plans).
type PlanRepository struct {
	db *gorm.DB
}

func NewPlanRepository(db *gorm.DB) *PlanRepository {
	return &PlanRepository{db: db}
}

func (r *PlanRepository) Save(ctx context.Context, p plan.Plan) error {
	m := model.ServicePlanModelFromDomain(p)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *PlanRepository) FindByID(ctx context.Context, id string) (plan.Plan, error) {
	var m model.ServicePlanModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return plan.Plan{}, plan.ErrNotFound
	}
	if err != nil {
		return plan.Plan{}, err
	}
	return m.ToDomain(), nil
}

func (r *PlanRepository) List(ctx context.Context, serviceType string, activeOnly bool, limit int) ([]plan.Plan, error) {
	q := r.db.WithContext(ctx).Model(&model.ServicePlanModel{})
	if serviceType != "" {
		q = q.Where("service_type = ?", serviceType)
	}
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	var mList []model.ServicePlanModel
	if err := q.Order("name asc").Find(&mList).Error; err != nil {
		return nil, err
	}
	plans := make([]plan.Plan, len(mList))
	for i, m := range mList {
		plans[i] = m.ToDomain()
	}
	return plans, nil
}

func (r *PlanRepository) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&model.ServicePlanModel{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return plan.ErrNotFound
	}
	return nil
}
