package device

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/fault"
	"github.com/quixiq/polyglot/pkg/ping"
	"github.com/quixiq/polyglot/pkg/response"
)

type deviceStatusStreamState struct {
	mu       sync.Mutex
	metrics  *devicepb.DeviceTestMetrics
	ifaceMap map[string]*devicepb.DeviceInterfaceInfo
}

// StreamDeviceStatus streams device status updates continuously using native RouterOS events.
func (h *DeviceConnectHandler) StreamDeviceStatus(
	ctx context.Context,
	req *connect.Request[devicepb.StreamDeviceStatusRequest],
	stream *connect.ServerStream[devicepb.DeviceStatusFrame],
) error {
	if req.Msg.Id == "" {
		return response.InvalidArgument("device id is required")
	}

	dev, err := h.useCase.GetDevice(ctx, req.Msg.Id)
	if err != nil {
		return response.MapDomainError(err)
	}

	frame := &devicepb.DeviceStatusFrame{
		Device: DomainToPb(dev),
	}

	var drv port.DeviceDriver
	var drvErr error
	if h.driverGetter != nil {
		d, err := h.driverGetter(ctx, dev.ID)
		if err == nil {
			drv = d
		} else {
			drvErr = err
		}
	}

	initialRes, _ := h.useCase.TestConnection(
		ctx,
		drv,
		dev.ID,
		req.Msg.SelectedInterface,
		req.Msg.InterfaceTypeFilter,
		req.Msg.InterfaceNameFilter,
	)
	if drv == nil && drvErr != nil {
		initialRes.Message = fmt.Sprintf("Failed to connect to device: %v", drvErr)
	}
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

	sDrv, ok := drv.(port.StreamingDeviceDriver)
	if !ok || !dev.Enabled {
		<-ctx.Done()
		return nil
	}

	ifaceMap := make(map[string]*devicepb.DeviceInterfaceInfo, len(pbIfaces))
	for _, ifc := range pbIfaces {
		if ifc != nil && ifc.Name != "" {
			ifaceMap[ifc.Name] = ifc
		}
	}

	state := &deviceStatusStreamState{
		metrics:  metrics,
		ifaceMap: ifaceMap,
	}

	notify := make(chan struct{}, 16)
	doneCh := make(chan struct{})
	var wg sync.WaitGroup

	startStream := func(openStream func() (port.StreamHandle, error), apply func(res command.Result)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			handle, sErr := openStream()
			if sErr != nil {
				return
			}
			defer func() { _ = handle.Cancel() }()

			for {
				select {
				case <-ctx.Done():
					return
				case res, channelOpen := <-handle.Chan():
					if !channelOpen {
						return
					}
					apply(res)
					select {
					case notify <- struct{}{}:
					default:
					}
				}
			}
		}()
	}

	// 1. Native stream /system/resource/print interval=1s (CPU, memory, uptime, version)
	startStream(func() (port.StreamHandle, error) {
		return h.streamGW.StreamResource(ctx, sDrv, "1s")
	}, func(res command.Result) {
		sysRes := h.streamGW.ParseResource(res)
		freeMem, _ := strconv.ParseInt(sysRes.FreeMemory, 10, 64)
		totalMem, _ := strconv.ParseInt(sysRes.TotalMemory, 10, 64)

		state.mu.Lock()
		state.metrics.Status = "connected"
		state.metrics.CpuLoad = int32(sysRes.CPULoad)
		state.metrics.FreeMemory = freeMem
		state.metrics.TotalMemory = totalMem
		if sysRes.Uptime != "" {
			state.metrics.Uptime = sysRes.Uptime
		}
		if sysRes.Version != "" {
			state.metrics.Version = sysRes.Version
		}
		if sysRes.BoardName != "" {
			state.metrics.BoardName = sysRes.BoardName
		}
		state.mu.Unlock()
	})

	// 2. Native stream /interface/print interval=2s (Interface list & running status)
	startStream(func() (port.StreamHandle, error) {
		return h.streamGW.StreamInterfaces(ctx, sDrv, req.Msg.InterfaceTypeFilter, req.Msg.InterfaceNameFilter, "2s")
	}, func(res command.Result) {
		ifaces := h.streamGW.ParseInterfaces(res)
		if len(ifaces) == 0 {
			return
		}

		state.mu.Lock()
		for _, ifc := range ifaces {
			if ifc.Name == "" {
				continue
			}
			state.ifaceMap[ifc.Name] = &devicepb.DeviceInterfaceInfo{
				Name:       ifc.Name,
				Type:       ifc.Type,
				Disabled:   ifc.Disabled,
				Running:    ifc.Running,
				MacAddress: ifc.MACAddress,
			}
		}

		names := make([]string, 0, len(state.ifaceMap))
		details := make([]*devicepb.DeviceInterfaceInfo, 0, len(state.ifaceMap))
		for _, ifc := range state.ifaceMap {
			names = append(names, ifc.Name)
			details = append(details, ifc)
		}
		sort.Strings(names)
		sort.Slice(details, func(i, j int) bool {
			return details[i].Name < details[j].Name
		})

		state.metrics.Interfaces = names
		state.metrics.InterfaceList = details
		state.mu.Unlock()
	})

	// 3. Native stream /system/identity/print interval=5s (Identity update)
	startStream(func() (port.StreamHandle, error) {
		return h.streamGW.StreamIdentity(ctx, sDrv, "5s")
	}, func(res command.Result) {
		ident := h.streamGW.ParseIdentity(res)
		if ident.Name != "" {
			state.mu.Lock()
			state.metrics.Identity = ident.Name
			state.mu.Unlock()
		}
	})

	go func() {
		wg.Wait()
		close(doneCh)
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-doneCh:
			return nil
		case <-notify:
			timer := time.NewTimer(50 * time.Millisecond)
		coalesce:
			for {
				select {
				case <-notify:
				case <-timer.C:
					break coalesce
				case <-ctx.Done():
					return nil
				case <-doneCh:
					return nil
				}
			}

			state.mu.Lock()
			clonedMetrics, ok := proto.Clone(state.metrics).(*devicepb.DeviceTestMetrics)
			state.mu.Unlock()

			if !ok || clonedMetrics == nil {
				continue
			}

			if err := stream.Send(&devicepb.DeviceStatusFrame{
				Device: DomainToPb(dev),
				Test:   clonedMetrics,
			}); err != nil {
				return err
			}
		}
	}
}

