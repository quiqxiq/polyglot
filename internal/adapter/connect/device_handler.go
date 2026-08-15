package connectadapter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
	"github.com/quixiq/polyglot/internal/adapter/connect/codec"
	"github.com/quixiq/polyglot/internal/adapter/connect/mapper"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/business"
	"github.com/quixiq/polyglot/pkg/response"
)

// DeviceConnectHandler handles ConnectRPC procedures for device inventory and streaming.
type DeviceConnectHandler struct {
	useCase        *business.ManageDeviceUseCase
	driverProvider port.DriverProvider
}

// NewDeviceConnectHandler creates a new DeviceConnectHandler.
func NewDeviceConnectHandler(uc *business.ManageDeviceUseCase, provider port.DriverProvider) *DeviceConnectHandler {
	return &DeviceConnectHandler{
		useCase:        uc,
		driverProvider: provider,
	}
}

// ListDevices returns all registered network devices.
func (h *DeviceConnectHandler) ListDevices(ctx context.Context, req *connect.Request[devicepb.ListDevicesRequest]) (*connect.Response[devicepb.ListDevicesResponse], error) {
	devices, err := h.useCase.ListDevices(ctx)
	if err != nil {
		return nil, response.ToConnectError(err)
	}

	return connect.NewResponse(&devicepb.ListDevicesResponse{
		Devices: mapper.DeviceListToProto(devices),
	}), nil
}

// GetDevice retrieves a specific device by ID.
func (h *DeviceConnectHandler) GetDevice(ctx context.Context, req *connect.Request[devicepb.GetDeviceRequest]) (*connect.Response[devicepb.GetDeviceResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device id is required"))
	}

	d, err := h.useCase.GetDevice(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.ToConnectError(err)
	}

	return connect.NewResponse(&devicepb.GetDeviceResponse{
		Device: mapper.DeviceToProto(d),
	}), nil
}

// UpdateDevice updates or creates device inventory information and credentials.
func (h *DeviceConnectHandler) UpdateDevice(ctx context.Context, req *connect.Request[devicepb.UpdateDeviceRequest]) (*connect.Response[devicepb.UpdateDeviceResponse], error) {
	if req.Msg.Device == nil || req.Msg.Device.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device with id is required"))
	}

	d := mapper.ProtoToDevice(req.Msg.Device)
	c := device.Credentials{
		Username: req.Msg.Username,
		Password: req.Msg.Password,
	}

	if err := h.useCase.UpdateDevice(ctx, d, c); err != nil {
		return nil, response.ToConnectError(err)
	}

	updated, err := h.useCase.GetDevice(ctx, d.ID)
	if err != nil {
		updated = d
	}

	return connect.NewResponse(&devicepb.UpdateDeviceResponse{
		Device:  mapper.DeviceToProto(updated),
		Message: "device updated successfully via ConnectRPC",
	}), nil
}

// DeleteDevice deletes a device by ID.
func (h *DeviceConnectHandler) DeleteDevice(ctx context.Context, req *connect.Request[devicepb.DeleteDeviceRequest]) (*connect.Response[devicepb.DeleteDeviceResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device id is required"))
	}

	if err := h.useCase.DeleteDevice(ctx, req.Msg.Id); err != nil {
		return nil, response.ToConnectError(err)
	}

	return connect.NewResponse(&devicepb.DeleteDeviceResponse{
		Message: "device deleted successfully",
	}), nil
}

// TestDeviceConnection tests connectivity and gathers system metadata from the device.
func (h *DeviceConnectHandler) TestDeviceConnection(ctx context.Context, req *connect.Request[devicepb.TestDeviceConnectionRequest]) (*connect.Response[devicepb.TestDeviceConnectionResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device id is required"))
	}

	var drv port.DeviceDriver
	if h.driverProvider != nil {
		if d, err := h.driverProvider(ctx, req.Msg.Id); err == nil {
			drv = d
		}
	}

	res, err := h.useCase.TestConnection(ctx, drv, req.Msg.Id)
	if err != nil {
		return connect.NewResponse(&devicepb.TestDeviceConnectionResponse{
			DeviceId: req.Msg.Id,
			Status:   "error",
			Message:  err.Error(),
			Success:  false,
		}), nil
	}

	return connect.NewResponse(mapper.ConnectionTestToProto(res)), nil
}

