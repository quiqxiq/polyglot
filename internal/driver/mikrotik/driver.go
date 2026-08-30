package mikrotik

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/quiqxiq/goros/v4"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// Driver implements port.DeviceDriver AND port.StreamingDeviceDriver for
// Mikrotik RouterOS via goros (github.com/quiqxiq/goros/v4).
//
// It holds TWO independent, persistent connections to the same device —
// exec for one-shot commands (Execute, via RunArgsContext) and stream for
// streaming commands (Stream, via ListenArgsContext) — deliberately never
// sharing one *routeros.Client between the two. See
// docs/adr/0003-mikrotik-dual-connection-streaming.md for the full
// rationale: without this separation, a long-running stream (/ping,
// /interface/monitor-traffic, or any print follow/follow-only/interval=)
// starves command execution on the same connection — confirmed directly
// from goros's own source, not assumed.
//
// DEVIASI: CLAUDE.md §1.2 describes vendor driver packages as normally
// just driver.go + commands.go. Mikrotik's dual-connection, reconnecting,
// streaming-capable design genuinely needs more structure than that
// two-file baseline, so it's split further: driver.go (this file — Driver
// type, Execute, Close, Classify/Translate delegation), connect.go (dial +
// auto-reconnect), stream.go (Stream + streamHandle). commands.go keeps
// its originally mandated role — vendor command catalog + risk
// classification — and nothing else.
type Driver struct {
	target device.Target

	exec   *managedConn
	stream *managedConn

	// execSem serializes wire USE of exec: held for the entire duration of
	// one RunArgsContext call, so two Execute calls can never interleave
	// their sentence bytes on the same connection.
	execSem chan struct{}
	// streamSem serializes only the brief moment of ISSUING a new Listen
	// on stream — see Stream's doc comment in stream.go for why it isn't
	// held for the stream's whole lifetime.
	streamSem chan struct{}

	stopCh chan struct{}
	wg     sync.WaitGroup

	streamsMu sync.Mutex
	streams   map[*streamHandle]struct{}

	closeOnce sync.Once
	closeErr  error
}

// Compile-time proof that *Driver actually satisfies both port.DeviceDriver
// (required per CLAUDE.md §1.2 for every vendor driver) and
// port.StreamingDeviceDriver (this vendor's opt-in — see
// internal/port/streaming_driver.go) and port.ValidatingDeviceDriver (this
// vendor's opt-in pre-flight validation — see internal/port/validating_driver.go).
var (
	_ port.DeviceDriver           = (*Driver)(nil)
	_ port.StreamingDeviceDriver  = (*Driver)(nil)
	_ port.ValidatingDeviceDriver = (*Driver)(nil)
)

// NewDriver connects to target and returns a ready, connected Driver. Both
// the exec and stream connections are dialed and logged in before NewDriver
// returns, so a Driver is only ever handed back already-usable. If the
// second dial fails, the first connection is closed before returning the
// error — no connection is left leaked.
func NewDriver(ctx context.Context, target device.Target) (*Driver, error) {
	execClient, execErrC, err := dialAndLogin(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("mikrotik: connect exec: %w", err)
	}

	streamClient, streamErrC, err := dialAndLogin(ctx, target)
	if err != nil {
		_ = execClient.Close()
		return nil, fmt.Errorf("mikrotik: connect stream: %w", err)
	}

	d := &Driver{
		target:    target,
		exec:      newManagedConn(execClient, execErrC),
		stream:    newManagedConn(streamClient, streamErrC),
		execSem:   make(chan struct{}, 1),
		streamSem: make(chan struct{}, 1),
		stopCh:    make(chan struct{}),
		streams:   make(map[*streamHandle]struct{}),
	}

	d.wg.Add(2)
	go d.superviseConnection(d.exec)
	go d.superviseConnection(d.stream)

	return d, nil
}

