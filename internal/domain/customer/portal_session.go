package customer

import "time"

// PortalSession represents an authenticated self-service portal session
// for a customer (DATABASE-SCHEMA-ISP.md §2.4 — customer_portal_sessions).
type PortalSession struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	CustomerID   string    `json:"customer_id"`
	SessionToken string    `json:"-"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}
