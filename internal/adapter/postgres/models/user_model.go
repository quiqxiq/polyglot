package models

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/customer"
)

// UserModel is the GORM database model for users.
// TableName eksplisit ke `users` (migrasi 000002) — tanpa ini GORM memakai
// `user_models` (AutoMigrate dev), divergen dari migrasi (prod).
type UserModel struct {
	ID uint `gorm:"primaryKey"`
	// Tipe eksplisit = migrasi 000002 (`username VARCHAR(100) UNIQUE`, dst.).
	// Tanpa `type:` GORM memetakan string -> `text` dan AutoMigrate mengubah
	// kolom dari migrasi (dev schema divergen dari prod). `unique` (bukan
	// `uniqueIndex`) karena migrasi memakai UNIQUE constraint per-kolom —
	// dengan `uniqueIndex` GORM membaca columnType.Unique()=true tapi
	// field.Unique=false → AutoMigrate crash saat DROP CONSTRAINT `uni_*`.
	Username     string `gorm:"type:varchar(100);unique;not null"`
	Email        string `gorm:"type:varchar(255);unique;not null"`
	PasswordHash string `gorm:"type:varchar(255);not null"`
	Role         string `gorm:"type:varchar(50);not null;default:agent"`
	TenantID     string `gorm:"type:varchar(100);not null;default:tenant-default"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName maps UserModel ke tabel migrasi `users`.
func (UserModel) TableName() string { return "users" }

func (m *UserModel) ToDomain() *customer.User {
	if m == nil {
		return nil
	}
	return &customer.User{
		ID:           m.ID,
		Username:     m.Username,
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
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		TenantID:     u.TenantID,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}
