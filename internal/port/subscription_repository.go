package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/subscription"
)

// SubscriptionRepository defines persistence operations for subscriptions.
// Mapped to the `subscriptions` table per migration 000009.
type SubscriptionRepository interface {
	// FindByID returns the subscription for the given id, or an error
	// wrapping subscription.ErrNotFound if no such subscription exists.
	FindByID(ctx context.Context, id string) (subscription.Subscription, error)

	// FindAll returns all subscriptions ordered by created_at desc.
	FindAll(ctx context.Context) ([]subscription.Subscription, error)

	// FindByCustomer returns subscriptions for a specific customer.
	FindByCustomer(ctx context.Context, customerID string) ([]subscription.Subscription, error)

	// FindByDevice returns subscriptions on a specific device.
	FindByDevice(ctx context.Context, deviceID string) ([]subscription.Subscription, error)

	// Create inserts a new subscription.
	Create(ctx context.Context, s subscription.Subscription) (subscription.Subscription, error)

	// Update modifies an existing subscription.
	Update(ctx context.Context, s subscription.Subscription) (subscription.Subscription, error)

	// Delete removes a subscription by id.
	Delete(ctx context.Context, id string) error
}
