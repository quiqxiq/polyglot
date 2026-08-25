package plan

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	planUC "github.com/quixiq/polyglot/internal/usecase/plan"
)

// NewPlanServiceHandler mounts the PlanService Connect handler onto http.ServeMux.
func NewPlanServiceHandler(uc *planUC.ManagePlansUseCase) (string, http.Handler) {
	handler := NewPlanConnectHandler(uc)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.PlanService"
	mux.Handle("/"+serviceName+"/CreatePlan", connect.NewUnaryHandler(
		"/"+serviceName+"/CreatePlan", handler.CreatePlan, codecOpt))
	mux.Handle("/"+serviceName+"/GetPlan", connect.NewUnaryHandler(
		"/"+serviceName+"/GetPlan", handler.GetPlan, codecOpt))
	mux.Handle("/"+serviceName+"/ListPlans", connect.NewUnaryHandler(
		"/"+serviceName+"/ListPlans", handler.ListPlans, codecOpt))
	mux.Handle("/"+serviceName+"/UpdatePlan", connect.NewUnaryHandler(
		"/"+serviceName+"/UpdatePlan", handler.UpdatePlan, codecOpt))
	mux.Handle("/"+serviceName+"/DeletePlan", connect.NewUnaryHandler(
		"/"+serviceName+"/DeletePlan", handler.DeletePlan, codecOpt))

	return "/" + serviceName + "/", mux
}
