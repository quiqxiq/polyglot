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
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.CustomerService"
	mux.Handle("/"+serviceName+"/ListCustomers", connect.NewUnaryHandler("/"+serviceName+"/ListCustomers", handler.ListCustomers, codecOpt))
	mux.Handle("/"+serviceName+"/GetCustomer", connect.NewUnaryHandler("/"+serviceName+"/GetCustomer", handler.GetCustomer, codecOpt))
	mux.Handle("/"+serviceName+"/CreateCustomer", connect.NewUnaryHandler("/"+serviceName+"/CreateCustomer", handler.CreateCustomer, codecOpt))
	mux.Handle("/"+serviceName+"/UpdateCustomer", connect.NewUnaryHandler("/"+serviceName+"/UpdateCustomer", handler.UpdateCustomer, codecOpt))
	mux.Handle("/"+serviceName+"/DeleteCustomer", connect.NewUnaryHandler("/"+serviceName+"/DeleteCustomer", handler.DeleteCustomer, codecOpt))

	mux.Handle("/"+serviceName+"/FindByPhone", connect.NewUnaryHandler("/"+serviceName+"/FindByPhone", handler.FindByPhone, codecOpt))
	mux.Handle("/"+serviceName+"/FindByCustomerCode", connect.NewUnaryHandler("/"+serviceName+"/FindByCustomerCode", handler.FindByCustomerCode, codecOpt))
	mux.Handle("/"+serviceName+"/FindByPortalCode", connect.NewUnaryHandler("/"+serviceName+"/FindByPortalCode", handler.FindByPortalCode, codecOpt))

	mux.Handle("/"+serviceName+"/ListSubscriptions", connect.NewUnaryHandler("/"+serviceName+"/ListSubscriptions", handler.ListSubscriptions, codecOpt))

	return "/" + serviceName + "/", mux
}
