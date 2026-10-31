package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/restorepoint"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// restorePointDefaultRetention adalah retensi pengaman bila usecase tidak
// mengisi ExpiresAt; kebijakan retensi sebenarnya tetap milik usecase.
const restorePointDefaultRetention = 30 * 24 * time.Hour

// RestorePointRepository adalah implementasi port.RestorePointRepository
// berbasis GORM/Postgres untuk bundle pemulihan sebelum mutasi router
// (PLAN-AI-ROUTER-OPS v2).
type RestorePointRepository struct {
	db *gorm.DB
}

var _ port.RestorePointRepository = (*RestorePointRepository)(nil)

// NewRestorePointRepository returns a port.RestorePointRepository backed by GORM.
func NewRestorePointRepository(db *gorm.DB) *RestorePointRepository {
	return &RestorePointRepository{db: db}
}

// Create menyimpan header restore point beserta seluruh artefaknya dalam satu
// transaksi; ID kosong dibuat otomatis.
func (r *RestorePointRepository) Create(ctx context.Context, rp restorepoint.RestorePoint) error {
	if rp.ID == "" {
		rp.ID = idgen.New("rp")
	}
	if rp.TenantID == "" {
		rp.TenantID = defaultTenantID
	}
	if rp.Status == "" {
		rp.Status = restorepoint.StatusCreating
	}
	if rp.CreatedAt.IsZero() {
		rp.CreatedAt = time.Now().UTC()
	}
	if rp.ExpiresAt.IsZero() {
		rp.ExpiresAt = rp.CreatedAt.Add(restorePointDefaultRetention)
	}
	header := model.RestorePointModelFromDomain(rp)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(header).Error; err != nil {
			return err
		}
		for _, artifact := range rp.Artifacts {
			if err := tx.Create(model.ArtifactModelFromDomain(rp.ID, artifact)).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// FindByID mengambil restore point beserta artefaknya.
func (r *RestorePointRepository) FindByID(ctx context.Context, id string) (restorepoint.RestorePoint, error) {
	var m model.RestorePointModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return restorepoint.RestorePoint{}, mapRestorePointError(err)
	}
	rp := m.ToDomain()
	if err := loadRestorePointArtifacts(ctx, r.db, []*restorepoint.RestorePoint{&rp}); err != nil {
		return restorepoint.RestorePoint{}, err
	}
	return rp, nil
}

// List mengembalikan header restore point sesuai filter beserta total baris
// (sebelum limit/offset), terbaru lebih dulu. Artefak tidak dimuat di sini —
// gunakan FindByID untuk bundle lengkap.
func (r *RestorePointRepository) List(ctx context.Context, f port.RestorePointFilter) ([]restorepoint.RestorePoint, int, error) {
	q := r.db.WithContext(ctx).Model(&model.RestorePointModel{}).
		Where("tenant_id = ?", orDefault(f.TenantID, defaultTenantID))
	if f.DeviceID != "" {
		q = q.Where("device_id = ?", f.DeviceID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	q = q.Order("created_at desc, id desc")
	if f.Limit > 0 {
		q = q.Limit(f.Limit)
	}
	if f.Offset > 0 {
		q = q.Offset(f.Offset)
	}
	var rows []model.RestorePointModel
	if err := q.Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]restorepoint.RestorePoint, len(rows))
	for i := range rows {
		out[i] = rows[i].ToDomain()
	}
	return out, int(total), nil
}

// UpdateStatus mengubah status bundle dan (opsional) used_at.
func (r *RestorePointRepository) UpdateStatus(ctx context.Context, id string, status restorepoint.Status, usedAt *time.Time) error {
	updates := map[string]any{"status": string(status)}
	if usedAt != nil {
		updates["used_at"] = *usedAt
	}
	res := r.db.WithContext(ctx).Model(&model.RestorePointModel{}).
		Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return restorepoint.ErrNotFound
	}
	return nil
}

