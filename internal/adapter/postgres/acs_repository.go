package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/ont"
	"github.com/quixiq/polyglot/internal/port"
)

// ACSRepository adalah implementasi port.ACSRepository berbasis GORM/Postgres
// untuk cache device GenieACS dan preset yang dikelola Polyglot.
type ACSRepository struct {
	db *gorm.DB
}

var _ port.ACSRepository = (*ACSRepository)(nil)

// NewACSRepository returns a port.ACSRepository backed by GORM/Postgres.
func NewACSRepository(db *gorm.DB) *ACSRepository {
	return &ACSRepository{db: db}
}

// acsDeviceUpdateColumns adalah kolom yang diperbarui saat sync ulang device
// yang sama. id/created_at/first_seen_at sengaja dipertahankan.
var acsDeviceUpdateColumns = []string{
	"tenant_id", "oui", "product_class", "serial_number", "software_version",
	"last_inform_at", "tags", "mac_candidates", "raw",
	"matched_ont_id", "match_source", "last_seen_at", "updated_at",
}

// UpsertDevice menyimpan cache device ACS secara idempoten berdasarkan
// acs_device_id. Nilai first_seen_at dan status match yang sudah ada
// dipertahankan bila data baru tidak membawanya.
func (r *ACSRepository) UpsertDevice(ctx context.Context, d ont.ACSDevice) error {
	if d.TenantID == "" {
		d.TenantID = defaultTenantID
	}
	var existing model.ACSDeviceModel
	err := r.db.WithContext(ctx).Where("acs_device_id = ?", d.ACSDeviceID).First(&existing).Error
	switch {
	case err == nil:
		if d.ID == "" {
			d.ID = existing.ID
		}
		if d.FirstSeenAt.IsZero() {
			d.FirstSeenAt = existing.FirstSeenAt
		}
		if d.LastSeenAt.IsZero() {
			d.LastSeenAt = existing.LastSeenAt
		}
		if d.LastInformAt == nil {
			d.LastInformAt = existing.LastInformAt
		}
		if d.MatchedONTID == nil {
			d.MatchedONTID = existing.MatchedONTID
			if d.MatchSource == "" {
				d.MatchSource = existing.MatchSource
			}
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		if d.ID == "" {
			d.ID = newID("acsdev")
		}
		if d.FirstSeenAt.IsZero() {
			d.FirstSeenAt = time.Now()
		}
		if d.LastSeenAt.IsZero() {
			d.LastSeenAt = d.FirstSeenAt
		}
	default:
		return err
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "acs_device_id"}},
		DoUpdates: clause.AssignmentColumns(acsDeviceUpdateColumns),
	}).Create(model.ACSDeviceModelFromDomain(d)).Error
}

// ListDevices mengembalikan cache device ACS dengan filter status match,
// pencarian teks, dan paginasi. Urutan: inform terbaru lebih dulu.
func (r *ACSRepository) ListDevices(ctx context.Context, f port.ACSDeviceFilter) ([]ont.ACSDevice, error) {
	q := r.db.WithContext(ctx).Model(&model.ACSDeviceModel{})
	if f.Matched != nil {
		if *f.Matched {
			q = q.Where("matched_ont_id IS NOT NULL")
		} else {
			q = q.Where("matched_ont_id IS NULL")
		}
	}
	if s := strings.TrimSpace(f.Search); s != "" {
		like := "%" + s + "%"
		op := likeOperator(r.db)
		q = q.Where(fmt.Sprintf(
			"acs_device_id %[1]s ? OR oui %[1]s ? OR product_class %[1]s ? OR serial_number %[1]s ? OR software_version %[1]s ?",
			op), like, like, like, like, like)
	}
	if f.Limit > 0 {
		q = q.Limit(f.Limit)
	}
	if f.Offset > 0 {
		q = q.Offset(f.Offset)
	}
	var rows []model.ACSDeviceModel
	err := q.Order("last_seen_at desc, acs_device_id asc").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]ont.ACSDevice, len(rows))
	for i := range rows {
		out[i] = rows[i].ToDomain()
	}
	return out, nil
}

// FindDeviceByACSID mencari satu device cache berdasarkan id GenieACS.
func (r *ACSRepository) FindDeviceByACSID(ctx context.Context, acsDeviceID string) (ont.ACSDevice, error) {
	var m model.ACSDeviceModel
	err := r.db.WithContext(ctx).Where("acs_device_id = ?", acsDeviceID).First(&m).Error
	if err != nil {
		return ont.ACSDevice{}, mapONTError(err)
	}
	return m.ToDomain(), nil
}

// UpdateDeviceMatch menetapkan (atau melepas, bila matchedONTID nil) kaitan
// device ACS ke satu ONT. Bila ONT tersebut masih dipegang device lain, kaitan
// device lain dilepas lebih dulu (D10: konflik dilepas).
func (r *ACSRepository) UpdateDeviceMatch(ctx context.Context, acsDeviceID string, matchedONTID *string, source string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if matchedONTID != nil {
			if err := tx.Model(&model.ACSDeviceModel{}).
				Where("matched_ont_id = ? AND acs_device_id <> ?", *matchedONTID, acsDeviceID).
				Updates(map[string]any{"matched_ont_id": nil, "match_source": nil}).Error; err != nil {
				return err
			}
		}
		var matchSource any
		if matchedONTID != nil && source != "" {
			matchSource = source
		}
		res := tx.Model(&model.ACSDeviceModel{}).
			Where("acs_device_id = ?", acsDeviceID).
			Updates(map[string]any{
				"matched_ont_id": matchedONTID,
				"match_source":   matchSource,
			})
		return mapONTResult(res)
	})
}

// acsPresetUpdateColumns adalah kolom mutable preset saat upsert by
// (tenant_id, name); id/created_at dipertahankan.
var acsPresetUpdateColumns = []string{
	"precondition", "configurations", "enabled", "last_synced_at", "last_error", "updated_at",
}

// SavePreset menyimpan preset ACS secara idempoten berdasarkan (tenant, name).
// ID dan created_at baris lama dipertahankan.
func (r *ACSRepository) SavePreset(ctx context.Context, p ont.ACSPreset) error {
	if p.TenantID == "" {
		p.TenantID = defaultTenantID
	}
	var existing model.ACSPresetModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND name = ?", p.TenantID, p.Name).
		First(&existing).Error
	switch {
	case err == nil:
		p.ID = existing.ID
		if p.CreatedAt.IsZero() {
			p.CreatedAt = existing.CreatedAt
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		if p.ID == "" {
			p.ID = newID("acsp")
		}
	default:
		return err
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tenant_id"}, {Name: "name"}},
		DoUpdates: clause.AssignmentColumns(acsPresetUpdateColumns),
	}).Create(model.ACSPresetModelFromDomain(p)).Error
}

// GetPreset mencari preset berdasarkan name pada tenant default.
func (r *ACSRepository) GetPreset(ctx context.Context, name string) (ont.ACSPreset, error) {
	var m model.ACSPresetModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND name = ?", defaultTenantID, name).
		First(&m).Error
	if err != nil {
		return ont.ACSPreset{}, mapONTError(err)
	}
	return m.ToDomain(), nil
}

// ListPresets mengembalikan seluruh preset tenant default terurut nama.
func (r *ACSRepository) ListPresets(ctx context.Context) ([]ont.ACSPreset, error) {
	var rows []model.ACSPresetModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", defaultTenantID).
		Order("name asc").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]ont.ACSPreset, len(rows))
	for i := range rows {
		out[i] = rows[i].ToDomain()
	}
	return out, nil
}
