package hotspot

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

func (h *HotspotConnectHandler) ListProfiles(ctx context.Context, req *connect.Request[devicepb.ListHotspotProfilesRequest]) (*connect.Response[devicepb.ListHotspotProfilesResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	profiles, err := h.useCase.GetProfiles(ctx, driver)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ListHotspotProfilesResponse{
		Profiles: toProtoHotspotProfiles(profiles),
	}), nil
}

func (h *HotspotConnectHandler) ListUsers(ctx context.Context, req *connect.Request[devicepb.ListHotspotUsersRequest]) (*connect.Response[devicepb.ListHotspotUsersResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	users, err := h.useCase.GetUsers(ctx, driver, port.ListUsersFilter{
		Profile:    req.Msg.Profile,
		Comment:    sanitizeBatchTag(req.Msg.Comment),
		OnlyUnused: req.Msg.OnlyUnused,
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ListHotspotUsersResponse{
		Users: toProtoHotspotUsers(users),
	}), nil
}
