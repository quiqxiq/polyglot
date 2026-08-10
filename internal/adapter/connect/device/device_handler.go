package device

import (
	"context"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
	deviceUC "github.com/quixiq/polyglot/internal/usecase/device"
)

type DeviceConnectHandler struct {
	useCase      *deviceUC.ManageDeviceUseCase
	driverGetter DriverGetter
}

func NewDeviceConnectHandler(uc *deviceUC.ManageDeviceUseCase, getter DriverGetter) *DeviceConnectHandler {
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
	if req.Msg.Device == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device payload is required"))
	}

	d := pbToDomain(req.Msg.Device)
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
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	} else {
		if err := h.useCase.UpdateDevice(ctx, d, c); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
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
		Device:  domainToPb(updated),
		Message: msg,
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

	var pbIfaces []*devicepb.DeviceInterfaceInfo
	for _, ifc := range res.InterfaceDetails {
		pbIfaces = append(pbIfaces, &devicepb.DeviceInterfaceInfo{
			Name:       ifc.Name,
			Type:       ifc.Type,
			Disabled:   ifc.Disabled,
			Running:    ifc.Running,
			MacAddress: ifc.MACAddress,
			RxBps:      ifc.RxBps,
			TxBps:      ifc.TxBps,
		})
	}

	return connect.NewResponse(&devicepb.TestDeviceConnectionResponse{
		DeviceId:      res.DeviceID,
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
		RxBps:         res.RxBps,
		TxBps:         res.TxBps,
		Message:       res.Message,
		Success:       res.Status == "connected" || res.Status == "online",
	}), nil
}

func (h *DeviceConnectHandler) StreamDeviceStatus(ctx context.Context, req *connect.Request[devicepb.StreamDeviceStatusRequest], stream *connect.ServerStream[devicepb.DeviceStatusFrame]) error {
	if req.Msg.Id == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device id is required"))
	}

	dev, err := h.useCase.GetDevice(ctx, req.Msg.Id)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, err)
	}

	frame := &devicepb.DeviceStatusFrame{
		Device: domainToPb(dev),
		Test: &devicepb.DeviceTestMetrics{
			DeviceId: dev.ID,
			Status:   "offline",
		},
	}

	if !dev.Enabled || h.driverGetter == nil {
		return stream.Send(frame)
	}

	drv, err := h.driverGetter(ctx, dev.ID)
	if err != nil || drv == nil {
		return stream.Send(frame)
	}

	// 1. Snapshot awal untuk data resource & daftar interface
	initialRes, _ := h.useCase.TestConnection(ctx, drv, dev.ID, req.Msg.SelectedInterface)
	var pbIfaces []*devicepb.DeviceInterfaceInfo
	for _, ifc := range initialRes.InterfaceDetails {
		pbIfaces = append(pbIfaces, &devicepb.DeviceInterfaceInfo{
			Name:       ifc.Name,
			Type:       ifc.Type,
			Disabled:   ifc.Disabled,
			Running:    ifc.Running,
			MacAddress: ifc.MACAddress,
			RxBps:      ifc.RxBps,
			TxBps:      ifc.TxBps,
		})
	}

	metrics := &devicepb.DeviceTestMetrics{
		DeviceId:      dev.ID,
		Status:        initialRes.Status,
		LatencyMs:     initialRes.LatencyMS,
		Uptime:        initialRes.Uptime,
		Version:       initialRes.Version,
		BoardName:     initialRes.BoardName,
		Identity:      initialRes.Identity,
		Message:       initialRes.Message,
		CpuLoad:       int32(initialRes.CPULoad),
		FreeMemory:    initialRes.FreeMemory,
		TotalMemory:   initialRes.TotalMemory,
		Interfaces:    initialRes.Interfaces,
		InterfaceList: pbIfaces,
		RxBps:         initialRes.RxBps,
		TxBps:         initialRes.TxBps,
	}
	frame.Test = metrics
	if err := stream.Send(frame); err != nil {
		return err
	}

	// Snapshot sent, wait until context cancelled for status stream
	<-ctx.Done()
	return nil
}

