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

// ARAging mengembalikan umur piutang tagihan pelanggan.
func (h *ReportConnectHandler) ARAging(ctx context.Context, req *connect.Request[devicepb.ARAgingRequest]) (*connect.Response[devicepb.ARAgingResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("report usecase unavailable")
	}
	report, err := h.useCase.ARAging(ctx, req.Msg.AsOf)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	buckets := make([]*devicepb.ARAgingBucket, len(report.Rows))
	for i := range report.Rows {
		buckets[i] = &devicepb.ARAgingBucket{
			Bucket:       report.Rows[i].Bucket,
			InvoiceCount: int32(report.Rows[i].InvoiceCount),
			Outstanding:  report.Rows[i].Outstanding,
		}
	}
	return connect.NewResponse(&devicepb.ARAgingResponse{
		AsOf:             report.AsOf.Format("2006-01-02"),
		Buckets:          buckets,
		TotalCount:       int32(report.TotalCount),
		TotalOutstanding: report.Total,
	}), nil
}

// GatewayReconciliation membandingkan transaksi gateway dengan buku kas.
func (h *ReportConnectHandler) GatewayReconciliation(ctx context.Context, req *connect.Request[devicepb.GatewayReconciliationRequest]) (*connect.Response[devicepb.GatewayReconciliationResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("report usecase unavailable")
	}
	report, err := h.useCase.GatewayReconciliation(ctx, req.Msg.From, req.Msg.To)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	rows := make([]*devicepb.GatewayReconciliationRow, len(report.Rows))
	for i := range report.Rows {
		row := report.Rows[i]
		rows[i] = &devicepb.GatewayReconciliationRow{
			Gateway:          row.Gateway,
			TransactionCount: int32(row.TransactionCount),
			GatewayAmount:    row.GatewayAmount,
			FeeAmount:        row.FeeAmount,
			RecordedAmount:   row.RecordedAmount,
			Difference:       row.GatewayAmount - row.RecordedAmount,
		}
	}
	return connect.NewResponse(&devicepb.GatewayReconciliationResponse{
		From:          report.From.Format("2006-01-02"),
		To:            report.To.Format("2006-01-02"),
		Rows:          rows,
		TotalGateway:  report.TotalGateway,
		TotalRecorded: report.TotalRecorded,
		TotalFee:      report.TotalFee,
	}), nil
}

// PSBStats mengembalikan statistik efisiensi & durasi PSB.
func (h *ReportConnectHandler) PSBStats(ctx context.Context, req *connect.Request[devicepb.PSBStatsRequest]) (*connect.Response[devicepb.PSBStatsResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("report usecase unavailable")
	}
	report, err := h.useCase.PSBStats(ctx, req.Msg.From, req.Msg.To)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	statuses := make([]*devicepb.PSBStatusCount, len(report.ByStatus))
	for i := range report.ByStatus {
		statuses[i] = &devicepb.PSBStatusCount{
			Status: report.ByStatus[i].Status,
			Count:  int32(report.ByStatus[i].Count),
		}
	}
	return connect.NewResponse(&devicepb.PSBStatsResponse{
		From:            report.From.Format("2006-01-02"),
		To:              report.To.Format("2006-01-02"),
		Total:           int32(report.Total),
		Converted:       int32(report.Converted),
		Closed:          int32(report.Closed),
		ByStatus:        statuses,
		AvgSurveyHours:  report.AvgSurveyHours,
		AvgInstallHours: report.AvgInstallHours,
		AvgLeadHours:    report.AvgLeadHours,
	}), nil
}

// TaxReport returns PPN tax recap for a billing period.
func (h *ReportConnectHandler) TaxReport(ctx context.Context, req *connect.Request[devicepb.TaxReportRequest]) (*connect.Response[devicepb.TaxReportResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("report usecase unavailable")
	}
	report, err := h.useCase.TaxReport(ctx, req.Msg.GetPeriod())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	rows := make([]*devicepb.TaxRow, len(report.Rows))
	for i, r := range report.Rows {
		rows[i] = &devicepb.TaxRow{
			Period:       r.Period,
			InvoiceCount: int32(r.InvoiceCount),
			Subtotal:     r.Subtotal,
			TaxPercent:   r.TaxPercent,
			TaxAmount:    r.TaxAmount,
			Total:        r.Total,
			PaidTax:      r.PaidTax,
		}
	}
	return connect.NewResponse(&devicepb.TaxReportResponse{
		Period:        report.Period,
		Rows:          rows,
		TotalSubtotal: report.TotalSubtotal,
		TotalTax:      report.TotalTax,
		TotalBilled:   report.TotalBilled,
		TotalPaidTax:  report.TotalPaidTax,
	}), nil
}

