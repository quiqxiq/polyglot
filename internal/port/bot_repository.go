package port

import (
	"github.com/quixiq/polyglot/internal/domain/bot"
)

// WASessionRepository defines the persistence interface for WA sessions.
type WASessionRepository interface {
	Create(session *bot.WASession) error
	FindByID(id uint) (*bot.WASession, error)
	FindAll() ([]bot.WASession, error)
	Update(session *bot.WASession) error
	Delete(id uint) error
}

// ConversationRepository defines the persistence interface for conversations.
type ConversationRepository interface {
	Create(conv *bot.Conversation) error
	FindByID(id uint) (*bot.Conversation, error)
	FindByIDWithMessages(id uint) (*bot.Conversation, error)
	FindBySessionID(sessionID uint) ([]bot.Conversation, error)
	FindByStatus(status bot.ConversationStatus) ([]bot.Conversation, error)
	FindAll() ([]bot.Conversation, error)
	FindActiveByCustomer(sessionID uint, customerNumber string) (*bot.Conversation, error)
	Update(conv *bot.Conversation) error
}

// MessageRepository defines the persistence interface for messages.
type MessageRepository interface {
	Create(msg *bot.Message) error
	FindByConversationID(conversationID uint) ([]bot.Message, error)
	FindRecent(conversationID uint, limit int) ([]bot.Message, error)
}
