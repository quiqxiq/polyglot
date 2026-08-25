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
	ID                 string    `json:"id"`
	TenantID           string    `json:"tenant_id"`
	Name               string    `json:"name"`
	ServiceType        string    `json:"service_type"`
	RateDownKbps       int       `json:"rate_down_kbps"`
	RateUpKbps         int       `json:"rate_up_kbps"`
	Price              float64   `json:"price"`
	IPPoolName         string    `json:"ip_pool_name,omitempty"`
	ParentQueue        string    `json:"parent_queue,omitempty"`
	AddressList        string    `json:"address_list,omitempty"`
	SharedUsers        int       `json:"shared_users"`
	BurstDownKbps      int       `json:"burst_down_kbps,omitempty"` // opsional; 0 = tanpa burst
	BurstUpKbps        int       `json:"burst_up_kbps,omitempty"`
	BurstThresholdKbps int       `json:"burst_threshold_kbps,omitempty"` // ambang burst (per arah)
	BurstTimeSeconds   int       `json:"burst_time_seconds,omitempty"`
	IsActive           bool      `json:"is_active"`
	Description        string    `json:"description,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// HasBurst reports whether burst parameters are configured (all-or-none).
func (p Plan) HasBurst() bool {
	return p.BurstDownKbps > 0 || p.BurstUpKbps > 0 || p.BurstThresholdKbps > 0 || p.BurstTimeSeconds > 0
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
	case !p.burstValid():
		return ErrInvalidBurst
	default:
		return nil
	}
}

// burstValid enforces the all-or-none rule and RouterOS sanity:
// burst rate >= base rate, threshold between base and burst, time > 0.
func (p Plan) burstValid() bool {
	set := []int{p.BurstDownKbps, p.BurstUpKbps, p.BurstThresholdKbps, p.BurstTimeSeconds}
	allSet := true
	for _, v := range set {
		if v <= 0 {
			allSet = false
		}
	}
	if !allSet {
		return !p.HasBurst() // semuanya kosong = tanpa burst (valid)
	}
	return p.BurstDownKbps >= p.RateDownKbps &&
		p.BurstUpKbps >= p.RateUpKbps &&
		p.BurstThresholdKbps >= p.RateDownKbps &&
		p.BurstThresholdKbps <= p.BurstDownKbps &&
		p.BurstThresholdKbps >= p.RateUpKbps &&
		p.BurstThresholdKbps <= p.BurstUpKbps
}
