package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/registration"
)

// RegistrationEventRepository menyimpan timeline/audit perjalanan PSB
// (tabel registration_events, planning v2 §3.7).
type RegistrationEventRepository interface {
	// Append menambahkan satu event ke timeline pendaftaran.
	Append(ctx context.Context, e registration.RegistrationEvent) error
	// ListByRegistration mengembalikan timeline pendaftaran, terurut lama→baru.
	ListByRegistration(ctx context.Context, registrationID string) ([]registration.RegistrationEvent, error)
}
