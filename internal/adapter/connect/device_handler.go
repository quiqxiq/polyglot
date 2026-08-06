package connectadapter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
	"github.com/quixiq/polyglot/internal/adapter/ws"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/usecase/business"
)

type DeviceConnectHandler struct {
	useCase       *business.ManageDeviceUseCase
	streamHandler *ws.DeviceStreamHandler
}

func NewDeviceConnectHandler(uc *business.ManageDeviceUseCase, streamH *ws.DeviceStreamHandler) *DeviceConnectHandler {
	return &DeviceConnectHandler{
		useCase:       uc,
		streamHandler: streamH,
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

func (h *DeviceConnectHandler) StreamDeviceStatus(ctx context.Context, req *connect.Request[devicepb.StreamDeviceStatusRequest], stream *connect.ServerStream[devicepb.DeviceStatusFrame]) error {
	outChan := make(chan []byte, 10)

	go func() {
		if req.Msg.Id != "" {
			_ = h.streamHandler.StreamSingleDeviceStatus(ctx, req.Msg.Id, outChan)
		} else {
			_ = h.streamHandler.StreamDevicesStatus(ctx, outChan)
		}
		close(outChan)
	}()

	for {
		select {
		case msg, ok := <-outChan:
			if !ok {
				return nil
			}

			var item ws.DeviceStatusItem
			if err := json.Unmarshal(msg, &item); err == nil {
				frame := &devicepb.DeviceStatusFrame{
					Device: domainToPb(item.Device),
					Test: &devicepb.DeviceTestMetrics{
						DeviceId:  item.Test.DeviceID,
						Status:    item.Test.Status,
						LatencyMs: item.Test.LatencyMS,
						Uptime:    item.Test.Uptime,
						Version:   item.Test.Version,
						BoardName: item.Test.BoardName,
						Identity:  item.Test.Identity,
						Message:   item.Test.Message,
					},
				}
				if err := stream.Send(frame); err != nil {
					return err
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type connectJSONCodec struct{}

func (connectJSONCodec) Name() string                     { return "json" }
func (connectJSONCodec) Marshal(v any) ([]byte, error)    { return json.Marshal(v) }
func (connectJSONCodec) Unmarshal(b []byte, v any) error { return json.Unmarshal(b, v) }

// NewDeviceServiceHandler creates the Connect http.Handler and registers procedures.
func NewDeviceServiceHandler(uc *business.ManageDeviceUseCase, streamH *ws.DeviceStreamHandler) (string, http.Handler) {
	handler := NewDeviceConnectHandler(uc, streamH)
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
	mux.Handle("/"+serviceName+"/StreamDeviceStatus", connect.NewServerStreamHandler(
		"/"+serviceName+"/StreamDeviceStatus",
		handler.StreamDeviceStatus,
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
