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

	// Bahan worker isolir/sinkronisasi router: cari langganan aktif
	// berdasarkan router (devices.id) + username PPP/hotspot.
	FindByDeviceAndUsername(ctx context.Context, deviceID, username string) (subscription.Subscription, error)

	// ListActive returns all non-deleted subscriptions with status ACTIVE —
	// input generator tagihan bulanan.
	ListActive(ctx context.Context) ([]subscription.Subscription, error)

	// ListLifecycle returns ACTIVE + ISOLATED non-deleted subscriptions —
	// domain kerja worker isolir/auto-suspend/provision-retry.
	ListLifecycle(ctx context.Context) ([]subscription.Subscription, error)

	// HasActiveForPlan melaporkan ada tidaknya langganan aktif/isolated
	// yang memakai plan — dipakai guard delete service plan.
	HasActiveForPlan(ctx context.Context, planID string) (bool, error)

	// Delete menghapus permanen baris langganan — dipanggil setelah guard
	// tagihan lolos pada alur manage_subscription.Delete.
	Delete(ctx context.Context, id string) error
}
