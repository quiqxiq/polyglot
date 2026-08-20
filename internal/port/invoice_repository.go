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
}
