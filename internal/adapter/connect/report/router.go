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

	return "/" + serviceName + "/", mux
}
