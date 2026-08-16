package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/bot"
)

var ErrNotFound = errors.New("record not found")

// WASessionRepository implementation
func (s *Store) CreateSession(ctx context.Context, session *bot.WASession) error {
	m := model.WASessionModelFromDomain(session)
	if err := s.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	session.ID = m.ID
	return nil
}

func (s *Store) FindSessionByID(ctx context.Context, id uint) (*bot.WASession, error) {
	var m model.WASessionModel
	if err := s.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (s *Store) FindAllSessions(ctx context.Context) ([]bot.WASession, error) {
	var mList []model.WASessionModel
	if err := s.db.WithContext(ctx).Order("created_at DESC").Find(&mList).Error; err != nil {
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

func (s *Store) UpdateSession(ctx context.Context, session *bot.WASession) error {
	m := model.WASessionModelFromDomain(session)
	return s.db.WithContext(ctx).Save(m).Error
}

func (s *Store) DeleteSession(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&model.WASessionModel{}, id).Error
}

// ConversationRepository implementation
func (s *Store) CreateConversation(ctx context.Context, conv *bot.Conversation) error {
	m := model.ConversationModelFromDomain(conv)
	if err := s.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	conv.ID = m.ID
	return nil
}

func (s *Store) FindConversationByID(ctx context.Context, id uint) (*bot.Conversation, error) {
	var m model.ConversationModel
	if err := s.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (s *Store) FindConversationByIDWithMessages(ctx context.Context, id uint) (*bot.Conversation, error) {
	var m model.ConversationModel
	if err := s.db.WithContext(ctx).Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (s *Store) FindConversationsBySessionID(ctx context.Context, sessionID uint) ([]bot.Conversation, error) {
	var mList []model.ConversationModel
	if err := s.db.WithContext(ctx).Where("session_id = ?", sessionID).Order("updated_at DESC").Find(&mList).Error; err != nil {
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

func (s *Store) FindConversationsByStatus(ctx context.Context, status bot.ConversationStatus) ([]bot.Conversation, error) {
	var mList []model.ConversationModel
	if err := s.db.WithContext(ctx).Where("status = ?", string(status)).Order("updated_at DESC").Find(&mList).Error; err != nil {
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

func (s *Store) FindAllConversations(ctx context.Context) ([]bot.Conversation, error) {
	var mList []model.ConversationModel
	if err := s.db.WithContext(ctx).Order("updated_at DESC").Find(&mList).Error; err != nil {
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

func (s *Store) FindActiveConversationByCustomer(ctx context.Context, sessionID uint, customerNumber string) (*bot.Conversation, error) {
	var m model.ConversationModel
	if err := s.db.WithContext(ctx).Where("session_id = ? AND customer_wa_number = ? AND status IN ('bot', 'escalation')", sessionID, customerNumber).
		Order("updated_at DESC").First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (s *Store) UpdateConversation(ctx context.Context, conv *bot.Conversation) error {
	m := model.ConversationModelFromDomain(conv)
	return s.db.WithContext(ctx).Save(m).Error
}

// MessageRepository implementation
func (s *Store) CreateMessage(ctx context.Context, msg *bot.Message) error {
	m := model.MessageModelFromDomain(msg)
	if err := s.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	msg.ID = m.ID

	_ = s.db.WithContext(ctx).Model(&model.ConversationModel{}).Where("id = ?", msg.ConversationID).Update("updated_at", msg.CreatedAt)
	return nil
}

func (s *Store) FindMessagesByConversationID(ctx context.Context, conversationID uint) ([]bot.Message, error) {
	var mList []model.MessageModel
	if err := s.db.WithContext(ctx).Where("conversation_id = ?", conversationID).Order("created_at ASC").Find(&mList).Error; err != nil {
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

func (s *Store) FindRecentMessages(ctx context.Context, conversationID uint, limit int) ([]bot.Message, error) {
	var mList []model.MessageModel
	if err := s.db.WithContext(ctx).Where("conversation_id = ?", conversationID).Order("created_at DESC").Limit(limit).Find(&mList).Error; err != nil {
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
