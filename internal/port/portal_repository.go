package port

import (
	"context"
	"errors"

	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
)

// Kesalahan alur OTP portal.
var (
	ErrOTPLocked   = errors.New("otp locked: too many failed attempts")
	ErrOTPExpired  = errors.New("otp expired")
	ErrOTPNotFound = errors.New("otp not found or already used")
)

// NotificationSender mengirim pesan WhatsApp secara langsung (bypass
// antrean) — dipakai OTP interaktif dan fallback worker.
type NotificationSender interface {
	Send(ctx context.Context, phone, content string) error
}

// PortalRepository menyimpan & memverifikasi sesi login portal dan OTP-nya.
type PortalRepository interface {
	// Sessions
	SaveSession(ctx context.Context, s domainCustomer.PortalSession) error
	FindValidSession(ctx context.Context, token string) (domainCustomer.PortalSession, error)
	DeleteSession(ctx context.Context, id string) error

	// OTP
	SaveOTP(ctx context.Context, o domainCustomer.PortalOTP) error
	// ConsumeOTP memverifikasi OTP terbaru belum-kedaluwarsa untuk phone:
	// cocok → tandai consumed, true; salah → attempts++, false;
	// attempts melebihi maxAttempts → ErrOTPLocked.
	ConsumeOTP(ctx context.Context, phone, codeHash string, maxAttempts int) (bool, error)
}
