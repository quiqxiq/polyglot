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

	// FindAll returns all devices ordered by name.
	FindAll(ctx context.Context) ([]device.Device, error)

	// Create inserts a new device into the inventory.
	Create(ctx context.Context, d device.Device) (device.Device, error)

	// Update modifies an existing device inventory record.
	Update(ctx context.Context, d device.Device) (device.Device, error)

	// Delete removes a device from the inventory.
	Delete(ctx context.Context, id string) error
}
