package model

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"

	"github.com/quixiq/polyglot/internal/domain/ont"
)

// ACSDeviceModel is the GORM model for the GenieACS device cache used to
// match CPE/ONT records to the ONT inventory
// (migrasi 000037_acs.up.sql — tabel acs_devices).
type ACSDeviceModel struct {
	ID       string `gorm:"column:id;primaryKey"`
	TenantID string `gorm:"column:tenant_id;type:text;not null"`
	// Tag unique (bukan uniqueIndex) agar AutoMigrate mengenali UNIQUE
	// constraint single-column dari migrasi SQL 000037 tanpa mencoba
	// men-drop constraint bernama lain.
	ACSDeviceID     string         `gorm:"column:acs_device_id;type:varchar(120);not null;unique"`
	OUI             string         `gorm:"column:oui;type:varchar(20)"`
	ProductClass    string         `gorm:"column:product_class;type:varchar(80)"`
	SerialNumber    string         `gorm:"column:serial_number;type:varchar(80);index"`
	SoftwareVersion string         `gorm:"column:software_version;type:varchar(80)"`
	LastInformAt    *time.Time     `gorm:"column:last_inform_at"`
	Tags            pq.StringArray `gorm:"column:tags;type:text[];not null"`
	MACCandidates   pq.StringArray `gorm:"column:mac_candidates;type:text[];not null"`
	Raw             *string        `gorm:"column:raw;type:jsonb"`
	MatchedONTID    *string        `gorm:"column:matched_ont_id;type:text;index"`
	MatchSource     string         `gorm:"column:match_source;type:varchar(20)"`
	FirstSeenAt     time.Time      `gorm:"column:first_seen_at"`
	LastSeenAt      time.Time      `gorm:"column:last_seen_at"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at"`
}

// TableName returns the database table name for cached ACS devices.
func (ACSDeviceModel) TableName() string {
	return "acs_devices"
}

// ToDomain converts a cached ACS device row to its domain representation.
func (m *ACSDeviceModel) ToDomain() ont.ACSDevice {
	if m == nil {
		return ont.ACSDevice{}
	}
	var raw json.RawMessage
	if m.Raw != nil && *m.Raw != "" {
		raw = json.RawMessage(*m.Raw)
	}
	return ont.ACSDevice{
		ID:              m.ID,
		TenantID:        m.TenantID,
		ACSDeviceID:     m.ACSDeviceID,
		OUI:             m.OUI,
		ProductClass:    m.ProductClass,
		SerialNumber:    m.SerialNumber,
		SoftwareVersion: m.SoftwareVersion,
		LastInformAt:    m.LastInformAt,
		Tags:            []string(m.Tags),
		MACCandidates:   []string(m.MACCandidates),
		Raw:             raw,
		MatchedONTID:    m.MatchedONTID,
		MatchSource:     m.MatchSource,
		FirstSeenAt:     m.FirstSeenAt,
		LastSeenAt:      m.LastSeenAt,
	}
}

// ACSDeviceModelFromDomain converts an ACS device domain entity to a database
// row. Tenant default disamakan dengan repository lainnya.
func ACSDeviceModelFromDomain(d ont.ACSDevice) *ACSDeviceModel {
	tenantID := d.TenantID
	if tenantID == "" {
		tenantID = "tenant-default"
	}
	tags := pq.StringArray(d.Tags)
	if tags == nil {
		tags = pq.StringArray{}
	}
	macCandidates := pq.StringArray(d.MACCandidates)
	if macCandidates == nil {
		macCandidates = pq.StringArray{}
	}
	return &ACSDeviceModel{
		ID:              d.ID,
		TenantID:        tenantID,
		ACSDeviceID:     d.ACSDeviceID,
		OUI:             d.OUI,
		ProductClass:    d.ProductClass,
		SerialNumber:    d.SerialNumber,
		SoftwareVersion: d.SoftwareVersion,
		LastInformAt:    d.LastInformAt,
		Tags:            tags,
		MACCandidates:   macCandidates,
		Raw:             jsonStringPtr(d.Raw),
		MatchedONTID:    d.MatchedONTID,
		MatchSource:     d.MatchSource,
		FirstSeenAt:     d.FirstSeenAt,
		LastSeenAt:      d.LastSeenAt,
	}
}

// ACSPresetModel is the GORM model for managed GenieACS presets
// (migrasi 000037_acs.up.sql — tabel acs_presets). Unik per (tenant, name).
type ACSPresetModel struct {
	ID             string     `gorm:"column:id;primaryKey"`
	TenantID       string     `gorm:"column:tenant_id;type:text;not null;uniqueIndex:uq_acs_presets_tenant_name"`
	Name           string     `gorm:"column:name;type:varchar(100);not null;uniqueIndex:uq_acs_presets_tenant_name"`
	Precondition   *string    `gorm:"column:precondition;type:jsonb"`
	Configurations string     `gorm:"column:configurations;type:jsonb;not null"`
	Enabled        bool       `gorm:"column:enabled;not null"`
	LastSyncedAt   *time.Time `gorm:"column:last_synced_at"`
	LastError      string     `gorm:"column:last_error;type:text"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

// TableName returns the database table name for ACS presets.
func (ACSPresetModel) TableName() string {
	return "acs_presets"
}

// ToDomain converts an ACS preset row to its domain representation.
func (m *ACSPresetModel) ToDomain() ont.ACSPreset {
	if m == nil {
		return ont.ACSPreset{}
	}
	var precondition json.RawMessage
	if m.Precondition != nil && *m.Precondition != "" {
		precondition = json.RawMessage(*m.Precondition)
	}
	var configurations json.RawMessage
	if m.Configurations != "" {
		configurations = json.RawMessage(m.Configurations)
	}
	return ont.ACSPreset{
		ID:             m.ID,
		TenantID:       m.TenantID,
		Name:           m.Name,
		Precondition:   precondition,
		Configurations: configurations,
		Enabled:        m.Enabled,
		LastSyncedAt:   m.LastSyncedAt,
		LastError:      m.LastError,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

// ACSPresetModelFromDomain converts an ACS preset domain entity to a database
// row. configurations NOT NULL sehingga kosong dinormalisasi ke '[]'.
func ACSPresetModelFromDomain(p ont.ACSPreset) *ACSPresetModel {
	tenantID := p.TenantID
	if tenantID == "" {
		tenantID = "tenant-default"
	}
	configurations := "[]"
	if len(p.Configurations) > 0 {
		configurations = string(p.Configurations)
	}
	return &ACSPresetModel{
		ID:             p.ID,
		TenantID:       tenantID,
		Name:           p.Name,
		Precondition:   jsonStringPtr(p.Precondition),
		Configurations: configurations,
		Enabled:        p.Enabled,
		LastSyncedAt:   p.LastSyncedAt,
		LastError:      p.LastError,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
