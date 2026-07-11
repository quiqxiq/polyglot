package auth

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v3"
)

// RBACEnforcer defines the interface for Role-Based Access Control.
type RBACEnforcer interface {
	Enforce(ctx context.Context, role, resource, action string) (bool, error)
}

type casbinEnforcer struct {
	enforcer *casbin.Enforcer
}

// NewRBAC builds the Casbin enforcer using the provided model and policy files.
func NewRBAC(modelPath, policyPath string) (RBACEnforcer, error) {
	e, err := casbin.NewEnforcer(modelPath, policyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	// Load the policy from file
	if err := e.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load casbin policy: %w", err)
	}

	return &casbinEnforcer{
		enforcer: e,
	}, nil
}

// Enforce checks if the given role is allowed to perform the action on the resource.
func (c *casbinEnforcer) Enforce(ctx context.Context, role, resource, action string) (bool, error) {
	allowed, err := c.enforcer.Enforce(role, resource, action)
	if err != nil {
		return false, fmt.Errorf("rbac enforce error: %w", err)
	}
	return allowed, nil
}
