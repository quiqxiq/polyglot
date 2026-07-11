package port

import (
	"context"
	"time"
)

// User represents an admin/staff user for authentication purposes. This
// struct lives in the port package because users are an infrastructure
// concern (auth), not a business domain entity. Mapped to the `users`
// table per migration 000004.
type User struct {
	ID           string
	Username     string
	PasswordHash string
	FullName     string
	Email        string
	Phone        string
	Role         string // "superadmin", "owner", "admin", "staff", "teknisi"
	IsActive     bool
	CreatedAt    time.Time
	LastLoginAt  *time.Time
}

// UserRepository defines persistence operations for admin/staff users.
type UserRepository interface {
	// FindByUsername returns the user with the given username, or an error
	// if not found.
	FindByUsername(ctx context.Context, username string) (User, error)

	// Create inserts a new user. The password should already be hashed.
	Create(ctx context.Context, user User) (User, error)

	// UpdatePassword changes the password hash for the user with the given ID.
	UpdatePassword(ctx context.Context, id string, passwordHash string) error

	// UpdateLastLogin records the current time as the last login for the
	// user with the given ID.
	UpdateLastLogin(ctx context.Context, id string) error
}
