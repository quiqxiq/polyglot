package monitor

import (
	"context"
	"time"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/response"
)

// StreamLogs streams /log/print follow natively from RouterOS.
func (h *NetworkMonitorConnectHandler) StreamLogs(ctx context.Context, req *connect.Request[polyglotv1.StreamLogsRequest], stream *connect.ServerStream[polyglotv1.LogsStreamFrame]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return response.Unimplemented("driver does not support streaming")
	}

	handle, err := h.streamGW.StreamLogs(ctx, sd, req.Msg.Topics)
	if err != nil {
		logger.WithComponent("NetworkMonitorConnectHandler").WithError(err).Warn("starting log stream failed")
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
			logs := h.streamGW.ParseLogEntries(res)
			if len(logs) == 0 {
				continue
			}

			items := make([]*polyglotv1.LogEntryItem, 0, len(logs))
			for _, l := range logs {
				items = append(items, &polyglotv1.LogEntryItem{
					Id:      l.RosID,
					Time:    l.Time,
					Topics:  l.Topics,
					Message: l.Message,
				})
			}

			err := stream.Send(&polyglotv1.LogsStreamFrame{
				DeviceId:      req.Msg.DeviceId,
				TimestampUnix: time.Now().Unix(),
				Logs:          items,
			})
			if err != nil {
				return err
			}
		}
	}
}
