package port

import (
	"context"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
)

// PaymentReader menyediakan riwayat pembayaran milik satu pelanggan
// (join payments → invoices).
type PaymentReader interface {
	ListByCustomer(ctx context.Context, customerID string, limit int) ([]domainBilling.Payment, error)
}
