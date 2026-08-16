package customer

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	customerUC "github.com/quixiq/polyglot/internal/usecase/customer"
)

// NewCustomerServiceHandler mounts the CustomerService Connect handler to http.ServeMux.
func NewCustomerServiceHandler(uc *customerUC.ManageCustomerUseCase) (string, http.Handler) {
	handler := NewCustomerConnectHandler(uc)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.CustomerService"
	mux.Handle("/"+serviceName+"/ListCustomers", connect.NewUnaryHandler(
		"/"+serviceName+"/ListCustomers",
		handler.ListCustomers,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/GetCustomer", connect.NewUnaryHandler(
		"/"+serviceName+"/GetCustomer",
		handler.GetCustomer,
		codecOpt,
	))

	return "/" + serviceName + "/", mux
}