func (h *DeviceConnectHandler) StreamPing(
	ctx context.Context,
	req *connect.Request[devicepb.StreamDevicePingRequest],
	stream *connect.ServerStream[devicepb.StreamDevicePingFrame],
) error {
	if req.Msg.Id == "" {
		return response.InvalidArgument("device id is required")
	}

	dev, err := h.useCase.GetDevice(ctx, req.Msg.Id)
	if err != nil {
		return response.MapDomainError(err)
	}

	if !dev.Enabled || h.driverGetter == nil {
		return response.Unavailable("device streaming is unavailable")
	}

	drv, err := h.driverGetter(ctx, dev.ID)
	if err != nil || drv == nil {
		if err != nil {
			return response.MapDomainError(fault.Wrap(fault.KindUnavailable, err))
		}
		return response.Unavailable("device driver is unavailable")
	}

	sDrv, ok := drv.(port.StreamingDeviceDriver)
	if !ok {
		return response.Unavailable("device driver does not support streaming")
	}

	pingTarget := req.Msg.Address
	if pingTarget == "" {
		pingTarget = dev.PingConfig().Target
	}
	if pingTarget == "" {
		pingTarget = dev.Host
	}
	if hostOnly, _, err := net.SplitHostPort(pingTarget); err == nil {
		pingTarget = hostOnly
	}

	pingHandle, err := h.streamGW.StreamPing(ctx, sDrv, pingTarget)
	if err != nil {
		return response.MapDomainError(fmt.Errorf("failed to start ping stream: %w", err))
	}
	defer func() { _ = pingHandle.Cancel() }()

	streamSeq := int32(0)
	for {
		select {
		case <-ctx.Done():
			return nil
		case res, ok := <-pingHandle.Chan():
			if !ok {
				if err := pingHandle.Err(); err != nil {
					return response.MapDomainError(fault.Wrap(fault.KindUnavailable, err))
				}
				return nil
			}
			if len(res.Rows) > 0 {
				row := res.Rows[0]
				latency, status := ping.ParsePingLatency(row)
				seq, _ := strconv.ParseInt(row["seq"], 10, 32)
				if seq == 0 {
					if s, ok := row["sequence"]; ok {
						seq, _ = strconv.ParseInt(s, 10, 32)
					}
					if seq == 0 {
						seq = int64(streamSeq)
					}
				}
				streamSeq = int32(seq + 1)
				ttl, _ := strconv.ParseInt(row["ttl"], 10, 32)
				size, _ := strconv.ParseInt(row["size"], 10, 32)
				sent, _ := strconv.ParseInt(row["sent"], 10, 32)
				recv, _ := strconv.ParseInt(row["received"], 10, 32)
				loss := ping.ParsePacketLoss(row["packet-loss"])
				minRTT := ping.ParseDurationMs(row["min-rtt"])
				avgRTT := ping.ParseDurationMs(row["avg-rtt"])
				maxRTT := ping.ParseDurationMs(row["max-rtt"])

				frame := &devicepb.StreamDevicePingFrame{
					DeviceId:   dev.ID,
					Address:    pingTarget,
					LatencyMs:  latency,
					Status:     status,
					Seq:        int32(seq),
					Ttl:        int32(ttl),
					Size:       int32(size),
					Sent:       int32(sent),
					Received:   int32(recv),
					PacketLoss: loss,
					MinRttMs:   minRTT,
					AvgRttMs:   avgRTT,
					MaxRttMs:   maxRTT,
				}
				if err := stream.Send(frame); err != nil {
					return err
				}
			}
		}
	}
}

