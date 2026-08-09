package subscription

import "time"

const (
	StatusActive    = "ACTIVE"
	StatusCancelled = "CANCELLED"
	StatusExpired   = "EXPIRED"
	StatusPending   = "PENDING"
)

// Subscription represents an active or historic plan subscription.
type Subscription struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	CustomerID string    `json:"customer_id" gorm:"not null;index"`
	PlanID     string    `json:"plan_id" gorm:"not null"`
	Status     string    `json:"status" gorm:"not null;default:ACTIVE"`
	StartDate  time.Time `json:"start_date" gorm:"not null"`
	EndDate    time.Time `json:"end_date" gorm:"not null"`
	Price      float64   `json:"price" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
