package device

import (
	"context"
	"time"
)

// Target describes how to reach and authenticate to a device — the minimal
// connection parameters a DeviceDriver needs. Distinct from Device below:
// Device is the stored inventory record (name, site, tags, ...); Target is
// just what NewDriver(ctx, target) needs to connect.
// Extra holds vendor-specific parameters that don't fit the common fields
// (e.g. the SNMP community string for zteolt).
type Target struct {
	Host     string
	Port     int
	Username string
	Password string
	Timeout  time.Duration
	Extra    map[string]string
}

// New creates a placeholder Device entity (the stored inventory record,
// mapped to the `devices` table per NetOps-Architecture.md §7.2).
// Named New, not NewDevice — CLAUDE.md §2.1: package name must not be
// stuttered in the identifier (device.New, not device.NewDevice).
// TODO: implement per NetOps-Architecture.md domain rules.
func New(ctx context.Context) error {
	return nil
}
