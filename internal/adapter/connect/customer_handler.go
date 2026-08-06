package connectadapter

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
	"github.com/quixiq/polyglot/internal/usecase/business"
)

type CustomerConnectHandler struct {
	useCase *business.ManageCustomerUseCase
}

func NewCustomerConnectHandler(uc *business.ManageCustomerUseCase) *CustomerConnectHandler {
	return &CustomerConnectHandler{useCase: uc}
}

func (h *CustomerConnectHandler) ListCustomers(ctx context.Context, req *connect.Request[devicepb.ListCustomersRequest]) (*connect.Response[devicepb.ListCustomersResponse], error) {
	customers, err := h.useCase.ListCustomers(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbCustomers := make([]*devicepb.Customer, len(customers))
	for i, c := range customers {
		pbCustomers[i] = &devicepb.Customer{
			Id:            c.ID,
			TenantId:      c.TenantID,
			Name:          c.Name,
			Email:         c.Email,
			Phone:         c.Phone,
			Address:       c.Address,
			Status:        c.Status,
			CreatedAtUnix: c.CreatedAt.Unix(),
		}
	}

	return connect.NewResponse(&devicepb.ListCustomersResponse{Customers: pbCustomers}), nil
}

func (h *CustomerConnectHandler) GetCustomer(ctx context.Context, req *connect.Request[devicepb.GetCustomerRequest]) (*connect.Response[devicepb.GetCustomerResponse], error) {
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("customer id is required"))
	}

	c, err := h.useCase.GetCustomer(ctx, req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewResponse(&devicepb.GetCustomerResponse{
		Customer: &devicepb.Customer{
			Id:            c.ID,
			TenantId:      c.TenantID,
			Name:          c.Name,
			Email:         c.Email,
			Phone:         c.Phone,
			Address:       c.Address,
			Status:        c.Status,
			CreatedAtUnix: c.CreatedAt.Unix(),
		},
	}), nil
}

func NewCustomerServiceHandler(uc *business.ManageCustomerUseCase) (string, http.Handler) {
	handler := NewCustomerConnectHandler(uc)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(connectJSONCodec{})

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
