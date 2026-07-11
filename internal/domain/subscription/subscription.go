package subscription

import "time"

// Subscription links a customer to a plan on a specific device. Mapped to
// the `subscriptions` table per migration 000009.
type Subscription struct {
	ID                     string
	CustomerID             string
	PlanID                 string
	ServiceType            string // "pppoe", "hotspot", "static_ip", "dhcp"
	Status                 string // "pending_install", "active", "suspended", "terminated"
	DeviceID               string
	ODPID                  *string
	ODPPort                string
	ONUSerialNumber        string
	IPPoolID               *string
	PPPoEUsername          string
	PPPoEPasswordEncrypted string
	StaticIP               string
	MACAddress             string
	InstalledAt            *time.Time
	ActivatedAt            *time.Time
	SuspendedAt            *time.Time
	TerminatedAt           *time.Time
	SuspensionReason       string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}
