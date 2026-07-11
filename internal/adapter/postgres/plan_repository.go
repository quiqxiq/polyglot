package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/domain/plan"
)

// planModel maps the `plans` table to a GORM-friendly struct per migration
// 000005.
type planModel struct {
	ID                  string    `gorm:"column:id;primaryKey;type:uuid"`
	Name                string    `gorm:"column:name;not null"`
	ServiceType         string    `gorm:"column:service_type;not null"`
	Description         string    `gorm:"column:description"`
	Price               float64   `gorm:"column:price;not null"`
	BillingPeriodMonths int       `gorm:"column:billing_period_months;not null;default:1"`
	BandwidthDownKbps   int       `gorm:"column:bandwidth_down_kbps;not null"`
	BandwidthUpKbps     int       `gorm:"column:bandwidth_up_kbps;not null"`
	BurstDownKbps       *int      `gorm:"column:burst_down_kbps"`
	BurstUpKbps         *int      `gorm:"column:burst_up_kbps"`
	BurstThresholdKbps  *int      `gorm:"column:burst_threshold_kbps"`
	BurstTimeSeconds    *int      `gorm:"column:burst_time_seconds"`
	FUPQuotaMB          *int      `gorm:"column:fup_quota_mb"`
	FUPThrottleDownKbps *int      `gorm:"column:fup_throttle_down_kbps"`
	FUPThrottleUpKbps   *int      `gorm:"column:fup_throttle_up_kbps"`
	IsActive            *bool     `gorm:"column:is_active;not null;default:true"`
	CreatedAt           time.Time `gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt           time.Time `gorm:"column:updated_at;not null;autoUpdateTime"`
}

// TableName returns the explicit table name for the plan model.
func (planModel) TableName() string {
	return "plans"
}

// toDomain maps a planModel to the domain plan.Plan.
func (m planModel) toDomain() plan.Plan {
	return plan.Plan{
		ID:                  m.ID,
		Name:                m.Name,
		ServiceType:         m.ServiceType,
		Description:         m.Description,
		Price:               m.Price,
		BillingPeriodMonths: m.BillingPeriodMonths,
		BandwidthDownKbps:   m.BandwidthDownKbps,
		BandwidthUpKbps:     m.BandwidthUpKbps,
		BurstDownKbps:       m.BurstDownKbps,
		BurstUpKbps:         m.BurstUpKbps,
		BurstThresholdKbps:  m.BurstThresholdKbps,
		BurstTimeSeconds:    m.BurstTimeSeconds,
		FUPQuotaMB:          m.FUPQuotaMB,
		FUPThrottleDownKbps: m.FUPThrottleDownKbps,
		FUPThrottleUpKbps:   m.FUPThrottleUpKbps,
		IsActive:            m.IsActive != nil && *m.IsActive,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}

// planFromDomain maps a domain plan.Plan to a planModel.
func planFromDomain(p plan.Plan) planModel {
	return planModel{
		ID:                  p.ID,
		Name:                p.Name,
		ServiceType:         p.ServiceType,
		Description:         p.Description,
		Price:               p.Price,
		BillingPeriodMonths: p.BillingPeriodMonths,
		BandwidthDownKbps:   p.BandwidthDownKbps,
		BandwidthUpKbps:     p.BandwidthUpKbps,
		BurstDownKbps:       p.BurstDownKbps,
		BurstUpKbps:         p.BurstUpKbps,
		BurstThresholdKbps:  p.BurstThresholdKbps,
		BurstTimeSeconds:    p.BurstTimeSeconds,
		FUPQuotaMB:          p.FUPQuotaMB,
		FUPThrottleDownKbps: p.FUPThrottleDownKbps,
		FUPThrottleUpKbps:   p.FUPThrottleUpKbps,
		IsActive:            &p.IsActive,
	}
}

// PlanRepository implements port.PlanRepository backed by PostgreSQL.
type PlanRepository struct {
	db *gorm.DB
}

// NewPlanRepository returns a port.PlanRepository backed by GORM/Postgres.
func NewPlanRepository(db *gorm.DB) *PlanRepository {
	return &PlanRepository{db: db}
}

// FindByID returns the plan for the given id, or plan.ErrNotFound.
func (r *PlanRepository) FindByID(ctx context.Context, id string) (plan.Plan, error) {
	var m planModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return plan.Plan{}, fmt.Errorf("plan %s: %w", id, plan.ErrNotFound)
		}
		return plan.Plan{}, fmt.Errorf("plan %s: %w", id, err)
	}
	return m.toDomain(), nil
}

// FindAll returns all plans ordered by name.
func (r *PlanRepository) FindAll(ctx context.Context) ([]plan.Plan, error) {
	var models []planModel
	if err := r.db.WithContext(ctx).Order("name").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list plans: %w", err)
	}

	plans := make([]plan.Plan, len(models))
	for i, m := range models {
		plans[i] = m.toDomain()
	}
	return plans, nil
}

// FindActive returns only active plans.
func (r *PlanRepository) FindActive(ctx context.Context) ([]plan.Plan, error) {
	var models []planModel
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("name").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list active plans: %w", err)
	}

	plans := make([]plan.Plan, len(models))
	for i, m := range models {
		plans[i] = m.toDomain()
	}
	return plans, nil
}

// Create inserts a new plan. Select("*") ensures zero-value fields (e.g.
// IsActive=false) are persisted rather than falling back to column defaults.
func (r *PlanRepository) Create(ctx context.Context, p plan.Plan) (plan.Plan, error) {
	m := planFromDomain(p)
	if err := r.db.WithContext(ctx).Select("*").Create(&m).Error; err != nil {
		return plan.Plan{}, fmt.Errorf("create plan: %w", err)
	}
	return m.toDomain(), nil
}

// Update modifies an existing plan.
func (r *PlanRepository) Update(ctx context.Context, p plan.Plan) (plan.Plan, error) {
	m := planFromDomain(p)
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return plan.Plan{}, fmt.Errorf("update plan: %w", err)
	}
	return m.toDomain(), nil
}

// Delete removes a plan by id.
func (r *PlanRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&planModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete plan: %w", err)
	}
	return nil
}
