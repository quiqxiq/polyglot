package port

import "context"

// Enforcer decides whether a role may perform an action on a resource. The
// concrete implementation (Casbin, etc.) lives in an adapter; callers
// depend only on this contract.
type Enforcer interface {
	// Enforce reports whether role is permitted to perform action on
	// resource.
	Enforce(ctx context.Context, role, resource, action string) (bool, error)
}
