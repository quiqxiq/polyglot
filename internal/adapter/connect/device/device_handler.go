package device

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"strings"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	iauth "github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
	deviceUC "github.com/quixiq/polyglot/internal/usecase/device"
	"github.com/quixiq/polyglot/internal/usecase/network"
	"github.com/quixiq/polyglot/pkg/response"
)

// DriverGetter fetches a connected port.DeviceDriver by device ID.
type DriverGetter func(ctx context.Context, deviceID string) (port.DeviceDriver, error)

// DeviceConnectHandler implements the device ConnectRPC service.
//
//nolint:revive // Explicit transport role is part of the project naming convention.
type DeviceConnectHandler struct {
	useCase      *deviceUC.ManageDeviceUseCase
	openTermUC   *network.OpenTerminalUseCase
	metricsUC    *deviceUC.ManageMetricsUseCase
	isolationUC  *deviceUC.ManageIsolationUseCase
	driverGetter DriverGetter
	streamGW     port.MonitorStreamGateway
}

// NewDeviceConnectHandler constructs a device ConnectRPC handler.
func NewDeviceConnectHandler(
	uc *deviceUC.ManageDeviceUseCase,
	openTermUC *network.OpenTerminalUseCase,
	getter DriverGetter,
	metricsUC *deviceUC.ManageMetricsUseCase,
	isolationUC *deviceUC.ManageIsolationUseCase,
	streamGW port.MonitorStreamGateway,
) *DeviceConnectHandler {
	return &DeviceConnectHandler{
		useCase:      uc,
		openTermUC:   openTermUC,
		driverGetter: getter,
		metricsUC:    metricsUC,
		isolationUC:  isolationUC,
		streamGW:     streamGW,
	}
}

// ListDevices returns devices visible to the caller.
func (h *DeviceConnectHandler) ListDevices(ctx context.Context, req *connect.Request[devicepb.ListDevicesRequest]) (*connect.Response[devicepb.ListDevicesResponse], error) {
	callerID, callerRoles, hasIdentity := iauth.IdentityFromContext(ctx)
	var devices []device.Device
	var err error
	if hasIdentity {
		devices, err = h.useCase.ListDevicesForUser(ctx, callerID, callerRoles)
	} else {
		devices, err = h.useCase.ListDevices(ctx)
	}
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	pbDevices := make([]*devicepb.Device, len(devices))
	for i, d := range devices {
		pbDevices[i] = DomainToPb(d)
	}

	return connect.NewResponse(&devicepb.ListDevicesResponse{Devices: pbDevices}), nil
}

// GetDevice returns one device by identifier.
func (h *DeviceConnectHandler) GetDevice(ctx context.Context, req *connect.Request[devicepb.GetDeviceRequest]) (*connect.Response[devicepb.GetDeviceResponse], error) {
	callerID, callerRoles, hasIdentity := iauth.IdentityFromContext(ctx)
	if hasIdentity && !isOwnerRole(callerRoles) {
		accessible, err := h.useCase.ListDevicesForUser(ctx, callerID, callerRoles)
		if err != nil {
			return nil, response.MapDomainError(err)
		}
		found := false
		for _, dev := range accessible {
			if dev.ID == req.Msg.Id {
				found = true
				break
			}
		}
		if !found {
			return nil, response.MapDomainError(device.ErrUnauthorized)
		}
	}

	d, err := h.useCase.GetDevice(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.GetDeviceResponse{Device: DomainToPb(d)}), nil
}

func isOwnerRole(roles []string) bool {
	for _, r := range roles {
		if strings.EqualFold(r, "owner") {
			return true
		}
	}
	return false
}

