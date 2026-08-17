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

// StreamPPPActive streams /ppp/active/print interval=1s natively from
// RouterOS — every tick RouterOS re-sends the full active PPPoE session
// list, so the frame always carries a complete snapshot.
func (h *HotspotConnectHandler) StreamPPPActive(ctx context.Context, req *connect.Request[devicepb.StreamPPPActiveRequest], stream *connect.ServerStream[devicepb.PPPActiveFrame]) error {
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

	handle, err := sd.Stream(ctx, mikrotik.NewStreamPPPActiveIntervalCommand("", interval))
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
			sessions := mikrotik.ParsePPPActiveSessions(res)
			items := make([]*devicepb.PPPActiveSessionItem, 0, len(sessions))
			for _, s := range sessions {
				items = append(items, &devicepb.PPPActiveSessionItem{
					Id:            s.RosID,
					Name:          s.Name,
					Service:       s.Service,
					CallerId:      s.CallerID,
					Address:       s.Address,
					Uptime:        s.Uptime,
					Encoding:      s.Encoding,
					SessionId:     s.SessionID,
					LimitBytesIn:  s.LimitBytesIn,
					LimitBytesOut: s.LimitBytesOut,
					Radius:        s.Radius,
				})
			}

			err := stream.Send(&devicepb.PPPActiveFrame{
				DeviceId:      req.Msg.DeviceId,
				TimestampUnix: time.Now().Unix(),
				Sessions:      items,
			})
			if err != nil {
				return err
			}
		}
	}
}

// pppSessionState keeps the latest full snapshots of the PPPoE secret
// directory and the active session list, fed by interval streams (full list
// per tick).
type pppSessionState struct {
	mu      sync.Mutex
	secrets []mikrotik.PPPoESecret
	active  []mikrotik.PPPActiveSession
}

func (s *pppSessionState) setSecrets(v []mikrotik.PPPoESecret) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.secrets = v
}

func (s *pppSessionState) setActive(v []mikrotik.PPPActiveSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = v
}

// inactive returns subscriber secrets without an active session.
func (s *pppSessionState) inactive() []mikrotik.PPPoESecret {
	s.mu.Lock()
	defer s.mu.Unlock()
	return mikrotik.FilterInactivePPPoESecrets(s.secrets, s.active)
}

// StreamPPPInactive computes the inactive PPPoE subscriber list from two
// native RouterOS interval streams — /ppp/secret/print and /ppp/active/print
// (both full snapshots per tick). Inactive secrets are those without a
// matching active session, mirroring legacy Mikhmon.
func (h *HotspotConnectHandler) StreamPPPInactive(ctx context.Context, req *connect.Request[devicepb.StreamPPPInactiveRequest], stream *connect.ServerStream[devicepb.PPPInactiveFrame]) error {
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

	state := &pppSessionState{}
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

	start(mikrotik.NewStreamPPPoESecretsIntervalCommand("", interval), func(res command.Result) {
		state.setSecrets(mikrotik.ParsePPPoESecrets(res))
	})
	start(mikrotik.NewStreamPPPActiveIntervalCommand("", interval), func(res command.Result) {
		state.setActive(mikrotik.ParsePPPActiveSessions(res))
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
			secrets := make([]*devicepb.PPPSecretItem, 0, len(inactive))
			for _, s := range inactive {
				secrets = append(secrets, &devicepb.PPPSecretItem{
					Id:            s.RosID,
					Name:          s.Name,
					Profile:       s.Profile,
					Service:       s.Service,
					LocalAddress:  s.LocalAddress,
					RemoteAddress: s.RemoteAddress,
					Comment:       s.Comment,
					Disabled:      s.Disabled,
					LastLoggedOut: s.LastLoggedOut,
					CallerId:      s.CallerID,
				})
			}

			err := stream.Send(&devicepb.PPPInactiveFrame{
				DeviceId:      req.Msg.DeviceId,
				TimestampUnix: time.Now().Unix(),
				Secrets:       secrets,
			})
			if err != nil {
				return err
			}
		}
	}
}
