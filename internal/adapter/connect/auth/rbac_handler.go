package auth

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

type RBACConnectHandler struct {
	enforcer port.RBACManager
}

// NewRBACConnectHandler constructs a new RBACConnectHandler.
func NewRBACConnectHandler(enforcer port.RBACManager) *RBACConnectHandler {
	return &RBACConnectHandler{enforcer: enforcer}
}

func (h *RBACConnectHandler) ListPolicies(ctx context.Context, req *connect.Request[polyglotv1.ListPoliciesRequest]) (*connect.Response[polyglotv1.ListPoliciesResponse], error) {
	if h.enforcer == nil {
		return connect.NewResponse(&polyglotv1.ListPoliciesResponse{Policies: []*polyglotv1.Policy{}}), nil
	}

	rawPolicies, err := h.enforcer.GetPolicies()
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	policies := make([]*polyglotv1.Policy, 0, len(rawPolicies))
	for _, p := range rawPolicies {
		if len(p) >= 3 {
			policies = append(policies, &polyglotv1.Policy{
				Sub: p[0],
				Obj: p[1],
				Act: p[2],
			})
		}
	}

	return connect.NewResponse(&polyglotv1.ListPoliciesResponse{Policies: policies}), nil
}

func (h *RBACConnectHandler) AddPolicy(ctx context.Context, req *connect.Request[polyglotv1.AddPolicyRequest]) (*connect.Response[polyglotv1.AddPolicyResponse], error) {
	if h.enforcer == nil {
		return nil, response.Unavailable("casbin enforcer unavailable")
	}

	ok, err := h.enforcer.AddPolicy(req.Msg.Policy.Sub, req.Msg.Policy.Obj, req.Msg.Policy.Act)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.AddPolicyResponse{Success: ok}), nil
}

func (h *RBACConnectHandler) RemovePolicy(ctx context.Context, req *connect.Request[polyglotv1.RemovePolicyRequest]) (*connect.Response[polyglotv1.RemovePolicyResponse], error) {
	if h.enforcer == nil {
		return nil, response.Unavailable("casbin enforcer unavailable")
	}

	ok, err := h.enforcer.RemovePolicy(req.Msg.Policy.Sub, req.Msg.Policy.Obj, req.Msg.Policy.Act)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.RemovePolicyResponse{Success: ok}), nil
}

func (h *RBACConnectHandler) ListRoleAssignments(ctx context.Context, req *connect.Request[polyglotv1.ListRoleAssignmentsRequest]) (*connect.Response[polyglotv1.ListRoleAssignmentsResponse], error) {
	if h.enforcer == nil {
		return connect.NewResponse(&polyglotv1.ListRoleAssignmentsResponse{RoleAssignments: []*polyglotv1.RoleAssignment{}}), nil
	}

	rawRoles, err := h.enforcer.GetGroupingPolicies()
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	assignments := make([]*polyglotv1.RoleAssignment, 0, len(rawRoles))
	for _, r := range rawRoles {
		if len(r) >= 2 {
			assignments = append(assignments, &polyglotv1.RoleAssignment{
				User: r[0],
				Role: r[1],
			})
		}
	}

	return connect.NewResponse(&polyglotv1.ListRoleAssignmentsResponse{RoleAssignments: assignments}), nil
}

func (h *RBACConnectHandler) AssignRole(ctx context.Context, req *connect.Request[polyglotv1.AssignRoleRequest]) (*connect.Response[polyglotv1.AssignRoleResponse], error) {
	if h.enforcer == nil {
		return nil, response.Unavailable("casbin enforcer unavailable")
	}

	ok, err := h.enforcer.AddRoleForUser(req.Msg.Assignment.User, req.Msg.Assignment.Role)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.AssignRoleResponse{Success: ok}), nil
}

func (h *RBACConnectHandler) UnassignRole(ctx context.Context, req *connect.Request[polyglotv1.UnassignRoleRequest]) (*connect.Response[polyglotv1.UnassignRoleResponse], error) {
	if h.enforcer == nil {
		return nil, response.Unavailable("casbin enforcer unavailable")
	}

	ok, err := h.enforcer.DeleteRoleForUser(req.Msg.Assignment.User, req.Msg.Assignment.Role)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.UnassignRoleResponse{Success: ok}), nil
}

func (h *RBACConnectHandler) SyncRolePermissions(ctx context.Context, req *connect.Request[polyglotv1.SyncRolePermissionsRequest]) (*connect.Response[polyglotv1.SyncRolePermissionsResponse], error) {
	if h.enforcer == nil {
		return nil, response.Unavailable("casbin enforcer unavailable")
	}

	if err := h.enforcer.SyncRolePermissions(req.Msg.Role, req.Msg.Permissions); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.SyncRolePermissionsResponse{
		Success: true,
		Message: fmt.Sprintf("permissions synced for role %s", req.Msg.Role),
	}), nil
}

func (h *RBACConnectHandler) DeleteRole(ctx context.Context, req *connect.Request[polyglotv1.DeleteRoleRequest]) (*connect.Response[polyglotv1.DeleteRoleResponse], error) {
	if req.Msg.Role == "owner" {
		return nil, response.PermissionDenied("role owner cannot be deleted")
	}

	if h.enforcer == nil {
		return nil, response.Unavailable("casbin enforcer unavailable")
	}

	if err := h.enforcer.DeleteRole(req.Msg.Role); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.DeleteRoleResponse{
		Success: true,
		Message: fmt.Sprintf("role %s deleted successfully", req.Msg.Role),
	}), nil
}

// NewRBACServiceHandler creates the Connect HTTP handler for RBACService.
func NewRBACServiceHandler(enforcer port.RBACManager) (string, http.Handler) {
	handler := NewRBACConnectHandler(enforcer)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.RBACService"
	mux.Handle("/"+serviceName+"/ListPolicies", connect.NewUnaryHandler("/"+serviceName+"/ListPolicies", handler.ListPolicies, opts...))
	mux.Handle("/"+serviceName+"/AddPolicy", connect.NewUnaryHandler("/"+serviceName+"/AddPolicy", handler.AddPolicy, opts...))
	mux.Handle("/"+serviceName+"/RemovePolicy", connect.NewUnaryHandler("/"+serviceName+"/RemovePolicy", handler.RemovePolicy, opts...))
	mux.Handle("/"+serviceName+"/ListRoleAssignments", connect.NewUnaryHandler("/"+serviceName+"/ListRoleAssignments", handler.ListRoleAssignments, opts...))
	mux.Handle("/"+serviceName+"/AssignRole", connect.NewUnaryHandler("/"+serviceName+"/AssignRole", handler.AssignRole, opts...))
	mux.Handle("/"+serviceName+"/UnassignRole", connect.NewUnaryHandler("/"+serviceName+"/UnassignRole", handler.UnassignRole, opts...))
	mux.Handle("/"+serviceName+"/SyncRolePermissions", connect.NewUnaryHandler("/"+serviceName+"/SyncRolePermissions", handler.SyncRolePermissions, opts...))
	mux.Handle("/"+serviceName+"/DeleteRole", connect.NewUnaryHandler("/"+serviceName+"/DeleteRole", handler.DeleteRole, opts...))

	return "/" + serviceName + "/", mux
}
