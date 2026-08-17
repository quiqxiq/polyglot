package hotspot

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

// StreamActiveSessions streams /ip/hotspot/active/print follow natively
// from RouterOS — maintains a local state map of active sessions updated
// on each event and pushes the updated snapshot to the client.
func (h *HotspotConnectHandler) StreamActiveSessions(ctx context.Context, req *connect.Request[devicepb.StreamActiveSessionsRequest], stream *connect.ServerStream[devicepb.ActiveSessionsStreamData]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("driver does not support streaming"))
	}

	activeMap := make(map[string]mikrotik.HotspotActiveSession)

	// Pre-populate with initial snapshot from one-shot query
	if initial, err := h.useCase.GetActiveSessions(ctx, driver); err == nil {
		for _, s := range initial {
			activeMap[s.RosID] = s
		}
	}

	initialSessions := make([]mikrotik.HotspotActiveSession, 0, len(activeMap))
	for _, s := range activeMap {
		initialSessions = append(initialSessions, s)
	}

	// Immediately send initial snapshot frame so frontend has data right away
	_ = stream.Send(&devicepb.ActiveSessionsStreamData{
		DeviceId:      req.Msg.DeviceId,
		Sessions:      ToProtoActiveSessions(initialSessions),
		TimestampUnix: time.Now().Unix(),
	})

	handle, err := sd.Stream(ctx, mikrotik.NewStreamHotspotActiveCommand(req.Msg.UserFilter))
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
			for _, row := range res.Rows {
				id := row[".id"]
				if id == "" {
					continue
				}
				if strings.EqualFold(row[".dead"], "yes") || strings.EqualFold(row[".dead"], "true") {
					delete(activeMap, id)
					continue
				}
				user := row["user"]
				if user == "" {
					continue
				}
				activeMap[id] = mikrotik.HotspotActiveSession{
					RosID:           id,
					Server:          row["server"],
					User:            user,
					Address:         row["address"],
					MACAddress:      row["mac-address"],
					LoginBy:         row["login-by"],
					Uptime:          row["uptime"],
					SessionTimeLeft: row["session-time-left"],
					IdleTime:        row["idle-time"],
					BytesIn:         row["bytes-in"],
					BytesOut:        row["bytes-out"],
					PacketsIn:       row["packets-in"],
					PacketsOut:      row["packets-out"],
				}
			}

			sessions := make([]mikrotik.HotspotActiveSession, 0, len(activeMap))
			for _, s := range activeMap {
				sessions = append(sessions, s)
			}

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

// hotspotSessionState keeps the latest full state of the hotspot user
// directory and the active session list via live map tracking.
type hotspotSessionState struct {
	mu        sync.Mutex
	userMap   map[string]mikrotik.HotspotUser
	activeMap map[string]mikrotik.HotspotActiveSession
}

func newHotspotSessionState() *hotspotSessionState {
	return &hotspotSessionState{
		userMap:   make(map[string]mikrotik.HotspotUser),
		activeMap: make(map[string]mikrotik.HotspotActiveSession),
	}
}

func (s *hotspotSessionState) updateUser(row map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := row[".id"]
	if id == "" {
		return
	}
	if strings.EqualFold(row[".dead"], "yes") || strings.EqualFold(row[".dead"], "true") {
		delete(s.userMap, id)
		return
	}
	name := row["name"]
	if name == "" {
		return
	}
	s.userMap[id] = mikrotik.HotspotUser{
		RosID:         id,
		Name:          name,
		Password:      row["password"],
		Profile:       row["profile"],
		Server:        row["server"],
		MACAddress:    row["mac-address"],
		Address:       row["address"],
		LimitUptime:   row["limit-uptime"],
		LimitBytesIn:  row["limit-bytes-in"],
		LimitBytesOut: row["limit-bytes-out"],
		Comment:       row["comment"],
		Disabled:      strings.EqualFold(row["disabled"], "true"),
	}
}

func (s *hotspotSessionState) updateActive(row map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := row[".id"]
	if id == "" {
		return
	}
	if strings.EqualFold(row[".dead"], "yes") || strings.EqualFold(row[".dead"], "true") {
		delete(s.activeMap, id)
		return
	}
	user := row["user"]
	if user == "" {
		return
	}
	s.activeMap[id] = mikrotik.HotspotActiveSession{
		RosID:           id,
		Server:          row["server"],
		User:            user,
		Address:         row["address"],
		MACAddress:      row["mac-address"],
		LoginBy:         row["login-by"],
		Uptime:          row["uptime"],
		SessionTimeLeft: row["session-time-left"],
		IdleTime:        row["idle-time"],
		BytesIn:         row["bytes-in"],
		BytesOut:        row["bytes-out"],
		PacketsIn:       row["packets-in"],
		PacketsOut:      row["packets-out"],
	}
}

// inactive returns users without an active session (legacy inactive list).
func (s *hotspotSessionState) inactive() []mikrotik.HotspotUser {
	s.mu.Lock()
	defer s.mu.Unlock()
	users := make([]mikrotik.HotspotUser, 0, len(s.userMap))
	for _, u := range s.userMap {
		users = append(users, u)
	}
	active := make([]mikrotik.HotspotActiveSession, 0, len(s.activeMap))
	for _, a := range s.activeMap {
		active = append(active, a)
	}
	return mikrotik.FilterInactiveHotspotUsers(users, active)
}

// StreamHotspotInactive computes the inactive hotspot user list from two
// native RouterOS follow streams — /ip/hotspot/user/print follow and
// /ip/hotspot/active/print follow. Inactive users are those without a matching
// active session, mirroring legacy Mikhmon.
func (h *HotspotConnectHandler) StreamHotspotInactive(ctx context.Context, req *connect.Request[devicepb.StreamHotspotInactiveRequest], stream *connect.ServerStream[devicepb.HotspotInactiveFrame]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("driver does not support streaming"))
	}

	state := newHotspotSessionState()

	// Pre-populate with initial snapshot from one-shot queries so inactive list is available instantly
	initialUsers, _ := h.useCase.GetUsers(ctx, driver, port.ListUsersFilter{})
	initialActive, _ := h.useCase.GetActiveSessions(ctx, driver)
	for _, u := range initialUsers {
		state.userMap[u.RosID] = u
	}
	for _, a := range initialActive {
		state.activeMap[a.RosID] = a
	}

	initialInactive := state.inactive()
	// Immediately send initial snapshot frame
	_ = stream.Send(&devicepb.HotspotInactiveFrame{
		DeviceId:      req.Msg.DeviceId,
		TimestampUnix: time.Now().Unix(),
		Users:         ToProtoHotspotUsers(initialInactive),
	})

	notify := make(chan struct{}, 10)
	doneCh := make(chan struct{})
	var wg sync.WaitGroup

	start := func(cmd command.Command, apply func(row map[string]string)) {
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
					for _, r := range res.Rows {
						apply(r)
					}
					select {
					case notify <- struct{}{}:
					default:
					}
				}
			}
		}()
	}

	start(mikrotik.NewStreamHotspotUsersCommand(""), func(row map[string]string) {
		state.updateUser(row)
	})
	start(mikrotik.NewStreamHotspotActiveCommand(""), func(row map[string]string) {
		state.updateActive(row)
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
