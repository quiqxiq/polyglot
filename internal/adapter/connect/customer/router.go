package customer

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	customerUC "github.com/quixiq/polyglot/internal/usecase/customer"
)

// NewCustomerServiceHandler mounts CustomerService Connect handlers to http.ServeMux.
func NewCustomerServiceHandler(uc *customerUC.ManageCustomerUseCase) (string, http.Handler) {
	handler := NewCustomerConnectHandler(uc)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.CustomerService"
	mux.Handle("/"+serviceName+"/ListCustomers", connect.NewUnaryHandler("/"+serviceName+"/ListCustomers", handler.ListCustomers, opts...))
	mux.Handle("/"+serviceName+"/GetCustomer", connect.NewUnaryHandler("/"+serviceName+"/GetCustomer", handler.GetCustomer, opts...))
	mux.Handle("/"+serviceName+"/CreateCustomer", connect.NewUnaryHandler("/"+serviceName+"/CreateCustomer", handler.CreateCustomer, opts...))
	mux.Handle("/"+serviceName+"/UpdateCustomer", connect.NewUnaryHandler("/"+serviceName+"/UpdateCustomer", handler.UpdateCustomer, opts...))
	mux.Handle("/"+serviceName+"/DeleteCustomer", connect.NewUnaryHandler("/"+serviceName+"/DeleteCustomer", handler.DeleteCustomer, opts...))

	mux.Handle("/"+serviceName+"/FindByPhone", connect.NewUnaryHandler("/"+serviceName+"/FindByPhone", handler.FindByPhone, opts...))
	mux.Handle("/"+serviceName+"/FindByCustomerCode", connect.NewUnaryHandler("/"+serviceName+"/FindByCustomerCode", handler.FindByCustomerCode, opts...))
	mux.Handle("/"+serviceName+"/FindByPortalCode", connect.NewUnaryHandler("/"+serviceName+"/FindByPortalCode", handler.FindByPortalCode, opts...))

	mux.Handle("/"+serviceName+"/ListSubscriptions", connect.NewUnaryHandler("/"+serviceName+"/ListSubscriptions", handler.ListSubscriptions, opts...))

	return "/" + serviceName + "/", mux
}
