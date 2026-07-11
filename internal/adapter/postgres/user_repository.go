package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/port"
)

// userModel maps the `users` table per migration 000004.
type userModel struct {
	ID           string     `gorm:"column:id;primaryKey"`
	Username     string     `gorm:"column:username;not null;unique"`
	PasswordHash string     `gorm:"column:password_hash;not null"`
	FullName     string     `gorm:"column:full_name;not null"`
	Email        string     `gorm:"column:email"`
	Phone        string     `gorm:"column:phone"`
	Role         string     `gorm:"column:role;not null"`
	IsActive     bool       `gorm:"column:is_active;not null;default:true"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null;autoCreateTime"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at"`
}

func (userModel) TableName() string {
	return "users"
}

func (m userModel) toPortUser() port.User {
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

func fromPortUser(u port.User) userModel {
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
			return port.User{}, fmt.Errorf("user %s: not found", username)
		}
		return port.User{}, fmt.Errorf("user %s: %w", username, err)
	}
	return m.toPortUser(), nil
}

// Create inserts a new user. The password must already be hashed.
func (r *UserRepository) Create(ctx context.Context, user port.User) (port.User, error) {
	m := fromPortUser(user)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return port.User{}, fmt.Errorf("create user: %w", err)
	}
	return m.toPortUser(), nil
}

// UpdatePassword changes the password hash for the user with the given ID.
func (r *UserRepository) UpdatePassword(ctx context.Context, id string, passwordHash string) error {
	res := r.db.WithContext(ctx).Model(&userModel{}).Where("id = ?", id).Update("password_hash", passwordHash)
	if res.Error != nil {
		return fmt.Errorf("update password: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("update password: user %s not found", id)
	}
	return nil
}

// UpdateLastLogin records the current time as the last login for the user
// with the given ID.
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&userModel{}).Where("id = ?", id).Update("last_login_at", now)
	if res.Error != nil {
		return fmt.Errorf("update last login: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("update last login: user %s not found", id)
	}
	return nil
}
