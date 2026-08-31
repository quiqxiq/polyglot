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
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.BillingService"

	mux.Handle("/"+serviceName+"/ListInvoices", connect.NewUnaryHandler("/"+serviceName+"/ListInvoices", handler.ListInvoices, opts...))
	mux.Handle("/"+serviceName+"/GetInvoice", connect.NewUnaryHandler("/"+serviceName+"/GetInvoice", handler.GetInvoice, opts...))

	mux.Handle("/"+serviceName+"/CashierResolve", connect.NewUnaryHandler("/"+serviceName+"/CashierResolve", handler.CashierResolve, opts...))
	mux.Handle("/"+serviceName+"/CashierPay", connect.NewUnaryHandler("/"+serviceName+"/CashierPay", handler.CashierPay, opts...))

	mux.Handle("/"+serviceName+"/ListSubscriptions", connect.NewUnaryHandler("/"+serviceName+"/ListSubscriptions", handler.ListSubscriptions, opts...))
	mux.Handle("/"+serviceName+"/GetSubscription", connect.NewUnaryHandler("/"+serviceName+"/GetSubscription", handler.GetSubscription, opts...))
	mux.Handle("/"+serviceName+"/ChangePlan", connect.NewUnaryHandler("/"+serviceName+"/ChangePlan", handler.ChangePlan, opts...))
	mux.Handle("/"+serviceName+"/CreateSubscription", connect.NewUnaryHandler("/"+serviceName+"/CreateSubscription", handler.CreateSubscription, opts...))
	mux.Handle("/"+serviceName+"/UpdateSubscription", connect.NewUnaryHandler("/"+serviceName+"/UpdateSubscription", handler.UpdateSubscription, opts...))
	mux.Handle("/"+serviceName+"/DeleteSubscription", connect.NewUnaryHandler("/"+serviceName+"/DeleteSubscription", handler.DeleteSubscription, opts...))
	mux.Handle("/"+serviceName+"/SuspendSubscription", connect.NewUnaryHandler("/"+serviceName+"/SuspendSubscription", handler.SuspendSubscription, opts...))
	mux.Handle("/"+serviceName+"/ResumeSubscription", connect.NewUnaryHandler("/"+serviceName+"/ResumeSubscription", handler.ResumeSubscription, opts...))
	mux.Handle("/"+serviceName+"/TerminateSubscription", connect.NewUnaryHandler("/"+serviceName+"/TerminateSubscription", handler.TerminateSubscription, opts...))
	mux.Handle("/"+serviceName+"/ActivateSubscription", connect.NewUnaryHandler("/"+serviceName+"/ActivateSubscription", handler.ActivateSubscription, opts...))
	mux.Handle("/"+serviceName+"/IsolateSubscription", connect.NewUnaryHandler("/"+serviceName+"/IsolateSubscription", handler.IsolateSubscription, opts...))
	mux.Handle("/"+serviceName+"/RestoreSubscription", connect.NewUnaryHandler("/"+serviceName+"/RestoreSubscription", handler.RestoreSubscription, opts...))

	mux.Handle("/"+serviceName+"/ListPlans", connect.NewUnaryHandler("/"+serviceName+"/ListPlans", handler.ListPlans, opts...))
	mux.Handle("/"+serviceName+"/GetPlan", connect.NewUnaryHandler("/"+serviceName+"/GetPlan", handler.GetPlan, opts...))
	mux.Handle("/"+serviceName+"/CreatePlan", connect.NewUnaryHandler("/"+serviceName+"/CreatePlan", handler.CreatePlan, opts...))
	mux.Handle("/"+serviceName+"/UpdatePlan", connect.NewUnaryHandler("/"+serviceName+"/UpdatePlan", handler.UpdatePlan, opts...))
	mux.Handle("/"+serviceName+"/DeletePlan", connect.NewUnaryHandler("/"+serviceName+"/DeletePlan", handler.DeletePlan, opts...))

	mux.Handle("/"+serviceName+"/GenerateInvoices", connect.NewUnaryHandler("/"+serviceName+"/GenerateInvoices", handler.GenerateInvoices, opts...))

	return "/" + serviceName + "/", mux
}
