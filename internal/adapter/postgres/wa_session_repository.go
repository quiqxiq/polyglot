package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/whatsapp"
	"github.com/quixiq/polyglot/internal/port"
)

// ErrNotFound indicates the requested record was not found.
var ErrNotFound = whatsapp.ErrNotFound

// WASessionRepository implements port.WASessionRepository for GORM/Postgres.
type WASessionRepository struct {
	db *gorm.DB
}

var _ port.WASessionRepository = (*WASessionRepository)(nil)

// NewWASessionRepository creates a new WASessionRepository.
func NewWASessionRepository(db *gorm.DB) *WASessionRepository {
	return &WASessionRepository{db: db}
}

// CreateSession inserts a new WhatsApp session into Postgres.
func (r *WASessionRepository) CreateSession(ctx context.Context, session *whatsapp.WASession) error {
	m := model.WASessionModelFromDomain(session)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	session.ID = m.ID
	return nil
}

// FindSessionByID retrieves a WhatsApp session by primary ID.
func (r *WASessionRepository) FindSessionByID(ctx context.Context, id uint) (*whatsapp.WASession, error) {
	var m model.WASessionModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

// FindAllSessions retrieves all WhatsApp sessions.
func (r *WASessionRepository) FindAllSessions(ctx context.Context) ([]whatsapp.WASession, error) {
	var models []model.WASessionModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}

	sessions := make([]whatsapp.WASession, len(models))
	for i, m := range models {
		sessions[i] = *m.ToDomain()
	}
	return sessions, nil
}

// UpdateSession saves updates to an existing WhatsApp session.
func (r *WASessionRepository) UpdateSession(ctx context.Context, session *whatsapp.WASession) error {
	m := model.WASessionModelFromDomain(session)
	// Update the existing row instead of Save: Save upserts, so an
	// asynchronous status callback that read the row just before a purge could
	// otherwise re-create the deleted session.
	res := r.db.WithContext(ctx).
		Model(m).
		Select("*").
		Omit("id").
		Updates(m)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteSession removes a WhatsApp session by ID.
func (r *WASessionRepository) DeleteSession(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&model.WASessionModel{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// PurgeSession removes a session together with its mirrored chat history in a
// single transaction. The runtime schema is created by AutoMigrate and does not
// carry the migration's ON DELETE CASCADE, so the child rows are deleted
// explicitly.
func (r *WASessionRepository) PurgeSession(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("session_id = ?", id).Delete(&model.WAMessageModel{}).Error; err != nil {
			return err
		}
		if err := tx.Where("session_id = ?", id).Delete(&model.WAChatModel{}).Error; err != nil {
			return err
		}
		res := tx.Delete(&model.WASessionModel{}, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

// ─── Backward compatibility delegations on Store ──────────────────────────────

// CreateSession inserts a new WhatsApp session into the store.
func (s *Store) CreateSession(ctx context.Context, session *whatsapp.WASession) error {
	return NewWASessionRepository(s.db).CreateSession(ctx, session)
}

// FindSessionByID retrieves a WhatsApp session from the store by ID.
func (s *Store) FindSessionByID(ctx context.Context, id uint) (*whatsapp.WASession, error) {
	return NewWASessionRepository(s.db).FindSessionByID(ctx, id)
}

// FindAllSessions retrieves all WhatsApp sessions from the store.
func (s *Store) FindAllSessions(ctx context.Context) ([]whatsapp.WASession, error) {
	return NewWASessionRepository(s.db).FindAllSessions(ctx)
}

// UpdateSession updates an existing WhatsApp session in the store.
func (s *Store) UpdateSession(ctx context.Context, session *whatsapp.WASession) error {
	return NewWASessionRepository(s.db).UpdateSession(ctx, session)
}

// PurgeSession removes a session and its mirrored chat history.
func (s *Store) PurgeSession(ctx context.Context, id uint) error {
	return NewWASessionRepository(s.db).PurgeSession(ctx, id)
}

// DeleteSession deletes a WhatsApp session from the store.
func (s *Store) DeleteSession(ctx context.Context, id uint) error {
	return NewWASessionRepository(s.db).DeleteSession(ctx, id)
}
