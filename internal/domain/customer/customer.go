package customer

import "time"

// Status pelanggan aktif (DATABASE-SCHEMA-ISP.md §2.4).
const (
	StatusActive     = "ACTIVE"
	StatusIsolated   = "ISOLATED"
	StatusSuspended  = "SUSPENDED"
	StatusTerminated = "TERMINATED"
)

// Customer represents an ISP subscriber entity.
type Customer struct {
	ID               string     `json:"id"`
	TenantID         string     `json:"tenant_id"`
	CustomerCode     string     `json:"customer_code"`
	Name             string     `json:"name"`
	Phone            string     `json:"phone"`
	Email            string     `json:"email,omitempty"`
	Address          string     `json:"address"`
	Latitude         *float64   `json:"latitude,omitempty"`
	Longitude        *float64   `json:"longitude,omitempty"`
	PortalAccessCode string     `json:"portal_access_code"`
	Status           string     `json:"status"`
	Notes            string     `json:"notes,omitempty"`
	RegisteredAt     time.Time  `json:"registered_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
