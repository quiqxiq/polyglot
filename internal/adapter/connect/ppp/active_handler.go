package ppp

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/pkg/response"
)

// ListActiveSessions retrieves currently active PPP sessions from the router.
func (h *PPPConnectHandler) ListActiveSessions(ctx context.Context, req *connect.Request[devicepb.ListPPPActiveSessionsRequest]) (*connect.Response[devicepb.ListPPPActiveSessionsResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	sessions, err := h.useCase.ListActive(ctx, driver, req.Msg.NameFilter)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ListPPPActiveSessionsResponse{
		Sessions: ToProtoPPPActiveSessions(sessions),
	}), nil
}

// KickActiveSession forcibly disconnects an active PPPoE session by RouterOS ID.
func (h *PPPConnectHandler) KickActiveSession(ctx context.Context, req *connect.Request[devicepb.KickPPPActiveSessionRequest]) (*connect.Response[devicepb.KickPPPActiveSessionResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if req.Msg.RosId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("ros_id is required"))
	}

	res, err := h.useCase.KickActive(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.KickPPPActiveSessionResponse{
		Message: fmt.Sprintf("session disconnected: output=%s", res.Output),
	}), nil
}

// KickActiveSessions forcibly disconnects multiple active PPPoE sessions.
func (h *PPPConnectHandler) KickActiveSessions(ctx context.Context, req *connect.Request[devicepb.KickPPPActiveSessionsRequest]) (*connect.Response[devicepb.KickPPPActiveSessionsResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if len(req.Msg.RosIds) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("ros_ids cannot be empty"))
	}

	count, err := h.useCase.KickActiveBatch(ctx, driver, req.Msg.RosIds)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.KickPPPActiveSessionsResponse{
		KickedCount: int32(count),
		Message:     fmt.Sprintf("%d sessions disconnected successfully", count),
	}), nil
}

// ListInactiveSecrets retrieves all registered secrets that do not currently have an active session.
func (h *PPPConnectHandler) ListInactiveSecrets(ctx context.Context, req *connect.Request[devicepb.ListPPPInactiveSecretsRequest]) (*connect.Response[devicepb.ListPPPInactiveSecretsResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	inactive, err := h.useCase.ListInactive(ctx, driver)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ListPPPInactiveSecretsResponse{
		Secrets: ToProtoPPPSecrets(inactive),
	}), nil
}
