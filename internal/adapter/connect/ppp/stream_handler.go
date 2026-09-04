package ppp

import (
	"context"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

// StreamActiveSessions streams /ppp/active/print follow natively from RouterOS.
func (h *PPPConnectHandler) StreamActiveSessions(ctx context.Context, req *connect.Request[devicepb.StreamPPPActiveSessionsRequest], stream *connect.ServerStream[devicepb.PPPActiveSessionsFrame]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return response.Unimplemented("driver does not support streaming")
	}

	activeMap := make(map[string]port.PPPActiveSession)

	// Pre-populate with initial snapshot
	if initial, err := h.useCase.ListActive(ctx, driver, req.Msg.NameFilter); err == nil {
		for _, s := range initial {
			activeMap[s.RosID] = s
		}
	}

	initialItems := make([]*devicepb.PPPActiveSession, 0, len(activeMap))
	for _, s := range activeMap {
		initialItems = append(initialItems, toProtoPPPActiveSession(s))
	}

	// Immediately send an initial frame
	_ = stream.Send(&devicepb.PPPActiveSessionsFrame{
		DeviceId:      req.Msg.DeviceId,
		TimestampUnix: time.Now().Unix(),
		Sessions:      initialItems,
	})

	handle, err := h.streamGW.StreamPPPActive(ctx, sd, req.Msg.NameFilter)
	if err != nil {
		return response.MapDomainError(err)
	}
	defer func() { _ = handle.Cancel() }()

	secretProfileMap := make(map[string]string)
	if secrets, err := h.useCase.ListSecrets(ctx, driver, req.Msg.NameFilter); err == nil {
		for _, s := range secrets {
			if s.Profile != "" {
				secretProfileMap[s.Name] = s.Profile
			}
		}
	}

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
				if strings.EqualFold(row[".dead"], "yes") || strings.EqualFold(row[".dead"], "true") || strings.EqualFold(row[".action"], "remove") {
					delete(activeMap, id)
					continue
				}
				name := row["name"]
				if name == "" {
					continue
				}
				profile := row["profile"]
				if profile == "" {
					if p, ok := secretProfileMap[name]; ok {
						profile = p
					}
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
					Profile:   profile,
					Uptime:    row["uptime"],
				}
			}

			items := make([]*devicepb.PPPActiveSession, 0, len(activeMap))
			for _, s := range activeMap {
				items = append(items, toProtoPPPActiveSession(s))
			}
			sort.Slice(items, func(i, j int) bool {
				return items[i].Name < items[j].Name
			})

			if err := stream.Send(&devicepb.PPPActiveSessionsFrame{
				DeviceId:      req.Msg.DeviceId,
				TimestampUnix: time.Now().Unix(),
				Sessions:      items,
			}); err != nil {
				return err
			}
		}
	}
}

