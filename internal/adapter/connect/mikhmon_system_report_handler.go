package connectadapter

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
)

func (h *MikhmonConnectHandler) GetDashboard(ctx context.Context, req *connect.Request[devicepb.GetMikhmonDashboardRequest]) (*connect.Response[devicepb.GetMikhmonDashboardResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	summary, err := h.useCase.GetDashboardSummary(ctx, driver)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&devicepb.GetMikhmonDashboardResponse{
		Summary: &devicepb.MikhmonDashboardSummary{
			TotalHotspotUsers: int32(summary.TotalUsers),
			TotalActiveUsers:  int32(summary.ActiveUsers),
			TodayIncome:       float64(summary.TodayIncome),
			Uptime:            summary.Uptime,
		},
	}), nil
}
