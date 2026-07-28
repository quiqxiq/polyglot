package plan

import "time"

// Plan is a service plan (ISP package) mapped to the `plans` table per
// migration 000005.
// Named Plan — the package itself is "plan", not "package" (reserved keyword).
type Plan struct {
	ID                  string
	Name                string
	ServiceType         string // "pppoe", "hotspot", "static_ip", "dhcp"
	Description         string
	Price               float64
	BillingPeriodMonths int
	BandwidthDownKbps   int
	BandwidthUpKbps     int
	BurstDownKbps       *int
	BurstUpKbps         *int
	BurstThresholdKbps  *int
	BurstTimeSeconds    *int
	FUPQuotaMB          *int
	FUPThrottleDownKbps *int
	FUPThrottleUpKbps   *int
	IsActive            bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
