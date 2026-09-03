package billing

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
)

// NewBillingServiceHandler mounts BillingService Connect handlers (Invoice & Cashier).
func NewBillingServiceHandler(
	invUC *billingUC.InvoiceUseCase,
	checkoutUC *billingUC.CheckoutUseCase,
	runBilling *billingUC.RunBillingUseCase,
) (string, http.Handler) {
	handler := NewBillingConnectHandler(invUC, checkoutUC, runBilling)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.BillingService"

	mux.Handle("/"+serviceName+"/ListInvoices", connect.NewUnaryHandler("/"+serviceName+"/ListInvoices", handler.ListInvoices, opts...))
	mux.Handle("/"+serviceName+"/GetInvoice", connect.NewUnaryHandler("/"+serviceName+"/GetInvoice", handler.GetInvoice, opts...))
	mux.Handle("/"+serviceName+"/GenerateInvoices", connect.NewUnaryHandler("/"+serviceName+"/GenerateInvoices", handler.GenerateInvoices, opts...))

	mux.Handle("/"+serviceName+"/CashierResolve", connect.NewUnaryHandler("/"+serviceName+"/CashierResolve", handler.CashierResolve, opts...))
	mux.Handle("/"+serviceName+"/CashierPay", connect.NewUnaryHandler("/"+serviceName+"/CashierPay", handler.CashierPay, opts...))

	return "/" + serviceName + "/", mux
}
