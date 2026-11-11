package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/ont"
	domainTenant "github.com/quixiq/polyglot/internal/domain/tenant"
	"github.com/quixiq/polyglot/internal/port"
)

// defaultTenantID adalah tenant tunggal deployment saat ini (D16
// PLAN-ACS-OLT-HIOSO.md), konsisten dengan default kolom tenant_id.
const defaultTenantID = domainTenant.DefaultID

// ONTRepository adalah implementasi port.ONTRepository berbasis GORM/Postgres
// untuk inventori ONT/ONU, timeline event, dan sampel metrik optik.
type ONTRepository struct {
	db *gorm.DB
}

var _ port.ONTRepository = (*ONTRepository)(nil)

// NewONTRepository returns a port.ONTRepository backed by GORM/Postgres.
func NewONTRepository(db *gorm.DB) *ONTRepository {
	return &ONTRepository{db: db}
}

// Save membuat atau memperbarui satu ONT (upsert berdasarkan id). ID baru
// dibuat dengan prefix "ont" bila pemanggil belum mengisinya.
func (r *ONTRepository) Save(ctx context.Context, o ont.ONT) error {
	if o.ID == "" {
		o.ID = newID("ont")
	}
	// created_at di-omit dari UPDATE agar upsert oleh pemanggil yang tidak
	// membawa timestamp lama tidak menimpa waktu pembuatan baris.
	return r.db.WithContext(ctx).Omit("created_at").Save(model.ONTModelFromDomain(o)).Error
}

// FindByID mencari ONT berdasarkan id.
func (r *ONTRepository) FindByID(ctx context.Context, id string) (ont.ONT, error) {
	var m model.ONTModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return ont.ONT{}, mapONTError(err)
	}
	return m.ToDomain(), nil
}

// FindByOLTAndMAC mencari ONT berdasarkan OLT + MAC (ternormalisasi).
// MAC dinormalisasi lebih dulu agar format "AA-BB-..." tetap cocok.
func (r *ONTRepository) FindByOLTAndMAC(ctx context.Context, oltDeviceID, mac string) (ont.ONT, error) {
	if norm, ok := ont.NormalizeMAC(mac); ok {
		mac = norm
	}
	var m model.ONTModel
	err := r.db.WithContext(ctx).
		Where("olt_device_id = ? AND mac_address = ?", oltDeviceID, mac).
		First(&m).Error
	if err != nil {
		return ont.ONT{}, mapONTError(err)
	}
	return m.ToDomain(), nil
}

// FindByOLTPosition mencari ONT berdasarkan posisi PON/ONU pada satu OLT.
func (r *ONTRepository) FindByOLTPosition(ctx context.Context, oltDeviceID string, ponIndex, onuIndex int) (ont.ONT, error) {
	var m model.ONTModel
	err := r.db.WithContext(ctx).
		Where("olt_device_id = ? AND pon_index = ? AND onu_index = ?", oltDeviceID, ponIndex, onuIndex).
		First(&m).Error
	if err != nil {
		return ont.ONT{}, mapONTError(err)
	}
	return m.ToDomain(), nil
}

// ListByOLT mengembalikan seluruh ONT satu OLT terurut posisi PON/ONU.
func (r *ONTRepository) ListByOLT(ctx context.Context, oltDeviceID string) ([]ont.ONT, error) {
	var rows []model.ONTModel
	err := r.db.WithContext(ctx).
		Where("olt_device_id = ?", oltDeviceID).
		Order("pon_index asc, onu_index asc").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]ont.ONT, len(rows))
	for i := range rows {
		out[i] = rows[i].ToDomain()
	}
	return out, nil
}

// List mengembalikan ONT dengan filter opsional (semua field kosong = tanpa
// filter), terurut OLT lalu posisi PON/ONU agar inventori deterministik.
func (r *ONTRepository) List(ctx context.Context, f port.ONTFilter) ([]ont.ONT, error) {
	q := r.db.WithContext(ctx).Model(&model.ONTModel{})
	if f.OLTDeviceID != "" {
		q = q.Where("olt_device_id = ?", f.OLTDeviceID)
	}
	if f.PONIndex > 0 {
		q = q.Where("pon_index = ?", f.PONIndex)
	}
	if f.State != "" {
		q = q.Where("state = ?", f.State)
	}
	if f.CustomerID != "" {
		q = q.Where("customer_id = ?", f.CustomerID)
	}
	if s := strings.TrimSpace(f.Search); s != "" {
		like := "%" + s + "%"
		op := likeOperator(r.db)
		q = q.Where(fmt.Sprintf(
			"mac_address %[1]s ? OR serial_number %[1]s ? OR model %[1]s ? OR pppoe_username %[1]s ?",
			op), like, like, like, like)
	}
	if f.Linked != nil {
		if *f.Linked {
			q = q.Where("(subscription_id IS NOT NULL OR customer_id IS NOT NULL)")
		} else {
			q = q.Where("subscription_id IS NULL AND customer_id IS NULL")
		}
	}
	if f.ACSMatched != nil {
		if *f.ACSMatched {
			q = q.Where("acs_device_id IS NOT NULL")
		} else {
			q = q.Where("acs_device_id IS NULL")
		}
	}
	if f.Limit > 0 {
		q = q.Limit(f.Limit)
	}
	if f.Offset > 0 {
		q = q.Offset(f.Offset)
	}
	var rows []model.ONTModel
	if err := q.Order("olt_device_id asc, pon_index asc, onu_index asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]ont.ONT, len(rows))
	for i := range rows {
		out[i] = rows[i].ToDomain()
	}
	return out, nil
}

// UpdateLink menautkan ONT ke pelanggan/langganan dan username PPPoE.
func (r *ONTRepository) UpdateLink(ctx context.Context, id string, customerID, subscriptionID *string, pppoeUsername string) error {
	res := r.db.WithContext(ctx).Model(&model.ONTModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"customer_id":     customerID,
			"subscription_id": subscriptionID,
			"pppoe_username":  pppoeUsername,
		})
	return mapONTResult(res)
}

