package customer

import "time"

// Customer is the stored record for a subscriber/customer, mapped to the
// `customers` table per migration 000008.
// Named Customer — the package is "customer", so callers write customer.Customer.
type Customer struct {
	ID             string
	FullName       string
	IDNumber       string
	Phone          string
	WhatsApp       string
	Email          string
	Address        string
	LocationLat    *float64
	LocationLng    *float64
	CustomerType   string // "residential" or "business"
	Status         string // "prospect", "active", "suspended", "terminated"
	ReferralSource string
	Notes          string
	RegisteredAt   time.Time
	TerminatedAt   *time.Time
}