// Execute runs cmd against the connected RouterOS device over the
// persistent exec connection and waits for its reply. cmd.Raw is a
// RouterOS API path (e.g. "/system/resource/print"), cmd.Args are sentence
// attributes. Execute refuses streaming commands (see isStreamingCommand
// in commands.go) — use Stream for those instead.
//
// ctx bounds how long Execute is willing to WAIT, but is deliberately
// never passed into the underlying goros call itself: if
// RunArgsContext's ctx is cancelled, it cancels the shared connection
// reader outright, taking down every other in-flight command and stream on
// this same connection — not just this one call. See connect.go's
// dialAndLogin doc comment and
// docs/adr/0003-mikrotik-dual-connection-streaming.md. Instead, the actual
// device call always runs against context.Background() in its own
// goroutine; ctx only controls how long THIS call waits for that
// goroutine's result before giving up on it — the goroutine itself is left
// to finish on its own even if Execute has already returned to its caller.
func (d *Driver) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
	if isStreamingCommand(cmd) {
		return command.Result{}, fmt.Errorf("%w: %q", ErrStreamingCommand, cmd.Raw)
	}

	select {
	case d.execSem <- struct{}{}:
	case <-ctx.Done():
		return command.Result{}, fmt.Errorf("mikrotik: waiting for execute slot: %w", ctx.Err())
	}

	client := d.exec.get()
	if client == nil {
		<-d.execSem
		return command.Result{}, ErrNotConnected
	}

	args := buildArgs(cmd)
	type outcome struct {
		reply *routeros.Reply
		err   error
	}
	done := make(chan outcome, 1)
	go func() {
		reply, err := client.RunArgsContext(context.Background(), args)
		done <- outcome{reply, err}
		<-d.execSem // released only once the wire operation truly finished — see doc comment above
	}()

	select {
	case <-ctx.Done():
		return command.Result{}, fmt.Errorf("mikrotik: execute %q cancelled: %w", cmd.Raw, ctx.Err())
	case o := <-done:
		if o.err != nil {
			return command.Result{}, fmt.Errorf("mikrotik: execute %q: %w", cmd.Raw, o.err)
		}
		return toResult(o.reply), nil
	}
}

// Validate dry-runs cmd against the connected RouterOS device's own parser
// (Gate 1 `:parse`) and attribute schema (Gate 2 `/console/inspect`) via
// goros, WITHOUT executing it — catching unknown paths, unknown attributes,
// and syntax errors before they reach the device. It returns nil when the
// command is valid, or when the session cannot validate (RouterOS v6
// degrades silently by design). Streaming commands are skipped here: they
// are refused by Execute anyway, and their flags (follow, interval) are
// stream-path concerns, not validation concerns.
//
// The underlying goros call is a network round-trip on the persistent exec
// connection, so — exactly like Execute — it deliberately runs against
// context.Background() in its own goroutine: if Validate's ctx is
// cancelled, only THIS call's wait is abandoned, never the shared
// connection reader. See connect.go's dialAndLogin doc comment and
// docs/adr/0003-mikrotik-dual-connection-streaming.md.
func (d *Driver) Validate(ctx context.Context, cmd command.Command) error {
	if isStreamingCommand(cmd) {
		return nil
	}

	select {
	case d.execSem <- struct{}{}:
	case <-ctx.Done():
		return fmt.Errorf("mikrotik: waiting for validate slot: %w", ctx.Err())
	}

	client := d.exec.get()
	if client == nil {
		<-d.execSem
		return ErrNotConnected
	}

	tc := toTransportCommand(cmd)
	done := make(chan error, 1)
	go func() {
		done <- client.Validate(context.Background(), tc)
		<-d.execSem // released only once the wire operation truly finished — see Execute's doc comment
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("mikrotik: validate %q cancelled: %w", cmd.Raw, ctx.Err())
	case err := <-done:
		if err == nil {
			return nil
		}
		return fmt.Errorf("mikrotik: validate %q: %w", cmd.Raw, err)
	}
}

// Classify reports the risk class of cmd according to RouterOS API conventions.
func (d *Driver) Classify(cmd command.Command) command.Class {
	return Classify(cmd)
}

// Translate maps an abstract Operation to a RouterOS-native Command.
func (d *Driver) Translate(op command.Operation) (command.Command, error) {
	return Translate(op)
}

// Close stops both persistent connections and their reconnect supervisors,
// cancelling any streams still open first so their pump goroutines exit
// cleanly instead of reading from a connection that's about to go away.
// Safe to call more than once — later calls return the same result as the
// first.
func (d *Driver) Close() error {
	d.closeOnce.Do(func() {
		d.streamsMu.Lock()
		streams := make([]*streamHandle, 0, len(d.streams))
		for h := range d.streams {
			streams = append(streams, h)
		}
		d.streamsMu.Unlock()
		for _, h := range streams {
			_ = h.Cancel()
		}

		close(d.stopCh)

		var errs []error
		if c := d.exec.get(); c != nil {
			if err := c.Close(); err != nil {
				errs = append(errs, fmt.Errorf("mikrotik: close exec connection: %w", err))
			}
		}
		if c := d.stream.get(); c != nil {
			if err := c.Close(); err != nil {
				errs = append(errs, fmt.Errorf("mikrotik: close stream connection: %w", err))
			}
		}

		d.wg.Wait()
		d.closeErr = errors.Join(errs...)
	})
	return d.closeErr
}

func (d *Driver) registerStream(h *streamHandle) {
	d.streamsMu.Lock()
	d.streams[h] = struct{}{}
	d.streamsMu.Unlock()
}

func (d *Driver) unregisterStream(h *streamHandle) {
	d.streamsMu.Lock()
	delete(d.streams, h)
	d.streamsMu.Unlock()
}
