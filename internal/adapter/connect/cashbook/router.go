package cashbook

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	cashbookUC "github.com/quixiq/polyglot/internal/usecase/cashbook"
)

// NewCashbookServiceHandler creates the cashbook ConnectRPC service handler.
func NewCashbookServiceHandler(useCase *cashbookUC.ManageCashbookUseCase) (string, http.Handler) {
	handler := NewCashbookConnectHandler(useCase)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.CashbookService"
	mux.Handle("/"+serviceName+"/ListAccounts", connect.NewUnaryHandler("/"+serviceName+"/ListAccounts", handler.ListAccounts, opts...))
	mux.Handle("/"+serviceName+"/SaveAccount", connect.NewUnaryHandler("/"+serviceName+"/SaveAccount", handler.SaveAccount, opts...))
	mux.Handle("/"+serviceName+"/ListCategories", connect.NewUnaryHandler("/"+serviceName+"/ListCategories", handler.ListCategories, opts...))
	mux.Handle("/"+serviceName+"/SaveCategory", connect.NewUnaryHandler("/"+serviceName+"/SaveCategory", handler.SaveCategory, opts...))
	mux.Handle("/"+serviceName+"/AddTransaction", connect.NewUnaryHandler("/"+serviceName+"/AddTransaction", handler.AddTransaction, opts...))
	mux.Handle("/"+serviceName+"/ListTransactions", connect.NewUnaryHandler("/"+serviceName+"/ListTransactions", handler.ListTransactions, opts...))
	mux.Handle("/"+serviceName+"/Balances", connect.NewUnaryHandler("/"+serviceName+"/Balances", handler.Balances, opts...))

	return "/" + serviceName + "/", mux
}
