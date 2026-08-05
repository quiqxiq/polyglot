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

// Compile-time assertion that *CasbinEnforcer implements port.Authorizer
var _ port.Authorizer = (*CasbinEnforcer)(nil)

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
	return ce.enforcer.Enforce(sub, obj, act)
}

// AddPolicy dynamically adds a new policy rule.
func (ce *CasbinEnforcer) AddPolicy(role, path, method string) (bool, error) {
	ok, err := ce.enforcer.AddPolicy(role, path, method)
	if err == nil && ok {
		_ = ce.enforcer.SavePolicy()
	}
	return ok, err
}

// RemovePolicy dynamically removes an existing policy rule.
func (ce *CasbinEnforcer) RemovePolicy(role, path, method string) (bool, error) {
	ok, err := ce.enforcer.RemovePolicy(role, path, method)
	if err == nil && ok {
		_ = ce.enforcer.SavePolicy()
	}
	return ok, err
}

// GetPolicies retrieves all policy rules.
func (ce *CasbinEnforcer) GetPolicies() ([][]string, error) {
	return ce.enforcer.GetPolicy()
}

// AddRoleForUser assigns a role to a user.
func (ce *CasbinEnforcer) AddRoleForUser(user, role string) (bool, error) {
	ok, err := ce.enforcer.AddRoleForUser(user, role)
	if err == nil && ok {
		_ = ce.enforcer.SavePolicy()
	}
	return ok, err
}

// DeleteRoleForUser revokes a role from a user.
func (ce *CasbinEnforcer) DeleteRoleForUser(user, role string) (bool, error) {
	ok, err := ce.enforcer.DeleteRoleForUser(user, role)
	if err == nil && ok {
		_ = ce.enforcer.SavePolicy()
	}
	return ok, err
}

// GetRolesForUser returns all roles assigned to a user.
func (ce *CasbinEnforcer) GetRolesForUser(user string) ([]string, error) {
	return ce.enforcer.GetRolesForUser(user)
}

// GetGroupingPolicies returns all user-role mappings.
func (ce *CasbinEnforcer) GetGroupingPolicies() ([][]string, error) {
	return ce.enforcer.GetGroupingPolicy()
}
