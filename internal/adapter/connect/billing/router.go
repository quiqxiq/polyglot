package billing

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
)

// NewBillingServiceHandler mounts all 17 BillingService Connect handlers.
func NewBillingServiceHandler(
	invUC *billingUC.InvoiceUseCase,
	checkoutUC *billingUC.CheckoutUseCase,
	subUC *billingUC.SubscriptionUseCase,
	lifecycleUC *billingUC.SubscriptionLifecycleUseCase,
	planUC *billingUC.PlanUseCase,
	runBilling *billingUC.RunBillingUseCase,
	manageSubUC *billingUC.ManageSubscriptionUseCase,
) (string, http.Handler) {
	handler := NewBillingConnectHandler(invUC, checkoutUC, subUC, lifecycleUC, planUC, runBilling, manageSubUC)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.BillingService"

	mux.Handle("/"+serviceName+"/ListInvoices", connect.NewUnaryHandler("/"+serviceName+"/ListInvoices", handler.ListInvoices, codecOpt))
	mux.Handle("/"+serviceName+"/GetInvoice", connect.NewUnaryHandler("/"+serviceName+"/GetInvoice", handler.GetInvoice, codecOpt))

	mux.Handle("/"+serviceName+"/CashierResolve", connect.NewUnaryHandler("/"+serviceName+"/CashierResolve", handler.CashierResolve, codecOpt))
	mux.Handle("/"+serviceName+"/CashierPay", connect.NewUnaryHandler("/"+serviceName+"/CashierPay", handler.CashierPay, codecOpt))

	mux.Handle("/"+serviceName+"/ListSubscriptions", connect.NewUnaryHandler("/"+serviceName+"/ListSubscriptions", handler.ListSubscriptions, codecOpt))
	mux.Handle("/"+serviceName+"/GetSubscription", connect.NewUnaryHandler("/"+serviceName+"/GetSubscription", handler.GetSubscription, codecOpt))
	mux.Handle("/"+serviceName+"/ChangePlan", connect.NewUnaryHandler("/"+serviceName+"/ChangePlan", handler.ChangePlan, codecOpt))
	mux.Handle("/"+serviceName+"/CreateSubscription", connect.NewUnaryHandler("/"+serviceName+"/CreateSubscription", handler.CreateSubscription, codecOpt))
	mux.Handle("/"+serviceName+"/UpdateSubscription", connect.NewUnaryHandler("/"+serviceName+"/UpdateSubscription", handler.UpdateSubscription, codecOpt))
	mux.Handle("/"+serviceName+"/DeleteSubscription", connect.NewUnaryHandler("/"+serviceName+"/DeleteSubscription", handler.DeleteSubscription, codecOpt))
	mux.Handle("/"+serviceName+"/SuspendSubscription", connect.NewUnaryHandler("/"+serviceName+"/SuspendSubscription", handler.SuspendSubscription, codecOpt))
	mux.Handle("/"+serviceName+"/ResumeSubscription", connect.NewUnaryHandler("/"+serviceName+"/ResumeSubscription", handler.ResumeSubscription, codecOpt))
	mux.Handle("/"+serviceName+"/TerminateSubscription", connect.NewUnaryHandler("/"+serviceName+"/TerminateSubscription", handler.TerminateSubscription, codecOpt))
	mux.Handle("/"+serviceName+"/ActivateSubscription", connect.NewUnaryHandler("/"+serviceName+"/ActivateSubscription", handler.ActivateSubscription, codecOpt))

	mux.Handle("/"+serviceName+"/ListPlans", connect.NewUnaryHandler("/"+serviceName+"/ListPlans", handler.ListPlans, codecOpt))
	mux.Handle("/"+serviceName+"/GetPlan", connect.NewUnaryHandler("/"+serviceName+"/GetPlan", handler.GetPlan, codecOpt))
	mux.Handle("/"+serviceName+"/CreatePlan", connect.NewUnaryHandler("/"+serviceName+"/CreatePlan", handler.CreatePlan, codecOpt))
	mux.Handle("/"+serviceName+"/UpdatePlan", connect.NewUnaryHandler("/"+serviceName+"/UpdatePlan", handler.UpdatePlan, codecOpt))
	mux.Handle("/"+serviceName+"/DeletePlan", connect.NewUnaryHandler("/"+serviceName+"/DeletePlan", handler.DeletePlan, codecOpt))

	mux.Handle("/"+serviceName+"/GenerateInvoices", connect.NewUnaryHandler("/"+serviceName+"/GenerateInvoices", handler.GenerateInvoices, codecOpt))

	return "/" + serviceName + "/", mux
}
