package hotspot

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/pkg/response"
)

// ListHotspotCookies returns all /ip/hotspot/cookie entries.
func (h *HotspotConnectHandler) ListHotspotCookies(ctx context.Context, req *connect.Request[devicepb.ListHotspotCookiesRequest]) (*connect.Response[devicepb.ListHotspotCookiesResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	cookies, err := h.useCase.GetCookies(ctx, driver)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	protoCookies := make([]*devicepb.HotspotCookie, 0, len(cookies))
	for _, c := range cookies {
		protoCookies = append(protoCookies, &devicepb.HotspotCookie{
			Id:         c.RosID,
			User:       c.User,
			MacAddress: c.MACAddress,
			ExpiresIn:  c.ExpiresIn,
			Domain:     c.Domain,
		})
	}

	return connect.NewResponse(&devicepb.ListHotspotCookiesResponse{
		Cookies: protoCookies,
	}), nil
}

// DeleteHotspotCookie removes a cookie by RouterOS .id or all cookies if ros_id is empty or "all".
func (h *HotspotConnectHandler) DeleteHotspotCookie(ctx context.Context, req *connect.Request[devicepb.DeleteHotspotCookieRequest]) (*connect.Response[devicepb.DeleteHotspotCookieResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	if _, err := h.useCase.DeleteCookie(ctx, driver, req.Msg.RosId); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.DeleteHotspotCookieResponse{
		Message: "Hotspot cookie(s) deleted successfully",
	}), nil
}
