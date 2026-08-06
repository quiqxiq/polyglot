package models

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/customer"
)

// UserModel is the GORM database model for users.
type UserModel struct {
	ID           uint   `gorm:"primaryKey"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	Role         string `gorm:"not null;default:agent"`
	TenantID     string `gorm:"not null;default:tenant-default"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (m *UserModel) ToDomain() *customer.User {
	if m == nil {
		return nil
	}
	return &customer.User{
		ID:           m.ID,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		Role:         m.Role,
		TenantID:     m.TenantID,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func UserModelFromDomain(u *customer.User) *UserModel {
	if u == nil {
		return nil
	}
	return &UserModel{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		TenantID:     u.TenantID,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}
