package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/bot"
)

// WhatsAppGateway defines the interface for WA connectivity operations.
//
// Semantic kontrak:
//   - Connect    : membuat/menghidupkan client & koneksi. Untuk session yang
//     belum ter-pair, memulai QR flow (QR disimpan di cache client, dibaca
//     via GetQRCode / di-broadcast via status callback).
//   - Disconnect : HANYA memutus koneksi. Session lokal & pairing dipertahankan.
//   - Logout     : unlink dari WhatsApp + hapus session lokal (store rows).
//     Slot di DB TETAP (re-pairable di bawah id yang sama).
//   - Purge      : logout + hapus session lokal. Penghapusan baris DB dan
//     mirror chat dilakukan oleh caller (handler) karena gateway tidak
//     memiliki akses ke repository session.
type WhatsAppGateway interface {
	Connect(session *bot.WASession) error
	Disconnect(sessionID uint) error
	Logout(sessionID uint) error
	Purge(sessionID uint) error
	Reconnect(sessionID uint) error
	SendMessage(sessionID uint, to string, content string) error
	SendDocument(ctx context.Context, sessionID uint, to string, fileBytes []byte, fileName string, contentType string, caption string) error
	SendImage(ctx context.Context, sessionID uint, to string, imageBytes []byte, contentType string, caption string) error
	GetStatus(sessionID uint) (string, error)
	GetQRCode(sessionID uint) (string, error)
	GetPairingCode(sessionID uint, phoneNumber string) (string, error)
	RestoreAllSessions(sessions []bot.WASession) error
}
