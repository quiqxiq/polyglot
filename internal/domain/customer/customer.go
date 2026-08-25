package customer

import "time"

// Customer is an ISP subscriber identity record. It carries NO network
// account data — those live on the Mikrotik device and are mapped through
// subscription.Subscription (mapping-only principle).
type Customer struct {
	ID           string     `json:"id"`
	TenantID     string     `json:"tenant_id"`
	CustomerCode string     `json:"customer_code,omitempty"`
	Name         string     `json:"name"`
	Email        string     `json:"email,omitempty"`
	Phone        string     `json:"phone"`
	Address      string     `json:"address,omitempty"`
	Latitude     float64    `json:"latitude,omitempty"`
	Longitude    float64    `json:"longitude,omitempty"`
	Status       string     `json:"status"`
	Notes        string     `json:"notes,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// Validate enforces the concise intake rules (nama + WA wajib).
func (c Customer) Validate() error {
	switch {
	case c.Name == "":
		return ErrNameRequired
	case c.Phone == "":
		return ErrPhoneRequired
	default:
		return nil
	}
}
