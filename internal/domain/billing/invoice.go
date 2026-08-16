package billing

import "time"

const (
	StatusUnpaid    = "UNPAID"
	StatusPaid      = "PAID"
	StatusOverdue   = "OVERDUE"
	StatusCancelled = "CANCELLED"
)

// Invoice represents a billing invoice for a customer.
type Invoice struct {
	ID         string     `json:"id"`
	CustomerID string     `json:"customer_id"`
	Amount     float64    `json:"amount"`
	Status     string     `json:"status"`
	DueDate    time.Time  `json:"due_date"`
	PaidAt     *time.Time `json:"paid_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

