package customer

import "time"

// User represents an admin or agent who can access the system.
type User struct {
	ID           uint      `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"` // admin, agent
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
