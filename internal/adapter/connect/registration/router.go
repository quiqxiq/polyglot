package registration

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/port"
	uc "github.com/quixiq/polyglot/internal/usecase/registration"
)

// NewRegistrationServiceHandler mounts staff-facing RegistrationService
// Connect handlers to http.ServeMux (di balik JWT+RBAC).
func NewRegistrationServiceHandler(
	managerUC *uc.ManageRegistrationUseCase,
	convertUC *uc.ConvertUseCase,
	repo port.RegistrationRepository,
) (string, http.Handler) {
	handler := NewRegistrationConnectHandler(managerUC, convertUC, repo)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.RegistrationService"
	mux.Handle("/"+serviceName+"/ListRegistrations", connect.NewUnaryHandler("/"+serviceName+"/ListRegistrations", handler.ListRegistrations, codecOpt))
	mux.Handle("/"+serviceName+"/GetRegistration", connect.NewUnaryHandler("/"+serviceName+"/GetRegistration", handler.GetRegistration, codecOpt))
	mux.Handle("/"+serviceName+"/ApproveRegistration", connect.NewUnaryHandler("/"+serviceName+"/ApproveRegistration", handler.ApproveRegistration, codecOpt))
	mux.Handle("/"+serviceName+"/ScheduleInstall", connect.NewUnaryHandler("/"+serviceName+"/ScheduleInstall", handler.ScheduleInstall, codecOpt))
	mux.Handle("/"+serviceName+"/MarkInstalled", connect.NewUnaryHandler("/"+serviceName+"/MarkInstalled", handler.MarkInstalled, codecOpt))
	mux.Handle("/"+serviceName+"/RejectRegistration", connect.NewUnaryHandler("/"+serviceName+"/RejectRegistration", handler.RejectRegistration, codecOpt))
	mux.Handle("/"+serviceName+"/CancelRegistration", connect.NewUnaryHandler("/"+serviceName+"/CancelRegistration", handler.CancelRegistration, codecOpt))
	mux.Handle("/"+serviceName+"/ConvertRegistration", connect.NewUnaryHandler("/"+serviceName+"/ConvertRegistration", handler.ConvertRegistration, codecOpt))

	return "/" + serviceName + "/", mux
}

// NewPublicSubmitHandler mounts SubmitRegistration secara publik (calon
// pelanggan tanpa JWT) pada rootMux.
func NewPublicSubmitHandler(managerUC *uc.ManageRegistrationUseCase) (string, http.Handler) {
	handler := NewRegistrationConnectHandler(managerUC, nil, nil)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.RegistrationService"
	mux.Handle("/"+serviceName+"/SubmitRegistration", connect.NewUnaryHandler(
		"/"+serviceName+"/SubmitRegistration", handler.SubmitRegistration, codecOpt))

	return "/" + serviceName + "/SubmitRegistration", mux
}
