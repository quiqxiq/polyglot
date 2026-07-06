package port

import "context"

// CustomerRepository defines persistence operations for customers.
type CustomerRepository interface {
	FindByID(ctx context.Context, id string) error
}
