package report

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	reportUC "github.com/quixiq/polyglot/internal/usecase/report"
	"github.com/quixiq/polyglot/pkg/response"
)

type ReportConnectHandler struct {
	useCase *reportUC.ManageReportUseCase
}

// NewReportConnectHandler constructs a new ReportConnectHandler.
func NewReportConnectHandler(uc *reportUC.ManageReportUseCase) *ReportConnectHandler {
	return &ReportConnectHandler{useCase: uc}
}

func (h *ReportConnectHandler) DailyReport(ctx context.Context, req *connect.Request[devicepb.DailyReportRequest]) (*connect.Response[devicepb.DailyReportResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("report usecase unavailable")
	}
	period, snaps, err := h.useCase.DailyReport(ctx, req.Msg.Date)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.DailyReportResponse{
		Summary: toProtoSummary(period, snaps),
	}), nil
}

func (h *ReportConnectHandler) MonthlyReport(ctx context.Context, req *connect.Request[devicepb.MonthlyReportRequest]) (*connect.Response[devicepb.MonthlyReportResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("report usecase unavailable")
	}
	period, snaps, err := h.useCase.MonthlyReport(ctx, req.Msg.Month)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.MonthlyReportResponse{
		Summary: toProtoSummary(period, snaps),
	}), nil
}

func (h *ReportConnectHandler) YearlyReport(ctx context.Context, req *connect.Request[devicepb.YearlyReportRequest]) (*connect.Response[devicepb.YearlyReportResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("report usecase unavailable")
	}
	period, snaps, err := h.useCase.YearlyReport(ctx, int(req.Msg.Year))
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.YearlyReportResponse{
		Summary: toProtoSummary(period, snaps),
	}), nil
}

func (h *ReportConnectHandler) RefreshSnapshot(ctx context.Context, req *connect.Request[devicepb.RefreshSnapshotRequest]) (*connect.Response[devicepb.RefreshSnapshotResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("report usecase unavailable")
	}
	dateStr, err := h.useCase.RefreshSnapshot(ctx, req.Msg.Date)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.RefreshSnapshotResponse{
		Message: "Snapshot refreshed",
		Date:    dateStr,
	}), nil
}
