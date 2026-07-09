package session

import "context"

// New creates a historical record of a connection made to a device — it
// maps to the `sessions` table described in Polyglot-Architecture.md §7.2
// ("Riwayat koneksi aktif/selesai"). This is NOT the same concept as a
// connection handle inside a DeviceDriver: per ADR 0002
// (docs/adr/0002-devicedriver-tanpa-session-terpisah.md), port.DeviceDriver
// itself has no separate Session type — NewDriver connects and returns an
// already-connected *Driver directly. This package only tracks the
// audit-facing history of when a session started/ended and its outcome,
// not the live connection object.
// Named New, not NewSession — CLAUDE.md §2.1 (avoid package name stutter).
// TODO: implement per Polyglot-Architecture.md §7.2 domain rules.
func New(ctx context.Context) error {
	return nil
}
