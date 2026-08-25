package isolir

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
)

// NewIsolationServiceHandler mounts the IsolationService Connect handler.
func NewIsolationServiceHandler(uc *networkUC.ManageIsolationUseCase) (string, http.Handler) {
	handler := NewIsolationConnectHandler(uc)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.IsolationService"
	mux.Handle("/"+serviceName+"/SetupIsolation", connect.NewUnaryHandler(
		"/"+serviceName+"/SetupIsolation", handler.SetupIsolation, codecOpt))
	mux.Handle("/"+serviceName+"/GetIsolationStatus", connect.NewUnaryHandler(
		"/"+serviceName+"/GetIsolationStatus", handler.GetIsolationStatus, codecOpt))
	mux.Handle("/"+serviceName+"/RemoveIsolation", connect.NewUnaryHandler(
		"/"+serviceName+"/RemoveIsolation", handler.RemoveIsolation, codecOpt))

	return "/" + serviceName + "/", mux
}
