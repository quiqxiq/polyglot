package cashbook

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/port"
)

func NewCashbookServiceHandler(repo port.CashbookRepository) (string, http.Handler) {
	handler := NewCashbookConnectHandler(repo)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.CashbookService"
	mux.Handle("/"+serviceName+"/ListAccounts", connect.NewUnaryHandler("/"+serviceName+"/ListAccounts", handler.ListAccounts, codecOpt))
	mux.Handle("/"+serviceName+"/SaveAccount", connect.NewUnaryHandler("/"+serviceName+"/SaveAccount", handler.SaveAccount, codecOpt))
	mux.Handle("/"+serviceName+"/ListCategories", connect.NewUnaryHandler("/"+serviceName+"/ListCategories", handler.ListCategories, codecOpt))
	mux.Handle("/"+serviceName+"/SaveCategory", connect.NewUnaryHandler("/"+serviceName+"/SaveCategory", handler.SaveCategory, codecOpt))
	mux.Handle("/"+serviceName+"/AddTransaction", connect.NewUnaryHandler("/"+serviceName+"/AddTransaction", handler.AddTransaction, codecOpt))
	mux.Handle("/"+serviceName+"/ListTransactions", connect.NewUnaryHandler("/"+serviceName+"/ListTransactions", handler.ListTransactions, codecOpt))
	mux.Handle("/"+serviceName+"/Balances", connect.NewUnaryHandler("/"+serviceName+"/Balances", handler.Balances, codecOpt))

	return "/" + serviceName + "/", mux
}
