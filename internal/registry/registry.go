package registry

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// DriverFactory builds a connected port.DeviceDriver for one device, given
// the connection target. Each vendor package's NewDriver satisfies this
// signature. The registry holds a map of driver_type → DriverFactory,
// populated at the composition root (main.go) with every supported vendor.
type DriverFactory func(ctx context.Context, target device.Target) (port.DeviceDriver, error)

// ErrDeviceDisabled indicates the device inventory record exists but is
// disabled, so the registry refuses to build a driver for it.
var ErrDeviceDisabled = errors.New("registry: device is disabled")

// ErrNoDriverFactory indicates no DriverFactory is registered for the
// device's driver_type.
var ErrNoDriverFactory = errors.New("registry: no driver factory for driver_type")

// Registry looks up device inventory records, resolves their encrypted
// credentials, and builds (and caches) long-lived DeviceDriver connections
// keyed by device ID. One driver per device is kept alive and reused across
// calls — this is the "avoid reconnect storms" lesson from go-routeros:
// never dial a fresh connection per command.
type Registry struct {
	repo      port.DeviceRepository
	vault     port.CredentialVault
	factories map[string]DriverFactory

	mu      sync.RWMutex
	drivers map[string]port.DeviceDriver
}

// New builds a Registry that resolves devices via repo, decrypts credentials
// via vault, and instantiates drivers via the provided factory map (keyed by
// device.DriverType). The caller (composition root) is responsible for
// populating factories with every supported vendor — the registry itself
// never imports driver packages, keeping it free of vendor-specific
// knowledge per Polyglot-Architecture.md §5.3.
func New(repo port.DeviceRepository, vault port.CredentialVault, factories map[string]DriverFactory) *Registry {
	return &Registry{
		repo:      repo,
		vault:     vault,
		factories: factories,
		drivers:   make(map[string]port.DeviceDriver),
	}
}

// Get returns the cached port.DeviceDriver for deviceID, or builds one if
// this is the first request for that device. The lookup chain: repo.FindByID
// (inventory record) → vault.Get (decrypted credentials) → Device.ToTarget
// (merged connection params) → factories[driver_type] (vendor NewDriver).
// The resulting driver is cached so subsequent calls reuse the same
// connection without re-dialing.
func (r *Registry) Get(ctx context.Context, deviceID string) (port.DeviceDriver, error) {
	r.mu.RLock()
	d, ok := r.drivers[deviceID]
	r.mu.RUnlock()
	if ok {
		return d, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock — another goroutine may have
	// built the driver between the read-lock check and here.
	if d, ok := r.drivers[deviceID]; ok {
		return d, nil
	}

	dev, err := r.repo.FindByID(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("registry: find device %s: %w", deviceID, err)
	}

	if !dev.Enabled {
		return nil, fmt.Errorf("registry: device %s: %w", deviceID, ErrDeviceDisabled)
	}

	creds, err := r.vault.Get(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("registry: get credentials for %s: %w", deviceID, err)
	}

	factory, ok := r.factories[dev.DriverType]
	if !ok {
		return nil, fmt.Errorf("registry: driver_type %q: %w", dev.DriverType, ErrNoDriverFactory)
	}

	target := dev.ToTarget(creds)
	driver, err := factory(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("registry: connect device %s: %w", deviceID, err)
	}

	r.drivers[deviceID] = driver
	return driver, nil
}

// Close releases all cached driver connections. Safe to call more than once.
// Called at server shutdown so long-lived device connections don't leak.
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error
	for id, d := range r.drivers {
		if err := d.Close(); err != nil {
			errs = append(errs, fmt.Errorf("registry: close driver for %s: %w", id, err))
		}
	}
	r.drivers = make(map[string]port.DeviceDriver)
	return errors.Join(errs...)
}
