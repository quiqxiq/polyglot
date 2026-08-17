package hotspot

import (
	"context"
	"fmt"
	"sync"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

// StreamActiveSessions streams /ip/hotspot/active/print interval=1s natively
// from RouterOS — every tick RouterOS re-sends the full active session list,
// so the frame always carries a complete snapshot (no backend polling).
func (h *HotspotConnectHandler) StreamActiveSessions(ctx context.Context, req *connect.Request[devicepb.StreamActiveSessionsRequest], stream *connect.ServerStream[devicepb.ActiveSessionsStreamData]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("driver does not support streaming"))
	}

	handle, err := sd.Stream(ctx, mikrotik.NewStreamHotspotActiveIntervalCommand(req.Msg.UserFilter, "1s"))
	if err != nil {
		return response.MapDomainError(err)
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
			sessions := mikrotik.ParseHotspotActiveSessions(res)
			err := stream.Send(&devicepb.ActiveSessionsStreamData{
				DeviceId:      req.Msg.DeviceId,
				Sessions:      ToProtoActiveSessions(sessions),
				TimestampUnix: time.Now().Unix(),
			})
			if err != nil {
				return err
			}
		}
	}
}

// hotspotSessionState keeps the latest full snapshots of the hotspot user
// directory and the active session list. Both are fed by interval streams
// (full list per tick), so the merged view is always complete.
type hotspotSessionState struct {
	mu     sync.Mutex
	users  []mikrotik.HotspotUser
	active []mikrotik.HotspotActiveSession
}

func (s *hotspotSessionState) setUsers(u []mikrotik.HotspotUser) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users = u
}

func (s *hotspotSessionState) setActive(a []mikrotik.HotspotActiveSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = a
}

// inactive returns users without an active session (legacy inactive list).
func (s *hotspotSessionState) inactive() []mikrotik.HotspotUser {
	s.mu.Lock()
	defer s.mu.Unlock()
	return mikrotik.FilterInactiveHotspotUsers(s.users, s.active)
}

// StreamHotspotInactive computes the inactive hotspot user list from two
// native RouterOS interval streams — /ip/hotspot/user/print and
// /ip/hotspot/active/print (both full snapshots per tick). Inactive users are
// those without a matching active session, mirroring legacy Mikhmon.
func (h *HotspotConnectHandler) StreamHotspotInactive(ctx context.Context, req *connect.Request[devicepb.StreamHotspotInactiveRequest], stream *connect.ServerStream[devicepb.HotspotInactiveFrame]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("driver does not support streaming"))
	}

	interval := req.Msg.Interval
	if interval == "" {
		interval = "1s"
	}

	state := &hotspotSessionState{}
	notify := make(chan struct{}, 4)
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
			defer handle.Cancel()
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

	start(mikrotik.NewStreamHotspotUsersIntervalCommand("", interval), func(res command.Result) {
		state.setUsers(mikrotik.ParseHotspotUsers(res))
	})
	start(mikrotik.NewStreamHotspotActiveIntervalCommand("", interval), func(res command.Result) {
		state.setActive(mikrotik.ParseHotspotActiveSessions(res))
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
			inactive := state.inactive()
			err := stream.Send(&devicepb.HotspotInactiveFrame{
				DeviceId:      req.Msg.DeviceId,
				TimestampUnix: time.Now().Unix(),
				Users:         ToProtoHotspotUsers(inactive),
			})
			if err != nil {
				return err
			}
		}
	}
}
