package hotspot

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/pkg/response"
)

// ListHosts returns all /ip/hotspot/host entries.
func (h *HotspotConnectHandler) ListHosts(ctx context.Context, req *connect.Request[devicepb.ListHotspotHostsRequest]) (*connect.Response[devicepb.ListHotspotHostsResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	rows, err := h.useCase.GetHosts(ctx, driver)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListHotspotHostsResponse{
		Hosts: toProtoHotspotHosts(rows),
	}), nil
}

// RemoveHost deletes a single /ip/hotspot/host entry by RouterOS .id.
// Note: this only removes the host entry, not the hotspot user.
func (h *HotspotConnectHandler) RemoveHost(ctx context.Context, req *connect.Request[devicepb.RemoveHotspotHostRequest]) (*connect.Response[devicepb.RemoveHotspotHostResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	res, err := h.useCase.RemoveHost(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.RemoveHotspotHostResponse{
		Message: fmt.Sprintf("host removed: output=%s", res.Output),
	}), nil
}

// ListHotspotServers returns all /ip/hotspot server instances.
func (h *HotspotConnectHandler) ListHotspotServers(ctx context.Context, req *connect.Request[devicepb.ListHotspotServersRequest]) (*connect.Response[devicepb.ListHotspotServersResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	rows, err := h.useCase.GetHotspotServers(ctx, driver)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListHotspotServersResponse{
		Servers: toProtoHotspotServers(rows),
	}), nil
}
