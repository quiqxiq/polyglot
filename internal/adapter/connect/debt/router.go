package debt

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	debtUC "github.com/quixiq/polyglot/internal/usecase/debt"
)

// NewDebtServiceHandler mounts DebtService Connect handlers.
func NewDebtServiceHandler(svc *debtUC.Service) (string, http.Handler) {
	handler := NewDebtConnectHandler(svc)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.DebtService"

	mux.Handle("/"+serviceName+"/ListDebts", connect.NewUnaryHandler("/"+serviceName+"/ListDebts", handler.ListDebts, opts...))
	mux.Handle("/"+serviceName+"/GetDebt", connect.NewUnaryHandler("/"+serviceName+"/GetDebt", handler.GetDebt, opts...))
	mux.Handle("/"+serviceName+"/CreateDebt", connect.NewUnaryHandler("/"+serviceName+"/CreateDebt", handler.CreateDebt, opts...))
	mux.Handle("/"+serviceName+"/UpdateDebt", connect.NewUnaryHandler("/"+serviceName+"/UpdateDebt", handler.UpdateDebt, opts...))
	mux.Handle("/"+serviceName+"/CancelDebt", connect.NewUnaryHandler("/"+serviceName+"/CancelDebt", handler.CancelDebt, opts...))
	mux.Handle("/"+serviceName+"/PayDebt", connect.NewUnaryHandler("/"+serviceName+"/PayDebt", handler.PayDebt, opts...))
	mux.Handle("/"+serviceName+"/ListDebtPayments", connect.NewUnaryHandler("/"+serviceName+"/ListDebtPayments", handler.ListDebtPayments, opts...))
	mux.Handle("/"+serviceName+"/GetDebtSummary", connect.NewUnaryHandler("/"+serviceName+"/GetDebtSummary", handler.GetDebtSummary, opts...))
	mux.Handle("/"+serviceName+"/ExportDebts", connect.NewUnaryHandler("/"+serviceName+"/ExportDebts", handler.ExportDebts, opts...))

	return "/" + serviceName + "/", mux
}
