package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/device"
)

// DeviceRepository defines persistence operations for device inventory
// records. Implementations (e.g. internal/adapter/postgres) map to the
// `devices` table per Polyglot-Architecture.md §7.2.
type DeviceRepository interface {
	// Save stores a new device inventory record in the database.
	Save(ctx context.Context, d device.Device) error
	// FindByID returns the device inventory record for id, or
	// device.ErrNotFound if no such device exists.
	FindByID(ctx context.Context, id string) (device.Device, error)
	// FindAll returns all registered device inventory records.
	FindAll(ctx context.Context) ([]device.Device, error)
	// Update modifies an existing device inventory record.
	Update(ctx context.Context, d device.Device) error
	// Delete removes a device inventory record by id.
	Delete(ctx context.Context, id string) error
	// FindByUserScope returns only devices assigned to the specified user.
	FindByUserScope(ctx context.Context, userID uint) ([]device.Device, error)
}
