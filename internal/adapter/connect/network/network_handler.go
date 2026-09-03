package network

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/port"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
	"github.com/quixiq/polyglot/pkg/fault"
	"github.com/quixiq/polyglot/pkg/response"
)

// ConnectDriverProvider signature to obtain a port.DeviceDriver for a given deviceId.
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
func (h *NetworkConnectHandler) ListDHCPLeases(ctx context.Context, req *connect.Request[devicepb.ListDHCPLeasesRequest]) (*connect.Response[devicepb.ListDHCPLeasesResponse], error) {
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

	return connect.NewResponse(&devicepb.ListDHCPLeasesResponse{
		Leases: ToProtoDHCPLeases(leases),
	}), nil
}

// BlockDHCPLease modifies the blocked status of a DHCP lease on the target router.
func (h *NetworkConnectHandler) BlockDHCPLease(ctx context.Context, req *connect.Request[devicepb.BlockDHCPLeaseRequest]) (*connect.Response[devicepb.BlockDHCPLeaseResponse], error) {
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

	return connect.NewResponse(&devicepb.BlockDHCPLeaseResponse{
		Message: fmt.Sprintf("lease block status updated: output=%s", res.Output),
	}), nil
}

// ListParentQueues lists parent simple queue entries from the target router.
func (h *NetworkConnectHandler) ListParentQueues(ctx context.Context, req *connect.Request[devicepb.ListParentQueuesRequest]) (*connect.Response[devicepb.ListParentQueuesResponse], error) {
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

// ListIPPools lists configured IP pools from the target router.
func (h *NetworkConnectHandler) ListIPPools(ctx context.Context, req *connect.Request[devicepb.ListIPPoolsRequest]) (*connect.Response[devicepb.ListIPPoolsResponse], error) {
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
