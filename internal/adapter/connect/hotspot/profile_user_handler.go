package hotspot

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/hotspot"
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
		Profiles: ToProtoHotspotProfiles(profiles),
	}), nil
}

func (h *HotspotConnectHandler) ListUsers(ctx context.Context, req *connect.Request[devicepb.ListHotspotUsersRequest]) (*connect.Response[devicepb.ListHotspotUsersResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	users, err := h.useCase.GetUsers(ctx, driver, req.Msg.Profile)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ListHotspotUsersResponse{
		Users: ToProtoHotspotUsers(users),
	}), nil
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

	params := hotspot.VoucherGenerateParams{
		Profile:    req.Msg.Profile,
		Prefix:     req.Msg.Prefix,
		UserLength: int(req.Msg.UserLength),
		CharSet:    hotspot.CharSet(req.Msg.CharacterSet),
	}

	batch, err := h.useCase.GenerateVouchers(ctx, driver, params, count)
	if err != nil {
		return nil, response.MapDomainError(err)
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
		Message:  fmt.Sprintf("successfully generated %d vouchers", len(pbVouchers)),
	}), nil
}
