package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// StreamHandle represents one open streaming command against a device (e.g.
// Mikrotik /ping, /interface/monitor-traffic, or any /.../print with
// follow/follow-only/interval=). Results arrive on Chan() as the underlying
// transport delivers them — there is no polling involved on either side of
// this interface.
type StreamHandle interface {
	// Chan returns the channel of streamed results. It is closed when the
	// stream ends: the device finished on its own (e.g. /ping with a
	// bounded count), Cancel was called, or the underlying connection was
	// lost. Call Err() after the channel closes to find out which.
	Chan() <-chan command.Result

	// Cancel stops the stream and releases its resources. Safe to call
	// more than once and safe to call after the stream already ended on
	// its own.
	Cancel() error

	// Err returns the first error encountered while streaming, if any.
	// Only meaningful after Chan() has been closed — reading it earlier
	// may miss an error that hasn't happened yet.
	Err() error
}

// StreamingDeviceDriver is an OPTIONAL capability a DeviceDriver may also
// implement, for vendors whose native streaming commands should be consumed
// as data arrives instead of polled repeatedly.
//
// This is deliberately a separate interface, not an addition to
// DeviceDriver itself: per ADR 0002, DeviceDriver's Stream method was
// explicitly deferred until a real implementation existed ("Fase 7").
// Mikrotik now has one (see internal/driver/mikrotik), but adding Stream
// directly onto DeviceDriver would force every other vendor driver (cisco,
// genericssh, netconf, zteolt, huaweiolt, genieacs) to grow a stub method
// they don't support yet — and their compile-time
// `var _ port.DeviceDriver = (*Driver)(nil)` assertions would keep passing
// vacuously either way, so the assertion would stop being a useful signal
// of real support.
//
// A separate interface + type assertion at the call site
// (`sd, ok := driver.(port.StreamingDeviceDriver)`) makes streaming support
// genuinely opt-in and visible per vendor. See docs/adr/0003-mikrotik-dual-connection-streaming.md.
type StreamingDeviceDriver interface {
	DeviceDriver

	// Stream starts cmd as a long-running/streaming operation and returns a
	// StreamHandle immediately (non-blocking). cmd must be a command this
	// vendor's driver recognizes as streaming — see that vendor's
	// commands.go for the exact rule (e.g. mikrotik.isStreamingCommand).
	// Passing a non-streaming command is an error, not silently accepted:
	// callers should use DeviceDriver.Execute for those instead.
	Stream(ctx context.Context, cmd command.Command) (StreamHandle, error)
}
