package connectadapter

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/mikhmon"
)

func (h *MikhmonConnectHandler) ListProfiles(ctx context.Context, req *connect.Request[devicepb.ListMikhmonProfilesRequest]) (*connect.Response[devicepb.ListMikhmonProfilesResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	profiles, err := h.useCase.GetProfiles(ctx, driver)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbProfiles := make([]*devicepb.MikhmonProfile, len(profiles))
	for i, p := range profiles {
		pbProfiles[i] = &devicepb.MikhmonProfile{
			Id:          p.RosID,
			Name:        p.Name,
			SharedUsers: p.SharedUsers,
			RateLimit:   p.RateLimit,
			ModeExpire:  p.OnLogin,
			ParentQueue: p.ParentQueue,
			Comment:     p.Comment,
		}
	}

	return connect.NewResponse(&devicepb.ListMikhmonProfilesResponse{Profiles: pbProfiles}), nil
}

func (h *MikhmonConnectHandler) ListUsers(ctx context.Context, req *connect.Request[devicepb.ListMikhmonUsersRequest]) (*connect.Response[devicepb.ListMikhmonUsersResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	users, err := h.useCase.GetUsers(ctx, driver, req.Msg.Profile)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbUsers := make([]*devicepb.MikhmonUser, len(users))
	for i, u := range users {
		pbUsers[i] = &devicepb.MikhmonUser{
			Id:          u.RosID,
			Name:        u.Name,
			Password:    u.Password,
			Profile:     u.Profile,
			LimitUptime: u.LimitUptime,
			LimitBytes:  u.LimitBytesIn,
			Comment:     u.Comment,
			Disabled:    u.Disabled,
		}
	}

	return connect.NewResponse(&devicepb.ListMikhmonUsersResponse{Users: pbUsers}), nil
}

func (h *MikhmonConnectHandler) GenerateVouchers(ctx context.Context, req *connect.Request[devicepb.GenerateVouchersRequest]) (*connect.Response[devicepb.GenerateVouchersResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	count := int(req.Msg.Count)
	if count <= 0 {
		count = 1
	}

	params := mikhmon.VoucherGenerateParams{
		Profile:    req.Msg.Profile,
		Prefix:     req.Msg.Prefix,
		UserLength: int(req.Msg.UserLength),
		CharSet:    mikhmon.CharSet(req.Msg.CharacterSet),
	}

	batch, err := h.useCase.GenerateVouchers(ctx, driver, params, count)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbVouchers := make([]*devicepb.MikhmonUser, len(batch.Vouchers))
	for i, u := range batch.Vouchers {
		pbVouchers[i] = &devicepb.MikhmonUser{
			Name:     u.Username,
			Password: u.Password,
			Profile:  req.Msg.Profile,
			Comment:  u.Comment,
		}
	}

	return connect.NewResponse(&devicepb.GenerateVouchersResponse{
		Vouchers: pbVouchers,
		Message:  "vouchers generated successfully via ConnectRPC",
	}), nil
}
