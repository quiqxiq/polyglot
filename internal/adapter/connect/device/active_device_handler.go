package device

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	iauth "github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/pkg/response"
)

// SetActiveDevice binds the operator's active router server-side so AI tools
// without an explicit device_id target it.
func (h *DeviceConnectHandler) SetActiveDevice(ctx context.Context, req *connect.Request[devicepb.SetActiveDeviceRequest]) (*connect.Response[devicepb.SetActiveDeviceResponse], error) {
	if h.activeDevice == nil {
		return nil, response.Unavailable("active device store unavailable")
	}
	userID, roles, ok := iauth.IdentityFromContext(ctx)
	if !ok || userID == 0 {
		return nil, response.Unauthenticated("user session required")
	}
	a, err := h.activeDevice.Set(ctx, "tenant-default", itoa(userID), roles, req.Msg.GetSessionId(), req.Msg.GetDeviceId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.SetActiveDeviceResponse{DeviceId: a.DeviceID, SessionId: a.SessionID}), nil
}

// GetActiveDevice returns the stored binding (empty when none).
func (h *DeviceConnectHandler) GetActiveDevice(ctx context.Context, req *connect.Request[devicepb.GetActiveDeviceRequest]) (*connect.Response[devicepb.GetActiveDeviceResponse], error) {
	if h.activeDevice == nil {
		return nil, response.Unavailable("active device store unavailable")
	}
	userID, _, ok := iauth.IdentityFromContext(ctx)
	if !ok || userID == 0 {
		return nil, response.Unauthenticated("user session required")
	}
	a, err := h.activeDevice.Get(ctx, "tenant-default", itoa(userID), req.Msg.GetSessionId())
	if err != nil {
		// No binding is not an error for UI purposes.
		return connect.NewResponse(&devicepb.GetActiveDeviceResponse{}), nil
	}
	resp := &devicepb.GetActiveDeviceResponse{DeviceId: a.DeviceID, SessionId: a.SessionID}
	if !a.UpdatedAt.IsZero() {
		resp.UpdatedAtUnix = a.UpdatedAt.Unix()
	}
	return connect.NewResponse(resp), nil
}

// ClearActiveDevice removes the binding (e.g. device deleted/disabled).
func (h *DeviceConnectHandler) ClearActiveDevice(ctx context.Context, req *connect.Request[devicepb.ClearActiveDeviceRequest]) (*connect.Response[devicepb.ClearActiveDeviceResponse], error) {
	if h.activeDevice == nil {
		return nil, response.Unavailable("active device store unavailable")
	}
	userID, _, ok := iauth.IdentityFromContext(ctx)
	if !ok || userID == 0 {
		return nil, response.Unauthenticated("user session required")
	}
	if err := h.activeDevice.Clear(ctx, "tenant-default", itoa(userID), req.Msg.GetSessionId()); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ClearActiveDeviceResponse{Cleared: true}), nil
}

func itoa(v uint) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
