package customer

import "time"

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
