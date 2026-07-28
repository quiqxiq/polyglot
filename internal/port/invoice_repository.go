package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

// InvoiceRepository defines persistence operations for invoices.
// Mapped to the `invoices` and `invoice_items` tables per migration 000014.
type InvoiceRepository interface {
	// FindByID returns the invoice (with items) for the given id, or an
	// error wrapping billing.ErrNotFound if no such invoice exists.
	FindByID(ctx context.Context, id string) (billing.Invoice, error)

	// FindAll returns all invoices ordered by created_at desc.
	FindAll(ctx context.Context) ([]billing.Invoice, error)

	// FindByCustomer returns invoices for a specific customer.
	FindByCustomer(ctx context.Context, customerID string) ([]billing.Invoice, error)

	// Create inserts a new invoice with its items.
	Create(ctx context.Context, inv billing.Invoice) (billing.Invoice, error)

	// Update modifies an existing invoice.
	Update(ctx context.Context, inv billing.Invoice) (billing.Invoice, error)

	// Delete removes an invoice (and its items via CASCADE) by id.
	Delete(ctx context.Context, id string) error
}
