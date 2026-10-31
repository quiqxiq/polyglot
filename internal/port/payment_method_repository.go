package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

// PaymentMethodRepository mengelola master metode pembayaran manual
// (tunai/bank/QRIS) — W3 PLAN-BILLING-OPS-V3.
type PaymentMethodRepository interface {
	Save(ctx context.Context, m billing.PaymentMethod) error
	FindByID(ctx context.Context, id string) (billing.PaymentMethod, error)
	List(ctx context.Context, tenantID string, activeOnly bool) ([]billing.PaymentMethod, error)
	SetActive(ctx context.Context, id string, active bool) error
}
