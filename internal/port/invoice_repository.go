package port

import "context"

// InvoiceRepository defines persistence operations for invoices.
type InvoiceRepository interface {
	FindByID(ctx context.Context, id string) error
}
