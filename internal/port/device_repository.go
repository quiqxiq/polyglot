package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/device"
)

// DeviceRepository defines persistence operations for device inventory
// records. Implementations (e.g. internal/adapter/postgres) map to the
// `devices` table per Polyglot-Architecture.md §7.2.
type DeviceRepository interface {
	// FindByID returns the device inventory record for id, or
	// device.ErrNotFound (wrapping it) if no such device exists.
	FindByID(ctx context.Context, id string) (device.Device, error)
}
