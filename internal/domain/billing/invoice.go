package billing

import "time"

// Status faktur (DATABASE-SCHEMA-ISP.md §2.6).
const (
	StatusUnpaid    = "UNPAID"
	StatusPartial   = "PARTIAL"
	StatusPaid      = "PAID"
	StatusOverdue   = "OVERDUE"
	StatusCancelled = "CANCELLED"
)

// Invoice represents a monthly billing invoice for a customer.
// Nominal dipecah dari Amount tunggal menjadi subtotal/discount/tax/total
// sesuai skema ISP; Total = Subtotal - Discount + TaxAmount.
type Invoice struct {
	ID                string     `json:"id"`
	TenantID          string     `json:"tenant_id"`
	InvoiceNumber     string     `json:"invoice_number"`
	CustomerID        string     `json:"customer_id"`
	SubscriptionID    *string    `json:"subscription_id,omitempty"`
	Period            string     `json:"period"` // 'YYYY-MM'
	Subtotal          float64    `json:"subtotal"`
	Discount          float64    `json:"discount"`
	TaxAmount         float64    `json:"tax_amount"`
	Total             float64    `json:"total"`
	PaidAmount        float64    `json:"paid_amount"`
	DueDate           time.Time  `json:"due_date"`
	PaidAt            *time.Time `json:"paid_at,omitempty"`
	Status            string     `json:"status"`
	QRPayload         string     `json:"qr_payload"`
	ManualPaymentCode string     `json:"manual_payment_code"`
	Notes             string     `json:"notes,omitempty"`
	CancelledAt       *time.Time `json:"cancelled_at,omitempty"`
	CancelReason      string     `json:"cancel_reason,omitempty"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}
