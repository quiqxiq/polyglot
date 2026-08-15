package postgres

import (
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/port"
)

var ErrNotFound = errors.New("record not found")

type waSessionRepository struct {
	db *gorm.DB
}

var _ port.WASessionRepository = (*waSessionRepository)(nil)

// NewWASessionRepository returns a port.WASessionRepository implementation.
func NewWASessionRepository(db *gorm.DB) port.WASessionRepository {
	return &waSessionRepository{db: db}
}

func (r *waSessionRepository) Create(session *bot.WASession) error {
	m := models.WASessionModelFromDomain(session)
	if err := r.db.Create(m).Error; err != nil {
		return err
	}
	session.ID = m.ID
	return nil
}

func (r *waSessionRepository) FindByID(id uint) (*bot.WASession, error) {
	var m models.WASessionModel
	if err := r.db.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *waSessionRepository) FindAll() ([]bot.WASession, error) {
	var mList []models.WASessionModel
	if err := r.db.Order("created_at DESC").Find(&mList).Error; err != nil {
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

func (r *waSessionRepository) Update(session *bot.WASession) error {
	m := models.WASessionModelFromDomain(session)
	return r.db.Save(m).Error
}

func (r *waSessionRepository) Delete(id uint) error {
	return r.db.Delete(&models.WASessionModel{}, id).Error
}

// Store backward-compatible delegations
func (s *Store) CreateSession(session *bot.WASession) error {
	return NewWASessionRepository(s.db).Create(session)
}

func (s *Store) FindSessionByID(id uint) (*bot.WASession, error) {
	return NewWASessionRepository(s.db).FindByID(id)
}

func (s *Store) FindAllSessions() ([]bot.WASession, error) {
	return NewWASessionRepository(s.db).FindAll()
}

func (s *Store) UpdateSession(session *bot.WASession) error {
	return NewWASessionRepository(s.db).Update(session)
}

func (s *Store) DeleteSession(id uint) error {
	return NewWASessionRepository(s.db).Delete(id)
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
	activeStatuses := []string{
		string(bot.StatusBot),
		string(bot.StatusEscalation),
	}
	err := s.db.Where("session_id = ? AND customer_number = ? AND status IN ?", sessionID, customerNumber, activeStatuses).
		Order("created_at DESC").
		First(&m).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
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