// UpdateDevice updates one device configuration.
func (h *DeviceConnectHandler) UpdateDevice(ctx context.Context, req *connect.Request[devicepb.UpdateDeviceRequest]) (*connect.Response[devicepb.UpdateDeviceResponse], error) {
	d := PbToDomain(req.Msg.Device)
	c := device.Credentials{
		Username: req.Msg.Username,
		Password: req.Msg.Password,
	}

	isNew := false
	if d.ID == "" {
		d.ID = uuid.NewString()
		isNew = true
	}

	if isNew {
		if err := h.useCase.CreateDevice(ctx, d, c); err != nil {
			return nil, response.MapDomainError(err)
		}
	} else {
		if err := h.useCase.UpdateDevice(ctx, d, c); err != nil {
			return nil, response.MapDomainError(err)
		}
	}

	updated, err := h.useCase.GetDevice(ctx, d.ID)
	if err != nil {
		updated = d
	}

	msg := "device updated successfully via ConnectRPC"
	if isNew {
		msg = "device created successfully via ConnectRPC"
	}

	return connect.NewResponse(&devicepb.UpdateDeviceResponse{
		Device:  DomainToPb(updated),
		Message: msg,
	}), nil
}

// DeleteDevice removes one device configuration.
func (h *DeviceConnectHandler) DeleteDevice(ctx context.Context, req *connect.Request[devicepb.DeleteDeviceRequest]) (*connect.Response[devicepb.DeleteDeviceResponse], error) {
	if err := h.useCase.DeleteDevice(ctx, req.Msg.Id); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.DeleteDeviceResponse{
		Message: "device deleted successfully",
	}), nil
}

// TestDeviceConnection checks connectivity to a device.
func (h *DeviceConnectHandler) TestDeviceConnection(ctx context.Context, req *connect.Request[devicepb.TestDeviceConnectionRequest]) (*connect.Response[devicepb.TestDeviceConnectionResponse], error) {
	var drv port.DeviceDriver
	if h.driverGetter != nil {
		d, err := h.driverGetter(ctx, req.Msg.Id)
		if err != nil {
			return connect.NewResponse(&devicepb.TestDeviceConnectionResponse{
				DeviceId: req.Msg.Id,
				Status:   "failed",
				Message:  fmt.Sprintf("Failed to connect to device: %v", err),
				Success:  false,
			}), nil
		}
		drv = d
	}

	res, err := h.useCase.TestConnection(
		ctx,
		drv,
		req.Msg.Id,
		req.Msg.SelectedInterface,
		req.Msg.InterfaceTypeFilter,
		req.Msg.InterfaceNameFilter,
	)
	if err != nil {
		return connect.NewResponse(&devicepb.TestDeviceConnectionResponse{
			DeviceId: req.Msg.Id,
			Status:   "error",
			Message:  err.Error(),
			Success:  false,
		}), nil
	}

	pbIfaces := ToProtoInterfaceDetails(res.InterfaceDetails)

	return connect.NewResponse(&devicepb.TestDeviceConnectionResponse{
		DeviceId:      req.Msg.Id,
		Status:        res.Status,
		LatencyMs:     res.LatencyMS,
		Uptime:        res.Uptime,
		Version:       res.Version,
		BoardName:     res.BoardName,
		Identity:      res.Identity,
		CpuLoad:       int32(res.CPULoad),
		FreeMemory:    res.FreeMemory,
		TotalMemory:   res.TotalMemory,
		Interfaces:    res.Interfaces,
		InterfaceList: pbIfaces,
		Success:       true,
		Message:       res.Message,
	}), nil
}

// GetIsolationStatus retrieves the router isolation infrastructure status.
func (h *DeviceConnectHandler) GetIsolationStatus(ctx context.Context, req *connect.Request[devicepb.GetIsolationStatusRequest]) (*connect.Response[devicepb.GetIsolationStatusResponse], error) {
	if h.isolationUC == nil {
		return nil, response.Unavailable("isolation usecase unavailable")
	}
	status, err := h.isolationUC.GetIsolationStatus(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetIsolationStatusResponse{
		Status: IsolationStatusToPb(status),
	}), nil
}

