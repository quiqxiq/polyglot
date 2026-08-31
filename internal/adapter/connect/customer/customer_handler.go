package customer

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	customerUC "github.com/quixiq/polyglot/internal/usecase/customer"
	"github.com/quixiq/polyglot/pkg/response"
)

type CustomerConnectHandler struct {
	customerUC *customerUC.ManageCustomerUseCase
}

func NewCustomerConnectHandler(uc *customerUC.ManageCustomerUseCase) *CustomerConnectHandler {
	return &CustomerConnectHandler{customerUC: uc}
}

func (h *CustomerConnectHandler) ListCustomers(ctx context.Context, req *connect.Request[devicepb.ListCustomersRequest]) (*connect.Response[devicepb.ListCustomersResponse], error) {
	if h.customerUC == nil {
		return nil, response.Unavailable("customer usecase unavailable")
	}
	customers, err := h.customerUC.ListCustomers(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListCustomersResponse{
		Customers: toProtoCustomerList(customers),
	}), nil
}

func (h *CustomerConnectHandler) GetCustomer(ctx context.Context, req *connect.Request[devicepb.GetCustomerRequest]) (*connect.Response[devicepb.GetCustomerResponse], error) {
	if h.customerUC == nil {
		return nil, response.Unavailable("customer usecase unavailable")
	}
	c, err := h.customerUC.GetCustomer(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetCustomerResponse{
		Customer: toProtoCustomer(&c),
	}), nil
}

func (h *CustomerConnectHandler) CreateCustomer(ctx context.Context, req *connect.Request[devicepb.CreateCustomerRequest]) (*connect.Response[devicepb.CreateCustomerResponse], error) {
	if h.customerUC == nil {
		return nil, response.Unavailable("customer usecase unavailable")
	}
	c := fromProtoCustomer(req.Msg.Customer)
	created, err := h.customerUC.CreateCustomer(ctx, c)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.CreateCustomerResponse{
		Customer: toProtoCustomer(&created),
		Message:  "Customer created successfully",
	}), nil
}

func (h *CustomerConnectHandler) UpdateCustomer(ctx context.Context, req *connect.Request[devicepb.UpdateCustomerRequest]) (*connect.Response[devicepb.UpdateCustomerResponse], error) {
	if h.customerUC == nil {
		return nil, response.Unavailable("customer usecase unavailable")
	}
	c := fromProtoCustomer(req.Msg.Customer)
	updated, err := h.customerUC.UpdateCustomer(ctx, c)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.UpdateCustomerResponse{
		Customer: toProtoCustomer(&updated),
		Message:  "Customer updated successfully",
	}), nil
}

func (h *CustomerConnectHandler) DeleteCustomer(ctx context.Context, req *connect.Request[devicepb.DeleteCustomerRequest]) (*connect.Response[devicepb.DeleteCustomerResponse], error) {
	if h.customerUC == nil {
		return nil, response.Unavailable("customer usecase unavailable")
	}
	if err := h.customerUC.DeleteCustomer(ctx, req.Msg.Id); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.DeleteCustomerResponse{
		Message: "Customer deleted successfully",
	}), nil
}

func (h *CustomerConnectHandler) FindByPhone(ctx context.Context, req *connect.Request[devicepb.FindCustomerByPhoneRequest]) (*connect.Response[devicepb.FindCustomerResponse], error) {
	c, err := h.customerUC.FindByPhone(ctx, req.Msg.Phone)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.FindCustomerResponse{Customer: toProtoCustomer(&c)}), nil
}

func (h *CustomerConnectHandler) FindByCustomerCode(ctx context.Context, req *connect.Request[devicepb.FindCustomerByCodeRequest]) (*connect.Response[devicepb.FindCustomerResponse], error) {
	c, err := h.customerUC.FindByCustomerCode(ctx, req.Msg.CustomerCode)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.FindCustomerResponse{Customer: toProtoCustomer(&c)}), nil
}

func (h *CustomerConnectHandler) FindByPortalCode(ctx context.Context, req *connect.Request[devicepb.FindCustomerByPortalCodeRequest]) (*connect.Response[devicepb.FindCustomerResponse], error) {
	c, err := h.customerUC.FindByPortalCode(ctx, req.Msg.PortalAccessCode)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.FindCustomerResponse{Customer: toProtoCustomer(&c)}), nil
}

func (h *CustomerConnectHandler) ListSubscriptions(ctx context.Context, req *connect.Request[devicepb.ListSubscriptionsRequest]) (*connect.Response[devicepb.ListSubscriptionsResponse], error) {
	return connect.NewResponse(&devicepb.ListSubscriptionsResponse{}), nil
}
