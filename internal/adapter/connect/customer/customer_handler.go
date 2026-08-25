package customer

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	customerUC "github.com/quixiq/polyglot/internal/usecase/customer"
	"github.com/quixiq/polyglot/pkg/response"
)

type CustomerConnectHandler struct {
	useCase *customerUC.ManageCustomerUseCase
}

func NewCustomerConnectHandler(uc *customerUC.ManageCustomerUseCase) *CustomerConnectHandler {
	return &CustomerConnectHandler{useCase: uc}
}

func (h *CustomerConnectHandler) ListCustomers(ctx context.Context, req *connect.Request[devicepb.ListCustomersRequest]) (*connect.Response[devicepb.ListCustomersResponse], error) {
	customers, err := h.useCase.ListCustomers(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ListCustomersResponse{
		Customers: toProtoCustomerList(customers),
	}), nil
}

func (h *CustomerConnectHandler) GetCustomer(ctx context.Context, req *connect.Request[devicepb.GetCustomerRequest]) (*connect.Response[devicepb.GetCustomerResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("customer id is required"))
	}

	c, err := h.useCase.GetCustomer(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.GetCustomerResponse{
		Customer: toProtoCustomer(&c),
	}), nil
}

func (h *CustomerConnectHandler) CreateCustomer(ctx context.Context, req *connect.Request[devicepb.CreateCustomerRequest]) (*connect.Response[devicepb.CreateCustomerResponse], error) {
	if req.Msg.Customer == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("customer is required"))
	}
	created, err := h.useCase.CreateCustomer(ctx, fromProtoCustomer(req.Msg.Customer))
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.CreateCustomerResponse{
		Customer: toProtoCustomer(&created),
		Message:  "customer created",
	}), nil
}

func (h *CustomerConnectHandler) UpdateCustomer(ctx context.Context, req *connect.Request[devicepb.UpdateCustomerRequest]) (*connect.Response[devicepb.UpdateCustomerResponse], error) {
	if req.Msg.Customer == nil || req.Msg.Customer.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("customer.id is required"))
	}
	updated, err := h.useCase.UpdateCustomer(ctx, fromProtoCustomer(req.Msg.Customer))
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.UpdateCustomerResponse{
		Customer: toProtoCustomer(&updated),
		Message:  "customer updated",
	}), nil
}

func (h *CustomerConnectHandler) DeleteCustomer(ctx context.Context, req *connect.Request[devicepb.DeleteCustomerRequest]) (*connect.Response[devicepb.DeleteCustomerResponse], error) {
	if err := h.useCase.DeleteCustomer(ctx, req.Msg.Id); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.DeleteCustomerResponse{
		Message: "customer deleted",
	}), nil
}
