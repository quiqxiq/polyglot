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

	// Pencarian cepat untuk portal & kasir (DATABASE-SCHEMA-ISP.md §4.2).
	FindByPortalAccessCode(ctx context.Context, code string) (customer.Customer, error)
	FindByPhone(ctx context.Context, phone string) (customer.Customer, error)
	FindByCustomerCode(ctx context.Context, code string) (customer.Customer, error)
}
