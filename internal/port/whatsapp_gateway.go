package port

import (
	"github.com/quixiq/polyglot/internal/domain/bot"
)

// WhatsAppGateway defines the interface for WA connectivity operations.
type WhatsAppGateway interface {
	Connect(session *bot.WASession) error
	Disconnect(sessionID uint) error
	Logout(sessionID uint) error
	Reconnect(sessionID uint) error
	SendMessage(sessionID uint, to string, content string) error
	GetStatus(sessionID uint) (string, error)
	GetQRCode(sessionID uint) (string, error)
	GetPairingCode(sessionID uint, phoneNumber string) (string, error)
	RestoreAllSessions(sessions []bot.WASession) error
}
