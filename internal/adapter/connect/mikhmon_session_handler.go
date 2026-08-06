package connectadapter

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
)

func (h *MikhmonConnectHandler) ListActiveSessions(ctx context.Context, req *connect.Request[devicepb.ListMikhmonActiveSessionsRequest]) (*connect.Response[devicepb.ListMikhmonActiveSessionsResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	sessions, err := h.useCase.GetActiveSessions(ctx, driver)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbSessions := make([]*devicepb.MikhmonActiveSession, len(sessions))
	for i, s := range sessions {
		pbSessions[i] = &devicepb.MikhmonActiveSession{
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

	return connect.NewResponse(&devicepb.ListMikhmonActiveSessionsResponse{Sessions: pbSessions}), nil
}

func (h *MikhmonConnectHandler) KickActiveSession(ctx context.Context, req *connect.Request[devicepb.KickMikhmonSessionRequest]) (*connect.Response[devicepb.KickMikhmonSessionResponse], error) {
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

	return connect.NewResponse(&devicepb.KickMikhmonSessionResponse{
		Message: fmt.Sprintf("session kicked: output=%s", res.Output),
	}), nil
}

func (h *MikhmonConnectHandler) ListDHCPLeases(ctx context.Context, req *connect.Request[devicepb.ListMikhmonDHCPLeasesRequest]) (*connect.Response[devicepb.ListMikhmonDHCPLeasesResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	leases, err := h.activeSessionsUseCase.GetDHCPLeases(ctx, driver, req.Msg.MacFilter)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbLeases := make([]*devicepb.MikhmonDHCPLease, len(leases))
	for i, l := range leases {
		pbLeases[i] = &devicepb.MikhmonDHCPLease{
			Id:         l.RosID,
			Address:    l.Address,
			MacAddress: l.MACAddress,
			HostName:   l.HostName,
			Status:     l.Status,
			Blocked:    l.Blocked,
			Comment:    l.Comment,
		}
	}

	return connect.NewResponse(&devicepb.ListMikhmonDHCPLeasesResponse{Leases: pbLeases}), nil
}

func (h *MikhmonConnectHandler) BlockDHCPLease(ctx context.Context, req *connect.Request[devicepb.BlockDHCPLeaseRequest]) (*connect.Response[devicepb.BlockDHCPLeaseResponse], error) {
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
