package business

import (
	"errors"
	"time"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/domain/bot"
)

type ConversationService struct {
	store *postgres.Store
}

type Service = ConversationService

func NewConversationService(store *postgres.Store) *ConversationService {
	return &ConversationService{store: store}
}

func NewService(store *postgres.Store) *ConversationService {
	return NewConversationService(store)
}

func (s *ConversationService) GetOrCreateConversation(sessionID uint, customerNumber string) (*bot.Conversation, error) {
	conv, err := s.store.FindActiveConversationByCustomer(sessionID, customerNumber)
	if err == nil {
		return conv, nil
	}
	if !errors.Is(err, postgres.ErrNotFound) {
		return nil, err
	}

	newConv := &bot.Conversation{
		SessionID:        sessionID,
		CustomerWANumber: customerNumber,
		Status:           bot.StatusBot,
		StartedAt:        time.Now(),
	}
	if err := s.store.CreateConversation(newConv); err != nil {
		return nil, err
	}
	return newConv, nil
}

func (s *ConversationService) AddMessageWithConfig(convID uint, senderType string, content string, tokenIn, tokenOut int, llmConfigID *uint) (*bot.Message, error) {
	msg := &bot.Message{
		ConversationID: convID,
		SenderType:     bot.SenderType(senderType),
		Content:        content,
		TokenIn:        tokenIn,
		TokenOut:       tokenOut,
		LLMConfigID:    llmConfigID,
		CreatedAt:      time.Now(),
	}
	if err := s.store.CreateMessage(msg); err != nil {
		return nil, err
	}

	conv, err := s.store.FindConversationByID(convID)
	if err == nil {
		conv.UpdatedAt = time.Now()
		_ = s.store.UpdateConversation(conv)
	}

	return msg, nil
}

func (s *ConversationService) AddMessage(convID uint, senderType, content string, tokenIn, tokenOut int) (*bot.Message, error) {
	return s.AddMessageWithConfig(convID, senderType, content, tokenIn, tokenOut, nil)
}

func (s *ConversationService) Escalate(convID uint) error {
	conv, err := s.store.FindConversationByID(convID)
	if err != nil {
		return err
	}
	conv.Status = bot.StatusEscalation
	return s.store.UpdateConversation(conv)
}

func (s *ConversationService) ResetBot(convID uint) error {
	conv, err := s.store.FindConversationByID(convID)
	if err != nil {
		return err
	}
	conv.Status = bot.StatusBot
	conv.AssignedAgentID = nil
	return s.store.UpdateConversation(conv)
}

func (s *ConversationService) TakeOver(convID uint, agentID uint) error {
	conv, err := s.store.FindConversationByID(convID)
	if err != nil {
		return err
	}
	conv.Status = bot.StatusEscalation
	conv.AssignedAgentID = &agentID
	return s.store.UpdateConversation(conv)
}

func (s *ConversationService) CloseConversation(convID uint) error {
	conv, err := s.store.FindConversationByID(convID)
	if err != nil {
		return err
	}
	conv.Status = bot.StatusDone
	return s.store.UpdateConversation(conv)
}

func (s *ConversationService) GetConversation(id uint) (*bot.Conversation, error) {
	return s.store.FindConversationByID(id)
}

func (s *ConversationService) GetConversationWithMessages(id uint) (*bot.Conversation, error) {
	return s.store.FindConversationByIDWithMessages(id)
}

func (s *ConversationService) ListConversations(status string) ([]bot.Conversation, error) {
	if status != "" {
		return s.store.FindConversationsByStatus(bot.ConversationStatus(status))
	}
	return s.store.FindAllConversations()
}
