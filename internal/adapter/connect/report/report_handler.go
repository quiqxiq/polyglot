package report

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainReporting "github.com/quixiq/polyglot/internal/domain/reporting"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

type ReportConnectHandler struct {
	repo        port.ReportingRepository
	snapshotter port.SnapshotComputer
}

func NewReportConnectHandler(repo port.ReportingRepository, snapshotter port.SnapshotComputer) *ReportConnectHandler {
	return &ReportConnectHandler{repo: repo, snapshotter: snapshotter}
}

func (h *ReportConnectHandler) DailyReport(ctx context.Context, req *connect.Request[devicepb.DailyReportRequest]) (*connect.Response[devicepb.DailyReportResponse], error) {
	if h.repo == nil {
		return nil, response.Unavailable("reporting repository unavailable")
	}
	day := time.Now().UTC()
	if req.Msg.Date != "" {
		d, err := time.Parse("2006-01-02", req.Msg.Date)
		if err != nil {
			return nil, response.InvalidArgument("date must be YYYY-MM-DD")
		}
		day = d
	}
	snap, err := h.repo.GetByDate(ctx, "tenant-default", day)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.DailyReportResponse{
		Summary: toProtoSummary(day.Format("2006-01-02"), []domainReporting.DailyFinancialSnapshot{snap}),
	}), nil
}

func (h *ReportConnectHandler) MonthlyReport(ctx context.Context, req *connect.Request[devicepb.MonthlyReportRequest]) (*connect.Response[devicepb.MonthlyReportResponse], error) {
	if h.repo == nil {
		return nil, response.Unavailable("reporting repository unavailable")
	}
	month := req.Msg.Month
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	from, err := time.Parse("2006-01", month)
	if err != nil {
		return nil, response.InvalidArgument("month must be YYYY-MM")
	}
	to := from.AddDate(0, 1, -1)
	snaps, err := h.repo.ListRange(ctx, "tenant-default", from, to)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.MonthlyReportResponse{
		Summary: toProtoSummary(month, snaps),
	}), nil
}

func (h *ReportConnectHandler) YearlyReport(ctx context.Context, req *connect.Request[devicepb.YearlyReportRequest]) (*connect.Response[devicepb.YearlyReportResponse], error) {
	if h.repo == nil {
		return nil, response.Unavailable("reporting repository unavailable")
	}
	year := int(req.Msg.Year)
	if year == 0 {
		year = time.Now().Year()
	}
	from := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(1, 0, -1)
	snaps, err := h.repo.ListRange(ctx, "tenant-default", from, to)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.YearlyReportResponse{
		Summary: toProtoSummary(strconv.Itoa(year), snaps),
	}), nil
}

func (h *ReportConnectHandler) RefreshSnapshot(ctx context.Context, req *connect.Request[devicepb.RefreshSnapshotRequest]) (*connect.Response[devicepb.RefreshSnapshotResponse], error) {
	if h.snapshotter == nil {
		return nil, response.Unavailable("snapshotter unavailable")
	}
	day := time.Now().UTC()
	if req.Msg.Date != "" {
		d, err := time.Parse("2006-01-02", req.Msg.Date)
		if err != nil {
			return nil, response.InvalidArgument("date must be YYYY-MM-DD")
		}
		day = d
	}
	if err := h.snapshotter.RecomputeDaily(ctx, "tenant-default", day); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.RefreshSnapshotResponse{
		Message: "Snapshot refreshed", Date: day.Format("2006-01-02"),
	}), nil
}

func toProtoSummary(period string, snaps []domainReporting.DailyFinancialSnapshot) *devicepb.ReportSummary {
	sum := &devicepb.ReportSummary{
		Period: period, CashBalances: map[string]doubleAlias{},
	}
	for _, s := range snaps {
		sum.InvoiceCount += int32(s.InvoiceCount)
		sum.InvoiceTotal += s.InvoiceTotal
		sum.PaymentCount += int32(s.PaymentCount)
		sum.PaymentTotal += s.PaymentTotal
		sum.OutstandingTotal = s.OutstandingTotal
		sum.ExpenseTotal += s.ExpenseTotal
		sum.ActiveSubscriptions = int32(s.ActiveSubscriptions)
		if len(s.CashBalanceJSON) > 0 {
			var m map[string]float64
			if err := json.Unmarshal(s.CashBalanceJSON, &m); err == nil {
				sum.CashBalances = m
			}
		}
	}
	return sum
}

type doubleAlias = float64
