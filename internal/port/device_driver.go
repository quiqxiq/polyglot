package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// DeviceDriver is the contract every vendor-specific driver must implement.
// Construction is NOT part of this interface: each vendor package exposes
// its own NewDriver(ctx, target device.Target) that connects immediately
// and returns an already-connected *Driver satisfying this interface —
// there is no separate Session type on the driver's connection lifecycle.
// See docs/adr/0002-devicedriver-tanpa-session-terpisah.md for the decision
// record, including its explicit divergence from the earlier
// Connect/Session/Stream sketch in NetOps-Architecture.md §5.3.
//
// Streaming (long-lived commands with live output) is intentionally absent
// here. It is deferred until Fase 7 (streaming) per NetOps-Architecture.md's
// roadmap is actually built, at which point this interface will be extended
// — not before, to avoid a half-implemented method existing today.
type DeviceDriver interface {
	// Execute runs cmd against the already-connected device.
	Execute(ctx context.Context, cmd command.Command) (command.Result, error)
	// Classify reports the risk class of cmd according to this vendor's own
	// conventions — domain/command/policy.go uses this result and never
	// inspects vendor command syntax itself.
	Classify(cmd command.Command) command.Class
	// Translate maps an abstract Operation to this vendor's native Command.
	// Returns an error if this vendor has no definition for op.
	Translate(op command.Operation) (command.Command, error)
	// Close releases the underlying connection.
	Close() error
}
