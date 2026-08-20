package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/subscription"
)

// SubscriptionRepository defines persistence operations for subscriptions.
type SubscriptionRepository interface {
	Save(ctx context.Context, sub subscription.Subscription) error
	FindByID(ctx context.Context, id string) (subscription.Subscription, error)
	FindByCustomerID(ctx context.Context, customerID string) ([]subscription.Subscription, error)
	FindAll(ctx context.Context) ([]subscription.Subscription, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}