// FindLatestByConfigHash mengembalikan bundle AVAILABLE terbaru untuk device
// dengan config_hash sama (dedupe binary backup).
func (r *RestorePointRepository) FindLatestByConfigHash(ctx context.Context, deviceID, configHash string) (restorepoint.RestorePoint, error) {
	if deviceID == "" || configHash == "" {
		return restorepoint.RestorePoint{}, restorepoint.ErrNotFound
	}
	var rows []model.RestorePointModel
	err := r.db.WithContext(ctx).
		Where("device_id = ? AND config_hash = ? AND status = ?",
			deviceID, configHash, string(restorepoint.StatusAvailable)).
		Order("created_at desc, id desc").
		Limit(1).
		Find(&rows).Error
	if err != nil {
		return restorepoint.RestorePoint{}, err
	}
	if len(rows) == 0 {
		return restorepoint.RestorePoint{}, restorepoint.ErrNotFound
	}
	rp := rows[0].ToDomain()
	if err := loadRestorePointArtifacts(ctx, r.db, []*restorepoint.RestorePoint{&rp}); err != nil {
		return restorepoint.RestorePoint{}, err
	}
	return rp, nil
}

// ListExpired mengembalikan bundle kedaluwarsa (expires_at < before) yang
// boleh dibersihkan retensi: paling lama lebih dulu, tetapi selalu
// mempertahankan `keepLast` bundle terbaru per device. Bundle DELETED
// dilewati. Artefak ikut dimuat agar pemanggil dapat menghapus file-nya.
func (r *RestorePointRepository) ListExpired(ctx context.Context, before time.Time, keepLast int) ([]restorepoint.RestorePoint, error) {
	if keepLast < 0 {
		keepLast = 0
	}
	const rankedSQL = `SELECT id FROM (
		SELECT id, ROW_NUMBER() OVER (PARTITION BY device_id ORDER BY created_at DESC, id DESC) AS rn
		FROM restore_points
		WHERE expires_at < ? AND status <> ?
	) ranked
	WHERE rn > ?
	ORDER BY id`
	ids := make([]string, 0)
	if err := r.db.WithContext(ctx).Raw(rankedSQL, before, string(restorepoint.StatusDeleted), keepLast).Scan(&ids).Error; err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []restorepoint.RestorePoint{}, nil
	}
	var rows []model.RestorePointModel
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).
		Order("created_at asc, id asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]restorepoint.RestorePoint, len(rows))
	for i := range rows {
		out[i] = rows[i].ToDomain()
	}
	pointers := make([]*restorepoint.RestorePoint, len(out))
	for i := range out {
		pointers[i] = &out[i]
	}
	if err := loadRestorePointArtifacts(ctx, r.db, pointers); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete menandai bundle DELETED (soft delete) dan mengembalikan bundle
// lengkap dengan artefaknya agar pemanggil dapat membersihkan file.
func (r *RestorePointRepository) Delete(ctx context.Context, id string) (restorepoint.RestorePoint, error) {
	var out restorepoint.RestorePoint
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := tx
		if r.db.Dialector.Name() == "postgres" {
			q = q.Clauses(lockingClause())
		}
		var m model.RestorePointModel
		if err := q.First(&m, "id = ?", id).Error; err != nil {
			return mapRestorePointError(err)
		}
		if err := tx.Model(&model.RestorePointModel{}).Where("id = ?", id).
			Update("status", string(restorepoint.StatusDeleted)).Error; err != nil {
			return err
		}
		m.Status = string(restorepoint.StatusDeleted)
		rp := m.ToDomain()
		if err := loadRestorePointArtifacts(ctx, tx, []*restorepoint.RestorePoint{&rp}); err != nil {
			return err
		}
		out = rp
		return nil
	})
	if err != nil {
		return restorepoint.RestorePoint{}, err
	}
	return out, nil
}

// loadRestorePointArtifacts memuat artefak untuk sekumpulan bundle sekaligus
// (satu query IN) lalu menempelkannya ke field Artifacts.
func loadRestorePointArtifacts(ctx context.Context, db *gorm.DB, points []*restorepoint.RestorePoint) error {
	if len(points) == 0 {
		return nil
	}
	ids := make([]string, len(points))
	for i, p := range points {
		ids[i] = p.ID
	}
	var rows []model.ArtifactModel
	if err := db.WithContext(ctx).Where("restore_point_id IN ?", ids).
		Order("id asc").Find(&rows).Error; err != nil {
		return err
	}
	byPoint := make(map[string][]restorepoint.Artifact, len(points))
	for i := range rows {
		byPoint[rows[i].RestorePointID] = append(byPoint[rows[i].RestorePointID], rows[i].ToDomain())
	}
	for _, p := range points {
		p.Artifacts = byPoint[p.ID]
	}
	return nil
}

// mapRestorePointError menerjemahkan gorm not-found menjadi sentinel domain.
func mapRestorePointError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return restorepoint.ErrNotFound
	}
	return err
}
