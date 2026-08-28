package hotspot

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	mikrotiksystem "github.com/quixiq/polyglot/internal/driver/mikrotik/system"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/response"
)

// StreamLogs streams /log/print follow natively from RouterOS — initial logs
// are streamed first and each new log line is pushed as it is written.
func (h *HotspotConnectHandler) StreamLogs(ctx context.Context, req *connect.Request[devicepb.StreamLogsRequest], stream *connect.ServerStream[devicepb.LogsStreamFrame]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("driver does not support streaming"))
	}

	handle, err := sd.Stream(ctx, mikrotiksystem.NewStreamLogsCommand(req.Msg.Topics))
	if err != nil {
		logger.WithComponent("HotspotConnectHandler").WithError(err).Warn("starting log stream failed")
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
			logs := mikrotiksystem.ParseLogs(res)
			if len(logs) == 0 {
				continue
			}

			items := make([]*devicepb.LogEntryItem, 0, len(logs))
			for _, l := range logs {
				items = append(items, &devicepb.LogEntryItem{
					Id:      l.RosID,
					Time:    l.Time,
					Topics:  l.Topics,
					Message: l.Message,
				})
			}

			err := stream.Send(&devicepb.LogsStreamFrame{
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