// CreateIsolationProfile provisions isolation profiles on the target router.
func (h *DeviceConnectHandler) CreateIsolationProfile(ctx context.Context, req *connect.Request[devicepb.CreateIsolationProfileRequest]) (*connect.Response[devicepb.CreateIsolationProfileResponse], error) {
	if h.isolationUC == nil {
		return nil, response.Unavailable("isolation usecase unavailable")
	}
	cfg := IsolationConfigToDomain(req.Msg.Config)
	status, err := h.isolationUC.CreateIsolationProfile(ctx, req.Msg.DeviceId, cfg)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.CreateIsolationProfileResponse{
		Status:  IsolationStatusToPb(status),
		Message: "Profil isolir dan firewall redirect berhasil dipasang di router",
	}), nil
}

// UpdateIsolationProfile updates isolation profiles and configuration on the target router.
func (h *DeviceConnectHandler) UpdateIsolationProfile(ctx context.Context, req *connect.Request[devicepb.UpdateIsolationProfileRequest]) (*connect.Response[devicepb.UpdateIsolationProfileResponse], error) {
	if h.isolationUC == nil {
		return nil, response.Unavailable("isolation usecase unavailable")
	}
	cfg := IsolationConfigToDomain(req.Msg.Config)
	status, err := h.isolationUC.UpdateIsolationProfile(ctx, req.Msg.DeviceId, cfg)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.UpdateIsolationProfileResponse{
		Status:  IsolationStatusToPb(status),
		Message: "Profil isolir berhasil diperbarui",
	}), nil
}

// DeleteIsolationProfile removes the isolation profile from the target router.
func (h *DeviceConnectHandler) DeleteIsolationProfile(ctx context.Context, req *connect.Request[devicepb.DeleteIsolationProfileRequest]) (*connect.Response[devicepb.DeleteIsolationProfileResponse], error) {
	if h.isolationUC == nil {
		return nil, response.Unavailable("isolation usecase unavailable")
	}
	if err := h.isolationUC.DeleteIsolationProfile(ctx, req.Msg.DeviceId, req.Msg.RemoveFirewallRules); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.DeleteIsolationProfileResponse{
		Message: "Profil isolir berhasil dihapus",
	}), nil
}

// GetRouterIntegrationScript generates webhook integration scripts for the router.
func (h *DeviceConnectHandler) GetRouterIntegrationScript(ctx context.Context, req *connect.Request[devicepb.GetRouterIntegrationScriptRequest]) (*connect.Response[devicepb.GetRouterIntegrationScriptResponse], error) {
	if h.isolationUC == nil {
		return nil, response.Unavailable("isolation usecase unavailable")
	}
	scripts, err := h.isolationUC.GetRouterIntegrationScript(ctx, req.Msg.DeviceId, req.Msg.ServiceType, req.Msg.WebhookUrl)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetRouterIntegrationScriptResponse{
		PppOnUpScript:         scripts.PPPOnUpScript,
		PppOnDownScript:       scripts.PPPOnDownScript,
		HotspotOnLoginScript:  scripts.HotspotOnLoginScript,
		HotspotOnLogoutScript: scripts.HotspotOnLogoutScript,
		WebhookToken:          scripts.WebhookToken,
	}), nil
}

// ApplyRouterIntegrationScript applies webhook integration scripts to a profile on the router.
func (h *DeviceConnectHandler) ApplyRouterIntegrationScript(ctx context.Context, req *connect.Request[devicepb.ApplyRouterIntegrationScriptRequest]) (*connect.Response[devicepb.ApplyRouterIntegrationScriptResponse], error) {
	if h.isolationUC == nil {
		return nil, response.Unavailable("isolation usecase unavailable")
	}
	if err := h.isolationUC.ApplyRouterIntegrationScript(ctx, req.Msg.DeviceId, req.Msg.ProfileName, req.Msg.ServiceType); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ApplyRouterIntegrationScriptResponse{
		Message: fmt.Sprintf("Script webhook berhasil diterapkan pada profile %s", req.Msg.ProfileName),
	}), nil
}
