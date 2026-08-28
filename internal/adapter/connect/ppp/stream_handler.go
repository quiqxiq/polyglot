package ppp

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	mikrotikppp "github.com/quixiq/polyglot/internal/driver/mikrotik/ppp"
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
		return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("driver does not support streaming"))
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
		initialItems = append(initialItems, ToProtoPPPActiveSession(s))
	}

	// Immediately send an initial frame
	_ = stream.Send(&devicepb.PPPActiveSessionsFrame{
		DeviceId:      req.Msg.DeviceId,
		TimestampUnix: time.Now().Unix(),
		Sessions:      initialItems,
	})

	handle, err := sd.Stream(ctx, mikrotikppp.NewStreamActiveCommand(req.Msg.NameFilter))
	if err != nil {
		return response.MapDomainError(err)
	}
	defer handle.Cancel()

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
				if strings.EqualFold(row[".dead"], "yes") || strings.EqualFold(row[".dead"], "true") {
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
				}
			}

			items := make([]*devicepb.PPPActiveSession, 0, len(activeMap))
			for _, s := range activeMap {
				items = append(items, ToProtoPPPActiveSession(s))
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

// StreamActiveStats streams /ppp/active/print stats interval=<interval> natively from RouterOS.
// It pumps dynamic telemetry stats (uptime, bytes-in/out, packets-in/out) per interval with minimal .proplist payload.
func (h *PPPConnectHandler) StreamActiveStats(ctx context.Context, req *connect.Request[devicepb.StreamPPPActiveStatsRequest], stream *connect.ServerStream[devicepb.PPPActiveStatsFrame]) error {
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

	handle, err := sd.Stream(ctx, mikrotikppp.NewStreamActiveStatsCommand(interval))
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
			stats := mikrotikppp.ParseActiveStats(res)
			if len(stats) == 0 {
				continue
			}

			if err := stream.Send(&devicepb.PPPActiveStatsFrame{
				DeviceId:      req.Msg.DeviceId,
				TimestampUnix: time.Now().Unix(),
				Stats:         ToProtoPPPActiveStats(stats),
			}); err != nil {
				return err
			}
		}
	}
}

// StreamInactiveSecrets streams offline subscribers via periodic push over ConnectRPC.
func (h *PPPConnectHandler) StreamInactiveSecrets(ctx context.Context, req *connect.Request[devicepb.StreamPPPInactiveSecretsRequest], stream *connect.ServerStream[devicepb.PPPInactiveSecretsFrame]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sendSnapshot := func() error {
		inactive, err := h.useCase.ListInactive(ctx, driver)
		if err != nil {
			return err
		}
		return stream.Send(&devicepb.PPPInactiveSecretsFrame{
			DeviceId:      req.Msg.DeviceId,
			TimestampUnix: time.Now().Unix(),
			Secrets:       ToProtoPPPSecrets(inactive),
		})
	}

	if err := sendSnapshot(); err != nil {
		return response.MapDomainError(err)
	}

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := sendSnapshot(); err != nil {
				return response.MapDomainError(err)
			}
		}
	}
}
