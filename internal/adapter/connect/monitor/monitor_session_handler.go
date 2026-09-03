package monitor

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/command"
	domainHotspot "github.com/quixiq/polyglot/internal/domain/hotspot"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

// StreamActiveSessions streams /ip/hotspot/active/print follow natively from RouterOS.
func (h *NetworkMonitorConnectHandler) StreamActiveSessions(ctx context.Context, req *connect.Request[devicepb.StreamActiveSessionsRequest], stream *connect.ServerStream[devicepb.ActiveSessionsStreamData]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return response.Unimplemented("driver does not support streaming")
	}

	activeMap := make(map[string]port.HotspotActiveSession)

	// Pre-populate with initial snapshot from one-shot query
	if h.hotspotUC != nil {
		if initial, err := h.hotspotUC.GetActiveSessions(ctx, driver); err == nil {
			for _, s := range initial {
				activeMap[s.RosID] = s
			}
		}
	}

	initialSessions := make([]port.HotspotActiveSession, 0, len(activeMap))
	for _, s := range activeMap {
		initialSessions = append(initialSessions, s)
	}
	sort.Slice(initialSessions, func(i, j int) bool {
		return initialSessions[i].User < initialSessions[j].User
	})

	// Immediately send initial snapshot frame so frontend has data right away
	_ = stream.Send(&devicepb.ActiveSessionsStreamData{
		DeviceId:      req.Msg.DeviceId,
		Sessions:      toProtoActiveSessions(initialSessions),
		TimestampUnix: time.Now().Unix(),
	})

	handle, err := h.streamGW.StreamHotspotActive(ctx, sd, req.Msg.UserFilter)
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
			rows := append([]map[string]string(nil), res.Rows...)
		drain:
			for {
				select {
				case next, nextOk := <-handle.Chan():
					if !nextOk {
						break drain
					}
					rows = append(rows, next.Rows...)
				default:
					break drain
				}
			}

			for _, row := range rows {
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
				activeMap[id] = port.HotspotActiveSession{
					RosID:      id,
					Server:     row["server"],
					User:       user,
					Address:    row["address"],
					MACAddress: row["mac-address"],
					LoginBy:    row["login-by"],
				}
			}

			sessions := make([]port.HotspotActiveSession, 0, len(activeMap))
			for _, s := range activeMap {
				sessions = append(sessions, s)
			}
			sort.Slice(sessions, func(i, j int) bool {
				return sessions[i].User < sessions[j].User
			})

			err := stream.Send(&devicepb.ActiveSessionsStreamData{
				DeviceId:      req.Msg.DeviceId,
				Sessions:      toProtoActiveSessions(sessions),
				TimestampUnix: time.Now().Unix(),
			})
			if err != nil {
				return err
			}
		}
	}
}

// StreamActiveStats streams /ip/hotspot/active/print stats interval=<interval> natively from RouterOS.
func (h *NetworkMonitorConnectHandler) StreamActiveStats(ctx context.Context, req *connect.Request[devicepb.StreamActiveStatsRequest], stream *connect.ServerStream[devicepb.ActiveStatsStreamData]) error {
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

	handle, err := h.streamGW.StreamHotspotActiveStats(ctx, sd, interval)
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
			stats := h.streamGW.ParseHotspotActiveStats(res)
			if len(stats) == 0 {
				continue
			}

			if err := stream.Send(&devicepb.ActiveStatsStreamData{
				DeviceId:      req.Msg.DeviceId,
				Stats:         toProtoActiveStats(stats),
				TimestampUnix: time.Now().Unix(),
			}); err != nil {
				return err
			}
		}
	}
}

type hotspotSessionState struct {
	mu        sync.Mutex
	userMap   map[string]port.HotspotUser
	activeMap map[string]port.HotspotActiveSession
}

func newHotspotSessionState() *hotspotSessionState {
	return &hotspotSessionState{
		userMap:   make(map[string]port.HotspotUser),
		activeMap: make(map[string]port.HotspotActiveSession),
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
	s.userMap[id] = port.HotspotUser{
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
		Uptime:        row["uptime"],
		BytesIn:       row["bytes-in"],
		BytesOut:      row["bytes-out"],
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
	s.activeMap[id] = port.HotspotActiveSession{
		RosID:      id,
		Server:     row["server"],
		User:       user,
		Address:    row["address"],
		MACAddress: row["mac-address"],
		LoginBy:    row["login-by"],
	}
}

func (s *hotspotSessionState) inactive() []port.HotspotUser {
	s.mu.Lock()
	defer s.mu.Unlock()
	users := make([]port.HotspotUser, 0, len(s.userMap))
	for _, u := range s.userMap {
		users = append(users, u)
	}
	active := make([]port.HotspotActiveSession, 0, len(s.activeMap))
	for _, a := range s.activeMap {
		active = append(active, a)
	}
	return domainHotspot.FilterInactiveUsers(users, active)
}

// StreamHotspotInactive computes the inactive hotspot user list from two
// native RouterOS follow streams.
func (h *NetworkMonitorConnectHandler) StreamHotspotInactive(ctx context.Context, req *connect.Request[devicepb.StreamHotspotInactiveRequest], stream *connect.ServerStream[devicepb.HotspotInactiveFrame]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return response.Unimplemented("driver does not support streaming")
	}

	state := newHotspotSessionState()

	if h.hotspotUC != nil {
		initialUsers, _ := h.hotspotUC.GetUsers(ctx, driver, port.ListUsersFilter{})
		initialActive, _ := h.hotspotUC.GetActiveSessions(ctx, driver)
		for _, u := range initialUsers {
			state.userMap[u.RosID] = u
		}
		for _, a := range initialActive {
			state.activeMap[a.RosID] = a
		}
	}

	initialInactive := state.inactive()
	_ = stream.Send(&devicepb.HotspotInactiveFrame{
		DeviceId:      req.Msg.DeviceId,
		TimestampUnix: time.Now().Unix(),
		Users:         toProtoHotspotUsers(initialInactive),
	})

	notify := make(chan struct{}, 10)
	doneCh := make(chan struct{})
	var wg sync.WaitGroup

	start := func(openStream func() (port.StreamHandle, error), apply func(row map[string]string)) {
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

	streamUsersCmd := command.Command{
		Raw:  "/ip/hotspot/user/print",
		Args: map[string]string{"follow": ""},
	}
	start(func() (port.StreamHandle, error) {
		return sd.Stream(ctx, streamUsersCmd)
	}, func(row map[string]string) {
		state.updateUser(row)
	})
	start(func() (port.StreamHandle, error) {
		return h.streamGW.StreamHotspotActive(ctx, sd, "")
	}, func(row map[string]string) {
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
				Users:         toProtoHotspotUsers(inactive),
			})
			if err != nil {
				return err
			}
		}
	}
}
