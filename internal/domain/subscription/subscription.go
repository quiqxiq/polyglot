package subscription

import "time"

const (
	StatusActive    = "ACTIVE"
	StatusCancelled = "CANCELLED"
	StatusExpired   = "EXPIRED"
	StatusPending   = "PENDING"
	StatusIsolated  = "ISOLATED"

	// StatusIsolated/Suspended/Terminated mengikuti skema ISP;
	// StatusExpired/Pending dipertahankan untuk kompatibilitas kode eksisting.
	StatusSuspended  = "SUSPENDED"
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
	LocalAddress       string     `json:"local_address,omitempty"`
	RemoteAddress      string     `json:"remote_address,omitempty"`
	ParentQueue        string     `json:"parent_queue,omitempty"`
	RateLimit          string     `json:"rate_limit,omitempty"`
	RouterProfile      string     `json:"router_profile,omitempty"` // profil paket aktif di router (basis restore)
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
}
