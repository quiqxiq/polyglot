package hotspot

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/pkg/response"
)

// ListReports returns Mikhmon transaction records from /system/script, with
// optional legacy filters (day → source, month → owner, year → date suffix).
// summary_only omits the record list, returning just total_income + total
// (efficient for dashboard polling). With no filters all mikhmon records are
// returned (decision #3).
func (h *HotspotConnectHandler) ListReports(ctx context.Context, req *connect.Request[devicepb.ListHotspotReportsRequest]) (*connect.Response[devicepb.ListHotspotReportsResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	records, err := h.useCase.GetReportsByFilter(ctx, driver, req.Msg.Day, req.Msg.Month, req.Msg.Year)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	resp := &devicepb.ListHotspotReportsResponse{
		TotalIncome: SumReportIncome(records),
		Total:       int32(len(records)),
	}
	if !req.Msg.SummaryOnly {
		resp.Reports = toProtoHotspotReports(records)
	}
	return connect.NewResponse(resp), nil
}

// DeleteReport removes a report script record by RouterOS .id. This is a new
// operation — legacy Mikhmon had no delete for report records.
func (h *HotspotConnectHandler) DeleteReport(ctx context.Context, req *connect.Request[devicepb.DeleteHotspotReportRequest]) (*connect.Response[devicepb.DeleteHotspotReportResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if req.Msg.RosId == "" {
		return nil, response.InvalidArgument("ros_id required")
	}

	if _, err := h.useCase.DeleteReport(ctx, driver, req.Msg.RosId); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.DeleteHotspotReportResponse{Message: "report deleted"}), nil
}