func (h *DeviceConnectHandler) StreamPing(ctx context.Context, req *connect.Request[devicepb.StreamDevicePingRequest], stream *connect.ServerStream[devicepb.StreamDevicePingFrame]) error {
	if req.Msg.Id == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device id is required"))
	}

	dev, err := h.useCase.GetDevice(ctx, req.Msg.Id)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, err)
	}

	if !dev.Enabled || h.driverGetter == nil {
		return nil
	}

	drv, err := h.driverGetter(ctx, dev.ID)
	if err != nil || drv == nil {
		return nil
	}

	sDrv, ok := drv.(port.StreamingDeviceDriver)
	if !ok {
		<-ctx.Done()
		return nil
	}

	pingTarget := req.Msg.Address
	if pingTarget == "" {
		pingTarget = dev.Host
	}
	if hostOnly, _, err := net.SplitHostPort(pingTarget); err == nil {
		pingTarget = hostOnly
	}

	pingCmd := mikrotik.NewPingStreamCommand(pingTarget)
	pingHandle, err := sDrv.Stream(ctx, pingCmd)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to start ping stream: %w", err))
	}
	defer pingHandle.Cancel()

	for {
		select {
		case <-ctx.Done():
			return nil
		case res, ok := <-pingHandle.Chan():
			if !ok {
				return nil
			}
			if len(res.Rows) > 0 {
				row := res.Rows[0]
				latency, status := parsePingLatency(row)
				frame := &devicepb.StreamDevicePingFrame{
					DeviceId:  dev.ID,
					Address:   pingTarget,
					LatencyMs: latency,
					Status:    status,
				}
				if err := stream.Send(frame); err != nil {
					return err
				}
			}
		}
	}
}

