package ppp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
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

	handle, err := sd.Stream(ctx, mikrotik.NewStreamPPPActiveCommand(req.Msg.NameFilter))
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
				name := row["name"]
				if name == "" {
					continue
				}
				activeMap[id] = port.PPPActiveSession{
					RosID:         id,
					Name:          name,
					Service:       row["service"],
					CallerID:      row["caller-id"],
					Address:       row["address"],
					Uptime:        row["uptime"],
					Encoding:      row["encoding"],
					SessionID:     row["session-id"],
					LimitBytesIn:  row["limit-bytes-in"],
					LimitBytesOut: row["limit-bytes-out"],
					Radius:        strings.EqualFold(row["radius"], "true"),
				}
			}

			items := make([]*devicepb.PPPActiveSession, 0, len(activeMap))
			for _, s := range activeMap {
				items = append(items, ToProtoPPPActiveSession(s))
			}

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
