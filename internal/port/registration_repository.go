package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/registration"
)

// RegistrationRepository persists new-subscriber application forms.
type RegistrationRepository interface {
	Save(ctx context.Context, r registration.Registration) error
	FindByID(ctx context.Context, id string) (registration.Registration, error)
	List(ctx context.Context, status string, limit int) ([]registration.Registration, error)
	// HasActiveByPhone reports whether a PENDING/APPROVED registration
	// already exists for the given phone number (duplicate guard).
	HasActiveByPhone(ctx context.Context, phone string) (bool, error)
}
