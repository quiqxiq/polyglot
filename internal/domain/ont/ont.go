// Package ont memodelkan ONT/ONU pelanggan pada OLT EPON: identitas (MAC
// LLID), status online hasil polling, kualitas optik, dan relasi ke
// layanan/pelanggan (PLAN-ACS-OLT-HIOSO.md).
package ont

import (
	"encoding/json"
	"time"
)

// State status online ONT.
type State string

const (
	StateOnline  State = "ONLINE"
	StateOffline State = "OFFLINE"
	StateUnknown State = "UNKNOWN"
	StateRemoved State = "REMOVED"
)

// AdminState status administrasi ONT di OLT.
type AdminState string

const (
	AdminEnable  AdminState = "ENABLE"
	AdminDisable AdminState = "DISABLE"
)

// MatchSource asal pencocokan ACS device ke ONT.
type MatchSource string

const (
	MatchSerial      MatchSource = "SERIAL"
	MatchMAC         MatchSource = "MAC"
	MatchMACNeighbor MatchSource = "MAC_NEIGHBOR"
	MatchManual      MatchSource = "MANUAL"
)

// ONT adalah satu perangkat ONU/ONT yang terdaftar di OLT.
type ONT struct {
	ID             string     `json:"id"`
	TenantID       string     `json:"tenant_id"`
	OLTDeviceID    string     `json:"olt_device_id"`
	PONIndex       int        `json:"pon_index"`
	ONUIndex       int        `json:"onu_index"`
	LLID           *int       `json:"llid,omitempty"`
	Name           string     `json:"name,omitempty"`
	MACAddress     string     `json:"mac_address"`
	SerialNumber   string     `json:"serial_number,omitempty"`
	Model          string     `json:"model,omitempty"`
	Firmware       string     `json:"firmware,omitempty"`
	State          State      `json:"state"`
	AdminState     AdminState `json:"admin_state"`
	RxPowerDBM     *float64   `json:"rx_power_dbm,omitempty"`
	TxPowerDBM     *float64   `json:"tx_power_dbm,omitempty"`
	DistanceM      *int       `json:"distance_m,omitempty"`
	TemperatureC   *float64   `json:"temperature_c,omitempty"`
	MissedPolls    int        `json:"missed_polls"`
	OfflineStrikes int        `json:"offline_strikes"`
	RxNAStrikes    int        `json:"rx_na_strikes"`
	LastSeenAt     *time.Time `json:"last_seen_at,omitempty"`
	LastPolledAt   *time.Time `json:"last_polled_at,omitempty"`

	CustomerID     *string `json:"customer_id,omitempty"`
	SubscriptionID *string `json:"subscription_id,omitempty"`
	PPPoEUsername  string  `json:"pppoe_username,omitempty"`

	ACSDeviceID    *string         `json:"acs_device_id,omitempty"`
	ACSMatchSource string          `json:"acs_match_source,omitempty"`
	ACSLastInform  *time.Time      `json:"acs_last_inform_at,omitempty"`
	ACSTelemetry   json.RawMessage `json:"acs_telemetry,omitempty"`
	WLANSlots      int             `json:"wlan_slots"`

	Source    string    `json:"source,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ISOLinked melaporkan apakah ONT sudah tertaut ke langganan/pelanggan.
func (o ONT) ISOLinked() bool {
	return o.SubscriptionID != nil || o.CustomerID != nil
}

// Event adalah satu entri timeline ONT.
type Event struct {
	ID        string          `json:"id"`
	TenantID  string          `json:"tenant_id"`
	ONTID     string          `json:"ont_id"`
	EventType string          `json:"event_type"`
	OldState  string          `json:"old_state,omitempty"`
	NewState  string          `json:"new_state,omitempty"`
	Message   string          `json:"message,omitempty"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

// MetricSample adalah satu sampel metrik optik ONT.
type MetricSample struct {
	ID           string    `json:"id"`
	ONTID        string    `json:"ont_id"`
	SampledAt    time.Time `json:"sampled_at"`
	RxPowerDBM   *float64  `json:"rx_power_dbm,omitempty"`
	TxPowerDBM   *float64  `json:"tx_power_dbm,omitempty"`
	TemperatureC *float64  `json:"temperature_c,omitempty"`
	DistanceM    *int      `json:"distance_m,omitempty"`
	Online       bool      `json:"online"`
}
