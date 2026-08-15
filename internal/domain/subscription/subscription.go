package subscription

import "time"

// Status represents the lifecycle status of a customer subscription.
type Status string

const (
	StatusActive    Status = "ACTIVE"
	StatusSuspended Status = "SUSPENDED"
	StatusExpired   Status = "EXPIRED"
	StatusPending   Status = "PENDING"
)

// Subscription represents an active ISP service plan subscription for a customer.
type Subscription struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	PlanID     string    `json:"plan_id"`
	Status     Status    `json:"status"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Price      float64   `json:"price"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
