package devicepb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Policy struct {
	Sub string `json:"sub"`
	Obj string `json:"obj"`
	Act string `json:"act"`
}

type RoleAssignment struct {
	User string `json:"user"`
	Role string `json:"role"`
}

type ListPoliciesRequest struct{}
type ListPoliciesResponse struct {
	Policies []*Policy `json:"policies"`
}

type AddPolicyRequest struct {
	Policy *Policy `json:"policy"`
}
type AddPolicyResponse struct {
	Success bool `json:"success"`
}

type RemovePolicyRequest struct {
	Policy *Policy `json:"policy"`
}
type RemovePolicyResponse struct {
	Success bool `json:"success"`
}

type ListRoleAssignmentsRequest struct{}
type ListRoleAssignmentsResponse struct {
	RoleAssignments []*RoleAssignment `json:"role_assignments"`
}

type AssignRoleRequest struct {
	Assignment *RoleAssignment `json:"assignment"`
}
type AssignRoleResponse struct {
	Success bool `json:"success"`
}

type UnassignRoleRequest struct {
	Assignment *RoleAssignment `json:"assignment"`
}
type UnassignRoleResponse struct {
	Success bool `json:"success"`
}

type RBACServiceClient interface {
	ListPolicies(ctx context.Context, in *ListPoliciesRequest, opts ...grpc.CallOption) (*ListPoliciesResponse, error)
	AddPolicy(ctx context.Context, in *AddPolicyRequest, opts ...grpc.CallOption) (*AddPolicyResponse, error)
	RemovePolicy(ctx context.Context, in *RemovePolicyRequest, opts ...grpc.CallOption) (*RemovePolicyResponse, error)
	ListRoleAssignments(ctx context.Context, in *ListRoleAssignmentsRequest, opts ...grpc.CallOption) (*ListRoleAssignmentsResponse, error)
	AssignRole(ctx context.Context, in *AssignRoleRequest, opts ...grpc.CallOption) (*AssignRoleResponse, error)
	UnassignRole(ctx context.Context, in *UnassignRoleRequest, opts ...grpc.CallOption) (*UnassignRoleResponse, error)
}

type rbacServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRBACServiceClient(cc grpc.ClientConnInterface) RBACServiceClient {
	return &rbacServiceClient{cc}
}