// StreamInterfaceTraffic streams real-time traffic statistics for a specific interface.
func (h *DeviceConnectHandler) StreamInterfaceTraffic(
	ctx context.Context,
	req *connect.Request[devicepb.StreamDeviceTrafficRequest],
	stream *connect.ServerStream[devicepb.StreamDeviceTrafficFrame],
) error {
	if req.Msg.Id == "" {
		return response.InvalidArgument("device id is required")
	}

	dev, err := h.useCase.GetDevice(ctx, req.Msg.Id)
	if err != nil {
		return response.MapDomainError(err)
	}

	drv, err := h.driverGetter(ctx, dev.ID)
	if err != nil || drv == nil {
		if err != nil {
			return response.MapDomainError(fault.Wrap(fault.KindUnavailable, err))
		}
		return response.Unavailable("device driver is unavailable")
	}

	sDrv, ok := drv.(port.StreamingDeviceDriver)
	if !ok {
		return response.Unavailable("device driver does not support streaming")
	}

	ifaceName := req.Msg.InterfaceName
	if ifaceName == "" || ifaceName == "default" {
		ifaceName = "ether1"
	}

	trafficHandle, err := h.streamGW.StreamTraffic(ctx, sDrv, ifaceName)
	if err != nil {
		return response.MapDomainError(fmt.Errorf("failed to start traffic stream: %w", err))
	}
	defer func() { _ = trafficHandle.Cancel() }()

	for {
		select {
		case <-ctx.Done():
			return nil
		case res, ok := <-trafficHandle.Chan():
			if !ok {
				if err := trafficHandle.Err(); err != nil {
					return response.MapDomainError(fault.Wrap(fault.KindUnavailable, err))
				}
				return nil
			}
			stats := h.streamGW.ParseInterfaceTraffic(res)
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
