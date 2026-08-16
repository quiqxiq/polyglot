package models

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/plan"
)

// PlanModel represents the GORM model for ISP plans.
type PlanModel struct {
	ID          string  `gorm:"primaryKey"`
	Name        string  `gorm:"not null"`
	SpeedMbps   int     `gorm:"not null"`
	Price       float64 `gorm:"not null"`
	Description string  `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (PlanModel) TableName() string {
	return "plans"
}

func (m *PlanModel) ToDomain() plan.Plan {
	if m == nil {
		return plan.Plan{}
	}
	return plan.Plan{
		ID:          m.ID,
		Name:        m.Name,
		SpeedMbps:   m.SpeedMbps,
		Price:       m.Price,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func PlanModelFromDomain(p plan.Plan) *PlanModel {
	return &PlanModel{
		ID:          p.ID,
		Name:        p.Name,
		SpeedMbps:   p.SpeedMbps,
		Price:       p.Price,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
