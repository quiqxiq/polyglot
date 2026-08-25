package network

import (
	"context"
	"fmt"

	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

// IsolirConfigOverride carries optional per-request overrides for the
// isolation infrastructure. Empty fields fall back to system_settings.
type IsolirConfigOverride struct {
	ProfileName    string
	PoolName       string
	PoolRange      string
	PortalIP       string
	PortalHTTPPort string
	RedirectPorts  string
}

// ManageIsolationUseCase exposes explicit device-level control over the
// isolation infrastructure (pool, suspended profile, payment-portal
// redirect rules). Setup/Remove are idempotent; Status is read-only.
// Subscriptions do not need this directly — IsolateSubscription ensures
// the same infrastructure lazily before suspending a PPPoE subscriber.
type ManageIsolationUseCase struct {
	settings  port.SettingRepository
	isolir    port.IsolationProvisioner
	driverFor func(ctx context.Context, deviceID string) (port.DeviceDriver, error)
}

func NewManageIsolationUseCase(
	settings port.SettingRepository,
	isolir port.IsolationProvisioner,
	driverFor func(ctx context.Context, deviceID string) (port.DeviceDriver, error),
) *ManageIsolationUseCase {
	return &ManageIsolationUseCase{settings: settings, isolir: isolir, driverFor: driverFor}
}

// Setup ensures the full infrastructure exists on the target router and
// returns the effective configuration used.
func (uc *ManageIsolationUseCase) Setup(ctx context.Context, deviceID string, override IsolirConfigOverride) (port.IsolirSetupResult, port.IsolirConfig, error) {
	cfg := uc.effectiveConfig(ctx, override)
	driver, err := uc.driverFor(ctx, deviceID)
	if err != nil {
		return port.IsolirSetupResult{}, cfg, fmt.Errorf("connect to device %s: %w", deviceID, err)
	}
	res, err := uc.isolir.EnsureIsolirInfrastructure(ctx, driver, cfg)
	if err != nil {
		return port.IsolirSetupResult{}, cfg, fmt.Errorf("setup isolation on %s: %w", deviceID, err)
	}
	logger.WithComponent("ManageIsolation").WithFields(map[string]any{
		"device_id": deviceID, "pool": cfg.PoolName, "profile": cfg.ProfileName,
	}).Info("isolation infrastructure ensured")
	return res, cfg, nil
}

// Status inspects the router and returns the health snapshot plus the
// effective configuration and human-readable warnings.
func (uc *ManageIsolationUseCase) Status(ctx context.Context, deviceID string) (port.IsolirInspection, port.IsolirConfig, []string, error) {
	cfg := uc.effectiveConfig(ctx, IsolirConfigOverride{})
	driver, err := uc.driverFor(ctx, deviceID)
	if err != nil {
		return port.IsolirInspection{}, cfg, nil, fmt.Errorf("connect to device %s: %w", deviceID, err)
	}
	ins, err := uc.isolir.InspectIsolirInfrastructure(ctx, driver, cfg)
	if err != nil {
		return ins, cfg, nil, fmt.Errorf("inspect isolation on %s: %w", deviceID, err)
	}

	warnings := make([]string, 0)
	if !ins.PoolExists {
		warnings = append(warnings, fmt.Sprintf("pool %q belum ada di router", cfg.PoolName))
	}
	if !ins.ProfileExists {
		warnings = append(warnings, fmt.Sprintf("profile isolir %q belum ada di router", cfg.ProfileName))
	}
	for _, r := range ins.NATRules {
		if !r.Exists {
			warnings = append(warnings, fmt.Sprintf("rule redirect port %s belum ada", r.Port))
		}
	}
	if cfg.PortalIP == "" {
		warnings = append(warnings, "isolir.portal_ip belum diisi — redirect hanya mengganti port, belum menuju portal pembayaran")
	}
	return ins, cfg, warnings, nil
}

// Remove tears down NAT redirect rules, the suspended profile, and the
// isolation pool on the target router.
func (uc *ManageIsolationUseCase) Remove(ctx context.Context, deviceID string) error {
	cfg := uc.effectiveConfig(ctx, IsolirConfigOverride{})
	driver, err := uc.driverFor(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("connect to device %s: %w", deviceID, err)
	}
	if err := uc.isolir.RemoveIsolirInfrastructure(ctx, driver, cfg); err != nil {
		return fmt.Errorf("remove isolation on %s: %w", deviceID, err)
	}
	logger.WithComponent("ManageIsolation").WithField("device_id", deviceID).
		Info("isolation infrastructure removed")
	return nil
}

// effectiveConfig merges system_settings values with non-empty overrides.
func (uc *ManageIsolationUseCase) effectiveConfig(ctx context.Context, o IsolirConfigOverride) port.IsolirConfig {
	return isolirConfigFromSettings(ctx, uc.settings, o)
}

// isolirConfigFromSettings is the single source for the "isolir.*" defaults
// and override merge — shared by the isolation endpoint and the lazy
// per-subscription isolation path.
func isolirConfigFromSettings(ctx context.Context, settings port.SettingRepository, o IsolirConfigOverride) port.IsolirConfig {
	get := func(key, fallback string) string {
		if settings == nil {
			return fallback
		}
		return settings.GetValue(ctx, key, fallback)
	}
	cfg := port.IsolirConfig{
		ProfileName:    get("isolir.profile_name", "isolir"),
		PoolName:       get("isolir.pool_name", "pool-isolir"),
		PoolRange:      get("isolir.pool_range", "172.16.99.10-172.16.99.254"),
		PortalIP:       get("isolir.portal_ip", ""),
		PortalHTTPPort: get("isolir.portal_http_port", "8080"),
		RedirectPorts:  get("isolir.redirect_ports", "80,443"),
	}
	if o.ProfileName != "" {
		cfg.ProfileName = o.ProfileName
	}
	if o.PoolName != "" {
		cfg.PoolName = o.PoolName
	}
	if o.PoolRange != "" {
		cfg.PoolRange = o.PoolRange
	}
	if o.PortalIP != "" {
		cfg.PortalIP = o.PortalIP
	}
	if o.PortalHTTPPort != "" {
		cfg.PortalHTTPPort = o.PortalHTTPPort
	}
	if o.RedirectPorts != "" {
		cfg.RedirectPorts = o.RedirectPorts
	}
	return cfg
}
