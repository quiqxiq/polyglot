package billing

import "time"

// Status represents the invoice payment status.
type Status string

const (
	StatusUnpaid    Status = "UNPAID"
	StatusPaid      Status = "PAID"
	StatusCancelled Status = "CANCELLED"
	StatusOverdue   Status = "OVERDUE"
)

// Invoice represents a billing invoice issued to a customer subscriber.
type Invoice struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	Amount     float64   `json:"amount"`
	Status     Status    `json:"status"`
	DueDate    time.Time `json:"due_date"`
	PaidAt     *time.Time `json:"paid_at,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
