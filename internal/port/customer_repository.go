package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/domain/subscription"
)

// CustomerRepository defines persistence operations for customer management.
type CustomerRepository interface {
	Save(ctx context.Context, c customer.Customer) error
	FindByID(ctx context.Context, id string) (customer.Customer, error)
	FindAll(ctx context.Context) ([]customer.Customer, error)
	Delete(ctx context.Context, id string) error
	FindSubscriptions(ctx context.Context, customerID string) ([]subscription.Subscription, error)
}

