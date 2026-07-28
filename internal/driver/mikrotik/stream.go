package mikrotik

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-routeros/routeros/v3"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// Stream starts cmd as a streaming (Listen) operation over the persistent
// stream connection and returns a StreamHandle immediately — non-blocking,
// no polling. cmd must satisfy isStreamingCommand (ping, monitor-traffic,
// or a print with follow/follow-only/interval=); anything else is refused,
// since Execute is the correct call for one-shot commands.
//
// Only the brief moment of ISSUING the underlying ListenArgsContext call is
// serialized against other Stream calls (via d.streamSem) — once issued,
// ListenArgsContext returns immediately (per go-routeros's own source, it
// registers a tag and returns; it does not block for data), so multiple
// streams (e.g. a /ping and a /interface/monitor-traffic at the same time)
// can run concurrently afterward without contending for that slot. This is
// intentionally different from Execute's semaphore, which stays held for
// the whole underlying RunArgsContext call — see driver.go's Execute for
// why.
//
// Like Execute, the actual go-routeros call always runs against
// context.Background(), never ctx: see connect.go's dialAndLogin doc
// comment and docs/adr/0003-mikrotik-dual-connection-streaming.md for why
// handing a request-scoped context straight to go-routeros is unsafe for a
// persistent connection. ctx here only bounds how long Stream is willing
// to wait for a free issuance slot; once a stream is running, stop it with
// StreamHandle.Cancel, not by cancelling ctx.
func (d *Driver) Stream(ctx context.Context, cmd command.Command) (port.StreamHandle, error) {
	if !isStreamingCommand(cmd) {
		return nil, fmt.Errorf("%w: %q", ErrNotStreamingCommand, cmd.Raw)
	}

	select {
	case d.streamSem <- struct{}{}:
	case <-ctx.Done():
		return nil, fmt.Errorf("mikrotik: waiting for stream issuance slot: %w", ctx.Err())
	}
	defer func() { <-d.streamSem }()

	client := d.stream.get()
	if client == nil {
		return nil, ErrNotConnected
	}

	args := buildArgs(cmd)
	listenReply, err := client.ListenArgsContext(context.Background(), args)
	if err != nil {
		return nil, fmt.Errorf("mikrotik: stream %q: %w", cmd.Raw, err)
	}

	h := newStreamHandle(listenReply)
	d.registerStream(h)
	go h.pump(d)

	return h, nil
}

// streamHandle implements port.StreamHandle over one go-routeros
// *ListenReply. It translates each streamed *proto.Sentence into a
// vendor-agnostic command.Result and forwards it on out — the only
// polling-free path from the wire to the caller.
type streamHandle struct {
	listen *routeros.ListenReply

	out    chan command.Result
	doneCh chan struct{} // closed once pump has fully exited

	mu     sync.Mutex
	err    error
	closed bool
}

func newStreamHandle(listen *routeros.ListenReply) *streamHandle {
	return &streamHandle{
		listen: listen,
		// Buffered generously so a momentarily slow consumer never stalls
		// the shared asyncLoop dispatcher for the whole stream connection
		// — same rationale as Client.Queue in connect.go.
		out:    make(chan command.Result, 1000),
		doneCh: make(chan struct{}),
	}
}

// Chan returns the channel of streamed results. It is closed when the
// stream ends: the device finished on its own (e.g. bounded /ping count),
// Cancel was called, or the underlying connection was lost.
func (h *streamHandle) Chan() <-chan command.Result {
	return h.out
}

// Cancel stops the stream and releases its resources. Safe to call more
// than once and safe to call after the stream already ended on its own.
// Blocks until the stream's pump goroutine has fully drained and exited.
func (h *streamHandle) Cancel() error {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return nil
	}
	h.closed = true
	h.mu.Unlock()

	_, err := h.listen.CancelContext(context.Background())
	<-h.doneCh
	if err != nil {
		return fmt.Errorf("mikrotik: cancel stream: %w", err)
	}
	return nil
}

// Err returns the first error encountered while streaming, if any. Only
// meaningful after Chan() has been closed.
func (h *streamHandle) Err() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.err
}

// pump copies sentences from the underlying go-routeros ListenReply into
// h.out as command.Result, translating each vendor-specific
// proto.Sentence word map into the vendor-agnostic Rows shape. Runs until
// the underlying channel closes (device ended the stream, Cancel was
// called, or the stream connection was lost) — it never polls anything.
func (h *streamHandle) pump(d *Driver) {
	defer close(h.out)
	defer close(h.doneCh)
	defer d.unregisterStream(h)

	for sen := range h.listen.Chan() {
		h.out <- command.Result{Rows: []map[string]string{sen.Map}}
	}

	if h.listen.Done != nil && h.listen.Done.Word == "!trap" {
		h.mu.Lock()
		// A trap that merely acknowledges our own Cancel is the expected end of
		// an unbounded stream, not a failure: RouterOS ends a cancelled
		// /interface/monitor-traffic with an "interrupted" trap (other commands
		// like "print follow" end cleanly). Record it as an error only when it
		// wasn't caused by us — h.closed is set by Cancel before it triggers the
		// cancellation that closes listen.Chan(), so it is reliably true here in
		// that case. Any trap on a stream we did NOT cancel is a real error.
		if !(h.closed && isInterruptedTrap(h.listen.Done.Map)) {
			h.err = fmt.Errorf("mikrotik: device reported error ending stream: %v", h.listen.Done.Map)
		}
		h.mu.Unlock()
	}
}

// isInterruptedTrap reports whether a RouterOS "!trap" sentence is the
// "interrupted" acknowledgement the device sends when a streaming command is
// cancelled, as opposed to a genuine error trap.
func isInterruptedTrap(trap map[string]string) bool {
	return trap["message"] == "interrupted"
}
