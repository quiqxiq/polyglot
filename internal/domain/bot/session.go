package bot

import "time"

// WASessionStatus defines status of WhatsApp session.
type WASessionStatus string

const (
	// StatusOnline berarti client terhubung DAN sudah ter-pair (Store.ID ada).
	StatusOnline WASessionStatus = "online"
	// StatusOffline berarti client terdaftar tapi tidak ada koneksi aktif.
	StatusOffline WASessionStatus = "offline"
	// StatusConnecting berarti client sedang dial ke server WhatsApp.
	StatusConnecting WASessionStatus = "connecting"
	// StatusNeedsRescan berarti client belum ter-pair / perlu scan QR ulang
	// (session baru, logout, atau remote logged-out).
	StatusNeedsRescan WASessionStatus = "needs_rescan"
)

// WASession represents a connected WhatsApp session/device in the domain layer.
type WASession struct {
	ID           uint            `json:"id"`
	DeviceName   string          `json:"device_name"`
	PhoneNumber  string          `json:"phone_number"`
	JID          string          `json:"jid"`
	Status       WASessionStatus `json:"status"`
	IsBotEnabled bool            `json:"is_bot_enabled"`
	ConnectedAt  time.Time       `json:"connected_at"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
