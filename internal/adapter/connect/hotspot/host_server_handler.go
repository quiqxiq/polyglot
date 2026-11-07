package hotspot

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/pkg/response"
)

// ListHosts returns all /ip/hotspot/host entries.
func (h *HotspotConnectHandler) ListHosts(ctx context.Context, req *connect.Request[polyglotv1.ListHotspotHostsRequest]) (*connect.Response[polyglotv1.ListHotspotHostsResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	rows, err := h.useCase.GetHosts(ctx, driver)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.ListHotspotHostsResponse{
		Hosts: toProtoHotspotHosts(rows),
	}), nil
}

// RemoveHost deletes a single /ip/hotspot/host entry by RouterOS .id.
// Note: this only removes the host entry, not the hotspot user.
func (h *HotspotConnectHandler) RemoveHost(ctx context.Context, req *connect.Request[polyglotv1.RemoveHotspotHostRequest]) (*connect.Response[polyglotv1.RemoveHotspotHostResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	res, err := h.useCase.RemoveHost(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.RemoveHotspotHostResponse{
		Message: fmt.Sprintf("host removed: output=%s", res.Output),
	}), nil
}

// ListHotspotServers returns all /ip/hotspot server instances.
func (h *HotspotConnectHandler) ListHotspotServers(ctx context.Context, req *connect.Request[polyglotv1.ListHotspotServersRequest]) (*connect.Response[polyglotv1.ListHotspotServersResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	rows, err := h.useCase.GetHotspotServers(ctx, driver)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.ListHotspotServersResponse{
		Servers: toProtoHotspotServers(rows),
	}), nil
}
