package port

import (
	"context"
	"time"

	"github.com/quixiq/polyglot/internal/domain/registration"
)

// RegistrationFilter narrows registration list queries. Zero values ignored.
type RegistrationFilter struct {
	Status    string
	Phone     string
	TechID    *uint
	Scheduled bool // hanya yang punya jadwal pasang
	Limit     int
}

// RegistrationRepository defines persistence operations for the new-customer
// signup flow: review → schedule → install → convert
// (DATABASE-SCHEMA-ISP.md §2.3).
type RegistrationRepository interface {
	Save(ctx context.Context, r registration.Registration) error
	FindByID(ctx context.Context, id string) (registration.Registration, error)
	FindByNo(ctx context.Context, regNo string) (registration.Registration, error)
	List(ctx context.Context, f RegistrationFilter) ([]registration.Registration, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	// CountPendingSince counts registrations created after since — untuk dashboard.
	CountPendingSince(ctx context.Context, since time.Time) (int64, error)
}
