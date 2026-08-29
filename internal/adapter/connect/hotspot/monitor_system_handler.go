package hotspot

import (
	"context"
	"sync"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/command"
	mikrotiksystem "github.com/quixiq/polyglot/internal/driver/mikrotik/system"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

// systemSnapshotState holds the latest row from each of the five system
// print streams. Every stream re-sends its full snapshot on its own tick,
// so this state is always a complete view of the last received frame.
type systemSnapshotState struct {
	mu          sync.Mutex
	clock       mikrotiksystem.SystemClock
	resource    mikrotiksystem.SystemResource
	routerboard mikrotiksystem.SystemRouterboard
	identity    string
	health      mikrotiksystem.SystemHealth
}

func (s *systemSnapshotState) setClock(c mikrotiksystem.SystemClock) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clock = c
}

func (s *systemSnapshotState) setResource(r mikrotiksystem.SystemResource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resource = r
}

func (s *systemSnapshotState) setRouterboard(r mikrotiksystem.SystemRouterboard) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routerboard = r
}

func (s *systemSnapshotState) setIdentity(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.identity = name
}

func (s *systemSnapshotState) setHealth(h mikrotiksystem.SystemHealth) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.health = h
}

// toFrame builds a combined SystemSnapshotFrame from the latest known data.
func (s *systemSnapshotState) toFrame(deviceID string) *devicepb.SystemSnapshotFrame {
	s.mu.Lock()
	defer s.mu.Unlock()
	return &devicepb.SystemSnapshotFrame{
		DeviceId:      deviceID,
		TimestampUnix: time.Now().Unix(),
		Clock: &devicepb.SystemClockInfo{
			Time:         s.clock.Time,
			Date:         s.clock.Date,
			TimeZoneName: s.clock.TimeZoneName,
			GmtOffset:    s.clock.GMTOffset,
		},
		Resource: &devicepb.SystemResourceInfo{
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
		Routerboard: &devicepb.SystemRouterboardInfo{
			BoardName:       s.routerboard.BoardName,
			Model:           s.routerboard.Model,
			SerialNumber:    s.routerboard.SerialNumber,
			FirmwareType:    s.routerboard.FirmwareType,
			FactoryFirmware: s.routerboard.FactoryFirmware,
			CurrentFirmware: s.routerboard.CurrentFirmware,
			UpgradeFirmware: s.routerboard.UpgradeFirmware,
		},
		Identity: s.identity,
		Health: &devicepb.SystemHealthInfo{
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

// StreamSystemSnapshot combines five native RouterOS streams — clock,
// resource, routerboard, identity, health (each `print interval=<n>`) — into
// a single combined frame pushed to the frontend. Every stream re-sends its
// full snapshot on its own tick, so no backend polling is involved: RouterOS
// itself keeps the stream alive and the backend merely forwards a merged
// frame. A stream that the board does not support (e.g. /system/health on
// some models) is skipped; the remaining sources keep streaming.
func (h *HotspotConnectHandler) StreamSystemSnapshot(ctx context.Context, req *connect.Request[devicepb.StreamSystemSnapshotRequest], stream *connect.ServerStream[devicepb.SystemSnapshotFrame]) error {
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

	start := func(cmd command.Command, apply func(res command.Result)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			handle, err := sd.Stream(ctx, cmd)
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

	start(mikrotiksystem.NewStreamClockCommand(interval), func(res command.Result) {
		state.setClock(mikrotiksystem.ParseClock(res))
	})
	start(mikrotiksystem.NewStreamResourceCommand(interval), func(res command.Result) {
		state.setResource(mikrotiksystem.ParseResource(res))
	})
	start(mikrotiksystem.NewStreamRouterboardCommand(interval), func(res command.Result) {
		state.setRouterboard(mikrotiksystem.ParseRouterboard(res))
	})
	start(mikrotiksystem.NewStreamIdentityCommand(interval), func(res command.Result) {
		state.setIdentity(mikrotiksystem.ParseIdentity(res).Name)
	})
	start(mikrotiksystem.NewStreamHealthCommand(interval), func(res command.Result) {
		state.setHealth(mikrotiksystem.ParseHealth(res))
	})

	go func() { wg.Wait(); close(doneCh) }()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-doneCh:
			return nil
		case <-notify:
			// Coalesce updates arriving in the same tick (all five streams
			// share the same interval) into a single frame.
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

// StreamResource streams /system/resource/print interval=1s natively from
// RouterOS (no backend polling) and forwards each full snapshot.
func (h *HotspotConnectHandler) StreamResource(ctx context.Context, req *connect.Request[devicepb.StreamResourceRequest], stream *connect.ServerStream[devicepb.ResourceStreamData]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return response.Unimplemented("driver does not support streaming")
	}

	handle, err := sd.Stream(ctx, mikrotiksystem.NewStreamResourceCommand("1s"))
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
			sysRes := mikrotiksystem.ParseResource(res)
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
