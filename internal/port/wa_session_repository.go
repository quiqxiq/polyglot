package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/whatsapp"
)

// WASessionRepository defines the persistence interface for WA sessions.
type WASessionRepository interface {
	CreateSession(ctx context.Context, session *whatsapp.WASession) error
	FindSessionByID(ctx context.Context, id uint) (*whatsapp.WASession, error)
	FindAllSessions(ctx context.Context) ([]whatsapp.WASession, error)
	UpdateSession(ctx context.Context, session *whatsapp.WASession) error
	DeleteSession(ctx context.Context, id uint) error
	// PurgeSession removes a session and its mirrored chat history atomically.
	PurgeSession(ctx context.Context, id uint) error
}
