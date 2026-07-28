package genericcli

import (
	"context"
	"fmt"
	"time"

	"github.com/scrapli/scrapligo/driver/opoptions"
	"github.com/scrapli/scrapligo/driver/options"
	"github.com/scrapli/scrapligo/platform"
	"github.com/scrapli/scrapligo/response"
	"github.com/scrapli/scrapligo/util"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
)

// defaultOpsTimeout is used when target.Timeout is zero — scrapligo's own
// SendCommand has no context.Context support at all and would otherwise
// wait on an unresponsive device with no bound.
const defaultOpsTimeout = 30 * time.Second

// cliDriver is the minimal surface Session needs from whichever concrete
// driver scrapligo's Platform actually builds. A platform definition (YAML)
// declares its own "driver-type": "network" (the convention used by every
// built-in platform and by scrapligo's own assets/platforms/example.yaml
// template for authoring new ones) or "generic". Platform.GetNetworkDriver
// and Platform.GetGenericDriver each error out if called against the type
// the YAML did NOT declare — see resolveDriver below — so Session cannot
// hardcode either one; it works with whichever this interface is
// satisfied by.
type cliDriver interface {
	Open() error
	Close() error
	SendCommand(command string, opts ...util.Option) (*response.Response, error)
}

// resolveDriver returns whichever concrete driver type p's underlying
// platform definition actually built. It tries GetNetworkDriver first,
// since that's the driver-type used by every scrapligo built-in platform
// and by scrapligo's own template for authoring custom ones; it falls back
// to GetGenericDriver for definitions that explicitly declared
// driver-type: generic instead.
func resolveDriver(p *platform.Platform) (cliDriver, error) {
	if nd, err := p.GetNetworkDriver(); err == nil {
		return nd, nil
	}
	gd, err := p.GetGenericDriver()
	if err != nil {
		return nil, fmt.Errorf("platform exposes neither a network nor a generic driver: %w", err)
	}
	return gd, nil
}

// Session wraps one persistent scrapligo connection and is the shared
// engine behind BOTH genericssh and generictelnet — they differ only in
// which transportType/defaultPort they pass to NewSession (see their own
// driver.go). All prompt detection, paging, and login-sequence handling is
// scrapligo's own engine, driven entirely by whatever platform definition
// is supplied; Session adds only what this project needs on top: risk
// classification and operation translation (via Catalog), and a
// context-aware call/retry wrapper, since scrapligo's own API has no
// context.Context support at all.
type Session struct {
	catalog       Catalog
	target        device.Target
	platformDef   any
	transportType string
	defaultPort   int
	opsTimeout    time.Duration

	// sem is a capacity-1 channel-based semaphore, not a sync.Mutex, so
	// that WAITING for a turn is itself cancellable via ctx — identical
	// pattern and rationale to mikrotik.Driver's execSem.
	sem chan struct{}

	driver cliDriver
}

// NewSession dials, logs in, and returns a ready-to-use persistent
// Session. transportType must be a scrapligo transport name — "standard"
// (pure-Go crypto/ssh) for genericssh, "telnet" for generictelnet.
// defaultPort is used when target.Port is zero (22 for SSH, 23 for
// Telnet). platformDef is anything platform.NewPlatform accepts: a
// built-in platform name, a custom YAML file path (see
// internal/platformdef/), or raw YAML/JSON bytes.
func NewSession(
	ctx context.Context,
	target device.Target,
	transportType string,
	defaultPort int,
	platformDef any,
	catalog Catalog,
) (*Session, error) {
	opsTimeout := target.Timeout
	if opsTimeout <= 0 {
		opsTimeout = defaultOpsTimeout
	}

	s := &Session{
		catalog:       catalog,
		target:        target,
		platformDef:   platformDef,
		transportType: transportType,
		defaultPort:   defaultPort,
		opsTimeout:    opsTimeout,
		sem:           make(chan struct{}, 1),
	}

	driver, err := s.dial(ctx)
	if err != nil {
		return nil, err
	}
	s.driver = driver

	return s, nil
}

