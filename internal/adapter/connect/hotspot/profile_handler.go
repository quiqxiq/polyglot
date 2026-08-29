package hotspot

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/hotspot"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

// CreateProfile creates a hotspot user profile with full Mikhmon metadata.
// The profile name is normalized (whitespace -> "-") exactly like legacy
// post_add_profile.php before it is embedded in report script names.
func (h *HotspotConnectHandler) CreateProfile(ctx context.Context, req *connect.Request[devicepb.CreateHotspotProfileRequest]) (*connect.Response[devicepb.CreateHotspotProfileResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	params := ProfileParamsFromProto(req.Msg.Profile)
	params.Name = hotspot.NormalizeProfileName(params.Name)

	if _, err := h.useCase.CreateProfile(ctx, driver, params); err != nil {
		return nil, response.MapDomainError(err)
	}

	created, err := h.findProfileByName(ctx, driver, params.Name)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.CreateHotspotProfileResponse{
		Profile: toProtoHotspotProfile(created),
		Message: fmt.Sprintf("profile %q created", params.Name),
	}), nil
}

// UpdateProfile modifies an existing hotspot user profile by RouterOS .id.
func (h *HotspotConnectHandler) UpdateProfile(ctx context.Context, req *connect.Request[devicepb.UpdateHotspotProfileRequest]) (*connect.Response[devicepb.UpdateHotspotProfileResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	params := ProfileParamsFromProto(req.Msg.Profile)
	if _, err := h.useCase.UpdateProfile(ctx, driver, req.Msg.RosId, params); err != nil {
		return nil, response.MapDomainError(err)
	}

	updated, err := h.findProfileByRosID(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.UpdateHotspotProfileResponse{
		Profile: toProtoHotspotProfile(updated),
		Message: fmt.Sprintf("profile %q updated", updated.Name),
	}), nil
}

// DeleteProfile removes a hotspot user profile by RouterOS .id. RouterOS will
// !trap when the profile is still referenced by users — the original message
// is propagated.
func (h *HotspotConnectHandler) DeleteProfile(ctx context.Context, req *connect.Request[devicepb.DeleteHotspotProfileRequest]) (*connect.Response[devicepb.DeleteHotspotProfileResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	res, err := h.useCase.DeleteProfile(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.DeleteHotspotProfileResponse{
		Message: fmt.Sprintf("profile deleted: output=%s", res.Output),
	}), nil
}

// findProfileByName re-prints all profiles and returns the one matching name.
func (h *HotspotConnectHandler) findProfileByName(ctx context.Context, driver port.DeviceDriver, name string) (port.HotspotUserProfile, error) {
	profiles, err := h.useCase.GetProfiles(ctx, driver)
	if err != nil {
		return port.HotspotUserProfile{}, err
	}
	for _, p := range profiles {
		if strings.EqualFold(p.Name, name) {
			return p, nil
		}
	}
	return port.HotspotUserProfile{}, fmt.Errorf("profile %q not found after create", name)
}

// findProfileByRosID re-prints all profiles and returns the one matching .id.
func (h *HotspotConnectHandler) findProfileByRosID(ctx context.Context, driver port.DeviceDriver, rosID string) (port.HotspotUserProfile, error) {
	profiles, err := h.useCase.GetProfiles(ctx, driver)
	if err != nil {
		return port.HotspotUserProfile{}, err
	}
	for _, p := range profiles {
		if p.RosID == rosID {
			return p, nil
		}
	}
	return port.HotspotUserProfile{}, fmt.Errorf("profile with .id %q not found after update", rosID)
}
