package report

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	reportUC "github.com/quixiq/polyglot/internal/usecase/report"
)

// NewReportServiceHandler creates the Connect HTTP handler for ReportService.
func NewReportServiceHandler(uc *reportUC.ManageReportUseCase) (string, http.Handler) {
	handler := NewReportConnectHandler(uc)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.ReportService"
	mux.Handle("/"+serviceName+"/DailyReport", connect.NewUnaryHandler("/"+serviceName+"/DailyReport", handler.DailyReport, opts...))
	mux.Handle("/"+serviceName+"/MonthlyReport", connect.NewUnaryHandler("/"+serviceName+"/MonthlyReport", handler.MonthlyReport, opts...))
	mux.Handle("/"+serviceName+"/YearlyReport", connect.NewUnaryHandler("/"+serviceName+"/YearlyReport", handler.YearlyReport, opts...))
	mux.Handle("/"+serviceName+"/RefreshSnapshot", connect.NewUnaryHandler("/"+serviceName+"/RefreshSnapshot", handler.RefreshSnapshot, opts...))
	mux.Handle("/"+serviceName+"/ARAging", connect.NewUnaryHandler("/"+serviceName+"/ARAging", handler.ARAging, opts...))
	mux.Handle("/"+serviceName+"/GatewayReconciliation", connect.NewUnaryHandler("/"+serviceName+"/GatewayReconciliation", handler.GatewayReconciliation, opts...))
	mux.Handle("/"+serviceName+"/PSBStats", connect.NewUnaryHandler("/"+serviceName+"/PSBStats", handler.PSBStats, opts...))
	mux.Handle("/"+serviceName+"/TaxReport", connect.NewUnaryHandler("/"+serviceName+"/TaxReport", handler.TaxReport, opts...))
	mux.Handle("/"+serviceName+"/SalesReport", connect.NewUnaryHandler("/"+serviceName+"/SalesReport", handler.SalesReport, opts...))
	mux.Handle("/"+serviceName+"/CollectionReport", connect.NewUnaryHandler("/"+serviceName+"/CollectionReport", handler.CollectionReport, opts...))
	mux.Handle("/"+serviceName+"/CashflowReport", connect.NewUnaryHandler("/"+serviceName+"/CashflowReport", handler.CashflowReport, opts...))
	mux.Handle("/"+serviceName+"/AgingDetail", connect.NewUnaryHandler("/"+serviceName+"/AgingDetail", handler.AgingDetail, opts...))
	mux.Handle("/"+serviceName+"/ClosingReport", connect.NewUnaryHandler("/"+serviceName+"/ClosingReport", handler.ClosingReport, opts...))
	mux.Handle("/"+serviceName+"/ExportReport", connect.NewUnaryHandler("/"+serviceName+"/ExportReport", handler.ExportReport, opts...))

	return "/" + serviceName + "/", mux
}
