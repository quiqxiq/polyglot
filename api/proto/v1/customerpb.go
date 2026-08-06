package devicepb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Customer struct {
	Id            string `json:"id"`
	TenantId      string `json:"tenant_id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	Address       string `json:"address"`
	Status        string `json:"status"`
	CreatedAtUnix int64  `json:"created_at_unix"`
}

type Subscription struct {
	Id            string  `json:"id"`
	CustomerId    string  `json:"customer_id"`
	PlanId        string  `json:"plan_id"`
	Status        string  `json:"status"`
	StartDateUnix int64   `json:"start_date_unix"`
	EndDateUnix   int64   `json:"end_date_unix"`
	Price         float64 `json:"price"`
}

type ListCustomersRequest struct{}

type ListCustomersResponse struct {
	Customers []*Customer `json:"customers"`
}

type GetCustomerRequest struct {
	Id string `json:"id"`
}

type GetCustomerResponse struct {
	Customer *Customer `json:"customer"`
}

type CreateCustomerRequest struct {
	Customer *Customer `json:"customer"`
}

type CreateCustomerResponse struct {
	Customer *Customer `json:"customer"`
	Message  string    `json:"message"`
}

type UpdateCustomerRequest struct {
	Customer *Customer `json:"customer"`
}

type UpdateCustomerResponse struct {
	Customer *Customer `json:"customer"`
	Message  string    `json:"message"`
}

type DeleteCustomerRequest struct {
	Id string `json:"id"`
}

type DeleteCustomerResponse struct {
	Message string `json:"message"`
}

type ListSubscriptionsRequest struct {
	CustomerId string `json:"customer_id"`
}

type ListSubscriptionsResponse struct {
	Subscriptions []*Subscription `json:"subscriptions"`
}

type CustomerServiceServer interface {
	ListCustomers(context.Context, *ListCustomersRequest) (*ListCustomersResponse, error)
	GetCustomer(context.Context, *GetCustomerRequest) (*GetCustomerResponse, error)
	CreateCustomer(context.Context, *CreateCustomerRequest) (*CreateCustomerResponse, error)
	UpdateCustomer(context.Context, *UpdateCustomerRequest) (*UpdateCustomerResponse, error)
	DeleteCustomer(context.Context, *DeleteCustomerRequest) (*DeleteCustomerResponse, error)
	ListSubscriptions(context.Context, *ListSubscriptionsRequest) (*ListSubscriptionsResponse, error)
}

type UnimplementedCustomerServiceServer struct{}

func (UnimplementedCustomerServiceServer) ListCustomers(context.Context, *ListCustomersRequest) (*ListCustomersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListCustomers not implemented")
}

func (UnimplementedCustomerServiceServer) GetCustomer(context.Context, *GetCustomerRequest) (*GetCustomerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCustomer not implemented")
}

func (UnimplementedCustomerServiceServer) CreateCustomer(context.Context, *CreateCustomerRequest) (*CreateCustomerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateCustomer not implemented")
}

func (UnimplementedCustomerServiceServer) UpdateCustomer(context.Context, *UpdateCustomerRequest) (*UpdateCustomerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateCustomer not implemented")
}

func (UnimplementedCustomerServiceServer) DeleteCustomer(context.Context, *DeleteCustomerRequest) (*DeleteCustomerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteCustomer not implemented")
}

func (UnimplementedCustomerServiceServer) ListSubscriptions(context.Context, *ListSubscriptionsRequest) (*ListSubscriptionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListSubscriptions not implemented")
}

func RegisterCustomerServiceServer(s *grpc.Server, srv CustomerServiceServer) {
	s.RegisterService(&CustomerService_ServiceDesc, srv)
}

var CustomerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "polyglot.v1.CustomerService",
	HandlerType: (*CustomerServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "api/proto/v1/customer.proto",
}
