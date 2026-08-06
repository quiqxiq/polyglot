package grpc

import (
	"context"
	"encoding/json"
	"fmt"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
	"github.com/quixiq/polyglot/internal/adapter/ws"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/usecase/business"
)

// DeviceServer implements devicepb.DeviceServiceServer.
type DeviceServer struct {
	devicepb.UnimplementedDeviceServiceServer
	useCase       *business.ManageDeviceUseCase
	streamHandler *ws.DeviceStreamHandler
}

// NewDeviceServer constructs a new gRPC DeviceServer.
func NewDeviceServer(uc *business.ManageDeviceUseCase, streamH *ws.DeviceStreamHandler) *DeviceServer {
	return &DeviceServer{
		useCase:       uc,
		streamHandler: streamH,
	}
}

func (s *DeviceServer) ListDevices(ctx context.Context, req *devicepb.ListDevicesRequest) (*devicepb.ListDevicesResponse, error) {
	devices, err := s.useCase.ListDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("list devices error: %w", err)
	}

	pbDevices := make([]*devicepb.Device, len(devices))
	for i, d := range devices {
		pbDevices[i] = domainToPb(d)
	}

	return &devicepb.ListDevicesResponse{Devices: pbDevices}, nil
}

func (s *DeviceServer) GetDevice(ctx context.Context, req *devicepb.GetDeviceRequest) (*devicepb.GetDeviceResponse, error) {
	if req.Id == "" {
		return nil, fmt.Errorf("device id is required")
	}

	d, err := s.useCase.GetDevice(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("get device error: %w", err)
	}

	return &devicepb.GetDeviceResponse{Device: domainToPb(d)}, nil
}

func (s *DeviceServer) UpdateDevice(ctx context.Context, req *devicepb.UpdateDeviceRequest) (*devicepb.UpdateDeviceResponse, error) {
	if req.Device == nil || req.Device.Id == "" {
		return nil, fmt.Errorf("device with id is required")
	}

	d := pbToDomain(req.Device)
	c := device.Credentials{
		Username: req.Username,
		Password: req.Password,
	}

	if err := s.useCase.UpdateDevice(ctx, d, c); err != nil {
		return nil, fmt.Errorf("update device error: %w", err)
	}

	updated, err := s.useCase.GetDevice(ctx, d.ID)
	if err != nil {
		updated = d
	}

	return &devicepb.UpdateDeviceResponse{
		Device:  domainToPb(updated),
		Message: "device updated successfully via gRPC",
	}, nil
}

func (s *DeviceServer) StreamDeviceStatus(req *devicepb.StreamDeviceStatusRequest, stream devicepb.DeviceService_StreamDeviceStatusServer) error {
	outChan := make(chan []byte, 10)
	ctx := stream.Context()

	go func() {
		if req.Id != "" {
			_ = s.streamHandler.StreamSingleDeviceStatus(ctx, req.Id, outChan)
		} else {
			_ = s.streamHandler.StreamDevicesStatus(ctx, outChan)
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