// StreamDeviceStatus pushes periodic status updates for devices.
func (h *DeviceConnectHandler) StreamDeviceStatus(ctx context.Context, req *connect.Request[devicepb.StreamDeviceStatusRequest], stream *connect.ServerStream[devicepb.DeviceStatusFrame]) error {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if req.Msg.Id != "" {
				dev, err := h.useCase.GetDevice(ctx, req.Msg.Id)
				if err == nil {
					frame := &devicepb.DeviceStatusFrame{Device: mapper.DeviceToProto(dev)}
					if err := stream.Send(frame); err != nil {
						return err
					}
				}
			} else {
				devices, err := h.useCase.ListDevices(ctx)
				if err == nil {
					for _, dev := range devices {
						frame := &devicepb.DeviceStatusFrame{Device: mapper.DeviceToProto(dev)}
						if err := stream.Send(frame); err != nil {
							return err
						}
					}
				}
			}
		}
	}
}

// StreamTerminal handles bidirectional PTY terminal streaming over ConnectRPC.
func (h *DeviceConnectHandler) StreamTerminal(ctx context.Context, stream *connect.BidiStream[devicepb.TerminalFrame, devicepb.TerminalFrame]) error {
	firstFrame, err := stream.Receive()
	if err != nil {
		return err
	}

	deviceID := firstFrame.DeviceId
	if deviceID == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device_id is required"))
	}

	if h.driverProvider == nil {
		return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("driver provider not configured"))
	}

	driver, err := h.driverProvider(ctx, deviceID)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get device driver: %w", err))
	}

	td, ok := driver.(port.TerminalDeviceDriver)
	if !ok {
		return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("driver for device %s does not support terminal streaming", deviceID))
	}

	cols := int(firstFrame.Cols)
	if cols <= 0 {
		cols = 80
	}
	rows := int(firstFrame.Rows)
	if rows <= 0 {
		rows = 24
	}

	session, err := td.OpenTerminalSession(ctx, cols, rows)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to open terminal session: %w", err))
	}
	defer session.Close()

	errChan := make(chan error, 2)

	// Goroutine 1: Read stdout from PTY -> send to client
	go func() {
		buf := make([]byte, 4096)
		stdout := session.Stdout()
		for {
			n, rErr := stdout.Read(buf)
			if n > 0 {
				outFrame := &devicepb.TerminalFrame{
					DeviceId:   deviceID,
					OutputData: buf[:n],
				}
				if sErr := stream.Send(outFrame); sErr != nil {
					errChan <- sErr
					return
				}
			}
			if rErr != nil {
				errChan <- rErr
				return
			}
		}
	}()

	// Goroutine 2: Receive client input frames -> write to PTY stdin
	go func() {
		stdin := session.Stdin()
		for {
			req, rErr := stream.Receive()
			if rErr != nil {
				errChan <- rErr
				return
			}
			if len(req.InputData) > 0 {
				if _, wErr := stdin.Write(req.InputData); wErr != nil {
					errChan <- wErr
					return
				}
			}
			if req.Cols > 0 && req.Rows > 0 {
				_ = session.Resize(int(req.Cols), int(req.Rows))
			}
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

// NewDeviceServiceHandler creates the Connect http.Handler and registers procedures.
func NewDeviceServiceHandler(uc *business.ManageDeviceUseCase, provider port.DriverProvider) (string, http.Handler) {
	handler := NewDeviceConnectHandler(uc, provider)
	mux := http.NewServeMux()
	codecOpt := codec.Option()

	serviceName := "polyglot.v1.DeviceService"
	mux.Handle("/"+serviceName+"/ListDevices", connect.NewUnaryHandler("/"+serviceName+"/ListDevices", handler.ListDevices, codecOpt))
	mux.Handle("/"+serviceName+"/GetDevice", connect.NewUnaryHandler("/"+serviceName+"/GetDevice", handler.GetDevice, codecOpt))
	mux.Handle("/"+serviceName+"/UpdateDevice", connect.NewUnaryHandler("/"+serviceName+"/UpdateDevice", handler.UpdateDevice, codecOpt))
	mux.Handle("/"+serviceName+"/DeleteDevice", connect.NewUnaryHandler("/"+serviceName+"/DeleteDevice", handler.DeleteDevice, codecOpt))
	mux.Handle("/"+serviceName+"/TestDeviceConnection", connect.NewUnaryHandler("/"+serviceName+"/TestDeviceConnection", handler.TestDeviceConnection, codecOpt))
	mux.Handle("/"+serviceName+"/StreamDeviceStatus", connect.NewServerStreamHandler("/"+serviceName+"/StreamDeviceStatus", handler.StreamDeviceStatus, codecOpt))
	mux.Handle("/"+serviceName+"/StreamTerminal", connect.NewBidiStreamHandler("/"+serviceName+"/StreamTerminal", handler.StreamTerminal, codecOpt))

	return "/" + serviceName + "/", mux
}
