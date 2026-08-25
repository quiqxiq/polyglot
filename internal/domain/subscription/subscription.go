package subscription

import "time"

// Lifecycle status values for a network subscription.
const (
	StatusPendingProvision = "PENDING_PROVISION"
	StatusActive           = "ACTIVE"
	StatusIsolated         = "ISOLATED"
	StatusSuspended        = "SUSPENDED"
	StatusTerminated       = "TERMINATED"
)

// Subscription maps one paying customer to ONE account provisioned on ONE
// Mikrotik router (PPPoE secret or permanent hotspot user).
//
// Mapping-only: the DB stores identity + relational mapping; the account's
// live details (password, assigned IP, active profile) live on the device
// and are read through the gateway when needed. The password itself is kept
// encrypted in SecretVault under key "subscription:<id>:password" — never in
// this struct.
type Subscription struct {
	ID              string     `json:"id"`
	TenantID        string     `json:"tenant_id"`
	CustomerID      string     `json:"customer_id"`
	PlanID          string     `json:"plan_id"`
	DeviceID        string     `json:"device_id,omitempty"`        // MAPPING: target router
	ServiceType     string     `json:"service_type"`               // 'PPPOE' | 'HOTSPOT'
	RemoteUsername  string     `json:"remote_username,omitempty"`  // MAPPING: akun di device
	RemoteID        string     `json:"remote_id,omitempty"`        // MAPPING: RouterOS .id
	BillingDay      int        `json:"billing_day"`
	Status          string     `json:"status"`
	StartDate       time.Time  `json:"start_date"`
	IsolatedAt      *time.Time `json:"isolated_at,omitempty"`
	IsolationReason string     `json:"isolation_reason,omitempty"`
	Notes           string     `json:"notes,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Validate enforces structural rules before persistence.
func (s Subscription) Validate() error {
	switch {
	case s.CustomerID == "":
		return ErrCustomerRequired
	case s.PlanID == "":
		return ErrPlanRequired
	case s.ServiceType != "PPPOE" && s.ServiceType != "HOTSPOT":
		return ErrInvalidServiceType
	case s.BillingDay < 1 || s.BillingDay > 28:
		return ErrInvalidBillingDay
	default:
		return nil
	}
}
