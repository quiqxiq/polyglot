package port

import (
	"context"

	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
)

// Error sentinels untuk alur OTP portal tinggal di
// internal/domain/customer/errors.go (ErrOTPLocked, ErrOTPExpired,
// ErrOTPNotFound) per DEVELOPMENT-GUIDELINES.md §6.

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
