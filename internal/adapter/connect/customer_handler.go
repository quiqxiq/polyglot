package connectadapter

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
	"github.com/quixiq/polyglot/internal/adapter/connect/codec"
	"github.com/quixiq/polyglot/internal/adapter/connect/mapper"
	"github.com/quixiq/polyglot/internal/usecase/business"
	"github.com/quixiq/polyglot/pkg/response"
)

// CustomerConnectHandler handles ConnectRPC procedures for customer subscriber accounts.
type CustomerConnectHandler struct {
	useCase *business.ManageCustomerUseCase
}

// NewCustomerConnectHandler creates a new CustomerConnectHandler.
func NewCustomerConnectHandler(uc *business.ManageCustomerUseCase) *CustomerConnectHandler {
	return &CustomerConnectHandler{useCase: uc}
}

// ListCustomers returns all registered customers.
func (h *CustomerConnectHandler) ListCustomers(ctx context.Context, req *connect.Request[devicepb.ListCustomersRequest]) (*connect.Response[devicepb.ListCustomersResponse], error) {
	customers, err := h.useCase.ListCustomers(ctx)
	if err != nil {
		return nil, response.ToConnectError(err)
	}

	return connect.NewResponse(&devicepb.ListCustomersResponse{
		Customers: mapper.CustomerListToProto(customers),
	}), nil
}

// GetCustomer retrieves a single customer by ID.
func (h *CustomerConnectHandler) GetCustomer(ctx context.Context, req *connect.Request[devicepb.GetCustomerRequest]) (*connect.Response[devicepb.GetCustomerResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("customer id is required"))
	}

	c, err := h.useCase.GetCustomer(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.ToConnectError(err)
	}

	return connect.NewResponse(&devicepb.GetCustomerResponse{
		Customer: mapper.CustomerToProto(c),
	}), nil
}

// NewCustomerServiceHandler creates the Connect http.Handler and registers procedures.
func NewCustomerServiceHandler(uc *business.ManageCustomerUseCase) (string, http.Handler) {
	handler := NewCustomerConnectHandler(uc)
	mux := http.NewServeMux()
	codecOpt := codec.Option()

	serviceName := "polyglot.v1.CustomerService"
	mux.Handle("/"+serviceName+"/ListCustomers", connect.NewUnaryHandler("/"+serviceName+"/ListCustomers", handler.ListCustomers, codecOpt))
	mux.Handle("/"+serviceName+"/GetCustomer", connect.NewUnaryHandler("/"+serviceName+"/GetCustomer", handler.GetCustomer, codecOpt))

	return "/" + serviceName + "/", mux
}
