package port

import "context"

// DeviceRepository defines persistence operations for devices.
type DeviceRepository interface {
	FindByID(ctx context.Context, id string) error
}
