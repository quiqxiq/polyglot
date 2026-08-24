package customer

import "time"

// PortalOTP adalah kode verifikasi sekali pakai untuk login portal,
// disimpan sebagai hash sha256 — bukan plaintext.
type PortalOTP struct {
	ID         string     `json:"id"`
	TenantID   string     `json:"tenant_id"`
	Phone      string     `json:"phone"`
	CodeHash   string     `json:"-"`
	Purpose    string     `json:"purpose"` // PORTAL_LOGIN
	Attempts   int        `json:"attempts"`
	ExpiresAt  time.Time  `json:"expires_at"`
	ConsumedAt *time.Time `json:"consumed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}
