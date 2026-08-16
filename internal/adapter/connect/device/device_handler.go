package device

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
	deviceUC "github.com/quixiq/polyglot/internal/usecase/device"
	"github.com/quixiq/polyglot/internal/usecase/network"
	"github.com/quixiq/polyglot/pkg/response"
)

// DriverGetter fetches a connected port.DeviceDriver by device ID.
type DriverGetter func(ctx context.Context, deviceID string) (port.DeviceDriver, error)

type DeviceConnectHandler struct {
	useCase      *deviceUC.ManageDeviceUseCase
	openTermUC   *network.OpenTerminalUseCase
	driverGetter DriverGetter
}

func NewDeviceConnectHandler(uc *deviceUC.ManageDeviceUseCase, openTermUC *network.OpenTerminalUseCase, getter DriverGetter) *DeviceConnectHandler {
	return &DeviceConnectHandler{
		useCase:      uc,
		openTermUC:   openTermUC,
		driverGetter: getter,
	}
}

func (h *DeviceConnectHandler) ListDevices(ctx context.Context, req *connect.Request[devicepb.ListDevicesRequest]) (*connect.Response[devicepb.ListDevicesResponse], error) {
	devices, err := h.useCase.ListDevices(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	pbDevices := make([]*devicepb.Device, len(devices))
	for i, d := range devices {
		pbDevices[i] = DomainToPb(d)
	}

	return connect.NewResponse(&devicepb.ListDevicesResponse{Devices: pbDevices}), nil
}

func (h *DeviceConnectHandler) GetDevice(ctx context.Context, req *connect.Request[devicepb.GetDeviceRequest]) (*connect.Response[devicepb.GetDeviceResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device id is required"))
	}

	d, err := h.useCase.GetDevice(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.GetDeviceResponse{Device: DomainToPb(d)}), nil
}

func (h *DeviceConnectHandler) UpdateDevice(ctx context.Context, req *connect.Request[devicepb.UpdateDeviceRequest]) (*connect.Response[devicepb.UpdateDeviceResponse], error) {
	if req.Msg.Device == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device payload is required"))
	}

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

func (h *DeviceConnectHandler) DeleteDevice(ctx context.Context, req *connect.Request[devicepb.DeleteDeviceRequest]) (*connect.Response[devicepb.DeleteDeviceResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device id is required"))
	}

	if err := h.useCase.DeleteDevice(ctx, req.Msg.Id); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.DeleteDeviceResponse{
		Message: "device deleted successfully",
	}), nil
}

func (h *DeviceConnectHandler) TestDeviceConnection(ctx context.Context, req *connect.Request[devicepb.TestDeviceConnectionRequest]) (*connect.Response[devicepb.TestDeviceConnectionResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device id is required"))
	}

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

	res, err := h.useCase.TestConnection(ctx, drv, req.Msg.Id, req.Msg.SelectedInterface)
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
