package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/port"
)

type PortalRepository struct {
	db *gorm.DB
}

var _ port.PortalRepository = (*PortalRepository)(nil)

// NewPortalRepository returns a port.PortalRepository backed by GORM/Postgres.
func NewPortalRepository(db *gorm.DB) *PortalRepository {
	return &PortalRepository{db: db}
}

// ─── Sessions ───────────────────────────────────────────────────────────

func (r *PortalRepository) SaveSession(ctx context.Context, s domainCustomer.PortalSession) error {
	m := model.PortalSessionModel{
		ID: s.ID, TenantID: s.TenantID, CustomerID: s.CustomerID,
		SessionToken: s.SessionToken, IPAddress: s.IPAddress, UserAgent: s.UserAgent,
		ExpiresAt: s.ExpiresAt, CreatedAt: s.CreatedAt,
	}
	return r.db.WithContext(ctx).Save(&m).Error
}

func (r *PortalRepository) FindValidSession(ctx context.Context, token string) (domainCustomer.PortalSession, error) {
	var m model.PortalSessionModel
	err := r.db.WithContext(ctx).
		Where("session_token = ? AND expires_at > ?", token, time.Now()).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domainCustomer.PortalSession{}, ErrNotFound
	}
	if err != nil {
		return domainCustomer.PortalSession{}, err
	}
	return domainCustomer.PortalSession{
		ID: m.ID, TenantID: m.TenantID, CustomerID: m.CustomerID,
		SessionToken: m.SessionToken, IPAddress: m.IPAddress, UserAgent: m.UserAgent,
		ExpiresAt: m.ExpiresAt, CreatedAt: m.CreatedAt,
	}, nil
}

func (r *PortalRepository) DeleteSession(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Delete(&model.PortalSessionModel{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ─── OTP ────────────────────────────────────────────────────────────────

func (r *PortalRepository) SaveOTP(ctx context.Context, o domainCustomer.PortalOTP) error {
	m := model.PortalOTPModelFromDomain(o)
	return r.db.WithContext(ctx).Create(m).Error
}

// ConsumeOTP implements port.PortalRepository with row-level locking:
// ambil OTP terbaru aktif untuk phone; cocok → consumed; salah → attempts++;
// melewati maxAttempts → consumed + ErrOTPLocked. Semua perubahan di-commit
// sebelum error dikembalikan agar lockout benar-benar tersimpan (F5-1).
func (r *PortalRepository) ConsumeOTP(ctx context.Context, phone, codeHash string, maxAttempts int) (bool, error) {
	var matched, locked, notFound bool
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var m model.PortalOTPModel
		q := tx
		if r.db.Dialector.Name() == "postgres" {
			q = q.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		err := q.
			Where("phone = ? AND consumed_at IS NULL AND expires_at > ?", phone, time.Now()).
			Order("created_at desc").First(&m).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			notFound = true
			return nil // commit transaksi kosong; not-found bukan kegagalan DB
		}
		if err != nil {
			return err
		}
		now := time.Now()
		if m.Attempts >= maxAttempts {
			locked = true
			return tx.Model(&m).Update("consumed_at", now).Error
		}
		if m.CodeHash != codeHash {
			newAttempts := m.Attempts + 1
			updates := map[string]any{"attempts": newAttempts}
			if newAttempts >= maxAttempts {
				updates["consumed_at"] = now
				locked = true
			}
			return tx.Model(&m).Updates(updates).Error
		}
		matched = true
		return tx.Model(&m).Update("consumed_at", now).Error
	})
	if err != nil {
		return false, err
	}
	switch {
	case matched:
		return true, nil
	case locked:
		return false, domainCustomer.ErrOTPLocked
	case notFound:
		return false, domainCustomer.ErrOTPNotFound
	default:
		return false, nil
	}
}
