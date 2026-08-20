package billing

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
)

// NewBillingServiceHandler creates the Connect http.Handler for BillingService.
func NewBillingServiceHandler(
	invUC *billingUC.InvoiceUseCase,
	subUC *billingUC.SubscriptionUseCase,
) (string, http.Handler) {
	handler := NewBillingConnectHandler(invUC, subUC)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.BillingService"
	mux.Handle("/"+serviceName+"/ListInvoices", connect.NewUnaryHandler("/"+serviceName+"/ListInvoices", handler.ListInvoices, codecOpt))
	mux.Handle("/"+serviceName+"/GetInvoice", connect.NewUnaryHandler("/"+serviceName+"/GetInvoice", handler.GetInvoice, codecOpt))
	mux.Handle("/"+serviceName+"/CreateInvoice", connect.NewUnaryHandler("/"+serviceName+"/CreateInvoice", handler.CreateInvoice, codecOpt))
	mux.Handle("/"+serviceName+"/PayInvoice", connect.NewUnaryHandler("/"+serviceName+"/PayInvoice", handler.PayInvoice, codecOpt))
	mux.Handle("/"+serviceName+"/ListSubscriptions", connect.NewUnaryHandler("/"+serviceName+"/ListSubscriptions", handler.ListSubscriptions, codecOpt))
	mux.Handle("/"+serviceName+"/CreateSubscription", connect.NewUnaryHandler("/"+serviceName+"/CreateSubscription", handler.CreateSubscription, codecOpt))
	mux.Handle("/"+serviceName+"/CancelSubscription", connect.NewUnaryHandler("/"+serviceName+"/CancelSubscription", handler.CancelSubscription, codecOpt))

	return "/" + serviceName + "/", mux
}
