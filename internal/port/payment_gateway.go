package port

import (
	"context"
	"encoding/json"
	"errors"
)

// Kesalahan umum payment gateway.
var (
	ErrGatewayDisabled   = errors.New("payment gateway tidak aktif")
	ErrGatewayBadSign    = errors.New("signature callback tidak valid")
	ErrGatewayUnknownRef = errors.New("external_id tidak dikenal")
)

// ChargeRequest adalah permintaan pembuatan tagihan online.
type ChargeRequest struct {
	TenantID      string
	InvoiceID     string
	InvoiceNumber string
	Amount        float64
	Channel       string // QRIS, BRIVA, BCAVA, ... (kosong = default settings)
	CustomerName  string
	CustomerPhone string
	CustomerEmail string
	ExpireMinutes int
}

// ChargeResult adalah hasil create-transaction dari provider.
type ChargeResult struct {
	ExternalID  string // merchant_ref / order id
	PaymentURL  string
	QRString    string
	VANumber    string
	FeeAmount   float64
	Status      string // PENDING | SETTLED | EXPIRED | FAILED
	RawResponse json.RawMessage
}

// WebhookEvent adalah callback tervalidasi dari provider.
type WebhookEvent struct {
	ExternalID  string
	MerchantRef string // invoice number / referensi merchant
	Status      string // PAID/SETTLED, EXPIRED, FAILED
	PaidAmount  float64
	Raw         json.RawMessage
}

// PaymentGateway adalah kontrak provider pembayaran online.
// Implementasi: adapter Tripay (fase 4); Midtrans/Xendit menyusul.
type PaymentGateway interface {
	Name() string
	Enabled(ctx context.Context) bool
	CreateCharge(ctx context.Context, req ChargeRequest) (ChargeResult, error)
	// ParseWebhook memvalidasi body + signature dan mengembalikan event.
	ParseWebhook(ctx context.Context, body []byte, signatureHeader string) (WebhookEvent, error)
}
