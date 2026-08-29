package device

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// DeviceInterfaceDetail represents structured details for a device interface.
type DeviceInterfaceDetail struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Disabled   bool   `json:"disabled"`
	Running    bool   `json:"running"`
	RxBps      int64  `json:"rx_bps"`
	TxBps      int64  `json:"tx_bps"`
	MACAddress string `json:"mac_address"`
}

// DeviceTestResult holds live connection test metrics for a device.
type DeviceTestResult struct {
	DeviceID         string                  `json:"device_id"`
	Status           string                  `json:"status"` // "connected" or "failed"
	LatencyMS        int64                   `json:"latency_ms"`
	Uptime           string                  `json:"uptime,omitempty"`
	Version          string                  `json:"version,omitempty"`
	BoardName        string                  `json:"board_name,omitempty"`
	Identity         string                  `json:"identity,omitempty"`
	CPULoad          int                     `json:"cpu_load,omitempty"`
	FreeMemory       int64                   `json:"free_memory,omitempty"`
	TotalMemory      int64                   `json:"total_memory,omitempty"`
	Interfaces       []string                `json:"interfaces,omitempty"`
	InterfaceDetails []DeviceInterfaceDetail `json:"interface_details,omitempty"`
	RxBps            int64                   `json:"rx_bps,omitempty"`
	TxBps            int64                   `json:"tx_bps,omitempty"`
	Message          string                  `json:"message,omitempty"`
}

// DriverEvicter defines the contract for evicting cached driver connections.
type DriverEvicter interface {
	Evict(deviceID string) error
}

// device.ErrDiagnosticsUnconfigured indicates TestConnection was called without a
// port.DeviceDiagnostics gateway — inventory CRUD still works without it.

// ManageDeviceUseCase manages device inventory CRUD and live connectivity testing.
type ManageDeviceUseCase struct {
	repo    port.DeviceRepository
	vault   port.CredentialVault
	evicter DriverEvicter
	diag    port.DeviceDiagnostics
}

// NewManageDeviceUseCase constructs a new ManageDeviceUseCase. diag may be
// nil when only inventory CRUD is needed; TestConnection then returns
// device.ErrDiagnosticsUnconfigured instead of panicking.
func NewManageDeviceUseCase(repo port.DeviceRepository, vault port.CredentialVault, evicter DriverEvicter, diag port.DeviceDiagnostics) *ManageDeviceUseCase {
	return &ManageDeviceUseCase{
		repo:    repo,
		vault:   vault,
		evicter: evicter,
		diag:    diag,
	}
}

// CreateDevice saves device inventory data and encrypts/stores its credentials in vault.
func (uc *ManageDeviceUseCase) CreateDevice(ctx context.Context, d device.Device, c device.Credentials) error {
	if d.Host == "" {
		return device.ErrInvalidInput
	}
	if err := uc.repo.Save(ctx, d); err != nil {
		return fmt.Errorf("failed to save device: %w", err)
	}
	if err := uc.vault.Save(ctx, d.ID, c); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}
	if uc.evicter != nil {
		_ = uc.evicter.Evict(d.ID) // best-effort: driver cache dibersihkan; kegagalan evict tidak membatalkan penyimpanan
	}
	return nil
}

// GetDevice fetches a single device inventory record by ID.
func (uc *ManageDeviceUseCase) GetDevice(ctx context.Context, id string) (device.Device, error) {
	result, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return device.Device{}, fmt.Errorf("find device %s: %w", id, err)
	}
	return result, nil
}

// ListDevices returns all registered device inventory records.
func (uc *ManageDeviceUseCase) ListDevices(ctx context.Context) ([]device.Device, error) {
	devices, err := uc.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	return devices, nil
}

// ListDevicesForUser returns device inventory records accessible to the specified user.
// Users with role 'owner' get all devices, while non-owners get only assigned devices.
func (uc *ManageDeviceUseCase) ListDevicesForUser(ctx context.Context, userID uint, userRoles []string) ([]device.Device, error) {
	for _, r := range userRoles {
		if strings.EqualFold(r, "owner") {
			devices, err := uc.repo.FindAll(ctx)
			if err != nil {
				return nil, fmt.Errorf("list devices for owner: %w", err)
			}
			return devices, nil
		}
	}
	if userID == 0 {
		return []device.Device{}, nil
	}
	return uc.repo.FindByUserScope(ctx, userID)
}

