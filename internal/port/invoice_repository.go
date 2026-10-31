package port

import (
	"context"
	"time"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

// InvoiceRepository defines persistence operations for invoices.
type InvoiceRepository interface {
	Save(ctx context.Context, inv billing.Invoice) error
	FindByID(ctx context.Context, id string) (billing.Invoice, error)
	FindByCustomerID(ctx context.Context, customerID string) ([]billing.Invoice, error)
	FindAll(ctx context.Context) ([]billing.Invoice, error)
	// FindPaged seperti FindAll dengan filter tenant + limit/offset (F6-8).
	FindPaged(ctx context.Context, f PageFilter) ([]billing.Invoice, error)
	UpdateStatus(ctx context.Context, id string, status string) error

	// Alur kasir scan & bayar cepat (DATABASE-SCHEMA-ISP.md §4.2).
	FindByPaymentCode(ctx context.Context, code string) (billing.Invoice, error)
	FindByQRPayload(ctx context.Context, qr string) (billing.Invoice, error)

	// Idempotensi generator tagihan: cek apakah periode sudah ditagih.
	FindBySubscriptionPeriod(ctx context.Context, subscriptionID, period string) (billing.Invoice, error)

	// SaveWithItems menyimpan faktur beserta seluruh baris item-nya dalam
	// satu transaksi (aggregate).
	SaveWithItems(ctx context.Context, inv billing.Invoice, items []billing.InvoiceItem) error

	// AddGatewayFeeItem menambahkan biaya layanan gateway (beban pelanggan)
	// ke invoice secara idempoten: bila item fee sudah ada, invoice
	// dikembalikan apa adanya. Mengembalikan invoice terbaru.
	AddGatewayFeeItem(ctx context.Context, invoiceID string, amount float64) (billing.Invoice, error)

	// HasForSubscription melaporkan ada tidaknya faktur yang menunjuk
	// langganan — guard delete subscription pada manage_subscription.
	HasForSubscription(ctx context.Context, subID string) (bool, error)

	// Delete menghapus permanen faktur dan item-nya berdasarkan id.
	Delete(ctx context.Context, id string) error

	// DeleteByCustomerID menghapus seluruh faktur dan item-nya untuk seorang pelanggan.
	DeleteByCustomerID(ctx context.Context, customerID string) error
}

// InvoiceFilter menyaring daftar invoice untuk halaman tagihan & laporan.
// Semua field opsional; Limit<=0 berarti tanpa batas (dipakai export).
type InvoiceFilter struct {
	TenantID          string
	CustomerID        string
	SubscriptionID    string
	Status            string
	Period            string
	InvoiceType       string
	Search            string // invoice number / nama pelanggan
	OverdueOnly       bool
	UnpaidOnly        bool
	UnpaidDaysDisplay int
	DueFrom           *time.Time
	DueTo             *time.Time
	CreatedFrom       *time.Time
	CreatedTo         *time.Time
	Limit             int
	Offset            int
}

// InvoiceSummary merangkum invoice yang cocok dengan filter.
type InvoiceSummary struct {
	Count            int
	TotalBilled      float64
	TotalPaid        float64
	TotalOutstanding float64
	TotalOverdue     float64
}

// InvoiceFilteredRepository adalah ekstensi InvoiceRepository untuk filter,
// ringkasan, dan kandidat overdue (W1-W2 PLAN-BILLING-OPS-V3).
type InvoiceFilteredRepository interface {
	FindFiltered(ctx context.Context, f InvoiceFilter) ([]billing.Invoice, int, error)
	Summary(ctx context.Context, f InvoiceFilter) (InvoiceSummary, error)
	FindOverdueCandidates(ctx context.Context, tenantID string, now time.Time, limit int) ([]billing.Invoice, error)
}