func (c *rbacServiceClient) ListPolicies(ctx context.Context, in *ListPoliciesRequest, opts ...grpc.CallOption) (*ListPoliciesResponse, error) {
	out := new(ListPoliciesResponse)
	err := c.cc.Invoke(ctx, "/polyglot.v1.RBACService/ListPolicies", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rbacServiceClient) AddPolicy(ctx context.Context, in *AddPolicyRequest, opts ...grpc.CallOption) (*AddPolicyResponse, error) {
	out := new(AddPolicyResponse)
	err := c.cc.Invoke(ctx, "/polyglot.v1.RBACService/AddPolicy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rbacServiceClient) RemovePolicy(ctx context.Context, in *RemovePolicyRequest, opts ...grpc.CallOption) (*RemovePolicyResponse, error) {
	out := new(RemovePolicyResponse)
	err := c.cc.Invoke(ctx, "/polyglot.v1.RBACService/RemovePolicy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rbacServiceClient) ListRoleAssignments(ctx context.Context, in *ListRoleAssignmentsRequest, opts ...grpc.CallOption) (*ListRoleAssignmentsResponse, error) {
	out := new(ListRoleAssignmentsResponse)
	err := c.cc.Invoke(ctx, "/polyglot.v1.RBACService/ListRoleAssignments", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rbacServiceClient) AssignRole(ctx context.Context, in *AssignRoleRequest, opts ...grpc.CallOption) (*AssignRoleResponse, error) {
	out := new(AssignRoleResponse)
	err := c.cc.Invoke(ctx, "/polyglot.v1.RBACService/AssignRole", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rbacServiceClient) UnassignRole(ctx context.Context, in *UnassignRoleRequest, opts ...grpc.CallOption) (*UnassignRoleResponse, error) {
	out := new(UnassignRoleResponse)
	err := c.cc.Invoke(ctx, "/polyglot.v1.RBACService/UnassignRole", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type RBACServiceServer interface {
	ListPolicies(context.Context, *ListPoliciesRequest) (*ListPoliciesResponse, error)
	AddPolicy(context.Context, *AddPolicyRequest) (*AddPolicyResponse, error)
	RemovePolicy(context.Context, *RemovePolicyRequest) (*RemovePolicyResponse, error)
	ListRoleAssignments(context.Context, *ListRoleAssignmentsRequest) (*ListRoleAssignmentsResponse, error)
	AssignRole(context.Context, *AssignRoleRequest) (*AssignRoleResponse, error)
	UnassignRole(context.Context, *UnassignRoleRequest) (*UnassignRoleResponse, error)
	mustEmbedUnimplementedRBACServiceServer()
}

type UnimplementedRBACServiceServer struct{}

func (UnimplementedRBACServiceServer) ListPolicies(context.Context, *ListPoliciesRequest) (*ListPoliciesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListPolicies not implemented")
}
func (UnimplementedRBACServiceServer) AddPolicy(context.Context, *AddPolicyRequest) (*AddPolicyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddPolicy not implemented")
}
func (UnimplementedRBACServiceServer) RemovePolicy(context.Context, *RemovePolicyRequest) (*RemovePolicyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemovePolicy not implemented")
}
func (UnimplementedRBACServiceServer) ListRoleAssignments(context.Context, *ListRoleAssignmentsRequest) (*ListRoleAssignmentsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListRoleAssignments not implemented")
}
func (UnimplementedRBACServiceServer) AssignRole(context.Context, *AssignRoleRequest) (*AssignRoleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AssignRole not implemented")
}
func (UnimplementedRBACServiceServer) UnassignRole(context.Context, *UnassignRoleRequest) (*UnassignRoleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnassignRole not implemented")
}
func (UnimplementedRBACServiceServer) mustEmbedUnimplementedRBACServiceServer() {}

func RegisterRBACServiceServer(s grpc.ServiceRegistrar, srv RBACServiceServer) {
	s.RegisterService(&RBACService_ServiceDesc, srv)
}

var RBACService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "polyglot.v1.RBACService",
	HandlerType: (*RBACServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "ListPolicies", Handler: _RBACService_ListPolicies_Handler},
		{MethodName: "AddPolicy", Handler: _RBACService_AddPolicy_Handler},
		{MethodName: "RemovePolicy", Handler: _RBACService_RemovePolicy_Handler},
		{MethodName: "ListRoleAssignments", Handler: _RBACService_ListRoleAssignments_Handler},
		{MethodName: "AssignRole", Handler: _RBACService_AssignRole_Handler},
		{MethodName: "UnassignRole", Handler: _RBACService_UnassignRole_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/proto/v1/rbac.proto",
}

func _RBACService_ListPolicies_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListPoliciesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RBACServiceServer).ListPolicies(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/polyglot.v1.RBACService/ListPolicies"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RBACServiceServer).ListPolicies(ctx, req.(*ListPoliciesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RBACService_AddPolicy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddPolicyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RBACServiceServer).AddPolicy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/polyglot.v1.RBACService/AddPolicy"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RBACServiceServer).AddPolicy(ctx, req.(*AddPolicyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RBACService_RemovePolicy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemovePolicyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RBACServiceServer).RemovePolicy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/polyglot.v1.RBACService/RemovePolicy"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RBACServiceServer).RemovePolicy(ctx, req.(*RemovePolicyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RBACService_ListRoleAssignments_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRoleAssignmentsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RBACServiceServer).ListRoleAssignments(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/polyglot.v1.RBACService/ListRoleAssignments"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RBACServiceServer).ListRoleAssignments(ctx, req.(*ListRoleAssignmentsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RBACService_AssignRole_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AssignRoleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RBACServiceServer).AssignRole(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/polyglot.v1.RBACService/AssignRole"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RBACServiceServer).AssignRole(ctx, req.(*AssignRoleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RBACService_UnassignRole_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UnassignRoleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RBACServiceServer).UnassignRole(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/polyglot.v1.RBACService/UnassignRole"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RBACServiceServer).UnassignRole(ctx, req.(*UnassignRoleRequest))
	}
	return interceptor(ctx, in, info, handler)
}
