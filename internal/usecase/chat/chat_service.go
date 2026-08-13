package chat

import (
	"errors"

	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/port"
)

// ErrEmptyChatJID menyatakan bahwa chat_jid tidak diberikan.
var ErrEmptyChatJID = errors.New("chat_jid is required")

// ChatService menyediakan operasi baca Inbox WhatsApp dari mirror chat.
type ChatService struct {
	repo port.ChatRepository
}

func NewChatService(repo port.ChatRepository) *ChatService {
	return &ChatService{repo: repo}
}

// ListChats mengembalikan daftar chat mirror sebuah perangkat.
func (s *ChatService) ListChats(sessionID uint, limit, offset int, search string) ([]bot.WAChat, error) {
	if s.repo == nil {
		return []bot.WAChat{}, nil
	}
	return s.repo.ListChats(sessionID, limit, offset, search)
}

// GetChatMessages mengembalikan pesan sebuah chat, ascending.
func (s *ChatService) GetChatMessages(sessionID uint, chatJID string, limit, offset int) ([]bot.WAMessage, error) {
	if s.repo == nil {
		return []bot.WAMessage{}, nil
	}
	if chatJID == "" {
		return nil, ErrEmptyChatJID
	}
	return s.repo.ListChatMessages(sessionID, chatJID, limit, offset)
}

// MarkRead mereset unread count sebuah chat.
func (s *ChatService) MarkRead(sessionID uint, chatJID string) error {
	if s.repo == nil {
		return nil
	}
	if chatJID == "" {
		return ErrEmptyChatJID
	}
	return s.repo.MarkChatRead(sessionID, chatJID)
}

// ToggleChatBot mengaktifkan/menonaktifkan auto-reply bot untuk satu chat.
func (s *ChatService) ToggleChatBot(sessionID uint, chatJID string, enabled bool) error {
	if s.repo == nil {
		return nil
	}
	if chatJID == "" {
		return ErrEmptyChatJID
	}
	return s.repo.SetChatBotEnabled(sessionID, chatJID, enabled)
}
