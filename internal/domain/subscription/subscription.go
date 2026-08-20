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
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	PlanID     string    `json:"plan_id"`
	Status     string    `json:"status"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Price      float64   `json:"price"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

