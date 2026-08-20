package mikrotik

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/quiqxiq/goros/v4"

	"github.com/quixiq/polyglot/internal/domain/device"
)

// defaultAPIPort is RouterOS's default plain (non-TLS) API port.
const defaultAPIPort = 8728

// managedConn holds one *routeros.Client plus its Async() error channel,
// safely swappable in place whenever superviseConnection reconnects it.
// Every read of the current client/errC goes through get()/errChan(), so
// Execute/Stream never observe a half-updated pointer.
type managedConn struct {
	mu     sync.RWMutex
	client *routeros.Client
	errC   <-chan error
}

func newManagedConn(client *routeros.Client, errC <-chan error) *managedConn {
	return &managedConn{client: client, errC: errC}
}

func (m *managedConn) get() *routeros.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.client
}

func (m *managedConn) set(client *routeros.Client, errC <-chan error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.client = client
	m.errC = errC
}

func (m *managedConn) errChan() <-chan error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.errC
}

// dialAndLogin opens one persistent, logged-in, async-enabled connection to
// target. Both the exec and stream connections are created this way — they
// are otherwise identical; only how Driver USES them (see driver.go,
// stream.go) differs.
//
// target.Timeout, if set, bounds only the dial+login phase (via ctx). The
// Async() call that follows deliberately always uses context.Background(),
// never ctx: AsyncContext's internal read loop watches whatever context it
// was started with and cancels the shared connection reader when that
// context is done — tearing down THIS ENTIRE CONNECTION, not just one
// command. A connection this driver intends to keep alive for its whole
// lifetime must never be handed a context that anything short-lived (a
// single dial's timeout, a single request) might cancel. See
// docs/adr/0003-mikrotik-dual-connection-streaming.md.
func dialAndLogin(ctx context.Context, target device.Target) (*routeros.Client, <-chan error, error) {
	dialCtx := ctx
	if target.Timeout > 0 {
		var cancel context.CancelFunc
		dialCtx, cancel = context.WithTimeout(ctx, target.Timeout)
		defer cancel()
	}

	address := fmt.Sprintf("%s:%d", target.Host, portOrDefault(target))

	client, err := routeros.DialContext(dialCtx, address, target.Username, target.Password)
	if err != nil {
		return nil, nil, fmt.Errorf("mikrotik: dial %s: %w", address, err)
	}

	// Buffer generously: Queue sizes the channel Listen() allocates for
	// streamed "!re" sentences. The default (0, unbuffered) would make the
	// shared asyncLoop dispatcher goroutine block on every single streamed
	// row until our own consumer goroutine reads it — stalling delivery
	// for every OTHER tag on this same connection too, exec or stream.
	client.Queue = 1000

	errC := client.AsyncContext(context.Background())

	return client, errC, nil
}

func portOrDefault(target device.Target) int {
	if target.Port != 0 {
		return target.Port
	}
	return defaultAPIPort
}

// superviseConnection watches conn's current Async() error channel and
// redials with capped exponential backoff whenever it fires, until
// d.stopCh is closed. One of these runs per connection (exec, stream) —
// started from NewDriver, stopped and waited-on from Driver.Close.
func (d *Driver) superviseConnection(conn *managedConn) {
	defer d.wg.Done()

	for {
		select {
		case <-d.stopCh:
			return
		case <-conn.errChan():
			// Fell through: the connection died. Reconnect below.
		}

		if !d.reconnectUntilSuccess(conn) {
			return // stopCh was closed while reconnecting
		}
	}
}

// reconnectUntilSuccess redials conn repeatedly with capped exponential
// backoff until it succeeds or d.stopCh is closed. Returns false if it gave
// up because of stopCh (caller should stop supervising), true once
// reconnected.
func (d *Driver) reconnectUntilSuccess(conn *managedConn) bool {
	backoff := time.Second
	const maxBackoff = 10 * time.Second

	for {
		select {
		case <-d.stopCh:
			return false
		default:
		}

		client, errC, err := dialAndLogin(context.Background(), d.target)
		if err != nil {
			select {
			case <-d.stopCh:
				return false
			case <-time.After(backoff):
			}
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		conn.set(client, errC)
		return true
	}
}
