package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/port"
)

// ErrNotFound indicates the requested bot record was not found.
var ErrNotFound = bot.ErrNotFound

// WASessionRepository implements port.WASessionRepository for GORM/Postgres.
type WASessionRepository struct {
	db *gorm.DB
}

var _ port.WASessionRepository = (*WASessionRepository)(nil)

// NewWASessionRepository creates a new WASessionRepository.
func NewWASessionRepository(db *gorm.DB) *WASessionRepository {
	return &WASessionRepository{db: db}
}

func (r *WASessionRepository) CreateSession(ctx context.Context, session *bot.WASession) error {
	m := model.WASessionModelFromDomain(session)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	session.ID = m.ID
	return nil
}

func (r *WASessionRepository) FindSessionByID(ctx context.Context, id uint) (*bot.WASession, error) {
	var m model.WASessionModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *WASessionRepository) FindAllSessions(ctx context.Context) ([]bot.WASession, error) {
	var mList []model.WASessionModel
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&mList).Error; err != nil {
		return nil, err
	}
	var res []bot.WASession
	for _, m := range mList {
		if d := m.ToDomain(); d != nil {
			res = append(res, *d)
		}
	}
	return res, nil
}

func (r *WASessionRepository) UpdateSession(ctx context.Context, session *bot.WASession) error {
	m := model.WASessionModelFromDomain(session)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *WASessionRepository) DeleteSession(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.WASessionModel{}, id).Error
}

// ConversationRepository implements port.ConversationRepository for GORM/Postgres.
type ConversationRepository struct {
	db *gorm.DB
}

var _ port.ConversationRepository = (*ConversationRepository)(nil)

// NewConversationRepository creates a new ConversationRepository.
func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) CreateConversation(ctx context.Context, conv *bot.Conversation) error {
	m := model.ConversationModelFromDomain(conv)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	conv.ID = m.ID
	return nil
}

func (r *ConversationRepository) FindConversationByID(ctx context.Context, id uint) (*bot.Conversation, error) {
	var m model.ConversationModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *ConversationRepository) FindConversationByIDWithMessages(ctx context.Context, id uint) (*bot.Conversation, error) {
	var m model.ConversationModel
	if err := r.db.WithContext(ctx).Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *ConversationRepository) FindConversationsBySessionID(ctx context.Context, sessionID uint) ([]bot.Conversation, error) {
	var mList []model.ConversationModel
	if err := r.db.WithContext(ctx).Where("session_id = ?", sessionID).Order("updated_at DESC").Find(&mList).Error; err != nil {
		return nil, err
	}
	var res []bot.Conversation
	for _, m := range mList {
		if d := m.ToDomain(); d != nil {
			res = append(res, *d)
		}
	}
	return res, nil
}

func (r *ConversationRepository) FindConversationsByStatus(ctx context.Context, status bot.ConversationStatus) ([]bot.Conversation, error) {
	var mList []model.ConversationModel
	if err := r.db.WithContext(ctx).Where("status = ?", string(status)).Order("updated_at DESC").Find(&mList).Error; err != nil {
		return nil, err
	}
	var res []bot.Conversation
	for _, m := range mList {
		if d := m.ToDomain(); d != nil {
			res = append(res, *d)
		}
	}
	return res, nil
}

func (r *ConversationRepository) FindAllConversations(ctx context.Context) ([]bot.Conversation, error) {
	var mList []model.ConversationModel
	if err := r.db.WithContext(ctx).Order("updated_at DESC").Find(&mList).Error; err != nil {
		return nil, err
	}
	var res []bot.Conversation
	for _, m := range mList {
		if d := m.ToDomain(); d != nil {
			res = append(res, *d)
		}
	}
	return res, nil
}

func (r *ConversationRepository) FindActiveConversationByCustomer(ctx context.Context, sessionID uint, customerNumber string) (*bot.Conversation, error) {
	var m model.ConversationModel
	query := r.db.WithContext(ctx).Where("customer_wa_number = ? AND status IN ('bot', 'escalation')", customerNumber)
	if sessionID > 0 {
		query = query.Where("session_id = ?", sessionID)
	}
	if err := query.Order("updated_at DESC").First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *ConversationRepository) UpdateConversation(ctx context.Context, conv *bot.Conversation) error {
	m := model.ConversationModelFromDomain(conv)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *ConversationRepository) CreateMessage(ctx context.Context, msg *bot.Message) error {
	m := model.MessageModelFromDomain(msg)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	msg.ID = m.ID

	_ = r.db.WithContext(ctx).Model(&model.ConversationModel{}).Where("id = ?", msg.ConversationID).Update("updated_at", msg.CreatedAt)
	return nil
}