// UpdateACS memperbarui kolom integrasi ACS. telemetry kosong berarti NULL.
func (r *ONTRepository) UpdateACS(ctx context.Context, id string, acsDeviceID *string, matchSource string, lastInform *time.Time, telemetry []byte, wlanSlots int) error {
	res := r.db.WithContext(ctx).Model(&model.ONTModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"acs_device_id":      acsDeviceID,
			"acs_matched_by":     matchSource,
			"acs_last_inform_at": lastInform,
			"acs_telemetry":      jsonbParam(telemetry),
			"wlan_slots":         wlanSlots,
		})
	return mapONTResult(res)
}

// RecordEvent menambahkan satu entri timeline ONT.
func (r *ONTRepository) RecordEvent(ctx context.Context, e ont.Event) error {
	if e.ID == "" {
		e.ID = newID("ontev")
	}
	return r.db.WithContext(ctx).Create(model.ONTEventModelFromDomain(e)).Error
}

// ListEvents mengembalikan timeline ONT terbaru lebih dulu; limit <= 0
// berarti tanpa batas.
func (r *ONTRepository) ListEvents(ctx context.Context, ontID string, limit int) ([]ont.Event, error) {
	q := r.db.WithContext(ctx).Model(&model.ONTEventModel{}).Where("ont_id = ?", ontID)
	if limit > 0 {
		q = q.Limit(limit)
	}
	var rows []model.ONTEventModel
	if err := q.Order("created_at desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]ont.Event, len(rows))
	for i := range rows {
		out[i] = rows[i].ToDomain()
	}
	return out, nil
}

// AddSample menambahkan satu sampel metrik optik ONT.
func (r *ONTRepository) AddSample(ctx context.Context, s ont.MetricSample) error {
	if s.ID == "" {
		s.ID = newID("ontsmp")
	}
	if s.SampledAt.IsZero() {
		s.SampledAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(model.ONTMetricSampleModelFromDomain(s)).Error
}

// ListSamples mengembalikan sampel metrik terbaru lebih dulu dengan filter
// waktu opsional (from nol = tanpa batas bawah); limit <= 0 = tanpa batas.
func (r *ONTRepository) ListSamples(ctx context.Context, ontID string, from time.Time, limit int) ([]ont.MetricSample, error) {
	q := r.db.WithContext(ctx).Model(&model.ONTMetricSampleModel{}).Where("ont_id = ?", ontID)
	if !from.IsZero() {
		q = q.Where("sampled_at >= ?", from)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	var rows []model.ONTMetricSampleModel
	if err := q.Order("sampled_at desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]ont.MetricSample, len(rows))
	for i := range rows {
		out[i] = rows[i].ToDomain()
	}
	return out, nil
}

// PurgeSamplesBefore menghapus sampel metrik lebih lama dari `before` dan
// mengembalikan jumlah baris yang terhapus (retensi harian).
func (r *ONTRepository) PurgeSamplesBefore(ctx context.Context, before time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("sampled_at < ?", before).
		Delete(&model.ONTMetricSampleModel{})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// CountByState menghitung jumlah ONT per state; oltDeviceID kosong berarti
// seluruh OLT.
func (r *ONTRepository) CountByState(ctx context.Context, oltDeviceID string) (port.ONTStateCounts, error) {
	q := r.db.WithContext(ctx).Model(&model.ONTModel{})
	if oltDeviceID != "" {
		q = q.Where("olt_device_id = ?", oltDeviceID)
	}
	var row struct {
		Total   int
		Online  int
		Offline int
		Unknown int
		Removed int
	}
	err := q.Select(`COUNT(*) AS total,
		COALESCE(SUM(CASE WHEN state = 'ONLINE' THEN 1 ELSE 0 END), 0) AS online,
		COALESCE(SUM(CASE WHEN state = 'OFFLINE' THEN 1 ELSE 0 END), 0) AS offline,
		COALESCE(SUM(CASE WHEN state = 'UNKNOWN' THEN 1 ELSE 0 END), 0) AS unknown,
		COALESCE(SUM(CASE WHEN state = 'REMOVED' THEN 1 ELSE 0 END), 0) AS removed`).
		Scan(&row).Error
	if err != nil {
		return port.ONTStateCounts{}, err
	}
	return port.ONTStateCounts{
		Total:   row.Total,
		Online:  row.Online,
		Offline: row.Offline,
		Unknown: row.Unknown,
		Removed: row.Removed,
	}, nil
}

// mapONTError menerjemahkan gorm.ErrRecordNotFound menjadi sentinel domain.
func mapONTError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ont.ErrNotFound
	}
	return err
}

// mapONTResult menerjemahkan hasil UPDATE tanpa baris menjadi sentinel domain.
func mapONTResult(res *gorm.DB) error {
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ont.ErrNotFound
	}
	return nil
}

// likeOperator memilih operator pencarian sesuai dialector: ILIKE untuk
// PostgreSQL produksi, LIKE untuk SQLite in-memory pada unit test.
func likeOperator(db *gorm.DB) string {
	if db.Dialector.Name() == "postgres" {
		return "ILIKE"
	}
	return "LIKE"
}

// jsonbParam mengubah JSON mentah menjadi parameter kolom JSONB: nil/Kosong
// menjadi NULL agar tidak menulis string kosong yang bukan JSON valid.
func jsonbParam(raw []byte) any {
	if len(raw) == 0 {
		return nil
	}
	return string(raw)
}
