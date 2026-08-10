package hotspot

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
	"github.com/quixiq/polyglot/internal/usecase/network"
)


// ConnectDriverProvider signature to obtain a port.DeviceDriver for a given deviceId.
type ConnectDriverProvider func(ctx context.Context, deviceID string) (port.DeviceDriver, error)

// HotspotConnectHandler orchestrates Hotspot ConnectRPC procedures across modular handler files.
type HotspotConnectHandler struct {
	useCase               *hotspotUC.HotspotUseCase
	activeSessionsUseCase *network.ActiveSessionsUseCase
	driverProvider        ConnectDriverProvider
}

// NewHotspotConnectHandler constructs a new HotspotConnectHandler.
func NewHotspotConnectHandler(uc *hotspotUC.HotspotUseCase, provider ConnectDriverProvider) *HotspotConnectHandler {
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

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("driver does not support streaming"))
	}

	cmd := mikrotik.NewStreamSystemResourceCommand("1s")
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
			sysRes := mikrotik.ParseSystemResource(res)
			err := stream.Send(&devicepb.ResourceStreamData{
				DeviceId:      req.Msg.DeviceId,
				CpuLoad:       int32(sysRes.CPULoad),
				FreeMemory:    sysRes.FreeMemory,
				Uptime:        sysRes.Uptime,
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

	ticker := time.NewTicker(3 * time.Second)
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
			pbSessions := make([]*devicepb.HotspotActiveSession, len(sessions))
			for i, s := range sessions {
				pbSessions[i] = &devicepb.HotspotActiveSession{
					Id:         s.RosID,
					Server:     s.Server,
					User:       s.User,
					Address:    s.Address,
					MacAddress: s.MACAddress,
					Uptime:     s.Uptime,
					BytesIn:    s.BytesIn,
					BytesOut:   s.BytesOut,
				}
			}
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

// NewHotspotServiceHandler creates the Connect http.Handler and registers procedures.
func NewHotspotServiceHandler(uc *hotspotUC.HotspotUseCase, provider ConnectDriverProvider) (string, http.Handler) {
	handler := NewHotspotConnectHandler(uc, provider)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.HotspotService"
	mux.Handle("/"+serviceName+"/GetDashboard", connect.NewUnaryHandler(
		"/"+serviceName+"/GetDashboard",
		handler.GetDashboard,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/ListProfiles", connect.NewUnaryHandler(
		"/"+serviceName+"/ListProfiles",
		handler.ListProfiles,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/ListUsers", connect.NewUnaryHandler(
		"/"+serviceName+"/ListUsers",
		handler.ListUsers,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/ListActiveSessions", connect.NewUnaryHandler(
		"/"+serviceName+"/ListActiveSessions",
		handler.ListActiveSessions,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/KickActiveSession", connect.NewUnaryHandler(
		"/"+serviceName+"/KickActiveSession",
		handler.KickActiveSession,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/ListDHCPLeases", connect.NewUnaryHandler(
		"/"+serviceName+"/ListDHCPLeases",
		handler.ListDHCPLeases,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/BlockDHCPLease", connect.NewUnaryHandler(
		"/"+serviceName+"/BlockDHCPLease",
		handler.BlockDHCPLease,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/GenerateVouchers", connect.NewUnaryHandler(
		"/"+serviceName+"/GenerateVouchers",
		handler.GenerateVouchers,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/StreamTraffic", connect.NewServerStreamHandler(
		"/"+serviceName+"/StreamTraffic",
		handler.StreamTraffic,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/StreamResource", connect.NewServerStreamHandler(
		"/"+serviceName+"/StreamResource",
		handler.StreamResource,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/StreamActiveSessions", connect.NewServerStreamHandler(
		"/"+serviceName+"/StreamActiveSessions",
		handler.StreamActiveSessions,
		codecOpt,
	))

	return "/" + serviceName + "/", mux
}
