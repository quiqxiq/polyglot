package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/bot"
)

// WASessionRepository defines the persistence interface for WA sessions.
type WASessionRepository interface {
	CreateSession(ctx context.Context, session *bot.WASession) error
	FindSessionByID(ctx context.Context, id uint) (*bot.WASession, error)
	FindAllSessions(ctx context.Context) ([]bot.WASession, error)
	UpdateSession(ctx context.Context, session *bot.WASession) error
	DeleteSession(ctx context.Context, id uint) error
}

// ConversationRepository defines the persistence interface for conversations and messages.
type ConversationRepository interface {
	FindActiveConversationByCustomer(ctx context.Context, sessionID uint, customerNumber string) (*bot.Conversation, error)
	CreateConversation(ctx context.Context, conv *bot.Conversation) error
	FindConversationByID(ctx context.Context, id uint) (*bot.Conversation, error)
	FindConversationByIDWithMessages(ctx context.Context, id uint) (*bot.Conversation, error)
	FindConversationsByStatus(ctx context.Context, status bot.ConversationStatus) ([]bot.Conversation, error)
	FindConversationsBySessionID(ctx context.Context, sessionID uint) ([]bot.Conversation, error)
	FindAllConversations(ctx context.Context) ([]bot.Conversation, error)
	UpdateConversation(ctx context.Context, conv *bot.Conversation) error
	CreateMessage(ctx context.Context, msg *bot.Message) error
	FindMessagesByConversationID(ctx context.Context, conversationID uint) ([]bot.Message, error)
	FindRecentMessages(ctx context.Context, conversationID uint, limit int) ([]bot.Message, error)
}
