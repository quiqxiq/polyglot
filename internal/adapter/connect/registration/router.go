package registration

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	regUC "github.com/quixiq/polyglot/internal/usecase/registration"
)

// NewRegistrationServiceHandler mounts the RegistrationService Connect handler.
func NewRegistrationServiceHandler(uc *regUC.RegistrationService) (string, http.Handler) {
	handler := NewRegistrationConnectHandler(uc)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.RegistrationService"
	mux.Handle("/"+serviceName+"/CreateRegistration", connect.NewUnaryHandler(
		"/"+serviceName+"/CreateRegistration", handler.CreateRegistration, codecOpt))
	mux.Handle("/"+serviceName+"/ListRegistrations", connect.NewUnaryHandler(
		"/"+serviceName+"/ListRegistrations", handler.ListRegistrations, codecOpt))
	mux.Handle("/"+serviceName+"/GetRegistration", connect.NewUnaryHandler(
		"/"+serviceName+"/GetRegistration", handler.GetRegistration, codecOpt))
	mux.Handle("/"+serviceName+"/ReviewRegistration", connect.NewUnaryHandler(
		"/"+serviceName+"/ReviewRegistration", handler.ReviewRegistration, codecOpt))
	mux.Handle("/"+serviceName+"/MarkInstalled", connect.NewUnaryHandler(
		"/"+serviceName+"/MarkInstalled", handler.MarkInstalled, codecOpt))
	mux.Handle("/"+serviceName+"/CancelRegistration", connect.NewUnaryHandler(
		"/"+serviceName+"/CancelRegistration", handler.CancelRegistration, codecOpt))

	return "/" + serviceName + "/", mux
}