// UpdateDevice updates a device inventory record and its credentials.
func (uc *ManageDeviceUseCase) UpdateDevice(ctx context.Context, d device.Device, c device.Credentials) error {
	if err := uc.repo.Update(ctx, d); err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}
	if c.Username != "" || c.Password != "" || len(c.Extra) > 0 {
		existingCreds, err := uc.vault.Get(ctx, d.ID)
		if err == nil {
			if c.Password == "" {
				c.Password = existingCreds.Password
			}
			if c.Username == "" {
				c.Username = existingCreds.Username
			}
			if c.Extra == nil {
				c.Extra = existingCreds.Extra
			}
		}
		if err := uc.vault.Save(ctx, d.ID, c); err != nil {
			return fmt.Errorf("failed to update credentials: %w", err)
		}
	}
	if uc.evicter != nil {
		_ = uc.evicter.Evict(d.ID) // best-effort: driver cache dibersihkan; kegagalan evict tidak membatalkan update
	}
	return nil
}

// DeleteDevice deletes a device inventory record.
func (uc *ManageDeviceUseCase) DeleteDevice(ctx context.Context, id string) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}
	if uc.evicter != nil {
		_ = uc.evicter.Evict(id) // best-effort: driver cache dibersihkan; kegagalan evict tidak membatalkan penghapusan
	}
	return nil
}

// TestConnection executes diagnostic checks against a live driver instance to verify connectivity and latency.
func (uc *ManageDeviceUseCase) TestConnection(
	ctx context.Context,
	driver port.DeviceDriver,
	deviceID string,
	selectedIface string,
	typeFilter string,
	nameFilter string,
) (DeviceTestResult, error) {
	if uc.diag == nil {
		return DeviceTestResult{}, device.ErrDiagnosticsUnconfigured
	}
	start := time.Now()

	// Execute /system/resource/print to test device response
	sysRes, err := uc.diag.GetSystemResource(ctx, driver)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return DeviceTestResult{
			DeviceID:  deviceID,
			Status:    "failed",
			LatencyMS: latency,
			Message:   err.Error(),
		}, nil
	}

	freeMem, _ := strconv.ParseInt(sysRes.FreeMemory, 10, 64)
	totalMem, _ := strconv.ParseInt(sysRes.TotalMemory, 10, 64)

	result := DeviceTestResult{
		DeviceID:    deviceID,
		Status:      "connected",
		LatencyMS:   latency,
		Uptime:      sysRes.Uptime,
		Version:     sysRes.Version,
		BoardName:   sysRes.BoardName,
		CPULoad:     sysRes.CPULoad,
		FreeMemory:  freeMem,
		TotalMemory: totalMem,
		Message:     "Connection test successful",
	}

	// Try fetching system identity if available
	if ident, err := uc.diag.GetSystemIdentity(ctx, driver); err == nil {
		result.Identity = ident
	}

	// Fetch interface list if available
	if ifaces, err := uc.diag.ListInterfaces(ctx, driver, typeFilter, nameFilter); err == nil {
		for _, ifc := range ifaces {
			result.Interfaces = append(result.Interfaces, ifc.Name)
			result.InterfaceDetails = append(result.InterfaceDetails, DeviceInterfaceDetail{
				Name:       ifc.Name,
				Type:       ifc.Type,
				Disabled:   ifc.Disabled,
				Running:    ifc.Running,
				MACAddress: ifc.MACAddress,
			})
		}
	}

	// Fetch selected interface traffic rates if specified
	if selectedIface != "" && selectedIface != "default" {
		if stats, err := uc.diag.MonitorTrafficOnce(ctx, driver, selectedIface); err == nil {
			rx, _ := strconv.ParseInt(stats.RxBitsPerSecond, 10, 64)
			tx, _ := strconv.ParseInt(stats.TxBitsPerSecond, 10, 64)
			result.RxBps = rx
			result.TxBps = tx
		}
	}

	return result, nil
}
