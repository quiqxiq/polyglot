package plan

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	planUC "github.com/quixiq/polyglot/internal/usecase/plan"
)

// NewPlanServiceHandler mounts PlanService Connect handlers.
func NewPlanServiceHandler(
	planUC *planUC.ManagePlanUseCase,
) (string, http.Handler) {
	handler := NewPlanConnectHandler(planUC)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.PlanService"

	mux.Handle("/"+serviceName+"/ListPlans", connect.NewUnaryHandler("/"+serviceName+"/ListPlans", handler.ListPlans, opts...))
	mux.Handle("/"+serviceName+"/GetPlan", connect.NewUnaryHandler("/"+serviceName+"/GetPlan", handler.GetPlan, opts...))
	mux.Handle("/"+serviceName+"/CreatePlan", connect.NewUnaryHandler("/"+serviceName+"/CreatePlan", handler.CreatePlan, opts...))
	mux.Handle("/"+serviceName+"/UpdatePlan", connect.NewUnaryHandler("/"+serviceName+"/UpdatePlan", handler.UpdatePlan, opts...))
	mux.Handle("/"+serviceName+"/DeletePlan", connect.NewUnaryHandler("/"+serviceName+"/DeletePlan", handler.DeletePlan, opts...))

	return "/" + serviceName + "/", mux
}
