package bot

import "time"

// WASessionStatus defines status of WhatsApp session.
type WASessionStatus string

const (
	StatusOnline      WASessionStatus = "online"
	StatusOffline     WASessionStatus = "offline"
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
	WebhookURL   string          `json:"webhook_url"`
	ConnectedAt  time.Time       `json:"connected_at"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
