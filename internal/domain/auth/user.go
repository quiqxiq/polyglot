package auth

import "time"

// User represents an authenticated system user (admin, technician, operator).
type User struct {
	ID           uint      `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"` // admin, operator, technician
	TenantID     string    `json:"tenant_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// LoginResult contains the result of a successful user authentication.
type LoginResult struct {
	Token         string    `json:"token"`
	User          *User     `json:"user"`
	ExpiresAtUnix int64     `json:"expires_at_unix"`
}