// dial builds a fresh scrapligo platform + driver and opens it. Used both
// by NewSession (initial connect) and by Execute's reconnect-on-failure
// path — always builds a brand-new driver rather than reusing state, since
// scrapligo drivers aren't designed to be reopened once closed.
func (s *Session) dial(ctx context.Context) (cliDriver, error) {
	port := s.target.Port
	if port == 0 {
		port = s.defaultPort
	}

	opts := []util.Option{
		options.WithAuthUsername(s.target.Username),
		options.WithAuthPassword(s.target.Password),
		options.WithTransportType(s.transportType),
		options.WithPort(port),
		options.WithTimeoutOps(s.opsTimeout),
		// TODO: make host-key strictness configurable via target.Extra —
		// defaulting to no-strict-key because most ISP network gear in
		// this project's context has no maintained known_hosts workflow.
		// This is a deliberate simplification, not an oversight.
		options.WithAuthNoStrictKey(),
	}
	if len(s.catalog.FailedWhenContains) > 0 {
		opts = append(opts, options.WithFailedWhenContains(s.catalog.FailedWhenContains))
	}
	if s.catalog.ReadDelay > 0 {
		opts = append(opts, options.WithReadDelay(s.catalog.ReadDelay))
	}

	p, err := platform.NewPlatform(s.platformDef, s.target.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf("genericcli: build platform: %w", err)
	}

	driver, err := resolveDriver(p)
	if err != nil {
		return nil, fmt.Errorf("genericcli: resolve driver: %w", err)
	}

	// scrapligo's Open() has no context.Context support at all; run it in
	// a goroutine so a caller-supplied ctx can still bound how long we
	// wait, without pretending we can cancel the library call itself.
	done := make(chan error, 1)
	go func() { done <- driver.Open() }()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("genericcli: open %s: %w", s.target.Host, ctx.Err())
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("genericcli: open %s: %w", s.target.Host, err)
		}
	}

	return driver, nil
}

// Execute runs cmd over the persistent session and waits for its result.
// Unlike mikrotik's Execute, passing ctx to the underlying library call is
// not a concern here at all — scrapligo's SendCommand has no ctx
// parameter, so ctx only ever controls how long THIS function waits; the
// underlying call is left to finish (or fail) on its own in the background
// either way, and the semaphore is released only once it truly has,
// mirroring mikrotik's Execute for the same reason: so the NEXT Execute
// never starts writing before the previous one has genuinely finished.
func (s *Session) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
	select {
	case s.sem <- struct{}{}:
	case <-ctx.Done():
		return command.Result{}, fmt.Errorf("genericcli: waiting for execute slot: %w", ctx.Err())
	}

	type outcome struct {
		resp *response.Response
		err  error
	}
	done := make(chan outcome, 1)
	go func() {
		resp, err := s.sendWithReconnect(ctx, cmd.Raw)
		done <- outcome{resp, err}
		<-s.sem
	}()

	select {
	case <-ctx.Done():
		return command.Result{}, fmt.Errorf("genericcli: execute %q cancelled: %w", cmd.Raw, ctx.Err())
	case o := <-done:
		if o.err != nil {
			return command.Result{}, fmt.Errorf("genericcli: execute %q: %w", cmd.Raw, o.err)
		}
		if o.resp.Failed != nil {
			return command.Result{}, fmt.Errorf("genericcli: execute %q rejected by device: %w", cmd.Raw, o.resp.Failed)
		}
		return command.Result{Output: o.resp.Result}, nil
	}
}

// sendWithReconnect sends raw once; if it fails, it assumes the session may
// have died, reconnects exactly once, and retries. This is a deliberately
// SIMPLER resilience model than mikrotik's background auto-reconnect
// supervisor: scrapligo exposes no async error channel to watch the way
// go-routeros does, so a background supervisor here would have nothing to
// wait on except polling — and this project does not poll. A synchronous
// retry-on-failure, bounded to exactly one attempt, is the honest
// alternative. See docs/adr/0004-generic-cli-driver-scrapligo.md.
func (s *Session) sendWithReconnect(ctx context.Context, raw string) (*response.Response, error) {
	resp, err := s.driver.SendCommand(raw, opoptions.WithTimeoutOps(s.opsTimeout))
	if err == nil {
		return resp, nil
	}

	newDriver, dialErr := s.dial(ctx)
	if dialErr != nil {
		return nil, fmt.Errorf("send failed (%w) and reconnect failed: %w", err, dialErr)
	}
	_ = s.driver.Close() // best-effort: old conn is being replaced, its close error does not affect the reconnect
	s.driver = newDriver

	return s.driver.SendCommand(raw, opoptions.WithTimeoutOps(s.opsTimeout))
}

// Classify reports the risk class of cmd according to this Session's Catalog.
func (s *Session) Classify(cmd command.Command) command.Class {
	return s.catalog.Classify(cmd)
}

// Translate maps an abstract Operation to this Session's Catalog's native Command.
func (s *Session) Translate(op command.Operation) (command.Command, error) {
	return s.catalog.Translate(op)
}

// Close releases the underlying connection. It waits for any in-flight
// Execute to finish first (by acquiring the same semaphore Execute uses),
// so Close never races a command that's still being sent.
func (s *Session) Close() error {
	s.sem <- struct{}{}
	defer func() { <-s.sem }()

	return s.driver.Close()
}
