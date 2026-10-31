package model

import (
	"encoding/json"
	"time"

	"github.com/quixiq/polyglot/internal/domain/ont"
)

// ONTModel is the GORM model for ONT/ONU inventory rows per OLT
// (migrasi 000036_olt_ont.up.sql — tabel onts). Identitas ONT adalah MAC
// LLID ternormalisasi yang unik per OLT (D5 PLAN-ACS-OLT-HIOSO.md).
type ONTModel struct {
	ID          string `gorm:"column:id;primaryKey"`
	TenantID    string `gorm:"column:tenant_id;type:text;not null"`
	OLTDeviceID string `gorm:"column:olt_device_id;type:uuid;not null;index"`
	PONIndex    int    `gorm:"column:pon_index;not null"`
	ONUIndex    int    `gorm:"column:onu_index;not null"`
	LLID        *int   `gorm:"column:llid"`
	Name        string `gorm:"column:name;type:varchar(100)"`

	MACAddress   string `gorm:"column:mac_address;type:varchar(20);not null"`
	SerialNumber string `gorm:"column:serial_number;type:varchar(50)"`
	Model        string `gorm:"column:model;type:varchar(50)"`
	Firmware     string `gorm:"column:firmware;type:varchar(50)"`
	State        string `gorm:"column:state;type:varchar(20);not null;index"`
	AdminState   string `gorm:"column:admin_state;type:varchar(10);not null"`

	RxPowerDBM   *float64 `gorm:"column:rx_power_dbm;type:numeric(6,2)"`
	TxPowerDBM   *float64 `gorm:"column:tx_power_dbm;type:numeric(6,2)"`
	DistanceM    *int     `gorm:"column:distance_m"`
	TemperatureC *float64 `gorm:"column:temperature_c;type:numeric(5,1)"`

	// Counter sengaja tanpa tag default agar nilai nol (reset strike setelah
	// recovery) tetap ikut ter-upsert oleh Save.
	MissedPolls    int `gorm:"column:missed_polls;not null"`
	OfflineStrikes int `gorm:"column:offline_strikes;not null"`
	RxNAStrikes    int `gorm:"column:rx_na_strikes;not null"`

	LastSeenAt   *time.Time `gorm:"column:last_seen_at"`
	LastPolledAt *time.Time `gorm:"column:last_polled_at"`

	CustomerID     *string `gorm:"column:customer_id;type:text;index"`
	SubscriptionID *string `gorm:"column:subscription_id;type:text;index"`
	PPPoEUsername  string  `gorm:"column:pppoe_username;type:varchar(100)"`

	// Kolom integrasi ACS (000036): acs_matched_by adalah asal pencocokan
	// (SERIAL/MAC/MAC_NEIGHBOR/MANUAL).
	ACSDeviceID    *string    `gorm:"column:acs_device_id;type:varchar(120);index"`
	ACSMatchSource string     `gorm:"column:acs_matched_by;type:varchar(20)"`
	ACSLastInform  *time.Time `gorm:"column:acs_last_inform_at"`
	ACSTelemetry   *string    `gorm:"column:acs_telemetry;type:jsonb"`
	WLANSlots      int        `gorm:"column:wlan_slots;not null"`

	Source    string    `gorm:"column:source;type:varchar(10);not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName returns the database table name for ONT inventory rows.
func (ONTModel) TableName() string {
	return "onts"
}

// ToDomain converts an ONT database row to its domain representation.
func (m *ONTModel) ToDomain() ont.ONT {
	if m == nil {
		return ont.ONT{}
	}
	var telemetry json.RawMessage
	if m.ACSTelemetry != nil && *m.ACSTelemetry != "" {
		telemetry = json.RawMessage(*m.ACSTelemetry)
	}
	return ont.ONT{
		ID:             m.ID,
		TenantID:       m.TenantID,
		OLTDeviceID:    m.OLTDeviceID,
		PONIndex:       m.PONIndex,
		ONUIndex:       m.ONUIndex,
		LLID:           m.LLID,
		Name:           m.Name,
		MACAddress:     m.MACAddress,
		SerialNumber:   m.SerialNumber,
		Model:          m.Model,
		Firmware:       m.Firmware,
		State:          ont.State(m.State),
		AdminState:     ont.AdminState(m.AdminState),
		RxPowerDBM:     m.RxPowerDBM,
		TxPowerDBM:     m.TxPowerDBM,
		DistanceM:      m.DistanceM,
		TemperatureC:   m.TemperatureC,
		MissedPolls:    m.MissedPolls,
		OfflineStrikes: m.OfflineStrikes,
		RxNAStrikes:    m.RxNAStrikes,
		LastSeenAt:     m.LastSeenAt,
		LastPolledAt:   m.LastPolledAt,
		CustomerID:     m.CustomerID,
		SubscriptionID: m.SubscriptionID,
		PPPoEUsername:  m.PPPoEUsername,
		ACSDeviceID:    m.ACSDeviceID,
		ACSMatchSource: m.ACSMatchSource,
		ACSLastInform:  m.ACSLastInform,
		ACSTelemetry:   telemetry,
		WLANSlots:      m.WLANSlots,
		Source:         m.Source,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

// ONTModelFromDomain converts an ONT domain entity to a database row.
// Tenant, state, admin state, dan source diberi default agar kolom NOT NULL
// selalu terisi tanpa bergantung pada default SQL.
func ONTModelFromDomain(o ont.ONT) *ONTModel {
	tenantID := o.TenantID
	if tenantID == "" {
		tenantID = "tenant-default"
	}
	state := o.State
	if state == "" {
		state = ont.StateUnknown
	}
	adminState := o.AdminState
	if adminState == "" {
		adminState = ont.AdminEnable
	}
	source := o.Source
	if source == "" {
		source = "SNMP"
	}
	return &ONTModel{
		ID:             o.ID,
		TenantID:       tenantID,
		OLTDeviceID:    o.OLTDeviceID,
		PONIndex:       o.PONIndex,
		ONUIndex:       o.ONUIndex,
		LLID:           o.LLID,
		Name:           o.Name,
		MACAddress:     o.MACAddress,
		SerialNumber:   o.SerialNumber,
		Model:          o.Model,
		Firmware:       o.Firmware,
		State:          string(state),
		AdminState:     string(adminState),
		RxPowerDBM:     o.RxPowerDBM,
		TxPowerDBM:     o.TxPowerDBM,
		DistanceM:      o.DistanceM,
		TemperatureC:   o.TemperatureC,
		MissedPolls:    o.MissedPolls,
		OfflineStrikes: o.OfflineStrikes,
		RxNAStrikes:    o.RxNAStrikes,
		LastSeenAt:     o.LastSeenAt,
		LastPolledAt:   o.LastPolledAt,
		CustomerID:     o.CustomerID,
		SubscriptionID: o.SubscriptionID,
		PPPoEUsername:  o.PPPoEUsername,
		ACSDeviceID:    o.ACSDeviceID,
		ACSMatchSource: o.ACSMatchSource,
		ACSLastInform:  o.ACSLastInform,
		ACSTelemetry:   jsonStringPtr(o.ACSTelemetry),
		WLANSlots:      o.WLANSlots,
		Source:         source,
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
	}
}

// ONTEventModel is the GORM model for the per-ONT timeline
// (migrasi 000036_olt_ont.up.sql — tabel ont_events).
type ONTEventModel struct {
	ID        string    `gorm:"column:id;primaryKey"`
	TenantID  string    `gorm:"column:tenant_id;type:text;not null"`
	ONTID     string    `gorm:"column:ont_id;type:text;not null;index"`
	EventType string    `gorm:"column:event_type;type:varchar(40);not null"`
	OldState  string    `gorm:"column:old_state;type:varchar(20)"`
	NewState  string    `gorm:"column:new_state;type:varchar(20)"`
	Message   string    `gorm:"column:message;type:text"`
	Metadata  *string   `gorm:"column:metadata;type:jsonb"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

// TableName returns the database table name for ONT timeline entries.
func (ONTEventModel) TableName() string {
	return "ont_events"
}

// ToDomain converts an ONT timeline row to its domain representation.
func (m *ONTEventModel) ToDomain() ont.Event {
	if m == nil {
		return ont.Event{}
	}
	var metadata json.RawMessage
	if m.Metadata != nil && *m.Metadata != "" {
		metadata = json.RawMessage(*m.Metadata)
	}
	return ont.Event{
		ID:        m.ID,
		TenantID:  m.TenantID,
		ONTID:     m.ONTID,
		EventType: m.EventType,
		OldState:  m.OldState,
		NewState:  m.NewState,
		Message:   m.Message,
		Metadata:  metadata,
		CreatedAt: m.CreatedAt,
	}
}

// ONTEventModelFromDomain converts an ONT event domain entity to a database row.
func ONTEventModelFromDomain(e ont.Event) *ONTEventModel {
	tenantID := e.TenantID
	if tenantID == "" {
		tenantID = "tenant-default"
	}
	return &ONTEventModel{
		ID:        e.ID,
		TenantID:  tenantID,
		ONTID:     e.ONTID,
		EventType: e.EventType,
		OldState:  e.OldState,
		NewState:  e.NewState,
		Message:   e.Message,
		Metadata:  jsonStringPtr(e.Metadata),
		CreatedAt: e.CreatedAt,
	}
}

// ONTMetricSampleModel is the GORM model for ONT optical metric samples
// (migrasi 000036_olt_ont.up.sql — tabel ont_metric_samples).
type ONTMetricSampleModel struct {
	ID           string    `gorm:"column:id;primaryKey"`
	ONTID        string    `gorm:"column:ont_id;type:text;not null;index"`
	SampledAt    time.Time `gorm:"column:sampled_at;not null"`
	RxPowerDBM   *float64  `gorm:"column:rx_power_dbm;type:numeric(6,2)"`
	TxPowerDBM   *float64  `gorm:"column:tx_power_dbm;type:numeric(6,2)"`
	TemperatureC *float64  `gorm:"column:temperature_c;type:numeric(5,1)"`
	DistanceM    *int      `gorm:"column:distance_m"`
	Online       bool      `gorm:"column:online;not null"`
}

// TableName returns the database table name for ONT metric samples.
func (ONTMetricSampleModel) TableName() string {
	return "ont_metric_samples"
}

// ToDomain converts an ONT metric sample row to its domain representation.
func (m *ONTMetricSampleModel) ToDomain() ont.MetricSample {
	if m == nil {
		return ont.MetricSample{}
	}
	return ont.MetricSample{
		ID:           m.ID,
		ONTID:        m.ONTID,
		SampledAt:    m.SampledAt,
		RxPowerDBM:   m.RxPowerDBM,
		TxPowerDBM:   m.TxPowerDBM,
		TemperatureC: m.TemperatureC,
		DistanceM:    m.DistanceM,
		Online:       m.Online,
	}
}

// ONTMetricSampleModelFromDomain converts an ONT metric sample domain entity
// to a database row.
func ONTMetricSampleModelFromDomain(s ont.MetricSample) *ONTMetricSampleModel {
	return &ONTMetricSampleModel{
		ID:           s.ID,
		ONTID:        s.ONTID,
		SampledAt:    s.SampledAt,
		RxPowerDBM:   s.RxPowerDBM,
		TxPowerDBM:   s.TxPowerDBM,
		TemperatureC: s.TemperatureC,
		DistanceM:    s.DistanceM,
		Online:       s.Online,
	}
}

// jsonStringPtr mengubah JSON mentah menjadi pointer string untuk kolom
// JSONB nullable (nilai kosong → NULL, bukan string kosong yang bukan JSON).
func jsonStringPtr(raw json.RawMessage) *string {
	if len(raw) == 0 {
		return nil
	}
	s := string(raw)
	return &s
}
