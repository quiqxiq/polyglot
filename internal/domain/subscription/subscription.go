package subscription

import "time"

const (
	// StatusActive indicates an active subscription.
	StatusActive = "ACTIVE"
	// StatusCancelled indicates a cancelled subscription.
	StatusCancelled = "CANCELLED"
	// StatusExpired indicates an expired subscription.
	StatusExpired = "EXPIRED"
	// StatusPending indicates a pending subscription.
	StatusPending = "PENDING"
	// StatusIsolated indicates an isolated subscription.
	StatusIsolated = "ISOLATED"

	// StatusSuspended and StatusTerminated follow the ISP schema;
	// StatusExpired/Pending dipertahankan untuk kompatibilitas kode eksisting.
	// StatusSuspended indicates a suspended subscription.
	StatusSuspended = "SUSPENDED"
	// StatusTerminated indicates a terminated subscription.
	StatusTerminated = "TERMINATED"
)

// Siklus penagihan. Label Indonesia ("1 Bulan") dirender di UI,
// nilai kanonik tetap ISO-like (fix review: '1 Bulan' -> MONTHLY).
const (
	CycleMonthly = "MONTHLY"
)

// Status provisi akun ke router.
const (
	ProvisionNone    = "NONE"    // belum diprovisikan (belum ada device)
	ProvisionPending = "PENDING" // menunggu worker mencoba
	ProvisionOK      = "OK"
	ProvisionFailed  = "FAILED"
)

// PPPoESubscriptionConfig mendefinisikan parameter spesifik akun PPPoE.
type PPPoESubscriptionConfig struct {
	LocalAddress  string `json:"local_address,omitempty"`
	RemoteAddress string `json:"remote_address,omitempty"`
	CallerID      string `json:"caller_id,omitempty"`
	Routes        string `json:"routes,omitempty"`
	RateLimit     string `json:"rate_limit,omitempty"`
	RouterProfile string `json:"router_profile,omitempty"`
}

// HotspotSubscriptionConfig mendefinisikan parameter spesifik user Hotspot tetap.
type HotspotSubscriptionConfig struct {
	Server        string `json:"server,omitempty"`
	MacAddress    string `json:"mac_address,omitempty"`
	IPAddress     string `json:"ip_address,omitempty"`
	RateLimit     string `json:"rate_limit,omitempty"`
	RouterProfile string `json:"router_profile,omitempty"`
	LimitUptime   string `json:"limit_uptime,omitempty"`
	LimitBytes    string `json:"limit_bytes,omitempty"`
}

// Subscription represents an active or historic plan subscription mapped to
// a MikroTik router (PPPoE secret / hotspot user) — DATABASE-SCHEMA-ISP.md §2.5.
type Subscription struct {
	ID                 string     `json:"id"`
	TenantID           string     `json:"tenant_id"`
	CustomerID         string     `json:"customer_id"`
	PlanID             string     `json:"plan_id"`
	DeviceID           *string    `json:"device_id,omitempty"` // Router BRAS target (devices.id UUID)
	ServiceType        string     `json:"service_type"`
	RemoteUsername     string     `json:"remote_username"`
	RemotePassword     string     `json:"-"` // plaintext hanya in-memory; persist sebagai ciphertext via vault
	ProvisionStatus    string     `json:"provision_status"`
	BillingCycle       string     `json:"billing_cycle"`
	BillingDay         int        `json:"billing_day"`
	AutoIsolate        bool       `json:"auto_isolate"`
	IsolationGraceDays int        `json:"isolation_grace_days"`
	Status             string     `json:"status"`
	StartDate          time.Time  `json:"start_date"`
	EndDate            *time.Time `json:"end_date,omitempty"`
	CustomPrice        *float64   `json:"custom_price,omitempty"`
	CurrentPeriodStart *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end,omitempty"`
	Notes              string     `json:"notes,omitempty"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`

	// Sub-konfigurasi bertipe
	PPPoE   *PPPoESubscriptionConfig   `json:"pppoe,omitempty"`
	Hotspot *HotspotSubscriptionConfig `json:"hotspot,omitempty"`

	// Flat fields dipertahankan untuk kompatibilitas DB & query
	LocalAddress  string `json:"local_address,omitempty"`
	RemoteAddress string `json:"remote_address,omitempty"`
	ParentQueue   string `json:"parent_queue,omitempty"`
	RateLimit     string `json:"rate_limit,omitempty"`
	RouterProfile string `json:"router_profile,omitempty"`
}