// SalesReport returns sales and revenue recap by service plan.
func (h *ReportConnectHandler) SalesReport(ctx context.Context, req *connect.Request[devicepb.SalesReportRequest]) (*connect.Response[devicepb.SalesReportResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("report usecase unavailable")
	}
	report, err := h.useCase.SalesReport(ctx, req.Msg.GetPeriod())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	rows := make([]*devicepb.SalesRow, len(report.Rows))
	for i, r := range report.Rows {
		rows[i] = &devicepb.SalesRow{
			PlanName:     r.PlanName,
			ServiceType:  r.ServiceType,
			InvoiceCount: int32(r.InvoiceCount),
			TotalBilled:  r.TotalBilled,
			TotalPaid:    r.TotalPaid,
		}
	}
	return connect.NewResponse(&devicepb.SalesReportResponse{
		Period:        report.Period,
		Rows:          rows,
		TotalBilled:   report.TotalBilled,
		TotalPaid:     report.TotalPaid,
		PaidCount:     int32(report.PaidCount),
		CustomerCount: int32(report.CustomerCount),
	}), nil
}

// CollectionReport returns collection performance and overdue rates.
func (h *ReportConnectHandler) CollectionReport(ctx context.Context, req *connect.Request[devicepb.CollectionReportRequest]) (*connect.Response[devicepb.CollectionReportResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("report usecase unavailable")
	}
	report, err := h.useCase.CollectionReport(ctx, req.Msg.GetPeriod())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.CollectionReportResponse{
		Period:           report.Period,
		TotalInvoices:    int32(report.TotalInvoices),
		PaidInvoices:     int32(report.PaidInvoices),
		OverdueInvoices:  int32(report.OverdueInvoices),
		TotalBilled:      report.TotalBilled,
		TotalCollected:   report.TotalCollected,
		TotalOutstanding: report.TotalOutstanding,
		CollectionRate:   report.CollectionRate,
	}), nil
}

// CashflowReport returns cash inflows and outflows by category.
func (h *ReportConnectHandler) CashflowReport(ctx context.Context, req *connect.Request[devicepb.CashflowReportRequest]) (*connect.Response[devicepb.CashflowReportResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("report usecase unavailable")
	}
	report, err := h.useCase.CashflowReport(ctx, req.Msg.GetFrom(), req.Msg.GetTo())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	rows := make([]*devicepb.CashflowRow, len(report.Rows))
	for i, r := range report.Rows {
		rows[i] = &devicepb.CashflowRow{
			CategoryName: r.CategoryName,
			Direction:    r.Direction,
			TrxCount:     int32(r.TrxCount),
			TotalAmount:  r.TotalAmount,
		}
	}
	return connect.NewResponse(&devicepb.CashflowReportResponse{
		From:        report.From.Format("2006-01-02"),
		To:          report.To.Format("2006-01-02"),
		Rows:        rows,
		TotalIn:     report.TotalIn,
		TotalOut:    report.TotalOut,
		NetCashflow: report.NetCashflow,
	}), nil
}

// AgingDetail returns detailed overdue balances per customer.
func (h *ReportConnectHandler) AgingDetail(ctx context.Context, req *connect.Request[devicepb.AgingDetailRequest]) (*connect.Response[devicepb.AgingDetailResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("report usecase unavailable")
	}
	report, err := h.useCase.AgingDetail(ctx, req.Msg.GetAsOf(), int(req.Msg.GetLimit()))
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	rows := make([]*devicepb.AgingDetailRow, len(report.Rows))
	for i, r := range report.Rows {
		rows[i] = &devicepb.AgingDetailRow{
			CustomerId:    r.CustomerID,
			CustomerName:  r.CustomerName,
			Phone:         r.Phone,
			InvoiceCount:  int32(r.InvoiceCount),
			Outstanding:   r.Outstanding,
			OldestDueUnix: r.OldestDue.Unix(),
			DaysOverdue:   int32(r.DaysOverdue),
		}
	}
	return connect.NewResponse(&devicepb.AgingDetailResponse{
		AsOf:             report.AsOf.Format("2006-01-02"),
		Rows:             rows,
		TotalCustomers:   int32(report.TotalCustomers),
		TotalOutstanding: report.TotalOutstanding,
	}), nil
}

// ClosingReport returns closing date summary and operational metrics.
func (h *ReportConnectHandler) ClosingReport(ctx context.Context, req *connect.Request[devicepb.ClosingReportRequest]) (*connect.Response[devicepb.ClosingReportResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("report usecase unavailable")
	}
	report, err := h.useCase.ClosingReport(ctx, req.Msg.GetPeriod())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ClosingReportResponse{
		Period:           report.Period,
		ClosingDateUnix:  report.ClosingDate.Unix(),
		TotalInvoices:    int32(report.TotalInvoices),
		TotalBilled:      report.TotalBilled,
		TotalCollected:   report.TotalCollected,
		TotalOutstanding: report.TotalOutstanding,
		ActiveCustomers:  int32(report.ActiveCustomers),
		NewSubscriptions: int32(report.NewSubscriptions),
		DebtsOutstanding: report.DebtsOutstanding,
	}), nil
}

// ExportReport exports a report in CSV or XLSX format.
func (h *ReportConnectHandler) ExportReport(ctx context.Context, req *connect.Request[devicepb.ExportReportRequest]) (*connect.Response[devicepb.ExportReportResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("report usecase unavailable")
	}
	payload, contentType, filename, err := h.useCase.ExportReport(ctx, req.Msg.GetReportType(), req.Msg.GetFormat(), req.Msg.GetParams())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ExportReportResponse{
		Payload:     payload,
		ContentType: contentType,
		Filename:    filename,
	}), nil
}
