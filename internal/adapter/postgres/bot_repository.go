package postgres

import (
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
	"github.com/quixiq/polyglot/internal/domain/bot"
)

var ErrNotFound = errors.New("record not found")

// WASessionRepository implementation
func (s *Store) CreateSession(session *bot.WASession) error {
	m := models.WASessionModelFromDomain(session)
	if err := s.db.Create(m).Error; err != nil {
		return err
	}
	session.ID = m.ID
	return nil
}

func (s *Store) FindSessionByID(id uint) (*bot.WASession, error) {
	var m models.WASessionModel
	if err := s.db.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (s *Store) FindAllSessions() ([]bot.WASession, error) {
	var mList []models.WASessionModel
	if err := s.db.Order("created_at DESC").Find(&mList).Error; err != nil {
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

func (s *Store) UpdateSession(session *bot.WASession) error {
	m := models.WASessionModelFromDomain(session)
	return s.db.Save(m).Error
}

func (s *Store) DeleteSession(id uint) error {
	return s.db.Delete(&models.WASessionModel{}, id).Error
}

// ConversationRepository implementation
func (s *Store) CreateConversation(conv *bot.Conversation) error {
	m := models.ConversationModelFromDomain(conv)
	if err := s.db.Create(m).Error; err != nil {
		return err
	}
	conv.ID = m.ID
	return nil
}

func (s *Store) FindConversationByID(id uint) (*bot.Conversation, error) {
	var m models.ConversationModel
	if err := s.db.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (s *Store) FindConversationByIDWithMessages(id uint) (*bot.Conversation, error) {
	var m models.ConversationModel
	if err := s.db.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (s *Store) FindConversationsBySessionID(sessionID uint) ([]bot.Conversation, error) {
	var mList []models.ConversationModel
	if err := s.db.Where("session_id = ?", sessionID).Order("updated_at DESC").Find(&mList).Error; err != nil {
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

func (s *Store) FindConversationsByStatus(status bot.ConversationStatus) ([]bot.Conversation, error) {
	var mList []models.ConversationModel
	if err := s.db.Where("status = ?", string(status)).Order("updated_at DESC").Find(&mList).Error; err != nil {
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

func (s *Store) FindAllConversations() ([]bot.Conversation, error) {
	var mList []models.ConversationModel
	if err := s.db.Order("updated_at DESC").Find(&mList).Error; err != nil {
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

func (s *Store) FindActiveConversationByCustomer(sessionID uint, customerNumber string) (*bot.Conversation, error) {
	var m models.ConversationModel
	if err := s.db.Where("session_id = ? AND customer_wa_number = ? AND status IN ('bot', 'escalation')", sessionID, customerNumber).
		Order("updated_at DESC").First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (s *Store) UpdateConversation(conv *bot.Conversation) error {
	m := models.ConversationModelFromDomain(conv)
	return s.db.Save(m).Error
}

// MessageRepository implementation
func (s *Store) CreateMessage(msg *bot.Message) error {
	m := models.MessageModelFromDomain(msg)
	if err := s.db.Create(m).Error; err != nil {
		return err
	}
	msg.ID = m.ID

	_ = s.db.Model(&models.ConversationModel{}).Where("id = ?", msg.ConversationID).Update("updated_at", msg.CreatedAt)
	return nil
}

func (s *Store) FindMessagesByConversationID(conversationID uint) ([]bot.Message, error) {
	var mList []models.MessageModel
	if err := s.db.Where("conversation_id = ?", conversationID).Order("created_at ASC").Find(&mList).Error; err != nil {
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

func (s *Store) FindRecentMessages(conversationID uint, limit int) ([]bot.Message, error) {
	var mList []models.MessageModel
	if err := s.db.Where("conversation_id = ?", conversationID).Order("created_at DESC").Limit(limit).Find(&mList).Error; err != nil {
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
