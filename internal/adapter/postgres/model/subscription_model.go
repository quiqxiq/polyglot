package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/subscription"
)

// SubscriptionModel represents the GORM model for customer plan subscriptions.
type SubscriptionModel struct {
	ID         string    `gorm:"primaryKey"`
	CustomerID string    `gorm:"not null;index"`
	PlanID     string    `gorm:"not null"`
	Status     string    `gorm:"not null;default:ACTIVE"`
	StartDate  time.Time `gorm:"not null"`
	EndDate    time.Time `gorm:"not null"`
	Price      float64   `gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (SubscriptionModel) TableName() string {
	return "subscriptions"
}

func (m *SubscriptionModel) ToDomain() subscription.Subscription {
	if m == nil {
		return subscription.Subscription{}
	}
	return subscription.Subscription{
		ID:         m.ID,
		CustomerID: m.CustomerID,
		PlanID:     m.PlanID,
		Status:     m.Status,
		StartDate:  m.StartDate,
		EndDate:    m.EndDate,
		Price:      m.Price,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

func SubscriptionModelFromDomain(s subscription.Subscription) *SubscriptionModel {
	return &SubscriptionModel{
		ID:         s.ID,
		CustomerID: s.CustomerID,
		PlanID:     s.PlanID,
		Status:     s.Status,
		StartDate:  s.StartDate,
		EndDate:    s.EndDate,
		Price:      s.Price,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}
