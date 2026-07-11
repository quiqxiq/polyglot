package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/port"
)

// userModel maps the `users` table to a GORM-friendly struct per migration
// 000004.
type userModel struct {
	ID           string     `gorm:"column:id;primaryKey;type:uuid"`
	Username     string     `gorm:"column:username;uniqueIndex;not null"`
	PasswordHash string     `gorm:"column:password_hash;not null"`
	FullName     string     `gorm:"column:full_name;not null"`
	Email        string     `gorm:"column:email"`
	Phone        string     `gorm:"column:phone"`
	Role         string     `gorm:"column:role;not null"`
	IsActive     bool       `gorm:"column:is_active;not null;default:true"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null;autoCreateTime"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at"`
}

// TableName returns the explicit table name for the user model.
func (userModel) TableName() string {
	return "users"
}

// toDomain maps a userModel to a port.User.
func (m userModel) toDomain() port.User {
	return port.User{
		ID:           m.ID,
		Username:     m.Username,
		PasswordHash: m.PasswordHash,
		FullName:     m.FullName,
		Email:        m.Email,
		Phone:        m.Phone,
		Role:         m.Role,
		IsActive:     m.IsActive,
		CreatedAt:    m.CreatedAt,
		LastLoginAt:  m.LastLoginAt,
	}
}

// userFromDomain maps a port.User to a userModel.
func userFromDomain(u port.User) userModel {
	return userModel{
		ID:           u.ID,
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		FullName:     u.FullName,
		Email:        u.Email,
		Phone:        u.Phone,
		Role:         u.Role,
		IsActive:     u.IsActive,
		CreatedAt:    u.CreatedAt,
		LastLoginAt:  u.LastLoginAt,
	}
}

// UserRepository implements port.UserRepository backed by PostgreSQL.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository returns a port.UserRepository backed by GORM/Postgres.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByUsername returns the user with the given username, or an error if
// not found.
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (port.User, error) {
	var m userModel
	if err := r.db.WithContext(ctx).First(&m, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return port.User{}, fmt.Errorf("user %q: not found", username)
		}
		return port.User{}, fmt.Errorf("user %q: %w", username, err)
	}
	return m.toDomain(), nil
}

// Create inserts a new user. The password should already be hashed.
func (r *UserRepository) Create(ctx context.Context, u port.User) (port.User, error) {
	m := userFromDomain(u)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return port.User{}, fmt.Errorf("create user: %w", err)
	}
	return m.toDomain(), nil
}

// UpdatePassword changes the password hash for the user with the given ID.
func (r *UserRepository) UpdatePassword(ctx context.Context, id string, passwordHash string) error {
	result := r.db.WithContext(ctx).Model(&userModel{}).Where("id = ?", id).Update("password_hash", passwordHash)
	if result.Error != nil {
		return fmt.Errorf("update password: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user %s: not found", id)
	}
	return nil
}

// UpdateLastLogin records the current time as the last login for the user.
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&userModel{}).Where("id = ?", id).Update("last_login_at", now)
	if result.Error != nil {
		return fmt.Errorf("update last login: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user %s: not found", id)
	}
	return nil
}
