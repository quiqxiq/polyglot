package port

import (
	"context"
	"time"

	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/domain/subscription"
)

type CustomerRepository interface {
	Save(ctx context.Context, c customer.Customer) error
	FindByID(ctx context.Context, id string) (customer.Customer, error)
	FindAll(ctx context.Context) ([]customer.Customer, error)
	Delete(ctx context.Context, id string) error
	// SoftDelete marks the record deleted (deleted_at set) without
	// destroying billing history.
	SoftDelete(ctx context.Context, id string, at time.Time) error
	FindByPhone(ctx context.Context, phone string) (customer.Customer, error)
	NextCustomerCode(ctx context.Context) (string, error)
	FindSubscriptions(ctx context.Context, customerID string) ([]subscription.Subscription, error)
}
