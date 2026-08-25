package plan

import "time"

// ServiceType enumerates supported network service kinds.
const (
	ServiceTypePPPoE   = "PPPOE"
	ServiceTypeHotspot = "HOTSPOT"
)

// Plan represents an ISP service offering (master paket layanan).
//
// Router-specific columns are NAME REFERENCES to objects that live on the
// Mikrotik device (IP pool, parent queue, address list) — never snapshots
// of router state, per the mapping-only principle in docs/database-schema.md.
// The effective rate-limit string is derived from RateDownKbps/RateUpKbps
// at provisioning time.
type Plan struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	Name          string    `json:"name"`
	ServiceType   string    `json:"service_type"`
	RateDownKbps  int       `json:"rate_down_kbps"`
	RateUpKbps    int       `json:"rate_up_kbps"`
	Price         float64   `json:"price"`
	IPPoolName    string    `json:"ip_pool_name,omitempty"`
	ParentQueue   string    `json:"parent_queue,omitempty"`
	AddressList   string    `json:"address_list,omitempty"`
	SharedUsers   int       `json:"shared_users"`
	IsActive      bool      `json:"is_active"`
	Description   string    `json:"description,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Validate enforces type-driven field rules without persisting anything.
func (p Plan) Validate() error {
	switch {
	case p.Name == "":
		return ErrNameRequired
	case p.ServiceType != ServiceTypePPPoE && p.ServiceType != ServiceTypeHotspot:
		return ErrInvalidServiceType
	case p.RateDownKbps <= 0 || p.RateUpKbps <= 0:
		return ErrInvalidRate
	case p.Price < 0:
		return ErrInvalidPrice
	default:
		return nil
	}
}
