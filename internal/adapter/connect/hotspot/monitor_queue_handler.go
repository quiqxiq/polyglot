package hotspot

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	mikrotikqueue "github.com/quixiq/polyglot/internal/driver/mikrotik/queue"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

// StreamQueueStats streams /queue/simple/print stats interval=<n> natively
// from RouterOS. Supports the same filters Mikhmon used for its queue
// dashboard: by queue name, by parent queue, or static parent queues only
// (dynamic=false). Each tick pushes the full list of matching queues.
func (h *HotspotConnectHandler) StreamQueueStats(ctx context.Context, req *connect.Request[devicepb.StreamQueueStatsRequest], stream *connect.ServerStream[devicepb.QueueStatsFrame]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("driver does not support streaming"))
	}

	params := mikrotikqueue.QueueStreamParams{
		NameFilter:   req.Msg.Name,
		ParentFilter: req.Msg.Parent,
		ParentsOnly:  req.Msg.ParentsOnly,
		Interval:     req.Msg.Interval,
	}
	if params.Interval == "" {
		params.Interval = "1s"
	}

	handle, err := sd.Stream(ctx, mikrotikqueue.NewStreamQueueStatsCommand(params))
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
			queues := mikrotikqueue.ParseSimpleQueues(res)
			items := make([]*devicepb.QueueStatsItem, 0, len(queues))
			for _, q := range queues {
				items = append(items, &devicepb.QueueStatsItem{
					Id:         q.RosID,
					Name:       q.Name,
					Target:     q.Target,
					Parent:     q.Parent,
					MaxLimit:   q.MaxLimit,
					LimitAt:    q.LimitAt,
					Queue:      q.Queue,
					Bytes:      q.Bytes,
					Packets:    q.Packets,
					Dropped:    q.Dropped,
					Rate:       q.Rate,
					PacketRate: q.PacketRate,
					Disabled:   q.Disabled,
				})
			}

			err := stream.Send(&devicepb.QueueStatsFrame{
				DeviceId:      req.Msg.DeviceId,
				TimestampUnix: time.Now().Unix(),
				Queues:        items,
			})
			if err != nil {
				return err
			}
		}
	}
}
