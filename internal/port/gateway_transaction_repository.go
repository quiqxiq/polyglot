package port

import (
	"context"
	"time"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
)

// GatewayTransactionRepository menyimpan transaksi payment gateway
// (DATABASE-SCHEMA-ISP.md §2.7 — gateway_transactions).
// GatewayTransactionFilter menyaring transaksi gateway.
type GatewayTransactionFilter struct {
	TenantID  string
	Gateway   string
	Status    string
	InvoiceID string
	From      *time.Time
	To        *time.Time
	Limit     int
	Offset    int
}

// GatewayTransactionFilteredRepository adalah ekstensi untuk daftar transaksi
// gateway & sinkronisasi transaksi PENDING (W4).
type GatewayTransactionFilteredRepository interface {
	ListFiltered(ctx context.Context, f GatewayTransactionFilter) ([]domainBilling.GatewayTransaction, int, error)
	ListPendingBefore(ctx context.Context, tenantID string, before time.Time, limit int) ([]domainBilling.GatewayTransaction, error)
}

type GatewayTransactionRepository interface {
	Save(ctx context.Context, t domainBilling.GatewayTransaction) error
	FindByExternalID(ctx context.Context, gateway, externalID string) (domainBilling.GatewayTransaction, error)
	FindByInvoice(ctx context.Context, invoiceID string) ([]domainBilling.GatewayTransaction, error)
	UpdateStatus(ctx context.Context, id, status string) error
	LinkPayment(ctx context.Context, id, paymentID string, feeAmount float64) error
}
