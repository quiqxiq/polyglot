package plan

import "time"

// Tipe layanan (DATABASE-SCHEMA-ISP.md §2.2).
const (
	TypePPPoE     = "PPPOE"
	TypeHotspot   = "HOTSPOT"
	TypeDedicated = "DEDICATED"
)

// Mode expired mengikuti konvensi Mikhmon/RouterOS agar perilaku
// migrasi dari Mikhmon 1:1.
const (
	ExpireNotFiltered    = "ntf"  // ntf  — nonaktifkan user
	ExpireNotFilteredCom = "ntfc" // ntfc — nonaktifkan + comment
	ExpireRemove         = "rem"  // rem  — hapus user
	ExpireRemoveComment  = "remc" // remc — hapus + comment
	ExpireNone           = "0"    // 0    — tanpa auto-expire
)

// Mode validity.
const (
	ValidityCalendar = "CALENDAR"
	ValidityUptime   = "UPTIME"
)

// PPPoEPlanConfig mendefinisikan parameter profil PPPoE di MikroTik.
type PPPoEPlanConfig struct {
	RemoteAddressPool string `json:"remote_address_pool,omitempty"`
	AddressList       string `json:"address_list,omitempty"`
}

// HotspotPlanConfig mendefinisikan parameter profil Hotspot (Mikhmon/RouterOS).
type HotspotPlanConfig struct {
	IPPoolName   string  `json:"ip_pool_name,omitempty"`
	SharedUsers  int     `json:"shared_users,omitempty"`
	Validity     string  `json:"validity,omitempty"`
	ValidityMode string  `json:"validity_mode,omitempty"`
	ExpireMode   string  `json:"expire_mode,omitempty"`
	LockUser     bool    `json:"lock_user,omitempty"`
	LockServer   bool    `json:"lock_server,omitempty"`
	SellingPrice float64 `json:"selling_price,omitempty"`
	LimitUptime  string  `json:"limit_uptime,omitempty"`
	LimitBytes   string  `json:"limit_bytes,omitempty"`
}

// ServicePlan represents an ISP service offering with full MikroTik profile
// parameters (PPPoE / fixed monthly hotspot) — DATABASE-SCHEMA-ISP.md §2.2.
type ServicePlan struct {
	ID                    string    `json:"id"`
	TenantID              string    `json:"tenant_id"`
	Name                  string    `json:"name"`
	ServiceType           string    `json:"service_type"`
	BandwidthDownloadKbps int       `json:"bandwidth_download_kbps"`
	BandwidthUploadKbps   int       `json:"bandwidth_upload_kbps"`
	BurstDownloadKbps     int       `json:"burst_download_kbps,omitempty"`
	BurstUploadKbps       int       `json:"burst_upload_kbps,omitempty"`
	BurstThresholdKbps    int       `json:"burst_threshold_kbps,omitempty"`
	BurstTimeSeconds      int       `json:"burst_time_seconds,omitempty"`
	Price                 float64   `json:"price"` // harga dasar tagihan bulanan
	InstallationFee       float64   `json:"installation_fee,omitempty"`
	TaxPercent            float64   `json:"tax_percent,omitempty"`
	ParentQueue           string    `json:"parent_queue,omitempty"`
	IsActive              bool      `json:"is_active"`
	Description           string    `json:"description,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`

	// Sub-konfigurasi bertipe
	PPPoE   *PPPoEPlanConfig   `json:"pppoe,omitempty"`
	Hotspot *HotspotPlanConfig `json:"hotspot,omitempty"`

	// Flat fields dipertahankan untuk kompatibilitas DB & query
	SellingPrice      float64 `json:"selling_price,omitempty"`
	Validity          string  `json:"validity,omitempty"`
	ValidityMode      string  `json:"validity_mode,omitempty"`
	SimultaneousUse   int     `json:"simultaneous_use,omitempty"`
	IPPoolName        string  `json:"ip_pool_name,omitempty"`
	RemoteAddressPool string  `json:"remote_address_pool,omitempty"`
	AddressList       string  `json:"address_list,omitempty"`
	SharedUsers       int     `json:"shared_users,omitempty"`
	ExpireMode        string  `json:"expire_mode,omitempty"`
	LockUser          bool    `json:"lock_user,omitempty"`
	LockServer        bool    `json:"lock_server,omitempty"`
	LimitUptime       string  `json:"limit_uptime,omitempty"`
	LimitBytes        string  `json:"limit_bytes,omitempty"`
}
