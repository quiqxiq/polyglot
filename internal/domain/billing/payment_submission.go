package billing

import "time"

// Status pembayaran manual (upload bukti transfer): PENDING, APPROVED, REJECTED.
const (
	PaymentSubmissionPending  = "PENDING"
	PaymentSubmissionApproved = "APPROVED"
	PaymentSubmissionRejected = "REJECTED"
)

// PaymentSubmission adalah bukti transfer manual yang diunggah pelanggan /
// kolektor dan direview staf finance (planning v2 §3.4).
type PaymentSubmission struct {
	ID              string     `json:"id"`
	TenantID        string     `json:"tenant_id"`
	InvoiceID       string     `json:"invoice_id"`
	Amount          float64    `json:"amount"`
	BankAccountID   *string    `json:"bank_account_id,omitempty"`
	SenderBank      string     `json:"sender_bank,omitempty"`
	SenderName      string     `json:"sender_name,omitempty"`
	ProofImageURL   string     `json:"proof_image_url"`
	Notes           string     `json:"notes,omitempty"`
	Status          string     `json:"status"`
	ReviewedBy      *uint      `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	RejectionReason string     `json:"rejection_reason,omitempty"`
	PaymentID       *string    `json:"payment_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
