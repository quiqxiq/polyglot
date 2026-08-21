package customer

import "time"

// User represents an admin, agent, or technician who can access the system.
type User struct {
	ID             uint      `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	PasswordHash   string    `json:"-"`
	Role           string    `json:"role"` // role utama: owner, admin, agent, teknisi
	FullName       string    `json:"full_name,omitempty"`
	PhoneNumber    string    `json:"phone_number,omitempty"`
	Specialization string    `json:"specialization,omitempty"`
	IsActive       bool      `json:"is_active"`
	TenantID       string    `json:"tenant_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
