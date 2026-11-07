package hotspot

import (
	"context"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

func (h *HotspotConnectHandler) ListProfiles(ctx context.Context, req *connect.Request[polyglotv1.ListHotspotProfilesRequest]) (*connect.Response[polyglotv1.ListHotspotProfilesResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	profiles, err := h.useCase.GetProfiles(ctx, driver)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.ListHotspotProfilesResponse{
		Profiles: toProtoHotspotProfiles(profiles),
	}), nil
}

func (h *HotspotConnectHandler) ListUsers(ctx context.Context, req *connect.Request[polyglotv1.ListHotspotUsersRequest]) (*connect.Response[polyglotv1.ListHotspotUsersResponse], error) {
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

	return connect.NewResponse(&polyglotv1.ListHotspotUsersResponse{
		Users: toProtoHotspotUsers(users),
	}), nil
}
