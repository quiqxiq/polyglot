package devicepb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Invoice struct {
	Id            string  `json:"id"`
	CustomerId    string  `json:"customer_id"`
	Amount        float64 `json:"amount"`
	Status        string  `json:"status"`
	DueDateUnix   int64   `json:"due_date_unix"`
	CreatedAtUnix int64   `json:"created_at_unix"`
}

type ListInvoicesRequest struct {
	CustomerId string `json:"customer_id"`
}
type ListInvoicesResponse struct {
	Invoices []*Invoice `json:"invoices"`
}

type GetInvoiceRequest struct {
	Id string `json:"id"`
}
type GetInvoiceResponse struct {
	Invoice *Invoice `json:"invoice"`
}

type CreateInvoiceRequest struct {
	CustomerId  string  `json:"customer_id"`
	Amount      float64 `json:"amount"`
	DueDateUnix int64   `json:"due_date_unix"`
}
type CreateInvoiceResponse struct {
	Invoice *Invoice `json:"invoice"`
}

type PayInvoiceRequest struct {
	Id            string `json:"id"`
	PaymentMethod string `json:"payment_method"`
}
type PayInvoiceResponse struct {
	Invoice *Invoice `json:"invoice"`
	Message string   `json:"message"`
}

type CreateSubscriptionRequest struct {
	CustomerId string  `json:"customer_id"`
	PlanId     string  `json:"plan_id"`
	Price      float64 `json:"price"`
}
type CreateSubscriptionResponse struct {
	Subscription *Subscription `json:"subscription"`
}

type CancelSubscriptionRequest struct {
	Id string `json:"id"`
}
type CancelSubscriptionResponse struct {
	Subscription *Subscription `json:"subscription"`
	Message      string        `json:"message"`
}

type BillingServiceClient interface {
	ListInvoices(ctx context.Context, in *ListInvoicesRequest, opts ...grpc.CallOption) (*ListInvoicesResponse, error)
	GetInvoice(ctx context.Context, in *GetInvoiceRequest, opts ...grpc.CallOption) (*GetInvoiceResponse, error)
	CreateInvoice(ctx context.Context, in *CreateInvoiceRequest, opts ...grpc.CallOption) (*CreateInvoiceResponse, error)
	PayInvoice(ctx context.Context, in *PayInvoiceRequest, opts ...grpc.CallOption) (*PayInvoiceResponse, error)
	ListSubscriptions(ctx context.Context, in *ListSubscriptionsRequest, opts ...grpc.CallOption) (*ListSubscriptionsResponse, error)
	CreateSubscription(ctx context.Context, in *CreateSubscriptionRequest, opts ...grpc.CallOption) (*CreateSubscriptionResponse, error)
	CancelSubscription(ctx context.Context, in *CancelSubscriptionRequest, opts ...grpc.CallOption) (*CancelSubscriptionResponse, error)
}

type billingServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBillingServiceClient(cc grpc.ClientConnInterface) BillingServiceClient {
	return &billingServiceClient{cc}
}

