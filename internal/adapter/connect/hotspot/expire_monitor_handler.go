package hotspot

import (
	"context"
	"regexp"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/pkg/response"
)

// expireIntervalRe validates the RouterOS duration format HH:MM:SS used by
// the expire monitor scheduler (legacy default "00:01:00").
var expireIntervalRe = regexp.MustCompile(`^\d{2}:\d{2}:\d{2}$`)

// GetExpireMonitorStatus reports the install/enabled state of the Mikhmon
// expire-monitor scheduler. Status is "ok" when installed and enabled,
// "not ready" otherwise (compatible with the legacy frontend).
func (h *HotspotConnectHandler) GetExpireMonitorStatus(ctx context.Context, req *connect.Request[devicepb.GetExpireMonitorStatusRequest]) (*connect.Response[devicepb.ExpireMonitorStatusResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	status, err := h.useCase.GetExpireMonitorStatus(ctx, driver)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	resp := &devicepb.ExpireMonitorStatusResponse{
		IsInstalled:   status.IsInstalled,
		IsEnabled:     status.IsEnabled,
		Status:        "not ready",
		SchedulerName: status.SchedulerName,
	}
	if status.IsInstalled && status.IsEnabled {
		resp.Status = "ok"
	}
	return connect.NewResponse(resp), nil
}

// SetupExpireMonitor installs (or updates, idempotently) the Mikhmon expire
// monitor. interval is a RouterOS duration HH:MM:SS; empty defaults to the
// legacy "00:01:00".
func (h *HotspotConnectHandler) SetupExpireMonitor(ctx context.Context, req *connect.Request[devicepb.SetupExpireMonitorRequest]) (*connect.Response[devicepb.SetupExpireMonitorResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	interval := req.Msg.Interval
	if interval != "" && !expireIntervalRe.MatchString(interval) {
		return nil, response.InvalidArgument("interval must be HH:MM:SS (e.g. 00:01:00)")
	}

	if _, err := h.useCase.SetupExpireMonitor(ctx, driver, interval); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.SetupExpireMonitorResponse{Message: "expire monitor installed"}), nil
}

// DisableExpireMonitor disables the found expire-monitor scheduler.
func (h *HotspotConnectHandler) DisableExpireMonitor(ctx context.Context, req *connect.Request[devicepb.DisableExpireMonitorRequest]) (*connect.Response[devicepb.DisableExpireMonitorResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	if _, err := h.useCase.DisableExpireMonitor(ctx, driver, true); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.DisableExpireMonitorResponse{Message: "expire monitor disabled"}), nil
}

// RemoveExpireMonitor deletes the found expire-monitor scheduler (the helper
// script is left behind).
func (h *HotspotConnectHandler) RemoveExpireMonitor(ctx context.Context, req *connect.Request[devicepb.RemoveExpireMonitorRequest]) (*connect.Response[devicepb.RemoveExpireMonitorResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	if _, err := h.useCase.RemoveExpireMonitor(ctx, driver); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.RemoveExpireMonitorResponse{Message: "expire monitor removed"}), nil
}
