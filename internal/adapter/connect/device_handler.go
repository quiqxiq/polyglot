package connectadapter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/business"
)

type DeviceConnectHandler struct {
	useCase      *business.ManageDeviceUseCase
	driverGetter DriverGetter
}

func NewDeviceConnectHandler(uc *business.ManageDeviceUseCase, getter DriverGetter) *DeviceConnectHandler {
	return &DeviceConnectHandler{
		useCase:      uc,
		driverGetter: getter,
	}
}

func (h *DeviceConnectHandler) ListDevices(ctx context.Context, req *connect.Request[devicepb.ListDevicesRequest]) (*connect.Response[devicepb.ListDevicesResponse], error) {
	devices, err := h.useCase.ListDevices(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbDevices := make([]*devicepb.Device, len(devices))
	for i, d := range devices {
		pbDevices[i] = domainToPb(d)
	}

	return connect.NewResponse(&devicepb.ListDevicesResponse{Devices: pbDevices}), nil
}

func (h *DeviceConnectHandler) GetDevice(ctx context.Context, req *connect.Request[devicepb.GetDeviceRequest]) (*connect.Response[devicepb.GetDeviceResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device id is required"))
	}

	d, err := h.useCase.GetDevice(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewResponse(&devicepb.GetDeviceResponse{Device: domainToPb(d)}), nil
}

func (h *DeviceConnectHandler) UpdateDevice(ctx context.Context, req *connect.Request[devicepb.UpdateDeviceRequest]) (*connect.Response[devicepb.UpdateDeviceResponse], error) {
	if req.Msg.Device == nil || req.Msg.Device.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device with id is required"))
	}

	d := pbToDomain(req.Msg.Device)
	c := device.Credentials{
		Username: req.Msg.Username,
		Password: req.Msg.Password,
	}

	if err := h.useCase.UpdateDevice(ctx, d, c); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	updated, err := h.useCase.GetDevice(ctx, d.ID)
	if err != nil {
		updated = d
	}

	return connect.NewResponse(&devicepb.UpdateDeviceResponse{
		Device:  domainToPb(updated),
		Message: "device updated successfully via ConnectRPC",
	}), nil
}

func (h *DeviceConnectHandler) DeleteDevice(ctx context.Context, req *connect.Request[devicepb.DeleteDeviceRequest]) (*connect.Response[devicepb.DeleteDeviceResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device id is required"))
	}

	if err := h.useCase.DeleteDevice(ctx, req.Msg.Id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
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
		if d, err := h.driverGetter(ctx, req.Msg.Id); err == nil {
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

	return connect.NewResponse(&devicepb.TestDeviceConnectionResponse{
		DeviceId:  res.DeviceID,
		Status:    res.Status,
		LatencyMs: res.LatencyMS,
		Uptime:    res.Uptime,
		Version:   res.Version,
		BoardName: res.BoardName,
		Identity:  res.Identity,
		Message:   res.Message,
		Success:   res.Status == "connected" || res.Status == "online",
	}), nil
}

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
					frame := &devicepb.DeviceStatusFrame{
						Device: domainToPb(dev),
					}
					if err := stream.Send(frame); err != nil {
						return err
					}
				}
			} else {
				devices, err := h.useCase.ListDevices(ctx)
				if err == nil {
					for _, dev := range devices {
						frame := &devicepb.DeviceStatusFrame{
							Device: domainToPb(dev),
						}
						if err := stream.Send(frame); err != nil {
							return err
						}
					}
				}
			}
		}
	}
}