func (c *billingServiceClient) ListInvoices(ctx context.Context, in *ListInvoicesRequest, opts ...grpc.CallOption) (*ListInvoicesResponse, error) {
	out := new(ListInvoicesResponse)
	err := c.cc.Invoke(ctx, "/polyglot.v1.BillingService/ListInvoices", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *billingServiceClient) GetInvoice(ctx context.Context, in *GetInvoiceRequest, opts ...grpc.CallOption) (*GetInvoiceResponse, error) {
	out := new(GetInvoiceResponse)
	err := c.cc.Invoke(ctx, "/polyglot.v1.BillingService/GetInvoice", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *billingServiceClient) CreateInvoice(ctx context.Context, in *CreateInvoiceRequest, opts ...grpc.CallOption) (*CreateInvoiceResponse, error) {
	out := new(CreateInvoiceResponse)
	err := c.cc.Invoke(ctx, "/polyglot.v1.BillingService/CreateInvoice", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *billingServiceClient) PayInvoice(ctx context.Context, in *PayInvoiceRequest, opts ...grpc.CallOption) (*PayInvoiceResponse, error) {
	out := new(PayInvoiceResponse)
	err := c.cc.Invoke(ctx, "/polyglot.v1.BillingService/PayInvoice", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *billingServiceClient) ListSubscriptions(ctx context.Context, in *ListSubscriptionsRequest, opts ...grpc.CallOption) (*ListSubscriptionsResponse, error) {
	out := new(ListSubscriptionsResponse)
	err := c.cc.Invoke(ctx, "/polyglot.v1.BillingService/ListSubscriptions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *billingServiceClient) CreateSubscription(ctx context.Context, in *CreateSubscriptionRequest, opts ...grpc.CallOption) (*CreateSubscriptionResponse, error) {
	out := new(CreateSubscriptionResponse)
	err := c.cc.Invoke(ctx, "/polyglot.v1.BillingService/CreateSubscription", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *billingServiceClient) CancelSubscription(ctx context.Context, in *CancelSubscriptionRequest, opts ...grpc.CallOption) (*CancelSubscriptionResponse, error) {
	out := new(CancelSubscriptionResponse)
	err := c.cc.Invoke(ctx, "/polyglot.v1.BillingService/CancelSubscription", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type BillingServiceServer interface {
	ListInvoices(context.Context, *ListInvoicesRequest) (*ListInvoicesResponse, error)
	GetInvoice(context.Context, *GetInvoiceRequest) (*GetInvoiceResponse, error)
	CreateInvoice(context.Context, *CreateInvoiceRequest) (*CreateInvoiceResponse, error)
	PayInvoice(context.Context, *PayInvoiceRequest) (*PayInvoiceResponse, error)
	ListSubscriptions(context.Context, *ListSubscriptionsRequest) (*ListSubscriptionsResponse, error)
	CreateSubscription(context.Context, *CreateSubscriptionRequest) (*CreateSubscriptionResponse, error)
	CancelSubscription(context.Context, *CancelSubscriptionRequest) (*CancelSubscriptionResponse, error)
	mustEmbedUnimplementedBillingServiceServer()
}

type UnimplementedBillingServiceServer struct{}

func (UnimplementedBillingServiceServer) ListInvoices(context.Context, *ListInvoicesRequest) (*ListInvoicesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListInvoices not implemented")
}
func (UnimplementedBillingServiceServer) GetInvoice(context.Context, *GetInvoiceRequest) (*GetInvoiceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInvoice not implemented")
}
func (UnimplementedBillingServiceServer) CreateInvoice(context.Context, *CreateInvoiceRequest) (*CreateInvoiceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateInvoice not implemented")
}
func (UnimplementedBillingServiceServer) PayInvoice(context.Context, *PayInvoiceRequest) (*PayInvoiceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PayInvoice not implemented")
}
func (UnimplementedBillingServiceServer) ListSubscriptions(context.Context, *ListSubscriptionsRequest) (*ListSubscriptionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListSubscriptions not implemented")
}
func (UnimplementedBillingServiceServer) CreateSubscription(context.Context, *CreateSubscriptionRequest) (*CreateSubscriptionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSubscription not implemented")
}
func (UnimplementedBillingServiceServer) CancelSubscription(context.Context, *CancelSubscriptionRequest) (*CancelSubscriptionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelSubscription not implemented")
}
func (UnimplementedBillingServiceServer) mustEmbedUnimplementedBillingServiceServer() {}

func RegisterBillingServiceServer(s grpc.ServiceRegistrar, srv BillingServiceServer) {
	s.RegisterService(&BillingService_ServiceDesc, srv)
}

var BillingService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "polyglot.v1.BillingService",
	HandlerType: (*BillingServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "ListInvoices", Handler: _BillingService_ListInvoices_Handler},
		{MethodName: "GetInvoice", Handler: _BillingService_GetInvoice_Handler},
		{MethodName: "CreateInvoice", Handler: _BillingService_CreateInvoice_Handler},
		{MethodName: "PayInvoice", Handler: _BillingService_PayInvoice_Handler},
		{MethodName: "ListSubscriptions", Handler: _BillingService_ListSubscriptions_Handler},
		{MethodName: "CreateSubscription", Handler: _BillingService_CreateSubscription_Handler},
		{MethodName: "CancelSubscription", Handler: _BillingService_CancelSubscription_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/proto/v1/billing.proto",
}

func _BillingService_ListInvoices_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListInvoicesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BillingServiceServer).ListInvoices(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/polyglot.v1.BillingService/ListInvoices"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BillingServiceServer).ListInvoices(ctx, req.(*ListInvoicesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BillingService_GetInvoice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetInvoiceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BillingServiceServer).GetInvoice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/polyglot.v1.BillingService/GetInvoice"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BillingServiceServer).GetInvoice(ctx, req.(*GetInvoiceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BillingService_CreateInvoice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateInvoiceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BillingServiceServer).CreateInvoice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/polyglot.v1.BillingService/CreateInvoice"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BillingServiceServer).CreateInvoice(ctx, req.(*CreateInvoiceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BillingService_PayInvoice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PayInvoiceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BillingServiceServer).PayInvoice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/polyglot.v1.BillingService/PayInvoice"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BillingServiceServer).PayInvoice(ctx, req.(*PayInvoiceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BillingService_ListSubscriptions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListSubscriptionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BillingServiceServer).ListSubscriptions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/polyglot.v1.BillingService/ListSubscriptions"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BillingServiceServer).ListSubscriptions(ctx, req.(*ListSubscriptionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BillingService_CreateSubscription_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateSubscriptionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BillingServiceServer).CreateSubscription(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/polyglot.v1.BillingService/CreateSubscription"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BillingServiceServer).CreateSubscription(ctx, req.(*CreateSubscriptionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BillingService_CancelSubscription_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CancelSubscriptionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BillingServiceServer).CancelSubscription(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/polyglot.v1.BillingService/CancelSubscription"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BillingServiceServer).CancelSubscription(ctx, req.(*CancelSubscriptionRequest))
	}
	return interceptor(ctx, in, info, handler)
}
