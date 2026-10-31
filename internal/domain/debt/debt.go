// Package debt memodelkan hutang-piutang manual pelanggan (di luar tagihan
// invoice) — fitur W5 PLAN-BILLING-OPS-V3.
package debt

import "time"

// Status hutang.
const (
	StatusOpen      = "OPEN"
	StatusPartial   = "PARTIAL"
	StatusPaid      = "PAID"
	StatusCancelled = "CANCELLED"
)

// Debt adalah satu catatan hutang pelanggan.
type Debt struct {
	ID              string     `json:"id"`
	TenantID        string     `json:"tenant_id"`
	CustomerID      string     `json:"customer_id"`
	SubscriptionID  *string    `json:"subscription_id,omitempty"`
	Amount          float64    `json:"amount"`
	RemainingAmount float64    `json:"remaining_amount"`
	Description     string     `json:"description,omitempty"`
	DebtDate        time.Time  `json:"debt_date"`
	DueDate         *time.Time `json:"due_date,omitempty"`
	Status          string     `json:"status"`
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	CreatedBy       *uint      `json:"created_by,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// IsOutstanding melaporkan hutang yang masih punya sisa & tidak dibatalkan.
func (d Debt) IsOutstanding() bool {
	return d.Status != StatusCancelled && d.Status != StatusPaid && d.RemainingAmount > 0
}

// Payment adalah pembayaran cicilan/lunas untuk satu hutang.
type Payment struct {
	ID                string    `json:"id"`
	TenantID          string    `json:"tenant_id"`
	DebtID            string    `json:"debt_id"`
	Amount            float64   `json:"amount"`
	PaymentMethodID   *string   `json:"payment_method_id,omitempty"`
	CashTransactionID *string   `json:"cash_transaction_id,omitempty"`
	PaymentDate       time.Time `json:"payment_date"`
	Notes             string    `json:"notes,omitempty"`
	RecordedBy        *uint     `json:"recorded_by,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

// Summary ringkasan hutang per pelanggan.
type Summary struct {
	Count       int     `json:"count"`
	TotalAmount float64 `json:"total_amount"`
	Outstanding float64 `json:"outstanding"`
	Overdue     int     `json:"overdue"`
}
