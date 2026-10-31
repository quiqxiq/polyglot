package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/restorepoint"
)

// RestorePointModel is the GORM model for restore_points (migrasi 000041
// PLAN-AI-ROUTER-OPS v2): bundle pemulihan sebelum mutasi router.
// Artifacts tidak dimuat di sini; repository menggabungkannya dari tabel
// restore_point_artifacts.
type RestorePointModel struct {
	ID       string  `gorm:"column:id;primaryKey"`
	TenantID string  `gorm:"column:tenant_id;type:text;not null;default:tenant-default"`
	DeviceID *string `gorm:"column:device_id;type:uuid"`

	Trigger         string  `gorm:"column:trigger;type:text;not null"`
	Scope           string  `gorm:"column:scope;type:jsonb;not null;default:'[]'"`
	RouterOSVersion string  `gorm:"column:routeros_version;type:text"`
	ConfigHash      string  `gorm:"column:config_hash;type:text"`
	RollbackScript  *string `gorm:"column:rollback_script;type:jsonb"`
	Manifest        string  `gorm:"column:manifest;type:jsonb;not null;default:'{}'"`

	TotalSizeBytes int64  `gorm:"column:total_size_bytes;not null;default:0"`
	Status         string `gorm:"column:status;type:text;not null;default:CREATING"`
	CreatedBy      string `gorm:"column:created_by;type:text;not null;default:''"`

	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	ExpiresAt time.Time  `gorm:"column:expires_at;not null"`
	UsedAt    *time.Time `gorm:"column:used_at"`
}

// TableName returns the database table name for restore points.
func (RestorePointModel) TableName() string {
	return "restore_points"
}

// ToDomain converts a restore point database row to its domain representation.
// Daftar Artifacts diisi oleh repository setelah baris artefak dimuat.
func (m *RestorePointModel) ToDomain() restorepoint.RestorePoint {
	if m == nil {
		return restorepoint.RestorePoint{}
	}
	return restorepoint.RestorePoint{
		ID:              m.ID,
		TenantID:        m.TenantID,
		DeviceID:        derefOrEmpty(m.DeviceID),
		Trigger:         restorepoint.Trigger(m.Trigger),
		Scope:           aiDecodeStringSlice(m.Scope),
		RouterOSVersion: m.RouterOSVersion,
		ConfigHash:      m.ConfigHash,
		RollbackScript:  aiRawFromPtr(m.RollbackScript),
		Manifest:        aiRawFromText(m.Manifest),
		TotalSizeBytes:  m.TotalSizeBytes,
		Status:          restorepoint.Status(m.Status),
		CreatedBy:       m.CreatedBy,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
		ExpiresAt:       m.ExpiresAt,
		UsedAt:          m.UsedAt,
	}
}

// RestorePointModelFromDomain converts a restore point domain entity to a
// database row. Nilai wajib diberi default agar kolom NOT NULL + CHECK
// constraint selalu terpenuhi.
func RestorePointModelFromDomain(rp restorepoint.RestorePoint) *RestorePointModel {
	tenantID := rp.TenantID
	if tenantID == "" {
		tenantID = aiDefaultTenantID
	}
	trigger := string(rp.Trigger)
	if trigger == "" {
		trigger = string(restorepoint.TriggerMutation)
	}
	status := string(rp.Status)
	if status == "" {
		status = string(restorepoint.StatusCreating)
	}
	expiresAt := rp.ExpiresAt
	if expiresAt.IsZero() {
		// Pengaman: expires_at NOT NULL; usecase selalu mengisi retensi nyata.
		expiresAt = time.Now().UTC().Add(30 * 24 * time.Hour)
	}
	return &RestorePointModel{
		ID:              rp.ID,
		TenantID:        tenantID,
		DeviceID:        aiNullableTextID(rp.DeviceID),
		Trigger:         trigger,
		Scope:           aiEncodeStringSlice(rp.Scope),
		RouterOSVersion: rp.RouterOSVersion,
		ConfigHash:      rp.ConfigHash,
		RollbackScript:  jsonStringPtr(rp.RollbackScript),
		Manifest:        aiJSONTextOr(rp.Manifest, "{}"),
		TotalSizeBytes:  rp.TotalSizeBytes,
		Status:          status,
		CreatedBy:       rp.CreatedBy,
		CreatedAt:       rp.CreatedAt,
		UpdatedAt:       rp.UpdatedAt,
		ExpiresAt:       expiresAt,
		UsedAt:          rp.UsedAt,
	}
}

// ArtifactModel is the GORM model for restore_point_artifacts (migrasi
// 000041): satu file terenkripsi di dalam bundle restore point.
type ArtifactModel struct {
	ID             int64  `gorm:"column:id;primaryKey;autoIncrement"`
	RestorePointID string `gorm:"column:restore_point_id;type:text;not null;index"`
	Kind           string `gorm:"column:kind;type:text;not null"`
	Scope          string `gorm:"column:scope;type:jsonb;not null;default:'[]'"`
	StoragePath    string `gorm:"column:storage_path;type:text;not null"`
	SHA256         string `gorm:"column:sha256;type:text;not null"`
	SizeBytes      int64  `gorm:"column:size_bytes;not null;default:0"`

	CreatedAt time.Time `gorm:"column:created_at"`
}

// TableName returns the database table name for restore point artifacts.
func (ArtifactModel) TableName() string {
	return "restore_point_artifacts"
}

// ToDomain converts an artifact database row to its domain representation.
func (m *ArtifactModel) ToDomain() restorepoint.Artifact {
	if m == nil {
		return restorepoint.Artifact{}
	}
	return restorepoint.Artifact{
		Kind:        restorepoint.ArtifactKind(m.Kind),
		Scope:       aiDecodeStringSlice(m.Scope),
		StoragePath: m.StoragePath,
		SHA256:      m.SHA256,
		SizeBytes:   m.SizeBytes,
	}
}

// ArtifactModelFromDomain converts an artifact domain entity to a database
// row that belongs to restorePointID.
func ArtifactModelFromDomain(restorePointID string, a restorepoint.Artifact) *ArtifactModel {
	return &ArtifactModel{
		RestorePointID: restorePointID,
		Kind:           string(a.Kind),
		Scope:          aiEncodeStringSlice(a.Scope),
		StoragePath:    a.StoragePath,
		SHA256:         a.SHA256,
		SizeBytes:      a.SizeBytes,
	}
}
