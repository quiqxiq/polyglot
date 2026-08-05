package business

import (
	"context"
	"fmt"
	"time"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
)

// DeviceTestResult holds live connection test metrics for a device.
type DeviceTestResult struct {
	DeviceID  string `json:"device_id"`
	Status    string `json:"status"` // "connected" or "failed"
	LatencyMS int64  `json:"latency_ms"`
	Uptime    string `json:"uptime,omitempty"`
	Version   string `json:"version,omitempty"`
	BoardName string `json:"board_name,omitempty"`
	Identity  string `json:"identity,omitempty"`
	Message   string `json:"message,omitempty"`
}

// ManageDeviceUseCase manages device inventory CRUD and live connectivity testing.
type ManageDeviceUseCase struct {
	repo  port.DeviceRepository
	vault port.CredentialVault
}

// NewManageDeviceUseCase constructs a new ManageDeviceUseCase.
func NewManageDeviceUseCase(repo port.DeviceRepository, vault port.CredentialVault) *ManageDeviceUseCase {
	return &ManageDeviceUseCase{
		repo:  repo,
		vault: vault,
	}
}

// CreateDevice saves device inventory data and encrypts/stores its credentials in vault.
func (uc *ManageDeviceUseCase) CreateDevice(ctx context.Context, d device.Device, c device.Credentials) error {
	if d.ID == "" {
		return fmt.Errorf("device ID cannot be empty")
	}
	if d.Host == "" {
		return fmt.Errorf("device host cannot be empty")
	}
	if err := uc.repo.Save(ctx, d); err != nil {
		return fmt.Errorf("failed to save device: %w", err)
	}
	if err := uc.vault.Save(ctx, d.ID, c); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}
	return nil
}

// GetDevice fetches a single device inventory record by ID.
func (uc *ManageDeviceUseCase) GetDevice(ctx context.Context, id string) (device.Device, error) {
	return uc.repo.FindByID(ctx, id)
}

// ListDevices returns all registered device inventory records.
func (uc *ManageDeviceUseCase) ListDevices(ctx context.Context) ([]device.Device, error) {
	return uc.repo.FindAll(ctx)
}

// UpdateDevice updates a device inventory record and its credentials.
func (uc *ManageDeviceUseCase) UpdateDevice(ctx context.Context, d device.Device, c device.Credentials) error {
	if err := uc.repo.Update(ctx, d); err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}
	if c.Username != "" || c.Password != "" || len(c.Extra) > 0 {
		if err := uc.vault.Save(ctx, d.ID, c); err != nil {
			return fmt.Errorf("failed to update credentials: %w", err)
		}
	}
	return nil
}

// DeleteDevice deletes a device inventory record.
func (uc *ManageDeviceUseCase) DeleteDevice(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

// TestConnection executes diagnostic checks against a live driver instance to verify connectivity and latency.
func (uc *ManageDeviceUseCase) TestConnection(ctx context.Context, driver port.DeviceDriver, deviceID string) (DeviceTestResult, error) {
	start := time.Now()

	// Execute /system/resource/print to test device response
	cmd := mikrotik.NewPrintSystemResourceCommand()
	res, err := driver.Execute(ctx, cmd)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return DeviceTestResult{
			DeviceID:  deviceID,
			Status:    "failed",
			LatencyMS: latency,
			Message:   err.Error(),
		}, nil
	}

	sysRes := mikrotik.ParseSystemResource(res)
	result := DeviceTestResult{
		DeviceID:  deviceID,
		Status:    "connected",
		LatencyMS: latency,
		Uptime:    sysRes.Uptime,
		Version:   sysRes.Version,
		BoardName: sysRes.BoardName,
		Message:   "Connection test successful",
	}

	// Try fetching system identity if available
	identCmd := mikrotik.NewPrintSystemIdentityCommand()
	if identRes, err := driver.Execute(ctx, identCmd); err == nil && len(identRes.Rows) > 0 {
		result.Identity = identRes.Rows[0]["name"]
	}

	return result, nil
}
