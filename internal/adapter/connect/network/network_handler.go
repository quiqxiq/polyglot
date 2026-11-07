package network

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/port"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
	"github.com/quixiq/polyglot/pkg/fault"
	"github.com/quixiq/polyglot/pkg/response"
)

// ConnectDriverProvider signature to obtain a port.DeviceDriver for a given deviceID.
type ConnectDriverProvider = iconnect.DriverProvider

// NetworkConnectHandler implements the NetworkService ConnectRPC service.
//
//nolint:revive // Explicit transport role is part of the project naming convention.
type NetworkConnectHandler struct {
	hotspotUC             *hotspotUC.UseCase
	activeSessionsUseCase *networkUC.ActiveSessionsUseCase
	driverProvider        ConnectDriverProvider
}

// NewNetworkConnectHandler constructs a NetworkConnectHandler.
func NewNetworkConnectHandler(
	hsUC *hotspotUC.UseCase,
	activeUC *networkUC.ActiveSessionsUseCase,
	provider ConnectDriverProvider,
) *NetworkConnectHandler {
	return &NetworkConnectHandler{
		hotspotUC:             hsUC,
		activeSessionsUseCase: activeUC,
		driverProvider:        provider,
	}
}

func (h *NetworkConnectHandler) getDriver(ctx context.Context, deviceID string) (port.DeviceDriver, error) {
	if deviceID == "" {
		return nil, response.InvalidArgument("device_id is required")
	}
	driver, err := h.driverProvider(ctx, deviceID)
	if err != nil {
		return nil, response.MapDomainError(fault.Wrap(fault.KindUnavailable, err))
	}
	return driver, nil
}

// ListDHCPLeases lists current DHCP leases from the target router.
func (h *NetworkConnectHandler) ListDHCPLeases(ctx context.Context, req *connect.Request[polyglotv1.ListDHCPLeasesRequest]) (*connect.Response[polyglotv1.ListDHCPLeasesResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if h.activeSessionsUseCase == nil {
		return nil, response.Unavailable("active sessions usecase unavailable")
	}

	leases, err := h.activeSessionsUseCase.GetDHCPLeases(ctx, driver, req.Msg.MacFilter)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.ListDHCPLeasesResponse{
		Leases: ToProtoDHCPLeases(leases),
	}), nil
}

// BlockDHCPLease modifies the blocked status of a DHCP lease on the target router.
func (h *NetworkConnectHandler) BlockDHCPLease(ctx context.Context, req *connect.Request[polyglotv1.BlockDHCPLeaseRequest]) (*connect.Response[polyglotv1.BlockDHCPLeaseResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if h.activeSessionsUseCase == nil {
		return nil, response.Unavailable("active sessions usecase unavailable")
	}

	res, err := h.activeSessionsUseCase.SetDHCPLeaseBlock(ctx, driver, req.Msg.RosId, req.Msg.Blocked, req.Msg.Comment)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.BlockDHCPLeaseResponse{
		Message: fmt.Sprintf("lease block status updated: output=%s", res.Output),
	}), nil
}

// ListParentQueues lists parent simple queue entries from the target router.
func (h *NetworkConnectHandler) ListParentQueues(ctx context.Context, req *connect.Request[polyglotv1.ListParentQueuesRequest]) (*connect.Response[polyglotv1.ListParentQueuesResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if h.hotspotUC == nil {
		return nil, response.Unavailable("hotspot usecase unavailable")
	}

	queues, err := h.hotspotUC.GetParentQueues(ctx, driver)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	entries := make([]*polyglotv1.ParentQueueEntry, 0, len(queues))
	for _, q := range queues {
		entries = append(entries, &polyglotv1.ParentQueueEntry{
			RosId:    q.RosID,
			Name:     q.Name,
			MaxLimit: q.MaxLimit,
		})
	}

	return connect.NewResponse(&polyglotv1.ListParentQueuesResponse{Queues: entries}), nil
}

// ListIPPools lists configured IP pools from the target router.
func (h *NetworkConnectHandler) ListIPPools(ctx context.Context, req *connect.Request[polyglotv1.ListIPPoolsRequest]) (*connect.Response[polyglotv1.ListIPPoolsResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if h.hotspotUC == nil {
		return nil, response.Unavailable("hotspot usecase unavailable")
	}

	pools, err := h.hotspotUC.GetIPPools(ctx, driver)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	entries := make([]*polyglotv1.IPPoolEntry, 0, len(pools))
	for _, p := range pools {
		entries = append(entries, &polyglotv1.IPPoolEntry{
			RosId:  p.RosID,
			Name:   p.Name,
			Ranges: p.Ranges,
		})
	}

	return connect.NewResponse(&polyglotv1.ListIPPoolsResponse{Pools: entries}), nil
}
