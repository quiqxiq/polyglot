package billing

import (
	"encoding/json"
	"time"
)

// Status transaksi payment gateway.
const (
	GatewayStatusPending = "PENDING"
	GatewayStatusSettled = "SETTLED"
	GatewayStatusExpired = "EXPIRED"
	GatewayStatusFailed  = "FAILED"
)

// GatewayTransaction represents an online payment-gateway transaction
// (Tripay / Midtrans / Xendit) incl. webhook callback bookkeeping
// (DATABASE-SCHEMA-ISP.md §2.7 — gateway_transactions).
type GatewayTransaction struct {
	ID             string          `json:"id"`
	TenantID       string          `json:"tenant_id"`
	Gateway        string          `json:"gateway"` // 'TRIPAY', 'MIDTRANS', 'XENDIT'
	ExternalID     string          `json:"external_id"`
	InvoiceID      *string         `json:"invoice_id,omitempty"`
	PaymentID      *string         `json:"payment_id,omitempty"`
	Amount         float64         `json:"amount"`
	FeeAmount      float64         `json:"fee_amount,omitempty"`
	Status         string          `json:"status"`
	PaymentChannel string          `json:"payment_channel,omitempty"` // 'BCA_VA', 'QRIS', ...
	PaymentURL     string          `json:"payment_url,omitempty"`
	QRString       string          `json:"qr_string,omitempty"`
	RawCallback    json.RawMessage `json:"raw_callback,omitempty"`
	CallbackCount  int             `json:"callback_count"`
	PaidAt         *time.Time      `json:"paid_at,omitempty"`
	ExpiresAt      *time.Time      `json:"expires_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}
