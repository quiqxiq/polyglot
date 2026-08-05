package customer

import "time"

// Technician represents an operational technician who receives escalation alerts.
type Technician struct {
	ID             uint      `json:"id"`
	FullName       string    `json:"full_name"`
	Username       string    `json:"username"`
	PhoneNumber    string    `json:"phone_number"` // WhatsApp number e.g. 628123456789
	Specialization string    `json:"specialization"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
