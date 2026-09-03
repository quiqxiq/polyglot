package auth

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/port"
)

// CasbinEnforcer manages dynamic role-based access control policies using Casbin v3.
type CasbinEnforcer struct {
	enforcer *casbin.Enforcer
}

// Compile-time assertion that *CasbinEnforcer implements port.Authorizer, port.RoleAuthorizer, and port.RBACManager
var (
	_ port.Authorizer     = (*CasbinEnforcer)(nil)
	_ port.RoleAuthorizer = (*CasbinEnforcer)(nil)
	_ port.RBACManager    = (*CasbinEnforcer)(nil)
)

// NewCasbinEnforcer initializes the Casbin Enforcer with GORM PostgreSQL storage.
func NewCasbinEnforcer(ctx context.Context, db *gorm.DB) (*CasbinEnforcer, error) {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin gorm adapter: %w", err)
	}

	m, err := model.NewModelFromString(config.RBACModelConf)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin rbac model: %w", err)
	}

	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize casbin enforcer: %w", err)
	}

	if err := e.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load casbin policies: %w", err)
	}

	ce := &CasbinEnforcer{enforcer: e}
	return ce, nil
}

// Enforce evaluates if a (subject, object, action) request is allowed.
func (ce *CasbinEnforcer) Enforce(sub, obj, act string) (bool, error) {
	if ce == nil || ce.enforcer == nil {
		return false, nil
	}
	return ce.enforcer.Enforce(sub, obj, act)
}

// AddPolicy dynamically adds a new policy rule.
func (ce *CasbinEnforcer) AddPolicy(role, obj, act string) (bool, error) {
	if ce == nil || ce.enforcer == nil {
		return false, fmt.Errorf("enforcer unavailable")
	}
	ok, err := ce.enforcer.AddPolicy(role, obj, act)
	if err == nil && ok {
		ce.persist()
	}
	return ok, err
}

// RemovePolicy dynamically removes an existing policy rule.
func (ce *CasbinEnforcer) RemovePolicy(role, obj, act string) (bool, error) {
	if ce == nil || ce.enforcer == nil {
		return false, fmt.Errorf("enforcer unavailable")
	}
	ok, err := ce.enforcer.RemovePolicy(role, obj, act)
	if err == nil && ok {
		ce.persist()
	}
	return ok, err
}

// GetPolicies retrieves all policy rules.
func (ce *CasbinEnforcer) GetPolicies() ([][]string, error) {
	if ce == nil || ce.enforcer == nil {
		return nil, nil
	}
	return ce.enforcer.GetPolicy()
}

// GetGroupingPolicies retrieves all role assignments (g policies).
func (ce *CasbinEnforcer) GetGroupingPolicies() ([][]string, error) {
	if ce == nil || ce.enforcer == nil {
		return nil, nil
	}
	return ce.enforcer.GetGroupingPolicy()
}

// SyncRolePermissions removes all existing permissions for a role and adds the provided list atomically.
func (ce *CasbinEnforcer) SyncRolePermissions(role string, permissions []string) error {
	if ce == nil || ce.enforcer == nil {
		return fmt.Errorf("enforcer unavailable")
	}
	if role == "" {
		return fmt.Errorf("role cannot be empty")
	}

	// Remove all previous policies for this role (field 0 is sub/role)
	if _, err := ce.enforcer.RemoveFilteredPolicy(0, role); err != nil {
		return fmt.Errorf("failed to clear old policies for role %s: %w", role, err)
	}

	// Add new policies
	rules := make([][]string, 0, len(permissions))
	for _, perm := range permissions {
		if perm != "" {
			rules = append(rules, []string{role, perm, "*"})
		}
	}
	if len(rules) > 0 {
		if _, err := ce.enforcer.AddPolicies(rules); err != nil {
			return fmt.Errorf("failed to add new policies for role %s: %w", role, err)
		}
	}

	ce.persist()
	return nil
}

// DeleteRole removes all policies and user assignments for the specified role.
func (ce *CasbinEnforcer) DeleteRole(role string) error {
	if ce == nil || ce.enforcer == nil {
		return fmt.Errorf("enforcer unavailable")
	}
	if role == "" || role == "owner" {
		return fmt.Errorf("cannot delete role %q", role)
	}

	if _, err := ce.enforcer.RemoveFilteredPolicy(0, role); err != nil {
		return fmt.Errorf("failed to remove policies for role %s: %w", role, err)
	}
	if _, err := ce.enforcer.RemoveFilteredGroupingPolicy(1, role); err != nil {
		return fmt.Errorf("failed to remove role assignments for role %s: %w", role, err)
	}

	ce.persist()
	return nil
}

// AddRoleForUser assigns a role to a user.
func (ce *CasbinEnforcer) AddRoleForUser(user, role string) (bool, error) {
	if ce == nil || ce.enforcer == nil {
		return false, fmt.Errorf("enforcer unavailable")
	}
	ok, err := ce.enforcer.AddRoleForUser(user, role)
	if err == nil && ok {
		ce.persist()
	}
	return ok, err
}

// DeleteRoleForUser revokes a role from a user.
func (ce *CasbinEnforcer) DeleteRoleForUser(user, role string) (bool, error) {
	if ce == nil || ce.enforcer == nil {
		return false, fmt.Errorf("enforcer unavailable")
	}
	ok, err := ce.enforcer.DeleteRoleForUser(user, role)
	if err == nil && ok {
		ce.persist()
	}
	return ok, err
}

// DeleteRolesForUser removes ALL role assignments of a user.
func (ce *CasbinEnforcer) DeleteRolesForUser(user string) (bool, error) {
	if ce == nil || ce.enforcer == nil {
		return false, fmt.Errorf("enforcer unavailable")
	}
	ok, err := ce.enforcer.DeleteRolesForUser(user)
	if err == nil && ok {
		ce.persist()
	}
	return ok, err
}

// GetRolesForUser retrieves all roles assigned to a user.
func (ce *CasbinEnforcer) GetRolesForUser(user string) ([]string, error) {
	if ce == nil || ce.enforcer == nil {
		return nil, nil
	}
	return ce.enforcer.GetRolesForUser(user)
}

// GetImplicitPermissionsForUser returns all permissions effective for the user.
func (ce *CasbinEnforcer) GetImplicitPermissionsForUser(user string) ([]string, error) {
	if ce == nil || ce.enforcer == nil {
		return nil, nil
	}
	rules, err := ce.enforcer.GetImplicitPermissionsForUser(user)
	if err != nil {
		return nil, err
	}
	perms := make([]string, 0, len(rules))
	for _, r := range rules {
		if len(r) >= 3 {
			perms = append(perms, r[1]+":"+r[2])
		}
	}
	return perms, nil
}

// persist safely saves changes to the persistent database store if an adapter exists.
func (ce *CasbinEnforcer) persist() {
	if ce != nil && ce.enforcer != nil && ce.enforcer.GetAdapter() != nil {
		// best-effort: kegagalan persist policy ke DB tidak mengubah hasil operasi di memori.
		_ = ce.enforcer.SavePolicy()
	}
}
