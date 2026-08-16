package auth

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/adapter/auth"
	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/pkg/response"
)

type RBACConnectHandler struct {
	enforcer *auth.CasbinEnforcer
}

func NewRBACConnectHandler(enforcer *auth.CasbinEnforcer) *RBACConnectHandler {
	return &RBACConnectHandler{enforcer: enforcer}
}

func (h *RBACConnectHandler) ListPolicies(ctx context.Context, req *connect.Request[devicepb.ListPoliciesRequest]) (*connect.Response[devicepb.ListPoliciesResponse], error) {
	if h.enforcer == nil {
		return connect.NewResponse(&devicepb.ListPoliciesResponse{Policies: []*devicepb.Policy{}}), nil
	}

	rawPolicies, err := h.enforcer.GetPolicies()
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	policies := make([]*devicepb.Policy, 0, len(rawPolicies))
	for _, p := range rawPolicies {
		if len(p) >= 3 {
			policies = append(policies, &devicepb.Policy{
				Sub: p[0],
				Obj: p[1],
				Act: p[2],
			})
		}
	}

	return connect.NewResponse(&devicepb.ListPoliciesResponse{Policies: policies}), nil
}

func (h *RBACConnectHandler) AddPolicy(ctx context.Context, req *connect.Request[devicepb.AddPolicyRequest]) (*connect.Response[devicepb.AddPolicyResponse], error) {
	if req.Msg.Policy == nil || req.Msg.Policy.Sub == "" || req.Msg.Policy.Obj == "" || req.Msg.Policy.Act == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("sub, obj, and act are required"))
	}

	if h.enforcer == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("casbin enforcer unavailable"))
	}

	ok, err := h.enforcer.AddPolicy(req.Msg.Policy.Sub, req.Msg.Policy.Obj, req.Msg.Policy.Act)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.AddPolicyResponse{Success: ok}), nil
}

func (h *RBACConnectHandler) RemovePolicy(ctx context.Context, req *connect.Request[devicepb.RemovePolicyRequest]) (*connect.Response[devicepb.RemovePolicyResponse], error) {
	if req.Msg.Policy == nil || req.Msg.Policy.Sub == "" || req.Msg.Policy.Obj == "" || req.Msg.Policy.Act == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("sub, obj, and act are required"))
	}

	if h.enforcer == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("casbin enforcer unavailable"))
	}

	ok, err := h.enforcer.RemovePolicy(req.Msg.Policy.Sub, req.Msg.Policy.Obj, req.Msg.Policy.Act)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.RemovePolicyResponse{Success: ok}), nil
}

func (h *RBACConnectHandler) ListRoleAssignments(ctx context.Context, req *connect.Request[devicepb.ListRoleAssignmentsRequest]) (*connect.Response[devicepb.ListRoleAssignmentsResponse], error) {
	if h.enforcer == nil {
		return connect.NewResponse(&devicepb.ListRoleAssignmentsResponse{RoleAssignments: []*devicepb.RoleAssignment{}}), nil
	}

	rawRoles, err := h.enforcer.GetGroupingPolicies()
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	assignments := make([]*devicepb.RoleAssignment, 0, len(rawRoles))
	for _, r := range rawRoles {
		if len(r) >= 2 {
			assignments = append(assignments, &devicepb.RoleAssignment{
				User: r[0],
				Role: r[1],
			})
		}
	}

	return connect.NewResponse(&devicepb.ListRoleAssignmentsResponse{RoleAssignments: assignments}), nil
}

func (h *RBACConnectHandler) AssignRole(ctx context.Context, req *connect.Request[devicepb.AssignRoleRequest]) (*connect.Response[devicepb.AssignRoleResponse], error) {
	if req.Msg.Assignment == nil || req.Msg.Assignment.User == "" || req.Msg.Assignment.Role == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user and role are required"))
	}

	if h.enforcer == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("casbin enforcer unavailable"))
	}

	ok, err := h.enforcer.AddRoleForUser(req.Msg.Assignment.User, req.Msg.Assignment.Role)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.AssignRoleResponse{Success: ok}), nil
}

func (h *RBACConnectHandler) UnassignRole(ctx context.Context, req *connect.Request[devicepb.UnassignRoleRequest]) (*connect.Response[devicepb.UnassignRoleResponse], error) {
	if req.Msg.Assignment == nil || req.Msg.Assignment.User == "" || req.Msg.Assignment.Role == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user and role are required"))
	}

	if h.enforcer == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("casbin enforcer unavailable"))
	}

	ok, err := h.enforcer.DeleteRoleForUser(req.Msg.Assignment.User, req.Msg.Assignment.Role)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.UnassignRoleResponse{Success: ok}), nil
}

func NewRBACServiceHandler(enforcer *auth.CasbinEnforcer) (string, http.Handler) {
	handler := NewRBACConnectHandler(enforcer)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.RBACService"
	mux.Handle("/"+serviceName+"/ListPolicies", connect.NewUnaryHandler("/"+serviceName+"/ListPolicies", handler.ListPolicies, codecOpt))
	mux.Handle("/"+serviceName+"/AddPolicy", connect.NewUnaryHandler("/"+serviceName+"/AddPolicy", handler.AddPolicy, codecOpt))
	mux.Handle("/"+serviceName+"/RemovePolicy", connect.NewUnaryHandler("/"+serviceName+"/RemovePolicy", handler.RemovePolicy, codecOpt))
	mux.Handle("/"+serviceName+"/ListRoleAssignments", connect.NewUnaryHandler("/"+serviceName+"/ListRoleAssignments", handler.ListRoleAssignments, codecOpt))
	mux.Handle("/"+serviceName+"/AssignRole", connect.NewUnaryHandler("/"+serviceName+"/AssignRole", handler.AssignRole, codecOpt))
	mux.Handle("/"+serviceName+"/UnassignRole", connect.NewUnaryHandler("/"+serviceName+"/UnassignRole", handler.UnassignRole, codecOpt))

	return "/" + serviceName + "/", mux
}
