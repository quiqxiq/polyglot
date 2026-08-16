package hotspot

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/pkg/response"
)

func (h *HotspotConnectHandler) GetDashboard(ctx context.Context, req *connect.Request[devicepb.GetHotspotDashboardRequest]) (*connect.Response[devicepb.GetHotspotDashboardResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	summary, err := h.useCase.GetDashboardSummary(ctx, driver)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.GetHotspotDashboardResponse{
		Summary: &devicepb.HotspotDashboardSummary{
			TotalHotspotUsers: int32(summary.TotalUsers),
			TotalActiveUsers:  int32(summary.ActiveUsers),
			TodayIncome:       float64(summary.TodayIncome),
			Uptime:            summary.Uptime,
		},
	}), nil
}
