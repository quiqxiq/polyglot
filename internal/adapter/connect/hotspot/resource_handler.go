package hotspot

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/pkg/response"
)

// ListParentQueues returns all static parent /queue/simple entries
// (dynamic=false). Used by the web UI to suggest parent queues when
// editing simple queues.
func (h *HotspotConnectHandler) ListParentQueues(ctx context.Context, req *connect.Request[devicepb.ListParentQueuesRequest]) (*connect.Response[devicepb.ListParentQueuesResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	queues, err := h.useCase.GetParentQueues(ctx, driver)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	entries := make([]*devicepb.ParentQueueEntry, 0, len(queues))
	for _, q := range queues {
		entries = append(entries, &devicepb.ParentQueueEntry{
			RosId:    q.RosID,
			Name:     q.Name,
			MaxLimit: q.MaxLimit,
		})
	}

	return connect.NewResponse(&devicepb.ListParentQueuesResponse{Queues: entries}), nil
}

// ListIPPools returns all IP pools (/ip/pool/print). Used by the web UI
// to suggest pool names and ranges when editing hotspot resources.
func (h *HotspotConnectHandler) ListIPPools(ctx context.Context, req *connect.Request[devicepb.ListIPPoolsRequest]) (*connect.Response[devicepb.ListIPPoolsResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	pools, err := h.useCase.GetIPPools(ctx, driver)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	entries := make([]*devicepb.IPPoolEntry, 0, len(pools))
	for _, p := range pools {
		entries = append(entries, &devicepb.IPPoolEntry{
			RosId:  p.RosID,
			Name:   p.Name,
			Ranges: p.Ranges,
		})
	}

	return connect.NewResponse(&devicepb.ListIPPoolsResponse{Pools: entries}), nil
}
