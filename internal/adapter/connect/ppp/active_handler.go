package ppp

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/pkg/response"
)

// ListActiveSessions retrieves currently active PPP sessions from the router.
func (h *PPPConnectHandler) ListActiveSessions(ctx context.Context, req *connect.Request[polyglotv1.ListPPPActiveSessionsRequest]) (*connect.Response[polyglotv1.ListPPPActiveSessionsResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	sessions, err := h.useCase.ListActive(ctx, driver, req.Msg.NameFilter)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.ListPPPActiveSessionsResponse{
		Sessions: toProtoPPPActiveSessions(sessions),
	}), nil
}

// KickActiveSession forcibly disconnects an active PPPoE session by RouterOS ID.
func (h *PPPConnectHandler) KickActiveSession(ctx context.Context, req *connect.Request[polyglotv1.KickPPPActiveSessionRequest]) (*connect.Response[polyglotv1.KickPPPActiveSessionResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	res, err := h.useCase.KickActive(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.KickPPPActiveSessionResponse{
		Message: fmt.Sprintf("session disconnected: output=%s", res.Output),
	}), nil
}

// KickActiveSessions forcibly disconnects multiple active PPPoE sessions.
func (h *PPPConnectHandler) KickActiveSessions(ctx context.Context, req *connect.Request[polyglotv1.KickPPPActiveSessionsRequest]) (*connect.Response[polyglotv1.KickPPPActiveSessionsResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	count, err := h.useCase.KickActiveBatch(ctx, driver, req.Msg.RosIds)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.KickPPPActiveSessionsResponse{
		KickedCount: int32(count),
		Message:     fmt.Sprintf("%d sessions disconnected successfully", count),
	}), nil
}

// ListInactiveSecrets retrieves all registered secrets that do not currently have an active session.
func (h *PPPConnectHandler) ListInactiveSecrets(ctx context.Context, req *connect.Request[polyglotv1.ListPPPInactiveSecretsRequest]) (*connect.Response[polyglotv1.ListPPPInactiveSecretsResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	inactive, err := h.useCase.ListInactive(ctx, driver)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.ListPPPInactiveSecretsResponse{
		Secrets: toProtoPPPSecrets(inactive),
	}), nil
}
