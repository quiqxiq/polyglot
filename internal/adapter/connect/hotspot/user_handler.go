package hotspot

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/hotspot"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

// GetUser returns a single hotspot user by RouterOS .id.
func (h *HotspotConnectHandler) GetUser(ctx context.Context, req *connect.Request[devicepb.GetHotspotUserRequest]) (*connect.Response[devicepb.GetHotspotUserResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if req.Msg.RosId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("ros_id required"))
	}

	user, err := h.useCase.GetUser(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetHotspotUserResponse{User: ToProtoHotspotUser(user)}), nil
}

// CreateUser adds a single hotspot user, auto-prefixing the comment with
// vc-/up- (legacy post_add_user.php) when comment is empty.
func (h *HotspotConnectHandler) CreateUser(ctx context.Context, req *connect.Request[devicepb.CreateHotspotUserRequest]) (*connect.Response[devicepb.CreateHotspotUserResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if req.Msg.Name == "" || req.Msg.Profile == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name and profile are required"))
	}

	comment := hotspot.BuildCreateUserComment(req.Msg.Name, req.Msg.Password, req.Msg.Comment, time.Now())
	params := HotspotUserParamsFromProto(req.Msg, comment)

	if _, err := h.useCase.AddUser(ctx, driver, params); err != nil {
		return nil, response.MapDomainError(err)
	}

	// Re-print to return the created user (with RouterOS .id) in the response.
	users, err := h.useCase.GetUsers(ctx, driver, port.ListUsersFilter{Name: req.Msg.Name})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	if len(users) == 0 {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("user %q created but not found", req.Msg.Name))
	}

	return connect.NewResponse(&devicepb.CreateHotspotUserResponse{
		User:    ToProtoHotspotUser(users[0]),
		Message: fmt.Sprintf("user %q created", req.Msg.Name),
	}), nil
}

// UpdateUser modifies an existing hotspot user. When reset_counter is set the
// byte/time counters are reset first (legacy reset=yes), then the user is set,
// then re-printed for the response.
func (h *HotspotConnectHandler) UpdateUser(ctx context.Context, req *connect.Request[devicepb.UpdateHotspotUserRequest]) (*connect.Response[devicepb.UpdateHotspotUserResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if req.Msg.RosId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("ros_id required"))
	}

	if req.Msg.ResetCounter {
		if _, err := h.useCase.ResetUserCounters(ctx, driver, req.Msg.RosId); err != nil {
			return nil, response.MapDomainError(err)
		}
	}

	comment := hotspot.BuildUpdatedComment(req.Msg.ExpireDate, req.Msg.UserCode, req.Msg.Comment, time.Now())
	params := HotspotUserParamsUpdateFromProto(req.Msg, comment)

	if _, err := h.useCase.UpdateUser(ctx, driver, req.Msg.RosId, params); err != nil {
		return nil, response.MapDomainError(err)
	}

	user, err := h.useCase.GetUser(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.UpdateHotspotUserResponse{
		User:    ToProtoHotspotUser(user),
		Message: fmt.Sprintf("user %q updated", user.Name),
	}), nil
}

// ResetUserCounters resets byte/time counters for a hotspot user.
func (h *HotspotConnectHandler) ResetUserCounters(ctx context.Context, req *connect.Request[devicepb.ResetHotspotUserCountersRequest]) (*connect.Response[devicepb.ResetHotspotUserCountersResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if req.Msg.RosId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("ros_id required"))
	}

	res, err := h.useCase.ResetUserCounters(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ResetHotspotUserCountersResponse{
		Message: fmt.Sprintf("counters reset: output=%s", res.Output),
	}), nil
}

// DeleteUser removes a hotspot user by RouterOS .id.
func (h *HotspotConnectHandler) DeleteUser(ctx context.Context, req *connect.Request[devicepb.DeleteHotspotUserRequest]) (*connect.Response[devicepb.DeleteHotspotUserResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if req.Msg.RosId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("ros_id required"))
	}

	res, err := h.useCase.RemoveUser(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.DeleteHotspotUserResponse{
		Message: fmt.Sprintf("user deleted: output=%s", res.Output),
	}), nil
}

// DeleteHotspotUsers deletes users matching filter mode ("profile", "comment", "expired").
func (h *HotspotConnectHandler) DeleteHotspotUsers(ctx context.Context, req *connect.Request[devicepb.DeleteHotspotUsersRequest]) (*connect.Response[devicepb.DeleteHotspotUsersResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if req.Msg.Mode == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("mode is required (profile, comment, or expired)"))
	}

	count, err := h.useCase.DeleteUsersByFilter(ctx, driver, req.Msg.Mode, req.Msg.Value)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.DeleteHotspotUsersResponse{
		DeletedCount: int32(count),
		Message:      fmt.Sprintf("%d users deleted successfully", count),
	}), nil
}

// CheckVoucherStatus inspects a voucher username and aggregates all relevant status.
func (h *HotspotConnectHandler) CheckVoucherStatus(ctx context.Context, req *connect.Request[devicepb.CheckVoucherStatusRequest]) (*connect.Response[devicepb.CheckVoucherStatusResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if req.Msg.Username == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("username is required"))
	}

	details, err := h.useCase.CheckVoucherStatus(ctx, driver, req.Msg.Username)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	resp := &devicepb.CheckVoucherStatusResponse{
		Found:      details.Found,
		IsOnline:   details.IsOnline,
		HasCookie:  details.HasCookie,
		Status:     details.Status,
		SisaWaktu:  details.SisaWaktu,
		SisaKuota:  details.SisaKuota,
		ExpireDate: details.ExpireDate,
		MacLocked:  details.MACLocked,
		Message:    details.Message,
	}

	if details.User != nil {
		resp.User = ToProtoHotspotUser(*details.User)
	}
	if details.Profile != nil {
		resp.Profile = ToProtoHotspotProfile(*details.Profile)
	}
	if details.ActiveSession != nil {
		resp.ActiveSession = ToProtoHotspotActiveSession(*details.ActiveSession)
	}
	if details.Cookie != nil {
		resp.Cookie = &devicepb.HotspotCookie{
			Id:         details.Cookie.RosID,
			User:       details.Cookie.User,
			MacAddress: details.Cookie.MACAddress,
			ExpiresIn:  details.Cookie.ExpiresIn,
			Domain:     details.Cookie.Domain,
		}
	}

	return connect.NewResponse(resp), nil
}
