package hotspot

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
	"github.com/quixiq/polyglot/internal/usecase/network"
)

// ConnectDriverProvider signature to obtain a port.DeviceDriver for a given deviceId.
type ConnectDriverProvider func(ctx context.Context, deviceID string) (port.DeviceDriver, error)

// HotspotConnectHandler orchestrates Hotspot ConnectRPC procedures across modular handler files.
type HotspotConnectHandler struct {
	useCase               *hotspotUC.UseCase
	activeSessionsUseCase *network.ActiveSessionsUseCase
	driverProvider        ConnectDriverProvider
}

// NewHotspotConnectHandler constructs a new HotspotConnectHandler.
func NewHotspotConnectHandler(uc *hotspotUC.UseCase, provider ConnectDriverProvider) *HotspotConnectHandler {
	return &HotspotConnectHandler{
		useCase:               uc,
		activeSessionsUseCase: network.NewActiveSessionsUseCase(),
		driverProvider:        provider,
	}
}

func (h *HotspotConnectHandler) getDriver(ctx context.Context, deviceID string) (port.DeviceDriver, error) {
	if deviceID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device_id is required"))
	}
	driver, err := h.driverProvider(ctx, deviceID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get device driver for %s: %w", deviceID, err))
	}
	return driver, nil
}

func (h *HotspotConnectHandler) StreamTraffic(ctx context.Context, req *connect.Request[devicepb.StreamTrafficRequest], stream *connect.ServerStream[devicepb.TrafficStreamData]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("driver does not support streaming"))
	}

	iface := req.Msg.Interface
	if iface == "" {
		iface = "ether1"
	}

	cmd := mikrotik.NewMonitorTrafficStreamCommand(iface)
	handle, err := sd.Stream(ctx, cmd)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	defer handle.Cancel()

	for {
		select {
		case <-ctx.Done():
			return nil
		case res, ok := <-handle.Chan():
			if !ok {
				return handle.Err()
			}
			stats := mikrotik.ParseInterfaceTrafficStats(res)
			rx, _ := strconv.ParseInt(stats.RxBitsPerSecond, 10, 64)
			tx, _ := strconv.ParseInt(stats.TxBitsPerSecond, 10, 64)

			err := stream.Send(&devicepb.TrafficStreamData{
				DeviceId:      req.Msg.DeviceId,
				Interface:     iface,
				RxBps:         rx,
				TxBps:         tx,
				TimestampUnix: time.Now().Unix(),
			})
			if err != nil {
				return err
			}
		}
	}
}

func (h *HotspotConnectHandler) StreamResource(ctx context.Context, req *connect.Request[devicepb.StreamResourceRequest], stream *connect.ServerStream[devicepb.ResourceStreamData]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			res, err := h.useCase.GetSystemResource(ctx, driver)
			if err != nil {
				continue
			}

			err = stream.Send(&devicepb.ResourceStreamData{
				DeviceId:      req.Msg.DeviceId,
				CpuLoad:       int32(res.CPULoad),
				FreeMemory:    res.FreeMemory,
				Uptime:        res.Uptime,
				TimestampUnix: time.Now().Unix(),
			})
			if err != nil {
				return err
			}
		}
	}
}

func (h *HotspotConnectHandler) StreamActiveSessions(ctx context.Context, req *connect.Request[devicepb.StreamActiveSessionsRequest], stream *connect.ServerStream[devicepb.ActiveSessionsStreamData]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			sessions, err := h.useCase.GetActiveSessions(ctx, driver)
			if err != nil {
				continue
			}

			pbSessions := ToProtoActiveSessions(sessions)
			err = stream.Send(&devicepb.ActiveSessionsStreamData{
				DeviceId:      req.Msg.DeviceId,
				Sessions:      pbSessions,
				TimestampUnix: time.Now().Unix(),
			})
			if err != nil {
				return err
			}
		}
	}
}
