package network

import "context"

// StreamOutput starts a long-running command and streams its output.
// Deferred: port.DeviceDriver has no Stream method yet (see
// NetOps-Architecture.md roadmap Fase 7 and port/device_driver.go) — this
// will be implemented once that interface extension lands, not before.
func StreamOutput(ctx context.Context) error {
	return nil
}
