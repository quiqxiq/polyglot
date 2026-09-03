package subscription

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	subUC "github.com/quixiq/polyglot/internal/usecase/subscription"
)

// NewSubscriptionServiceHandler mounts SubscriptionService Connect handlers.
func NewSubscriptionServiceHandler(
	subUC *subUC.ManageSubscriptionUseCase,
	lifecycleUC *subUC.LifecycleUseCase,
) (string, http.Handler) {
	handler := NewSubscriptionConnectHandler(subUC, lifecycleUC)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.SubscriptionService"

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

	return "/" + serviceName + "/", mux
}
