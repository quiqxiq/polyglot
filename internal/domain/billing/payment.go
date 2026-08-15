package billing

import "time"

// PaymentMethod represents the method used for settlement.
type PaymentMethod string

const (
	PaymentMethodBankTransfer PaymentMethod = "BANK_TRANSFER"
	PaymentMethodEWallet      PaymentMethod = "EWALLET"
	PaymentMethodCash         PaymentMethod = "CASH"
	PaymentMethodQRIS         PaymentMethod = "QRIS"
)

// Payment represents a recorded payment transaction against an invoice.
type Payment struct {
	ID            string        `json:"id"`
	InvoiceID     string        `json:"invoice_id"`
	Amount        float64       `json:"amount"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	TransactionRef string       `json:"transaction_ref"`
	PaidAt        time.Time     `json:"paid_at"`
	CreatedAt     time.Time     `json:"created_at"`
}
