package port

import (
	"context"
	"time"

	"github.com/quixiq/polyglot/internal/domain/restorepoint"
)

// RestorePointFilter narrows restore point listings.
type RestorePointFilter struct {
	TenantID string
	DeviceID string
	Status   string
	Limit    int
	Offset   int
}

// RestorePointRepository persists restore point bundles and artifacts.
type RestorePointRepository interface {
	Create(ctx context.Context, rp restorepoint.RestorePoint) error
	FindByID(ctx context.Context, id string) (restorepoint.RestorePoint, error)
	List(ctx context.Context, f RestorePointFilter) ([]restorepoint.RestorePoint, int, error)
	UpdateStatus(ctx context.Context, id string, status restorepoint.Status, usedAt *time.Time) error
	// FindLatestByConfigHash returns the newest AVAILABLE bundle for device
	// with the same config hash (binary backup dedupe).
	FindLatestByConfigHash(ctx context.Context, deviceID, configHash string) (restorepoint.RestorePoint, error)
	// ListExpired returns bundles eligible for retention cleanup.
	ListExpired(ctx context.Context, before time.Time, keepLast int) ([]restorepoint.RestorePoint, error)
	// Delete marks a bundle DELETED (soft) and returns its artifacts for file cleanup.
	Delete(ctx context.Context, id string) (restorepoint.RestorePoint, error)
}
