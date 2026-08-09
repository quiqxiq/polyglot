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
	ID         string    `json:"id" gorm:"primaryKey"`
	CustomerID string    `json:"customer_id" gorm:"not null;index"`
	Amount     float64   `json:"amount" gorm:"not null"`
	Status     string    `json:"status" gorm:"not null;default:UNPAID"`
	DueDate    time.Time `json:"due_date" gorm:"not null"`
	PaidAt     *time.Time `json:"paid_at,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Plan represents an ISP service offering.
type Plan struct {
	ID          string  `json:"id" gorm:"primaryKey"`
	Name        string  `json:"name" gorm:"not null"`
	SpeedMbps   int     `json:"speed_mbps" gorm:"not null"`
	Price       float64 `json:"price" gorm:"not null"`
	Description string  `json:"description"`
}
