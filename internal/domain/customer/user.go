package customer

import (
	"strings"
	"time"
)

// BuiltInRoles are the roles managed by the system policy seeder. Role
// assignments for these roles are reconciled with users.role at startup and
// login; custom roles created through the RBAC service are preserved.
var BuiltInRoles = map[string]struct{}{
	"owner":      {},
	"superadmin": {},
	"admin":      {},
	"agent":      {},
	"teknisi":    {},
	"finance":    {},
	"technician": {}, // legacy alias, migrated to teknisi
}

// IsBuiltInRole reports whether role belongs to the system-managed role set.
func IsBuiltInRole(role string) bool {
	_, ok := BuiltInRoles[strings.ToLower(strings.TrimSpace(role))]
	return ok
}

// IsAdminRole reports whether role is one of the harness/NetOps administrator
// roles: owner, superadmin, or admin.
func IsAdminRole(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "owner", "superadmin", "admin":
		return true
	default:
		return false
	}
}

// HasAdminRole reports whether any role in roles is an administrator role.
func HasAdminRole(roles []string) bool {
	for _, role := range roles {
		if IsAdminRole(role) {
			return true
		}
	}
	return false
}

// User represents an admin, agent, or technician who can access the system.
type User struct {
	ID                uint      `json:"id"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	PasswordHash      string    `json:"-"`
	Role              string    `json:"role"` // role utama: owner, admin, agent, teknisi
	FullName          string    `json:"full_name,omitempty"`
	PhoneNumber       string    `json:"phone_number,omitempty"`
	Specialization    string    `json:"specialization,omitempty"`
	IsActive          bool      `json:"is_active"`
	TenantID          string    `json:"tenant_id"`
	AssignedDeviceIDs []string  `json:"assigned_device_ids,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// UserDeviceAssignment represents a relation between a user and an assigned device.
type UserDeviceAssignment struct {
	UserID     uint      `json:"user_id"`
	DeviceID   string    `json:"device_id"`
	AssignedBy *uint     `json:"assigned_by,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
