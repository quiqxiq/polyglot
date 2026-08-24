package port

import (
	"context"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
)

// GatewayTransactionRepository menyimpan transaksi payment gateway
// (DATABASE-SCHEMA-ISP.md §2.7 — gateway_transactions).
type GatewayTransactionRepository interface {
	Save(ctx context.Context, t domainBilling.GatewayTransaction) error
	FindByExternalID(ctx context.Context, gateway, externalID string) (domainBilling.GatewayTransaction, error)
	FindByInvoice(ctx context.Context, invoiceID string) ([]domainBilling.GatewayTransaction, error)
	UpdateStatus(ctx context.Context, id, status string) error
	LinkPayment(ctx context.Context, id, paymentID string, feeAmount float64) error
}
