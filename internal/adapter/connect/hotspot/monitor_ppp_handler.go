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
	mikrotikppp "github.com/quixiq/polyglot/internal/driver/mikrotik/ppp"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

// StreamPPPActive streams /ppp/active/print follow natively from
// RouterOS — maintains a local state map of active PPPoE sessions updated
// on each event and pushes the updated snapshot to the client.
func (h *HotspotConnectHandler) StreamPPPActive(ctx context.Context, req *connect.Request[devicepb.StreamPPPActiveRequest], stream *connect.ServerStream[devicepb.PPPActiveFrame]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("driver does not support streaming"))
	}

	activeMap := make(map[string]port.PPPActiveSession)

	// Pre-populate with initial snapshot from one-shot query
	if initial, err := h.activeSessionsUseCase.GetPPPActiveSessions(ctx, driver); err == nil {
		for _, s := range initial {
			activeMap[s.RosID] = s
		}
	}

	initialItems := make([]*devicepb.PPPActiveSessionItem, 0, len(activeMap))
	for _, s := range activeMap {
		initialItems = append(initialItems, &devicepb.PPPActiveSessionItem{
			Id:        s.RosID,
			Name:      s.Name,
			Service:   s.Service,
			CallerId:  s.CallerID,
			Address:   s.Address,
			Encoding:  s.Encoding,
			SessionId: s.SessionID,
			Radius:    s.Radius,
		})
	}

	// Immediately send an initial frame
	_ = stream.Send(&devicepb.PPPActiveFrame{
		DeviceId:      req.Msg.DeviceId,
		TimestampUnix: time.Now().Unix(),
		Sessions:      initialItems,
	})

	handle, err := sd.Stream(ctx, mikrotikppp.NewStreamActiveCommand(""))
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
			for _, row := range res.Rows {
				id := row[".id"]
				if id == "" {
					continue
				}
				if strings.EqualFold(row[".dead"], "yes") || strings.EqualFold(row[".dead"], "true") {
					delete(activeMap, id)
					continue
				}
				name := row["name"]
				if name == "" {
					continue
				}
				activeMap[id] = port.PPPActiveSession{
					RosID:     id,
					Name:      name,
					Service:   row["service"],
					CallerID:  row["caller-id"],
					Address:   row["address"],
					Encoding:  row["encoding"],
					SessionID: row["session-id"],
					Radius:    strings.EqualFold(row["radius"], "true"),
				}
			}

			items := make([]*devicepb.PPPActiveSessionItem, 0, len(activeMap))
			for _, s := range activeMap {
				items = append(items, &devicepb.PPPActiveSessionItem{
					Id:        s.RosID,
					Name:      s.Name,
					Service:   s.Service,
					CallerId:  s.CallerID,
					Address:   s.Address,
					Encoding:  s.Encoding,
					SessionId: s.SessionID,
					Radius:    s.Radius,
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

// pppSessionState keeps the latest full state of the PPPoE secret
// directory and the active session list via live map tracking.
type pppSessionState struct {
	mu        sync.Mutex
	secretMap map[string]port.PPPoESecret
	activeMap map[string]port.PPPActiveSession
}

func newPPPSessionState() *pppSessionState {
	return &pppSessionState{
		secretMap: make(map[string]port.PPPoESecret),
		activeMap: make(map[string]port.PPPActiveSession),
	}
}

func (s *pppSessionState) updateSecret(row map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := row[".id"]
	if id == "" {
		return
	}
	if strings.EqualFold(row[".dead"], "yes") || strings.EqualFold(row[".dead"], "true") {
		delete(s.secretMap, id)
		return
	}
	name := row["name"]
	if name == "" {
		return
	}
	s.secretMap[id] = port.PPPoESecret{
		RosID:         id,
		Name:          name,
		Profile:       row["profile"],
		Service:       row["service"],
		LocalAddress:  row["local-address"],
		RemoteAddress: row["remote-address"],
		Comment:       row["comment"],
		Disabled:      strings.EqualFold(row["disabled"], "true"),
		LastLoggedOut: row["last-logged-out"],
		CallerID:      row["caller-id"],
	}
}

func (s *pppSessionState) updateActive(row map[string]string) {
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
	name := row["name"]
	if name == "" {
		return
	}
	s.activeMap[id] = port.PPPActiveSession{
		RosID:     id,
		Name:      name,
		Service:   row["service"],
		CallerID:  row["caller-id"],
		Address:   row["address"],
		Encoding:  row["encoding"],
		SessionID: row["session-id"],
		Radius:    strings.EqualFold(row["radius"], "true"),
	}
}

// inactive returns subscriber secrets without an active session.
func (s *pppSessionState) inactive() []port.PPPoESecret {
	s.mu.Lock()
	defer s.mu.Unlock()
	secrets := make([]port.PPPoESecret, 0, len(s.secretMap))
	for _, sec := range s.secretMap {
		secrets = append(secrets, sec)
	}
	active := make([]port.PPPActiveSession, 0, len(s.activeMap))
	for _, a := range s.activeMap {
		active = append(active, a)
	}
	return mikrotikppp.FilterInactiveSecrets(secrets, active)
}

// StreamPPPInactive computes the inactive PPPoE subscriber list from two
// native RouterOS follow streams — /ppp/secret/print follow and
// /ppp/active/print follow. Inactive secrets are those without a
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

	state := newPPPSessionState()

	// Pre-populate with initial snapshot from one-shot queries so inactive list is available instantly
	initialInactive, _ := h.activeSessionsUseCase.GetPPPInactiveSessions(ctx, driver)
	initialActive, _ := h.activeSessionsUseCase.GetPPPActiveSessions(ctx, driver)
	for _, s := range initialInactive {
		state.secretMap[s.RosID] = s
	}
	for _, a := range initialActive {
		state.activeMap[a.RosID] = a
	}

	initialItems := make([]*devicepb.PPPSecretItem, 0, len(initialInactive))
	for _, s := range initialInactive {
		initialItems = append(initialItems, &devicepb.PPPSecretItem{
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

	// Immediately send an initial frame
	_ = stream.Send(&devicepb.PPPInactiveFrame{
		DeviceId:      req.Msg.DeviceId,
		TimestampUnix: time.Now().Unix(),
		Secrets:       initialItems,
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

	streamSecretsCmd := command.Command{
		Raw:  "/ppp/secret/print",
		Args: map[string]string{"follow": ""},
	}
	start(streamSecretsCmd, func(row map[string]string) {
		state.updateSecret(row)
	})
	start(mikrotikppp.NewStreamActiveCommand(""), func(row map[string]string) {
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
