package port

import "context"

// SubscriptionRepository defines persistence operations for subscriptions.
type SubscriptionRepository interface {
	FindByID(ctx context.Context, id string) error
}
