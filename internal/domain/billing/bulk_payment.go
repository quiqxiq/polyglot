package billing

import "time"

// Status batch & item mass payment.
const (
	BulkStatusQueued   = "QUEUED"
	BulkStatusRunning  = "RUNNING"
	BulkStatusDone     = "DONE"
	BulkStatusPartial  = "PARTIAL"
	BulkStatusFailed   = "FAILED"
	BulkItemQueued     = "QUEUED"
	BulkItemProcessing = "PROCESSING"
	BulkItemSucceeded  = "SUCCEEDED"
	BulkItemFailed     = "FAILED"
)

// BulkPaymentBatch adalah satu batch pembayaran massal (job queue).
type BulkPaymentBatch struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	CreatedBy       *uint     `json:"created_by,omitempty"`
	PaymentMethodID *string   `json:"payment_method_id,omitempty"`
	InvoiceCount    int       `json:"invoice_count"`
	Status          string    `json:"status"`
	JobID           string    `json:"job_id,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// BulkPaymentItem adalah satu invoice di dalam batch mass payment.
type BulkPaymentItem struct {
	ID          string     `json:"id"`
	BatchID     string     `json:"batch_id"`
	InvoiceID   string     `json:"invoice_id"`
	Amount      float64    `json:"amount"`
	Status      string     `json:"status"`
	Error       string     `json:"error,omitempty"`
	PaymentID   *string    `json:"payment_id,omitempty"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