func (h *DeviceConnectHandler) StreamInterfaceTraffic(ctx context.Context, req *connect.Request[devicepb.StreamDeviceTrafficRequest], stream *connect.ServerStream[devicepb.StreamDeviceTrafficFrame]) error {
	if req.Msg.Id == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device id is required"))
	}

	dev, err := h.useCase.GetDevice(ctx, req.Msg.Id)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, err)
	}

	if !dev.Enabled || h.driverGetter == nil {
		return nil
	}

	drv, err := h.driverGetter(ctx, dev.ID)
	if err != nil || drv == nil {
		return nil
	}

	sDrv, ok := drv.(port.StreamingDeviceDriver)
	if !ok {
		<-ctx.Done()
		return nil
	}

	ifaceName := req.Msg.InterfaceName
	if ifaceName == "" || ifaceName == "default" {
		ifaceName = "ether1"
	}

	log.Printf("[StreamInterfaceTraffic] Starting stream for device=%s iface=%s", dev.ID, ifaceName)

	trafficCmd := mikrotik.NewMonitorTrafficStreamCommand(ifaceName)
	trafficHandle, err := sDrv.Stream(ctx, trafficCmd)
	if err != nil {
		log.Printf("[StreamInterfaceTraffic] Error starting stream for iface=%s: %v", ifaceName, err)
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to start traffic stream: %w", err))
	}
	defer func() {
		log.Printf("[StreamInterfaceTraffic] Cancelling stream for iface=%s", ifaceName)
		trafficHandle.Cancel()
	}()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[StreamInterfaceTraffic] Context done for iface=%s", ifaceName)
			return nil
		case res, ok := <-trafficHandle.Chan():
			if !ok {
				log.Printf("[StreamInterfaceTraffic] Channel closed for iface=%s", ifaceName)
				return nil
			}
			stats := mikrotik.ParseInterfaceTrafficStats(res)
			rx, _ := strconv.ParseInt(stats.RxBitsPerSecond, 10, 64)
			tx, _ := strconv.ParseInt(stats.TxBitsPerSecond, 10, 64)
			log.Printf("[StreamInterfaceTraffic] Sending frame iface=%s rx=%d tx=%d raw=%+v", ifaceName, rx, tx, stats)
			frame := &devicepb.StreamDeviceTrafficFrame{
				DeviceId:      dev.ID,
				InterfaceName: ifaceName,
				RxBps:         rx,
				TxBps:         tx,
			}
			if err := stream.Send(frame); err != nil {
				return err
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

// DriverGetter fetches a connected port.DeviceDriver by device ID.
type DriverGetter func(ctx context.Context, deviceID string) (port.DeviceDriver, error)

// NewDeviceServiceHandler creates the Connect http.Handler and registers procedures.
func NewDeviceServiceHandler(uc *deviceUC.ManageDeviceUseCase, getter DriverGetter) (string, http.Handler) {
	handler := &DeviceConnectHandler{
		useCase:      uc,
		driverGetter: getter,
	}
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

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
	mux.Handle("/"+serviceName+"/StreamPing", connect.NewServerStreamHandler(
		"/"+serviceName+"/StreamPing",
		handler.StreamPing,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/StreamInterfaceTraffic", connect.NewServerStreamHandler(
		"/"+serviceName+"/StreamInterfaceTraffic",
		handler.StreamInterfaceTraffic,
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
	sshPort := int32(d.SSHPort)
	if sshPort <= 0 {
		sshPort = 22
	}
	return &devicepb.Device{
		Id:             d.ID,
		TenantId:       d.TenantID,
		Name:           d.Name,
		Vendor:         d.Vendor,
		DriverType:     d.DriverType,
		Host:           d.Host,
		Port:           int32(d.Port),
		SshPort:        sshPort,
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
	sshPort := int(pb.SshPort)
	if sshPort <= 0 {
		sshPort = 22
	}
	return device.Device{
		ID:             pb.Id,
		TenantID:       pb.TenantId,
		Name:           pb.Name,
		Vendor:         pb.Vendor,
		DriverType:     pb.DriverType,
		Host:           pb.Host,
		Port:           int(pb.Port),
		SSHPort:        sshPort,
		TimeoutMS:      int(pb.TimeoutMs),
		PollIntervalMS: int(pb.PollIntervalMs),
		Tags:           pb.Tags,
		Enabled:        pb.Enabled,
	}
}

func parsePingLatency(row map[string]string) (int64, string) {
	status := row["status"]
	if status == "timeout" || status == "host unreachable" || status == "net unreachable" {
		return 0, status
	}
	if status == "" {
		status = "connected"
	}

	timeStr := row["time"]
	if timeStr == "" {
		timeStr = row["avg-rtt"]
	}
	if timeStr == "" {
		timeStr = row["min-rtt"]
	}
	if timeStr == "" {
		timeStr = row["rtt"]
	}
	if timeStr == "" {
		timeStr = row["response-time"]
	}

	if timeStr == "" {
		if _, hasSeq := row["seq"]; hasSeq {
			return 1, status
		}
		if _, hasHost := row["host"]; hasHost {
			return 1, status
		}
		return 0, status
	}

	timeStr = strings.TrimSpace(timeStr)
	timeStr = strings.TrimPrefix(timeStr, "<")
	timeStr = strings.TrimPrefix(timeStr, ">")

	// Try Go stdlib time.ParseDuration (handles "15ms", "1ms200us", "230us", etc.)
	if d, err := time.ParseDuration(timeStr); err == nil {
		ms := float64(d) / float64(time.Millisecond)
		if ms > 0 && ms < 1.0 {
			return 1, status
		}
		return int64(math.Round(ms)), status
	}

	// Check HH:MM:SS.microsecond format (e.g. 00:00:00.023000)
	if strings.Contains(timeStr, ":") {
		parts := strings.Split(timeStr, ":")
		if len(parts) == 3 {
			secStr := parts[2]
			if sec, err := strconv.ParseFloat(secStr, 64); err == nil {
				ms := sec * 1000.0
				if ms > 0 && ms < 1.0 {
					return 1, status
				}
				return int64(math.Round(ms)), status
			}
		}
	}

	cleanStr := strings.TrimSuffix(timeStr, "ms")
	cleanStr = strings.TrimSuffix(cleanStr, "s")
	cleanStr = strings.TrimSpace(cleanStr)

	if f, err := strconv.ParseFloat(cleanStr, 64); err == nil {
		if f > 0 && f < 1.0 {
			return 1, status
		}
		return int64(math.Round(f)), status
	}

	return 1, status
}
