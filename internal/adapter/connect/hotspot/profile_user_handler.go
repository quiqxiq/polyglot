package hotspot

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	mikhmon "github.com/quixiq/polyglot/internal/driver/mikrotik/hotspot"
)

func (h *HotspotConnectHandler) ListProfiles(ctx context.Context, req *connect.Request[devicepb.ListHotspotProfilesRequest]) (*connect.Response[devicepb.ListHotspotProfilesResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	profiles, err := h.useCase.GetProfiles(ctx, driver)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbProfiles := make([]*devicepb.HotspotProfile, len(profiles))
	for i, p := range profiles {
		pbProfiles[i] = &devicepb.HotspotProfile{
			Id:          p.RosID,
			Name:        p.Name,
			SharedUsers: p.SharedUsers,
			RateLimit:   p.RateLimit,
			ModeExpire:  p.OnLogin,
			ParentQueue: p.ParentQueue,
			Comment:     p.Comment,
		}
	}

	return connect.NewResponse(&devicepb.ListHotspotProfilesResponse{Profiles: pbProfiles}), nil
}

func (h *HotspotConnectHandler) ListUsers(ctx context.Context, req *connect.Request[devicepb.ListHotspotUsersRequest]) (*connect.Response[devicepb.ListHotspotUsersResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	users, err := h.useCase.GetUsers(ctx, driver, req.Msg.Profile)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbUsers := make([]*devicepb.HotspotUser, len(users))
	for i, u := range users {
		pbUsers[i] = &devicepb.HotspotUser{
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

	return connect.NewResponse(&devicepb.ListHotspotUsersResponse{Users: pbUsers}), nil
}

func (h *HotspotConnectHandler) GenerateVouchers(ctx context.Context, req *connect.Request[devicepb.GenerateVouchersRequest]) (*connect.Response[devicepb.GenerateVouchersResponse], error) {
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

	pbVouchers := make([]*devicepb.HotspotUser, len(batch.Vouchers))
	for i, u := range batch.Vouchers {
		pbVouchers[i] = &devicepb.HotspotUser{
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
