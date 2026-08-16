package device

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"connectrpc.com/connect"
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/ping"
	"github.com/quixiq/polyglot/pkg/response"
)

func (h *DeviceConnectHandler) StreamDeviceStatus(
	ctx context.Context,
	req *connect.Request[devicepb.StreamDeviceStatusRequest],
	stream *connect.ServerStream[devicepb.DeviceStatusFrame],
) error {
	if req.Msg.Id == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device id is required"))
	}

	dev, err := h.useCase.GetDevice(ctx, req.Msg.Id)
	if err != nil {
		return response.MapDomainError(err)
	}

	frame := &devicepb.DeviceStatusFrame{
		Device: DomainToPb(dev),
	}

	var drv port.DeviceDriver
	if h.driverGetter != nil {
		d, err := h.driverGetter(ctx, dev.ID)
		if err == nil {
			drv = d
		}
	}

	initialRes, _ := h.useCase.TestConnection(ctx, drv, dev.ID, req.Msg.SelectedInterface)
	pbIfaces := ToProtoInterfaceDetails(initialRes.InterfaceDetails)

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

	<-ctx.Done()
	return nil
}

func (h *DeviceConnectHandler) StreamPing(
	ctx context.Context,
	req *connect.Request[devicepb.StreamDevicePingRequest],
	stream *connect.ServerStream[devicepb.StreamDevicePingFrame],
) error {
	if req.Msg.Id == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device id is required"))
	}

	dev, err := h.useCase.GetDevice(ctx, req.Msg.Id)
	if err != nil {
		return response.MapDomainError(err)
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
		return response.MapDomainError(fmt.Errorf("failed to start ping stream: %w", err))
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
				latency, status := ping.ParsePingLatency(row)
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

func (h *DeviceConnectHandler) StreamInterfaceTraffic(
	ctx context.Context,
	req *connect.Request[devicepb.StreamDeviceTrafficRequest],
	stream *connect.ServerStream[devicepb.StreamDeviceTrafficFrame],
) error {
	if req.Msg.Id == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device id is required"))
	}

	dev, err := h.useCase.GetDevice(ctx, req.Msg.Id)
	if err != nil {
		return response.MapDomainError(err)
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

	trafficCmd := mikrotik.NewMonitorTrafficStreamCommand(ifaceName)
	trafficHandle, err := sDrv.Stream(ctx, trafficCmd)
	if err != nil {
		return response.MapDomainError(fmt.Errorf("failed to start traffic stream: %w", err))
	}
	defer trafficHandle.Cancel()

	for {
		select {
		case <-ctx.Done():
			return nil
		case res, ok := <-trafficHandle.Chan():
			if !ok {
				return nil
			}
			stats := mikrotik.ParseInterfaceTrafficStats(res)
			rx, _ := strconv.ParseInt(stats.RxBitsPerSecond, 10, 64)
			tx, _ := strconv.ParseInt(stats.TxBitsPerSecond, 10, 64)
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

func (h *DeviceConnectHandler) StreamTerminal(
	ctx context.Context,
	stream *connect.BidiStream[devicepb.TerminalFrame, devicepb.TerminalFrame],
) error {
	firstFrame, err := stream.Receive()
	if err != nil {
		return err
	}
	if firstFrame.DeviceId == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device_id is required in the initial frame"))
	}

	if h.openTermUC == nil {
		return connect.NewError(connect.CodeUnavailable, fmt.Errorf("open terminal use case is not initialized"))
	}

	cols := int(firstFrame.Cols)
	rows := int(firstFrame.Rows)
	if cols <= 0 {
		cols = 120
	}
	if rows <= 0 {
		rows = 35
	}

	session, err := h.openTermUC.Execute(ctx, firstFrame.DeviceId, cols, rows)
	if err != nil {
		logger.WithComponent("DeviceConnectHandler").Errorf("SSH PTY connection failed for device %s: %v", firstFrame.DeviceId, err)
		return connect.NewError(connect.CodeUnavailable, fmt.Errorf("failed to open SSH terminal: %w", err))
	}
	defer session.Close()

	errChan := make(chan error, 2)

	// Goroutine 1: PTY stdout -> Client stream
	go func() {
		buf := make([]byte, 4096)
		stdout := session.Stdout()
		for {
			n, rErr := stdout.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				if sErr := stream.Send(&devicepb.TerminalFrame{
					DeviceId:   firstFrame.DeviceId,
					OutputData: chunk,
				}); sErr != nil {
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

	// Goroutine 2: Client stream -> PTY stdin
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
				// best-effort: gagal resize tidak menghentikan aliran data terminal.
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
