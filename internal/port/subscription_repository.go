package port

import (
	"context"
	"time"

	"github.com/quixiq/polyglot/internal/domain/subscription"
)

type SubscriptionRepository interface {
	Save(ctx context.Context, sub subscription.Subscription) error
	FindByID(ctx context.Context, id string) (subscription.Subscription, error)
	FindByCustomerID(ctx context.Context, customerID string) ([]subscription.Subscription, error)
	FindAll(ctx context.Context) ([]subscription.Subscription, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	// UpdateMapping persists provisioning mapping results (remote_id and
	// friends) after the device-side account has been created.
	UpdateMapping(ctx context.Context, id, remoteUsername, remoteID, deviceID string) error
	// FindByDeviceAndUsername resolves the subscription that owns a given
	// account on a router — used for idempotent re-provisioning.
	FindByDeviceAndUsername(ctx context.Context, deviceID, username string) (subscription.Subscription, error)
	// SetIsolation records the isolation lifecycle fields atomically with
	// the status change.
	SetIsolation(ctx context.Context, id string, status string, isolatedAt *time.Time, reason string) error
}
