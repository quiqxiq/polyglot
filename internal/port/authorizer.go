package port

// Authorizer defines the interface for role-based access control evaluations.
// In Clean Architecture, domain and usecase layers depend on this interface,
// while internal/adapter/auth/casbin.go provides the concrete Casbin v3 implementation.
type Authorizer interface {
	Enforce(sub, obj, act string) (bool, error)
	AddPolicy(role, path, method string) (bool, error)
	RemovePolicy(role, path, method string) (bool, error)
	GetPolicies() ([][]string, error)
	AddRoleForUser(user, role string) (bool, error)
	DeleteRoleForUser(user, role string) (bool, error)
	GetRolesForUser(user string) ([]string, error)
	GetGroupingPolicies() ([][]string, error)
}
