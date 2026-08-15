package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/auth"
)

// UserRepository provides persistence access for system user accounts.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*auth.User, error)
	FindByID(ctx context.Context, id uint) (*auth.User, error)
	Save(ctx context.Context, user *auth.User) error
	Count(ctx context.Context) (int64, error)
}
