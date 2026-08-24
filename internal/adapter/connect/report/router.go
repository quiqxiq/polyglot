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
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.ReportService"
	mux.Handle("/"+serviceName+"/DailyReport", connect.NewUnaryHandler("/"+serviceName+"/DailyReport", handler.DailyReport, codecOpt))
	mux.Handle("/"+serviceName+"/MonthlyReport", connect.NewUnaryHandler("/"+serviceName+"/MonthlyReport", handler.MonthlyReport, codecOpt))
	mux.Handle("/"+serviceName+"/YearlyReport", connect.NewUnaryHandler("/"+serviceName+"/YearlyReport", handler.YearlyReport, codecOpt))
	mux.Handle("/"+serviceName+"/RefreshSnapshot", connect.NewUnaryHandler("/"+serviceName+"/RefreshSnapshot", handler.RefreshSnapshot, codecOpt))

	return "/" + serviceName + "/", mux
}
