package report

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/port"
)

func NewReportServiceHandler(repo port.ReportingRepository, snapshotter port.SnapshotComputer) (string, http.Handler) {
	handler := NewReportConnectHandler(repo, snapshotter)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.ReportService"
	mux.Handle("/"+serviceName+"/DailyReport", connect.NewUnaryHandler("/"+serviceName+"/DailyReport", handler.DailyReport, opts...))
	mux.Handle("/"+serviceName+"/MonthlyReport", connect.NewUnaryHandler("/"+serviceName+"/MonthlyReport", handler.MonthlyReport, opts...))
	mux.Handle("/"+serviceName+"/YearlyReport", connect.NewUnaryHandler("/"+serviceName+"/YearlyReport", handler.YearlyReport, opts...))
	mux.Handle("/"+serviceName+"/RefreshSnapshot", connect.NewUnaryHandler("/"+serviceName+"/RefreshSnapshot", handler.RefreshSnapshot, opts...))

	return "/" + serviceName + "/", mux
}
