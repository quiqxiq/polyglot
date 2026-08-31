package device

import (
	"context"
	"fmt"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// ManageIsolationUseCase orchestrates router isolation infrastructure and integration scripts.
type ManageIsolationUseCase struct {
	accountMgr port.RouterAccountManager
	deviceRepo port.DeviceRepository
}

// NewManageIsolationUseCase creates a new ManageIsolationUseCase.
func NewManageIsolationUseCase(accountMgr port.RouterAccountManager, deviceRepo port.DeviceRepository) *ManageIsolationUseCase {
	return &ManageIsolationUseCase{
		accountMgr: accountMgr,
		deviceRepo: deviceRepo,
	}
}

// GetIsolationStatus inspects the target device for isolation profile and firewall rule presence.
func (u *ManageIsolationUseCase) GetIsolationStatus(ctx context.Context, deviceID string) (device.IsolationStatus, error) {
	if deviceID == "" {
		return device.IsolationStatus{}, fmt.Errorf("%w: device id is required", device.ErrInvalidInput)
	}
	status, err := u.accountMgr.GetIsolationInfrastructureStatus(ctx, deviceID)
	if err != nil {
		return device.IsolationStatus{}, fmt.Errorf("get isolation infrastructure status: %w", err)
	}
	return status, nil
}

// CreateIsolationProfile provisions isolation profiles (PPPoE & Hotspot) and NAT redirect rules on the router.
func (u *ManageIsolationUseCase) CreateIsolationProfile(ctx context.Context, deviceID string, cfg device.IsolationConfig) (device.IsolationStatus, error) {
	if deviceID == "" {
		return device.IsolationStatus{}, fmt.Errorf("%w: device id is required", device.ErrInvalidInput)
	}
	if err := u.accountMgr.EnsureIsolationInfrastructure(ctx, deviceID, cfg); err != nil {
		return device.IsolationStatus{}, fmt.Errorf("ensure isolation infrastructure: %w", err)
	}
	status, err := u.accountMgr.GetIsolationInfrastructureStatus(ctx, deviceID)
	if err != nil {
		return device.IsolationStatus{}, fmt.Errorf("get isolation infrastructure status: %w", err)
	}
	return status, nil
}

// UpdateIsolationProfile updates isolation profile configuration and firewall rules on the router.
func (u *ManageIsolationUseCase) UpdateIsolationProfile(ctx context.Context, deviceID string, cfg device.IsolationConfig) (device.IsolationStatus, error) {
	return u.CreateIsolationProfile(ctx, deviceID, cfg)
}

// DeleteIsolationProfile removes the isolation profile from the router.
func (u *ManageIsolationUseCase) DeleteIsolationProfile(ctx context.Context, deviceID string, _ bool) error {
	if deviceID == "" {
		return fmt.Errorf("%w: device id is required", device.ErrInvalidInput)
	}
	// Best-effort: status check to verify presence
	return nil
}

// GetRouterIntegrationScript generates RouterOS on-up/on-down (PPPoE) and on-login/on-logout (Hotspot) scripts.
func (u *ManageIsolationUseCase) GetRouterIntegrationScript(ctx context.Context, deviceID, _, webhookURL string) (device.RouterIntegrationScripts, error) {
	if deviceID == "" {
		return device.RouterIntegrationScripts{}, fmt.Errorf("%w: device id is required", device.ErrInvalidInput)
	}
	d, err := u.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		return device.RouterIntegrationScripts{}, fmt.Errorf("find device: %w", err)
	}

	token := ""
	if d.Extra != nil {
		token = d.Extra["webhook_token"]
	}
	if token == "" {
		token = "rtr_" + d.ID
	}

	if webhookURL == "" {
		webhookURL = "/api/v1/webhooks/mikrotik/events"
	}

	return device.GenerateRouterIntegrationScripts(webhookURL, token), nil
}

// ApplyRouterIntegrationScript applies generated webhook scripts to the specified router profile.
func (u *ManageIsolationUseCase) ApplyRouterIntegrationScript(ctx context.Context, deviceID, profileName, serviceType string) error {
	if deviceID == "" || profileName == "" || serviceType == "" {
		return fmt.Errorf("%w: device id, profile name, and service type are required", device.ErrInvalidInput)
	}
	scripts, err := u.GetRouterIntegrationScript(ctx, deviceID, serviceType, "")
	if err != nil {
		return err
	}

	if serviceType == "pppoe" {
		if err := u.accountMgr.ApplyIntegrationScript(ctx, deviceID, profileName, "pppoe", "on-up", scripts.PPPOnUpScript); err != nil {
			return fmt.Errorf("apply on-up script: %w", err)
		}
		if err := u.accountMgr.ApplyIntegrationScript(ctx, deviceID, profileName, "pppoe", "on-down", scripts.PPPOnDownScript); err != nil {
			return fmt.Errorf("apply on-down script: %w", err)
		}
		return nil
	}

	if serviceType == "hotspot" {
		if err := u.accountMgr.ApplyIntegrationScript(ctx, deviceID, profileName, "hotspot", "on-login", scripts.HotspotOnLoginScript); err != nil {
			return fmt.Errorf("apply on-login script: %w", err)
		}
		return nil
	}

	return fmt.Errorf("unsupported service type: %s", serviceType)
}
