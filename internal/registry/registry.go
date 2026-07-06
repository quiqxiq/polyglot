package registry

import "context"

// New builds the device registry and connection pool dispatcher.
// Named New, not NewRegistry — CLAUDE.md §2.1 (avoid package name stutter).
// TODO: hold a map of deviceID -> port.DeviceDriver, one long-lived
// connection per device (reused across calls, closed on idle timeout) —
// this is where the "avoid reconnect storms" lesson from go-routeros
// applies: never dial a fresh connection per command.
func New(ctx context.Context) error {
	return nil
}