func (h *DeviceConnectHandler) StreamTerminal(ctx context.Context, stream *connect.BidiStream[devicepb.TerminalFrame, devicepb.TerminalFrame]) error {
	firstFrame, err := stream.Receive()
	if err != nil {
		return err
	}

	deviceID := firstFrame.DeviceId
	if deviceID == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device_id is required"))
	}

	if h.driverGetter == nil {
		return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("driver getter not configured"))
	}

	driver, err := h.driverGetter(ctx, deviceID)
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

	// Goroutine 1: Read PTY stdout and send to client
	errChan := make(chan error, 2)
	go func() {
		buf := make([]byte, 4096)
		stdout := session.Stdout()
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				outFrame := &devicepb.TerminalFrame{
					DeviceId:   deviceID,
					OutputData: buf[:n],
				}
				if sendErr := stream.Send(outFrame); sendErr != nil {
					errChan <- sendErr
					return
				}
			}
			if err != nil {
				errChan <- err
				return
			}
		}
	}()

	// Goroutine 2: Receive client frames and write to PTY stdin / handle resize
	go func() {
		stdin := session.Stdin()
		for {
			req, err := stream.Receive()
			if err != nil {
				errChan <- err
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

type connectJSONCodec struct{}

func (connectJSONCodec) Name() string                     { return "json" }
func (connectJSONCodec) Marshal(v any) ([]byte, error)    { return json.Marshal(v) }
func (connectJSONCodec) Unmarshal(b []byte, v any) error { return json.Unmarshal(b, v) }

// DriverGetter fetches a connected port.DeviceDriver by device ID.
type DriverGetter func(ctx context.Context, deviceID string) (port.DeviceDriver, error)

// NewDeviceServiceHandler creates the Connect http.Handler and registers procedures.
func NewDeviceServiceHandler(uc *business.ManageDeviceUseCase, getter DriverGetter) (string, http.Handler) {
	handler := &DeviceConnectHandler{
		useCase:      uc,
		driverGetter: getter,
	}
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(connectJSONCodec{})

	serviceName := "polyglot.v1.DeviceService"
	mux.Handle("/"+serviceName+"/ListDevices", connect.NewUnaryHandler(
		"/"+serviceName+"/ListDevices",
		handler.ListDevices,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/GetDevice", connect.NewUnaryHandler(
		"/"+serviceName+"/GetDevice",
		handler.GetDevice,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/UpdateDevice", connect.NewUnaryHandler(
		"/"+serviceName+"/UpdateDevice",
		handler.UpdateDevice,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/DeleteDevice", connect.NewUnaryHandler(
		"/"+serviceName+"/DeleteDevice",
		handler.DeleteDevice,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/TestDeviceConnection", connect.NewUnaryHandler(
		"/"+serviceName+"/TestDeviceConnection",
		handler.TestDeviceConnection,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/StreamDeviceStatus", connect.NewServerStreamHandler(
		"/"+serviceName+"/StreamDeviceStatus",
		handler.StreamDeviceStatus,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/StreamTerminal", connect.NewBidiStreamHandler(
		"/"+serviceName+"/StreamTerminal",
		handler.StreamTerminal,
		codecOpt,
	))

	return "/" + serviceName + "/", mux
}

func domainToPb(d device.Device) *devicepb.Device {
	return &devicepb.Device{
		Id:             d.ID,
		TenantId:       d.TenantID,
		Name:           d.Name,
		Vendor:         d.Vendor,
		DriverType:     d.DriverType,
		Host:           d.Host,
		Port:           int32(d.Port),
		TimeoutMs:      int32(d.TimeoutMS),
		PollIntervalMs: int32(d.PollIntervalMS),
		Tags:           d.Tags,
		Enabled:        d.Enabled,
	}
}

func pbToDomain(pb *devicepb.Device) device.Device {
	if pb == nil {
		return device.Device{}
	}
	return device.Device{
		ID:             pb.Id,
		TenantID:       pb.TenantId,
		Name:           pb.Name,
		Vendor:         pb.Vendor,
		DriverType:     pb.DriverType,
		Host:           pb.Host,
		Port:           int(pb.Port),
		TimeoutMS:      int(pb.TimeoutMs),
		PollIntervalMS: int(pb.PollIntervalMs),
		Tags:           pb.Tags,
		Enabled:        pb.Enabled,
	}
}
