package monitor

import (
	"context"
	"sync"
	"time"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/command"
	domainDevice "github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

// systemSnapshotState holds the latest row from each of the five system
// print streams. Every stream re-sends its full snapshot on its own tick,
// so this state is always a complete view of the last received frame.
type systemSnapshotState struct {
	mu          sync.Mutex
	clock       domainDevice.SystemClock
	resource    port.SystemResource
	routerboard domainDevice.SystemRouterboard
	identity    string
	health      domainDevice.SystemHealth
}

func (s *systemSnapshotState) setClock(c domainDevice.SystemClock) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clock = c
}

func (s *systemSnapshotState) setResource(r port.SystemResource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resource = r
}

func (s *systemSnapshotState) setRouterboard(r domainDevice.SystemRouterboard) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routerboard = r
}

func (s *systemSnapshotState) setIdentity(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.identity = name
}

func (s *systemSnapshotState) setHealth(h domainDevice.SystemHealth) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.health = h
}

// toFrame builds a combined SystemSnapshotFrame from the latest known data.
func (s *systemSnapshotState) toFrame(deviceID string) *polyglotv1.SystemSnapshotFrame {
	s.mu.Lock()
	defer s.mu.Unlock()
	return &polyglotv1.SystemSnapshotFrame{
		DeviceId:      deviceID,
		TimestampUnix: time.Now().Unix(),
		Clock: &polyglotv1.SystemClockInfo{
			Time:         s.clock.Time,
			Date:         s.clock.Date,
			TimeZoneName: s.clock.TimeZoneName,
			GmtOffset:    s.clock.GMTOffset,
		},
		Resource: &polyglotv1.SystemResourceInfo{
			CpuLoad:       int32(s.resource.CPULoad),
			CpuCount:      int32(s.resource.CPUCount),
			CpuFrequency:  s.resource.CPUFrequency,
			FreeMemory:    s.resource.FreeMemory,
			TotalMemory:   s.resource.TotalMemory,
			FreeHddSpace:  s.resource.FreeHDDSpace,
			TotalHddSpace: s.resource.TotalHDDSpace,
			Architecture:  s.resource.Architecture,
			Model:         s.resource.Model,
			SerialNumber:  s.resource.SerialNumber,
			FirmwareType:  s.resource.FirmwareType,
			Voltage:       s.resource.Voltage,
			Temperature:   s.resource.Temperature,
			BadBlocks:     s.resource.BadBlocks,
			Uptime:        s.resource.Uptime,
			Version:       s.resource.Version,
			BoardName:     s.resource.BoardName,
		},
		Routerboard: &polyglotv1.SystemRouterboardInfo{
			BoardName:       s.routerboard.BoardName,
			Model:           s.routerboard.Model,
			SerialNumber:    s.routerboard.SerialNumber,
			FirmwareType:    s.routerboard.FirmwareType,
			FactoryFirmware: s.routerboard.FactoryFirmware,
			CurrentFirmware: s.routerboard.CurrentFirmware,
			UpgradeFirmware: s.routerboard.UpgradeFirmware,
		},
		Identity: s.identity,
		Health: &polyglotv1.SystemHealthInfo{
			Voltage:        s.health.Voltage,
			Temperature:    s.health.Temperature,
			CpuTemperature: s.health.CPUTemperature,
			PsuVoltage:     s.health.PSUVoltage,
			PsuCurrent:     s.health.PSUCurrent,
			PsuTemperature: s.health.PSUTemperature,
			Fan1Speed:      s.health.Fan1Speed,
			Fan2Speed:      s.health.Fan2Speed,
		},
	}
}

// StreamSystemSnapshot combines five native RouterOS streams.
func (h *NetworkMonitorConnectHandler) StreamSystemSnapshot(ctx context.Context, req *connect.Request[polyglotv1.StreamSystemSnapshotRequest], stream *connect.ServerStream[polyglotv1.SystemSnapshotFrame]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return response.Unimplemented("driver does not support streaming")
	}

	interval := req.Msg.Interval
	if interval == "" {
		interval = "1s"
	}

	state := &systemSnapshotState{}
	notify := make(chan struct{}, 8)
	doneCh := make(chan struct{})
	var wg sync.WaitGroup

	start := func(openStream func() (port.StreamHandle, error), apply func(res command.Result)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			handle, err := openStream()
			if err != nil {
				return
			}
			defer func() { _ = handle.Cancel() }()
			for {
				select {
				case <-ctx.Done():
					return
				case res, ok := <-handle.Chan():
					if !ok {
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

	start(func() (port.StreamHandle, error) {
		return h.streamGW.StreamClock(ctx, sd, interval)
	}, func(res command.Result) {
		state.setClock(h.streamGW.ParseClock(res))
	})
	start(func() (port.StreamHandle, error) {
		return h.streamGW.StreamResource(ctx, sd, interval)
	}, func(res command.Result) {
		state.setResource(h.streamGW.ParseResource(res))
	})
	start(func() (port.StreamHandle, error) {
		return h.streamGW.StreamRouterboard(ctx, sd, interval)
	}, func(res command.Result) {
		state.setRouterboard(h.streamGW.ParseRouterboard(res))
	})
	start(func() (port.StreamHandle, error) {
		return h.streamGW.StreamIdentity(ctx, sd, interval)
	}, func(res command.Result) {
		state.setIdentity(h.streamGW.ParseIdentity(res).Name)
	})
	start(func() (port.StreamHandle, error) {
		return h.streamGW.StreamHealth(ctx, sd, interval)
	}, func(res command.Result) {
		state.setHealth(h.streamGW.ParseHealth(res))
	})

	go func() { wg.Wait(); close(doneCh) }()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-doneCh:
			return nil
		case <-notify:
			timer := time.NewTimer(100 * time.Millisecond)
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
			if err := stream.Send(state.toFrame(req.Msg.DeviceId)); err != nil {
				return err
			}
		}
	}
}

// StreamResource streams /system/resource/print interval=1s natively from RouterOS.
func (h *NetworkMonitorConnectHandler) StreamResource(ctx context.Context, req *connect.Request[polyglotv1.StreamResourceRequest], stream *connect.ServerStream[polyglotv1.ResourceStreamData]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return response.Unimplemented("driver does not support streaming")
	}

	handle, err := h.streamGW.StreamResource(ctx, sd, "1s")
	if err != nil {
		return response.MapDomainError(err)
	}
	defer func() { _ = handle.Cancel() }()

	for {
		select {
		case <-ctx.Done():
			return nil
		case res, ok := <-handle.Chan():
			if !ok {
				return handle.Err()
			}
			sysRes := h.streamGW.ParseResource(res)
			err := stream.Send(&polyglotv1.ResourceStreamData{
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
