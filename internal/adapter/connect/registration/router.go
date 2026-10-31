package registration

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	uc "github.com/quixiq/polyglot/internal/usecase/registration"
)

// NewRegistrationServiceHandler mounts staff-facing RegistrationService
// Connect handlers to http.ServeMux (di balik JWT+RBAC).
func NewRegistrationServiceHandler(
	managerUC *uc.ManageRegistrationUseCase,
	convertUC *uc.ConvertUseCase,
) (string, http.Handler) {
	handler := NewRegistrationConnectHandler(managerUC, convertUC)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.RegistrationService"
	mux.Handle("/"+serviceName+"/ListRegistrations", connect.NewUnaryHandler("/"+serviceName+"/ListRegistrations", handler.ListRegistrations, opts...))
	mux.Handle("/"+serviceName+"/GetRegistration", connect.NewUnaryHandler("/"+serviceName+"/GetRegistration", handler.GetRegistration, opts...))
	mux.Handle("/"+serviceName+"/ApproveRegistration", connect.NewUnaryHandler("/"+serviceName+"/ApproveRegistration", handler.ApproveRegistration, opts...))
	mux.Handle("/"+serviceName+"/ScheduleInstall", connect.NewUnaryHandler("/"+serviceName+"/ScheduleInstall", handler.ScheduleInstall, opts...))
	mux.Handle("/"+serviceName+"/RecordSurvey", connect.NewUnaryHandler("/"+serviceName+"/RecordSurvey", handler.RecordSurvey, opts...))
	mux.Handle("/"+serviceName+"/StartInstallation", connect.NewUnaryHandler("/"+serviceName+"/StartInstallation", handler.StartInstallation, opts...))
	mux.Handle("/"+serviceName+"/MarkInstalled", connect.NewUnaryHandler("/"+serviceName+"/MarkInstalled", handler.MarkInstalled, opts...))
	mux.Handle("/"+serviceName+"/RejectRegistration", connect.NewUnaryHandler("/"+serviceName+"/RejectRegistration", handler.RejectRegistration, opts...))
	mux.Handle("/"+serviceName+"/CancelRegistration", connect.NewUnaryHandler("/"+serviceName+"/CancelRegistration", handler.CancelRegistration, opts...))
	mux.Handle("/"+serviceName+"/ConvertRegistration", connect.NewUnaryHandler("/"+serviceName+"/ConvertRegistration", handler.ConvertRegistration, opts...))
	mux.Handle("/"+serviceName+"/ListRegistrationEvents", connect.NewUnaryHandler("/"+serviceName+"/ListRegistrationEvents", handler.ListRegistrationEvents, opts...))

	return "/" + serviceName + "/", mux
}

// RegisterPublic mounts public registration endpoints (SubmitRegistration, TrackRegistration, ListPublicPlans)
// directly on mux (allowing unauthenticated access without conflicting with protected RegistrationService routes).
func RegisterPublic(mux *http.ServeMux, managerUC *uc.ManageRegistrationUseCase) {
	handler := NewRegistrationConnectHandler(managerUC, nil)
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.RegistrationService"
	mux.Handle("/"+serviceName+"/SubmitRegistration", connect.NewUnaryHandler(
		"/"+serviceName+"/SubmitRegistration", handler.SubmitRegistration, opts...))
	mux.Handle("/"+serviceName+"/TrackRegistration", connect.NewUnaryHandler(
		"/"+serviceName+"/TrackRegistration", handler.TrackRegistration, opts...))
	mux.Handle("/"+serviceName+"/ListPublicPlans", connect.NewUnaryHandler(
		"/"+serviceName+"/ListPublicPlans", handler.ListPublicPlans, opts...))
}

// NewPublicSubmitHandler mounts SubmitRegistration secara publik (calon
// pelanggan tanpa JWT) pada rootMux.
//
// Deprecated: use RegisterPublic directly on rootMux to avoid ServeMux subtree pattern conflicts.
func NewPublicSubmitHandler(managerUC *uc.ManageRegistrationUseCase) (string, http.Handler) {
	mux := http.NewServeMux()
	RegisterPublic(mux, managerUC)
	return "/polyglot.v1.RegistrationService/", mux
}
