package hotspot

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/pkg/response"
)

func (h *HotspotConnectHandler) ListActiveSessions(ctx context.Context, req *connect.Request[devicepb.ListHotspotActiveSessionsRequest]) (*connect.Response[devicepb.ListHotspotActiveSessionsResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	sessions, err := h.useCase.GetActiveSessions(ctx, driver)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ListHotspotActiveSessionsResponse{
		Sessions: toProtoActiveSessions(sessions),
	}), nil
}

func (h *HotspotConnectHandler) KickActiveSession(ctx context.Context, req *connect.Request[devicepb.KickHotspotSessionRequest]) (*connect.Response[devicepb.KickHotspotSessionResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	if req.Msg.RosId == "" {
		return nil, response.InvalidArgument("ros_id required")
	}

	res, err := h.useCase.RemoveActiveSession(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.KickHotspotSessionResponse{
		Message: fmt.Sprintf("session kicked: output=%s", res.Output),
	}), nil
}
