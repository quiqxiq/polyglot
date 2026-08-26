package device

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

// ManageMetricsUseCase orchestrates telemetry settings and historical time-series queries.
type ManageMetricsUseCase struct {
	deviceRepo  port.DeviceRepository
	metricsRepo port.MetricsRepository
	authorizer  port.DeviceAuthorizer
}

// NewManageMetricsUseCase creates a new ManageMetricsUseCase instance.
func NewManageMetricsUseCase(
	deviceRepo port.DeviceRepository,
	metricsRepo port.MetricsRepository,
	authorizer port.DeviceAuthorizer,
) *ManageMetricsUseCase {
	return &ManageMetricsUseCase{
		deviceRepo:  deviceRepo,
		metricsRepo: metricsRepo,
		authorizer:  authorizer,
	}
}

// isDeviceAuthorized verifies if the caller has access to the device.
func (uc *ManageMetricsUseCase) isDeviceAuthorized(ctx context.Context, deviceID string, callerID uint, callerRoles []string) (bool, error) {
	for _, r := range callerRoles {
		if strings.EqualFold(r, "owner") {
			return true, nil
		}
	}
	if uc.authorizer == nil {
		return true, nil
	}
	return uc.authorizer.CanAccessDevice(ctx, callerID, callerRoles, deviceID)
}

// GetPingConfig retrieves the ping telemetry configuration for a device.
func (uc *ManageMetricsUseCase) GetPingConfig(
	ctx context.Context,
	deviceID string,
	callerID uint,
	callerRoles []string,
) (device.DevicePingConfig, bool, error) {
	if deviceID == "" {
		return device.DefaultPingConfig(), false, errors.New("device_id is required")
	}

	ok, err := uc.isDeviceAuthorized(ctx, deviceID, callerID, callerRoles)
	if err != nil {
		return device.DefaultPingConfig(), false, fmt.Errorf("check device access: %w", err)
	}
	if !ok {
		return device.DefaultPingConfig(), false, device.ErrUnauthorized
	}

	dev, err := uc.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		return device.DefaultPingConfig(), false, err
	}

	timescaleAvailable := false
	if uc.metricsRepo != nil {
		timescaleAvailable, _ = uc.metricsRepo.IsTimescaleDBAvailable(ctx)
	}

	return dev.PingConfig(), timescaleAvailable, nil
}

// UpdatePingConfig updates the ping telemetry settings for a device.
func (uc *ManageMetricsUseCase) UpdatePingConfig(
	ctx context.Context,
	deviceID string,
	cfg device.DevicePingConfig,
	callerID uint,
	callerRoles []string,
) (device.DevicePingConfig, error) {
	if deviceID == "" {
		return device.DefaultPingConfig(), errors.New("device_id is required")
	}

	ok, err := uc.isDeviceAuthorized(ctx, deviceID, callerID, callerRoles)
	if err != nil {
		return device.DefaultPingConfig(), fmt.Errorf("check device access: %w", err)
	}
	if !ok {
		return device.DefaultPingConfig(), device.ErrUnauthorized
	}

	dev, err := uc.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		return device.DefaultPingConfig(), err
	}

	dev.SetPingConfig(cfg)
	if err := uc.deviceRepo.Update(ctx, dev); err != nil {
		return device.DefaultPingConfig(), fmt.Errorf("update device ping config: %w", err)
	}

	logger.WithComponent("MetricsUseCase").WithFields(map[string]any{
		"actor_id":       callerID,
		"device_id":      deviceID,
		"ping_enabled":   cfg.Enabled,
		"ping_target":    cfg.Target,
		"retention_days": cfg.RetentionDays,
	}).Info("device ping configuration updated")

	return dev.PingConfig(), nil
}

// QueryPingMetrics fetches historical time-series ping telemetry within the filter window.
func (uc *ManageMetricsUseCase) QueryPingMetrics(
	ctx context.Context,
	filter device.PingMetricsFilter,
	callerID uint,
	callerRoles []string,
) ([]device.PingMetricPoint, device.PingSummary, bool, error) {
	if filter.DeviceID == "" {
		return nil, device.PingSummary{}, false, errors.New("device_id is required")
	}

	ok, err := uc.isDeviceAuthorized(ctx, filter.DeviceID, callerID, callerRoles)
	if err != nil {
		return nil, device.PingSummary{}, false, fmt.Errorf("check device access: %w", err)
	}
	if !ok {
		return nil, device.PingSummary{}, false, device.ErrUnauthorized
	}

	if uc.metricsRepo == nil {
		return nil, device.PingSummary{}, false, errors.New("metrics repository is not configured")
	}

	timescaleAvailable, err := uc.metricsRepo.IsTimescaleDBAvailable(ctx)
	if err != nil || !timescaleAvailable {
		return nil, device.PingSummary{}, false, nil
	}

	points, summary, err := uc.metricsRepo.QueryPingMetrics(ctx, filter)
	if err != nil {
		return nil, device.PingSummary{}, true, err
	}

	return points, summary, true, nil
}

// SavePingMetric stores a ping telemetry frame to the metrics repository.
func (uc *ManageMetricsUseCase) SavePingMetric(ctx context.Context, point device.PingMetricPoint) error {
	if uc.metricsRepo == nil {
		return nil
	}
	return uc.metricsRepo.SavePingMetric(ctx, point)
}
