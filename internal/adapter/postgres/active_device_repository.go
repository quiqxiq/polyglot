package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/port"
)

// activeDeviceNoTTL adalah horizon "tanpa kedaluwarsa" bila TTL <= 0, supaya
// binding tidak langsung dianggap kedaluwarsa oleh filter Get.
const activeDeviceNoTTL = 100 * 365 * 24 * time.Hour

// ActiveDeviceRepository adalah implementasi port.ActiveDeviceStore berbasis
// GORM/Postgres: fallback/audit di samping Redis untuk binding "router aktif"
// per user/sesi. TTL disimpan pada kolom updated_at sebagai batas
// kedaluwarsa (now + ttl) sehingga Get cukup memfilter updated_at > now.
type ActiveDeviceRepository struct {
	db *gorm.DB
}

var _ port.ActiveDeviceStore = (*ActiveDeviceRepository)(nil)

// NewActiveDeviceRepository returns a port.ActiveDeviceStore backed by GORM.
func NewActiveDeviceRepository(db *gorm.DB) *ActiveDeviceRepository {
	return &ActiveDeviceRepository{db: db}
}

// Set meng-upsert binding router aktif untuk (tenant, user, sesi) dan
// menetapkan masa berlaku now+ttl.
func (r *ActiveDeviceRepository) Set(ctx context.Context, a port.ActiveDevice, ttl time.Duration) error {
	now := time.Now().UTC()
	expiresAt := now.Add(activeDeviceNoTTL)
	if ttl > 0 {
		expiresAt = now.Add(ttl)
	}
	m := model.ActiveDeviceModelFromDomain(a, expiresAt)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "tenant_id"}, {Name: "user_id"}, {Name: "session_id"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"device_id":  gorm.Expr("EXCLUDED.device_id"),
			"updated_at": gorm.Expr("EXCLUDED.updated_at"),
		}),
	}).Create(m).Error
}

// Get mengambil binding yang masih berlaku; binding kedaluwarsa dianggap
// tidak ada.
func (r *ActiveDeviceRepository) Get(ctx context.Context, tenantID, userID, sessionID string) (port.ActiveDevice, error) {
	var m model.ActiveDeviceModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ? AND session_id = ? AND updated_at > ?",
			orDefault(tenantID, defaultTenantID), userID, sessionID, time.Now().UTC()).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return port.ActiveDevice{}, ErrNotFound
	}
	if err != nil {
		return port.ActiveDevice{}, err
	}
	return m.ToDomain(), nil
}

// Clear menghapus binding (idempoten; tidak error bila tidak ada baris).
func (r *ActiveDeviceRepository) Clear(ctx context.Context, tenantID, userID, sessionID string) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ? AND session_id = ?",
			orDefault(tenantID, defaultTenantID), userID, sessionID).
		Delete(&model.ActiveDeviceModel{}).Error
}