// StreamInactiveSecrets streams offline subscribers via dual native RouterOS follow (/ppp/active/print follow and /ppp/secret/print follow).
func (h *PPPConnectHandler) StreamInactiveSecrets(ctx context.Context, req *connect.Request[devicepb.StreamPPPInactiveSecretsRequest], stream *connect.ServerStream[devicepb.PPPInactiveSecretsFrame]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return response.Unimplemented("driver does not support streaming")
	}

	secretsList, err := h.useCase.ListSecrets(ctx, driver, "")
	if err != nil {
		return response.MapDomainError(err)
	}
	activeList, err := h.useCase.ListActive(ctx, driver, "")
	if err != nil {
		return response.MapDomainError(err)
	}

	secretsMap := make(map[string]port.PPPoESecret)
	for _, s := range secretsList {
		secretsMap[s.RosID] = s
	}

	activeNames := make(map[string]bool)
	for _, a := range activeList {
		activeNames[a.Name] = true
	}

	getInactiveList := func() []*devicepb.PPPSecret {
		inactive := make([]*devicepb.PPPSecret, 0)
		for _, s := range secretsMap {
			if !activeNames[s.Name] {
				inactive = append(inactive, toProtoPPPSecret(s))
			}
		}
		sort.Slice(inactive, func(i, j int) bool {
			return inactive[i].Name < inactive[j].Name
		})
		return inactive
	}

	// Send initial snapshot immediately
	if err := stream.Send(&devicepb.PPPInactiveSecretsFrame{
		DeviceId:      req.Msg.DeviceId,
		TimestampUnix: time.Now().Unix(),
		Secrets:       getInactiveList(),
	}); err != nil {
		return err
	}

	handleActive, err := h.streamGW.StreamPPPActive(ctx, sd, "")
	if err != nil {
		return response.MapDomainError(err)
	}
	defer func() { _ = handleActive.Cancel() }()

	handleSecrets, err := h.streamGW.StreamPPPSecrets(ctx, sd, "")
	if err != nil {
		return response.MapDomainError(err)
	}
	defer func() { _ = handleSecrets.Cancel() }()

	for {
		select {
		case <-ctx.Done():
			return nil
		case res, ok := <-handleActive.Chan():
			if !ok {
				return handleActive.Err()
			}
			rows := append([]map[string]string(nil), res.Rows...)
		drainActive:
			for {
				select {
				case next, nextOk := <-handleActive.Chan():
					if !nextOk {
						break drainActive
					}
					rows = append(rows, next.Rows...)
				default:
					break drainActive
				}
			}
			changed := false
			for _, row := range rows {
				name := row["name"]
				if name == "" {
					continue
				}
				isDead := strings.EqualFold(row[".dead"], "yes") || strings.EqualFold(row[".dead"], "true") || strings.EqualFold(row[".action"], "remove")
				if isDead {
					if activeNames[name] {
						delete(activeNames, name)
						changed = true
					}
				} else {
					if !activeNames[name] {
						activeNames[name] = true
						changed = true
					}
				}
			}
			if changed {
				if err := stream.Send(&devicepb.PPPInactiveSecretsFrame{
					DeviceId:      req.Msg.DeviceId,
					TimestampUnix: time.Now().Unix(),
					Secrets:       getInactiveList(),
				}); err != nil {
					return err
				}
			}
		case res, ok := <-handleSecrets.Chan():
			if !ok {
				return handleSecrets.Err()
			}
			rows := append([]map[string]string(nil), res.Rows...)
		drainSecrets:
			for {
				select {
				case next, nextOk := <-handleSecrets.Chan():
					if !nextOk {
						break drainSecrets
					}
					rows = append(rows, next.Rows...)
				default:
					break drainSecrets
				}
			}
			changed := false
			for _, row := range rows {
				id := row[".id"]
				if id == "" {
					continue
				}
				isDead := strings.EqualFold(row[".dead"], "yes") || strings.EqualFold(row[".dead"], "true") || strings.EqualFold(row[".action"], "remove")
				if isDead {
					if _, ok := secretsMap[id]; ok {
						delete(secretsMap, id)
						changed = true
					}
					continue
				}
				name := row["name"]
				if name == "" {
					continue
				}
				callerID := row["caller-id"]
				if callerID == "" {
					callerID = row["last-caller-id"]
				}
				disabled := strings.EqualFold(row["disabled"], "yes") || strings.EqualFold(row["disabled"], "true")
				secretsMap[id] = port.PPPoESecret{
					RosID:         id,
					Name:          name,
					Profile:       row["profile"],
					Service:       row["service"],
					LocalAddress:  row["local-address"],
					RemoteAddress: row["remote-address"],
					Comment:       row["comment"],
					Disabled:      disabled,
					LastLoggedOut: row["last-logged-out"],
					CallerID:      callerID,
				}
				changed = true
			}
			if changed {
				if err := stream.Send(&devicepb.PPPInactiveSecretsFrame{
					DeviceId:      req.Msg.DeviceId,
					TimestampUnix: time.Now().Unix(),
					Secrets:       getInactiveList(),
				}); err != nil {
					return err
				}
			}
		}
	}
}
