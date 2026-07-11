package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/customer"
)

// CustomerRepository defines persistence operations for customers.
// Mapped to the `customers` table per migration 000008.
type CustomerRepository interface {
	// FindByID returns the customer for the given id, or an error wrapping
	// customer.ErrNotFound if no such customer exists.
	FindByID(ctx context.Context, id string) (customer.Customer, error)

	// FindAll returns all customers ordered by full_name.
	FindAll(ctx context.Context) ([]customer.Customer, error)

	// Create inserts a new customer.
	Create(ctx context.Context, c customer.Customer) (customer.Customer, error)

	// Update modifies an existing customer.
	Update(ctx context.Context, c customer.Customer) (customer.Customer, error)

	// Delete removes a customer by id.
	Delete(ctx context.Context, id string) error
}
