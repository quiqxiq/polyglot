package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

// Error sentinels untuk alur pembayaran kasir tinggal di
// internal/domain/billing/errors.go (ErrInvoiceAlreadyPaid,
// ErrInvoiceCancelled, ErrOverpayment) per DEVELOPMENT-GUIDELINES.md §6.

// CashPaymentCommand is the atomic cashier payment request: satu transaksi
// DB memproses invoice + kwitansi + jurnal kas + antrean WA sekaligus
// (DATABASE-SCHEMA-ISP.md §4.2).
type CashPaymentCommand struct {
	TenantID         string
	InvoiceID        string
	PaymentMethodID  *string
	Amount           float64
	CashAccountID    string // rekening kas tujuan (mis. 'ca-1001-kas-kantor')
	IncomeCategoryID string // kategori kas IN (mis. 'cc-tagihan')
	ReceivedBy       *uint  // users.id kasir
	ScanMethod       string // QR_SCAN | CODE_INPUT | MANUAL | PAYMENT_GATEWAY
	Reference        string
	Notes            string
}

// PaymentProcessor executes an all-or-nothing cashier payment:
// update invoices → insert payments → insert cash_transactions →
// queue wa_notifications, committed atomically.
type PaymentProcessor interface {
	ProcessCashPayment(ctx context.Context, cmd CashPaymentCommand) (billing.Payment, error)
}
