package connect

import (
	"context"

	"github.com/quixiq/polyglot/internal/port"
)

// DriverProvider is the standard function signature to obtain a port.DeviceDriver for a given device ID.
type DriverProvider func(ctx context.Context, deviceID string) (port.DeviceDriver, error)