func (r *ConversationRepository) FindMessagesByConversationID(ctx context.Context, conversationID uint) ([]bot.Message, error) {
	var mList []model.MessageModel
	if err := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).Order("created_at ASC").Find(&mList).Error; err != nil {
		return nil, err
	}
	var res []bot.Message
	for _, m := range mList {
		if d := m.ToDomain(); d != nil {
			res = append(res, *d)
		}
	}
	return res, nil
}

func (r *ConversationRepository) FindRecentMessages(ctx context.Context, conversationID uint, limit int) ([]bot.Message, error) {
	var mList []model.MessageModel
	if err := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).Order("created_at DESC").Limit(limit).Find(&mList).Error; err != nil {
		return nil, err
	}
	var res []bot.Message
	for i := len(mList) - 1; i >= 0; i-- {
		if d := mList[i].ToDomain(); d != nil {
			res = append(res, *d)
		}
	}
	return res, nil
}

// ─── Backward compatibility delegations on Store ──────────────────────────────

func (s *Store) CreateSession(ctx context.Context, session *bot.WASession) error {
	return NewWASessionRepository(s.db).CreateSession(ctx, session)
}

func (s *Store) FindSessionByID(ctx context.Context, id uint) (*bot.WASession, error) {
	return NewWASessionRepository(s.db).FindSessionByID(ctx, id)
}

func (s *Store) FindAllSessions(ctx context.Context) ([]bot.WASession, error) {
	return NewWASessionRepository(s.db).FindAllSessions(ctx)
}

func (s *Store) UpdateSession(ctx context.Context, session *bot.WASession) error {
	return NewWASessionRepository(s.db).UpdateSession(ctx, session)
}

func (s *Store) DeleteSession(ctx context.Context, id uint) error {
	return NewWASessionRepository(s.db).DeleteSession(ctx, id)
}

func (s *Store) CreateConversation(ctx context.Context, conv *bot.Conversation) error {
	return NewConversationRepository(s.db).CreateConversation(ctx, conv)
}

func (s *Store) FindConversationByID(ctx context.Context, id uint) (*bot.Conversation, error) {
	return NewConversationRepository(s.db).FindConversationByID(ctx, id)
}

func (s *Store) FindConversationByIDWithMessages(ctx context.Context, id uint) (*bot.Conversation, error) {
	return NewConversationRepository(s.db).FindConversationByIDWithMessages(ctx, id)
}

func (s *Store) FindConversationsBySessionID(ctx context.Context, sessionID uint) ([]bot.Conversation, error) {
	return NewConversationRepository(s.db).FindConversationsBySessionID(ctx, sessionID)
}

func (s *Store) FindConversationsByStatus(ctx context.Context, status bot.ConversationStatus) ([]bot.Conversation, error) {
	return NewConversationRepository(s.db).FindConversationsByStatus(ctx, status)
}

func (s *Store) FindAllConversations(ctx context.Context) ([]bot.Conversation, error) {
	return NewConversationRepository(s.db).FindAllConversations(ctx)
}

func (s *Store) FindActiveConversationByCustomer(ctx context.Context, sessionID uint, customerNumber string) (*bot.Conversation, error) {
	return NewConversationRepository(s.db).FindActiveConversationByCustomer(ctx, sessionID, customerNumber)
}

func (s *Store) UpdateConversation(ctx context.Context, conv *bot.Conversation) error {
	return NewConversationRepository(s.db).UpdateConversation(ctx, conv)
}

func (s *Store) CreateMessage(ctx context.Context, msg *bot.Message) error {
	return NewConversationRepository(s.db).CreateMessage(ctx, msg)
}

func (s *Store) FindMessagesByConversationID(ctx context.Context, conversationID uint) ([]bot.Message, error) {
	return NewConversationRepository(s.db).FindMessagesByConversationID(ctx, conversationID)
}

func (s *Store) FindRecentMessages(ctx context.Context, conversationID uint, limit int) ([]bot.Message, error) {
	return NewConversationRepository(s.db).FindRecentMessages(ctx, conversationID, limit)
}
