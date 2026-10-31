package device

import (
	"context"
	"fmt"
	"time"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// ActiveDeviceConfig controls active-router resolution.
type ActiveDeviceConfig struct {
	TTL                  time.Duration
	SingleDeviceFallback bool
}

// ActiveDeviceUseCase resolves and persists the operator's active router.
type ActiveDeviceUseCase struct {
	store      port.ActiveDeviceStore
	devices    port.DeviceRepository
	authorizer port.DeviceAuthorizer
	cfg        ActiveDeviceConfig
}

// NewActiveDeviceUseCase builds the resolver.
func NewActiveDeviceUseCase(store port.ActiveDeviceStore, devices port.DeviceRepository, authorizer port.DeviceAuthorizer, cfg ActiveDeviceConfig) *ActiveDeviceUseCase {
	if cfg.TTL <= 0 {
		cfg.TTL = 24 * time.Hour
	}
	return &ActiveDeviceUseCase{store: store, devices: devices, authorizer: authorizer, cfg: cfg}
}

// Set binds a router to the user/session after validating device access.
func (uc *ActiveDeviceUseCase) Set(ctx context.Context, tenantID, userID string, roles []string, sessionID, deviceID string) (port.ActiveDevice, error) {
	if deviceID == "" {
		return port.ActiveDevice{}, fmt.Errorf("device_id is required")
	}
	dev, err := uc.devices.FindByID(ctx, deviceID)
	if err != nil {
		return port.ActiveDevice{}, err
	}
	if !dev.Enabled {
		return port.ActiveDevice{}, fmt.Errorf("device %s is disabled", deviceID)
	}
	if uid, ok := parseUint(userID); ok && uc.authorizer != nil {
		allowed, aerr := uc.authorizer.CanAccessDevice(ctx, uid, roles, deviceID)
		if aerr != nil {
			return port.ActiveDevice{}, aerr
		}
		if !allowed {
			return port.ActiveDevice{}, fmt.Errorf("no access to device %s", deviceID)
		}
	}
	a := port.ActiveDevice{TenantID: tenantID, UserID: userID, SessionID: sessionID, DeviceID: deviceID, UpdatedAt: time.Now().UTC()}
	if err := uc.store.Set(ctx, a, uc.cfg.TTL); err != nil {
		return port.ActiveDevice{}, err
	}
	return a, nil
}

// Get returns the active binding (not resolved).
func (uc *ActiveDeviceUseCase) Get(ctx context.Context, tenantID, userID, sessionID string) (port.ActiveDevice, error) {
	return uc.store.Get(ctx, tenantID, userID, sessionID)
}

// Clear removes the binding.
func (uc *ActiveDeviceUseCase) Clear(ctx context.Context, tenantID, userID, sessionID string) error {
	return uc.store.Clear(ctx, tenantID, userID, sessionID)
}

// Resolve returns the active device ID using the documented priority:
// explicit device_id is handled by the caller; then session binding, then
// user binding, then optional single-device fallback. It never picks a random
// device.
func (uc *ActiveDeviceUseCase) Resolve(ctx context.Context, tenantID, userID, sessionID string) (string, error) {
	if sessionID != "" {
		if a, err := uc.store.Get(ctx, tenantID, userID, sessionID); err == nil && a.DeviceID != "" {
			return a.DeviceID, uc.validate(ctx, a.DeviceID)
		}
	}
	if a, err := uc.store.Get(ctx, tenantID, userID, ""); err == nil && a.DeviceID != "" {
		return a.DeviceID, uc.validate(ctx, a.DeviceID)
	}
	if uc.cfg.SingleDeviceFallback {
		devs, err := uc.devices.FindAll(ctx)
		if err == nil {
			var mikrotiks []device.Device
			for _, d := range devs {
				if d.Enabled && d.DriverType == "mikrotik" {
					mikrotiks = append(mikrotiks, d)
				}
			}
			if len(mikrotiks) == 1 {
				return mikrotiks[0].ID, nil
			}
		}
	}
	return "", fmt.Errorf("no active router selected: open Devices and select a router, or pass device_id")
}

func (uc *ActiveDeviceUseCase) validate(ctx context.Context, deviceID string) error {
	dev, err := uc.devices.FindByID(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("active device %s no longer exists: %w", deviceID, err)
	}
	if !dev.Enabled {
		return fmt.Errorf("active device %s is disabled", deviceID)
	}
	return nil
}

func parseUint(s string) (uint, bool) {
	var v uint
	if s == "" {
		return 0, false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		v = v*10 + uint(r-'0')
	}
	return v, true
}
