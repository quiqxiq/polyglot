package billing

import "time"

// Tipe kanal pembayaran.
const (
	MethodTypeCash    = "CASH"
	MethodTypeBank    = "BANK"
	MethodTypeQRIS    = "QRIS"
	MethodTypeGateway = "GATEWAY"
)

// Cara kasir menemukan tagihan saat pembayaran cepat.
const (
	ScanQRScan         = "QR_SCAN"
	ScanCodeInput      = "CODE_INPUT"
	ScanManual         = "MANUAL"
	ScanPaymentGateway = "PAYMENT_GATEWAY"
)

// PaymentMethod represents a payment channel master
// (DATABASE-SCHEMA-ISP.md §2.7 — payment_methods).
type PaymentMethod struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"` // 'TUNAI', 'TRANSFER_BCA', 'QRIS', 'TRIPAY_VA'
	Type      string    `json:"type"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// Payment represents a payment receipt settling an invoice
// (DATABASE-SCHEMA-ISP.md §2.7 — payments).
type Payment struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	PaymentNo       string    `json:"payment_no"` // "PAY-202608-0012"
	InvoiceID       string    `json:"invoice_id"`
	PaymentMethodID *string   `json:"payment_method_id,omitempty"`
	Amount          float64   `json:"amount"`
	PaymentDate     time.Time `json:"payment_date"`
	ReceivedBy      *uint     `json:"received_by,omitempty"` // users.id kasir
	ScanMethod      string    `json:"scan_method"`
	Reference       string    `json:"reference,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}
