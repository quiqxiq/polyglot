package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/customer"
)

// UserRepository defines the persistence contract for User accounts.
type UserRepository interface {
	Create(ctx context.Context, user *customer.User) error
	FindByID(ctx context.Context, id uint) (*customer.User, error)
	FindByUsername(ctx context.Context, username string) (*customer.User, error)
	FindByEmail(ctx context.Context, email string) (*customer.User, error)
	FindByPhone(ctx context.Context, phone string) (*customer.User, error)
	// FindAllByPhone returns every user with an exact phone match. It returns an
	// empty slice (not an error) when no row matches, so callers can detect
	// ambiguous phone ownership instead of silently picking one account.
	FindAllByPhone(ctx context.Context, phone string) ([]*customer.User, error)
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context, page, pageSize int, search string) ([]*customer.User, int64, error)
	FindAll(ctx context.Context) ([]*customer.User, error)
	FindByRoles(ctx context.Context, roles []string, activeOnly bool) ([]*customer.User, error)
	Update(ctx context.Context, user *customer.User) error
	Delete(ctx context.Context, id uint) error
	UpdatePassword(ctx context.Context, id uint, passwordHash string) error
	UpdateStatus(ctx context.Context, id uint, isActive bool) error
	AssignDevices(ctx context.Context, userID uint, deviceIDs []string, assignedBy *uint) error
	GetAssignedDeviceIDs(ctx context.Context, userID uint) ([]string, error)
	IsDeviceAccessibleByUser(ctx context.Context, userID uint, deviceID string) (bool, error)
}
