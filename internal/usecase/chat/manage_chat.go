package chat

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/port"
)

// bot.ErrEmptyChatJID menyatakan bahwa chat_jid tidak diberikan.

// ChatUseCase menyediakan operasi baca Inbox WhatsApp dari mirror chat.
type ChatUseCase struct {
	repo port.ChatRepository
}

func NewChatUseCase(repo port.ChatRepository) *ChatUseCase {
	return &ChatUseCase{repo: repo}
}

// ListChats mengembalikan daftar chat mirror sebuah perangkat.
func (s *ChatUseCase) ListChats(ctx context.Context, sessionID uint, limit, offset int, search string) ([]bot.WAChat, error) {
	if s.repo == nil {
		return []bot.WAChat{}, nil
	}
	return s.repo.ListChats(ctx, sessionID, limit, offset, search)
}

// GetChatMessages mengembalikan pesan sebuah chat, ascending.
func (s *ChatUseCase) GetChatMessages(ctx context.Context, sessionID uint, chatJID string, limit, offset int) ([]bot.WAMessage, error) {
	if s.repo == nil {
		return []bot.WAMessage{}, nil
	}
	if chatJID == "" {
		return nil, bot.ErrEmptyChatJID
	}
	return s.repo.ListChatMessages(ctx, sessionID, chatJID, limit, offset)
}

// MarkRead mereset unread count sebuah chat.
func (s *ChatUseCase) MarkRead(ctx context.Context, sessionID uint, chatJID string) error {
	if s.repo == nil {
		return nil
	}
	if chatJID == "" {
		return bot.ErrEmptyChatJID
	}
	return s.repo.MarkChatRead(ctx, sessionID, chatJID)
}

// ToggleChatBot mengaktifkan/menonaktifkan auto-reply bot untuk satu chat.
func (s *ChatUseCase) ToggleChatBot(ctx context.Context, sessionID uint, chatJID string, enabled bool) error {
	if s.repo == nil {
		return nil
	}
	if chatJID == "" {
		return bot.ErrEmptyChatJID
	}
	return s.repo.SetChatBotEnabled(ctx, sessionID, chatJID, enabled)
}
