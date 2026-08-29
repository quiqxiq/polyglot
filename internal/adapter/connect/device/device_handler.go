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
	driverGetter DriverGetter
}

// NewDeviceConnectHandler constructs a device ConnectRPC handler.
func NewDeviceConnectHandler(
	uc *deviceUC.ManageDeviceUseCase,
	openTermUC *network.OpenTerminalUseCase,
	getter DriverGetter,
	metricsUC *deviceUC.ManageMetricsUseCase,
) *DeviceConnectHandler {
	return &DeviceConnectHandler{
		useCase:      uc,
		openTermUC:   openTermUC,
		driverGetter: getter,
		metricsUC:    metricsUC,
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
			return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("access to device %s denied", req.Msg.Id))
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
