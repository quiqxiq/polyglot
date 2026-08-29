package ppp

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/pkg/response"
)

// ListProfiles retrieves all PPP profiles on the router.
func (h *PPPConnectHandler) ListProfiles(ctx context.Context, req *connect.Request[devicepb.ListPPPProfilesRequest]) (*connect.Response[devicepb.ListPPPProfilesResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	profiles, err := h.useCase.ListProfiles(ctx, driver, req.Msg.NameFilter)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ListPPPProfilesResponse{
		Profiles: toProtoPPPProfiles(profiles),
	}), nil
}

// GetProfile fetches a single PPP profile by its RouterOS ID.
func (h *PPPConnectHandler) GetProfile(ctx context.Context, req *connect.Request[devicepb.GetPPPProfileRequest]) (*connect.Response[devicepb.GetPPPProfileResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	profile, err := h.useCase.GetProfile(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.GetPPPProfileResponse{
		Profile: toProtoPPPProfile(profile),
	}), nil
}

// CreateProfile adds a new PPP profile to the router.
func (h *PPPConnectHandler) CreateProfile(ctx context.Context, req *connect.Request[devicepb.CreatePPPProfileRequest]) (*connect.Response[devicepb.CreatePPPProfileResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	params := FromProtoCreateProfileRequest(req.Msg)
	if _, err := h.useCase.AddProfile(ctx, driver, params); err != nil {
		return nil, response.MapDomainError(err)
	}

	profiles, err := h.useCase.ListProfiles(ctx, driver, req.Msg.Name)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	if len(profiles) == 0 {
		return nil, response.Internal("profile created but could not be loaded")
	}

	return connect.NewResponse(&devicepb.CreatePPPProfileResponse{
		Profile: toProtoPPPProfile(profiles[0]),
		Message: fmt.Sprintf("profile %q created successfully", req.Msg.Name),
	}), nil
}

// UpdateProfile modifies an existing PPP profile on the router.
func (h *PPPConnectHandler) UpdateProfile(ctx context.Context, req *connect.Request[devicepb.UpdatePPPProfileRequest]) (*connect.Response[devicepb.UpdatePPPProfileResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	params := FromProtoUpdateProfileRequest(req.Msg)
	if _, err := h.useCase.UpdateProfile(ctx, driver, req.Msg.RosId, params); err != nil {
		return nil, response.MapDomainError(err)
	}

	profile, err := h.useCase.GetProfile(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.UpdatePPPProfileResponse{
		Profile: toProtoPPPProfile(profile),
		Message: fmt.Sprintf("profile %q updated successfully", profile.Name),
	}), nil
}

// DeleteProfile removes a PPP profile from the router.
func (h *PPPConnectHandler) DeleteProfile(ctx context.Context, req *connect.Request[devicepb.DeletePPPProfileRequest]) (*connect.Response[devicepb.DeletePPPProfileResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	res, err := h.useCase.RemoveProfile(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.DeletePPPProfileResponse{
		Message: fmt.Sprintf("profile deleted: output=%s", res.Output),
	}), nil
}
