package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

// InvoiceRepository defines persistence operations for invoices.
type InvoiceRepository interface {
	Save(ctx context.Context, inv billing.Invoice) error
	FindByID(ctx context.Context, id string) (billing.Invoice, error)
	FindByCustomerID(ctx context.Context, customerID string) ([]billing.Invoice, error)
	FindAll(ctx context.Context) ([]billing.Invoice, error)
	UpdateStatus(ctx context.Context, id string, status string) error

	// Alur kasir scan & bayar cepat (DATABASE-SCHEMA-ISP.md §4.2).
	FindByPaymentCode(ctx context.Context, code string) (billing.Invoice, error)
	FindByQRPayload(ctx context.Context, qr string) (billing.Invoice, error)

	// Idempotensi generator tagihan: cek apakah periode sudah ditagih.
	FindBySubscriptionPeriod(ctx context.Context, subscriptionID, period string) (billing.Invoice, error)

	// SaveWithItems menyimpan faktur beserta seluruh baris item-nya dalam
	// satu transaksi (aggregate).
	SaveWithItems(ctx context.Context, inv billing.Invoice, items []billing.InvoiceItem) error

	// HasForSubscription melaporkan ada tidaknya faktur yang menunjuk
	// langganan — guard delete subscription pada manage_subscription.
	HasForSubscription(ctx context.Context, subID string) (bool, error)

	// Delete menghapus permanen faktur dan item-nya berdasarkan id.
	Delete(ctx context.Context, id string) error

	// DeleteByCustomerID menghapus seluruh faktur dan item-nya untuk seorang pelanggan.
	DeleteByCustomerID(ctx context.Context, customerID string) error
}
