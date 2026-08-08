package hotspot

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
)

func (h *HotspotConnectHandler) ListActiveSessions(ctx context.Context, req *connect.Request[devicepb.ListHotspotActiveSessionsRequest]) (*connect.Response[devicepb.ListHotspotActiveSessionsResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	sessions, err := h.useCase.GetActiveSessions(ctx, driver)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbSessions := make([]*devicepb.HotspotActiveSession, len(sessions))
	for i, s := range sessions {
		pbSessions[i] = &devicepb.HotspotActiveSession{
			Id:         s.RosID,
			Server:     s.Server,
			User:       s.User,
			Address:    s.Address,
			MacAddress: s.MACAddress,
			Uptime:     s.Uptime,
			BytesIn:    s.BytesIn,
			BytesOut:   s.BytesOut,
		}
	}

	return connect.NewResponse(&devicepb.ListHotspotActiveSessionsResponse{Sessions: pbSessions}), nil
}

func (h *HotspotConnectHandler) KickActiveSession(ctx context.Context, req *connect.Request[devicepb.KickHotspotSessionRequest]) (*connect.Response[devicepb.KickHotspotSessionResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	if req.Msg.RosId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("ros_id required"))
	}

	res, err := h.useCase.RemoveActiveSession(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&devicepb.KickHotspotSessionResponse{
		Message: fmt.Sprintf("session kicked: output=%s", res.Output),
	}), nil
}

func (h *HotspotConnectHandler) ListDHCPLeases(ctx context.Context, req *connect.Request[devicepb.ListDHCPLeasesRequest]) (*connect.Response[devicepb.ListDHCPLeasesResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	leases, err := h.activeSessionsUseCase.GetDHCPLeases(ctx, driver, req.Msg.MacFilter)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbLeases := make([]*devicepb.DHCPLease, len(leases))
	for i, l := range leases {
		pbLeases[i] = &devicepb.DHCPLease{
			Id:         l.RosID,
			Address:    l.Address,
			MacAddress: l.MACAddress,
			HostName:   l.HostName,
			Status:     l.Status,
			Blocked:    l.Blocked,
			Comment:    l.Comment,
		}
	}

	return connect.NewResponse(&devicepb.ListDHCPLeasesResponse{Leases: pbLeases}), nil
}

func (h *HotspotConnectHandler) BlockDHCPLease(ctx context.Context, req *connect.Request[devicepb.BlockDHCPLeaseRequest]) (*connect.Response[devicepb.BlockDHCPLeaseResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	res, err := h.activeSessionsUseCase.SetDHCPLeaseBlock(ctx, driver, req.Msg.RosId, req.Msg.Blocked, req.Msg.Comment)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&devicepb.BlockDHCPLeaseResponse{
		Message: fmt.Sprintf("lease block status updated: output=%s", res.Output),
	}), nil
}
