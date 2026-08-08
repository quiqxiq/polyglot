package conversation

import (
	"errors"
	"time"

	"github.com/quixiq/polyglot/internal/domain/bot"
)

var ErrNotFound = errors.New("conversation not found")

type ConversationRepository interface {
	FindActiveConversationByCustomer(sessionID uint, customerNumber string) (*bot.Conversation, error)
	CreateConversation(conv *bot.Conversation) error
	FindConversationByID(id uint) (*bot.Conversation, error)
	FindConversationByIDWithMessages(id uint) (*bot.Conversation, error)
	FindConversationsByStatus(status bot.ConversationStatus) ([]bot.Conversation, error)
	FindAllConversations() ([]bot.Conversation, error)
	UpdateConversation(conv *bot.Conversation) error
}

type ConversationService struct {
	repo ConversationRepository
}

type Service = ConversationService

func NewConversationService(repo ConversationRepository) *ConversationService {
	return &ConversationService{repo: repo}
}

func NewService(repo ConversationRepository) *ConversationService {
	return NewConversationService(repo)
}

func (s *ConversationService) GetOrCreateConversation(sessionID uint, customerNumber string) (*bot.Conversation, error) {
	conv, err := s.repo.FindActiveConversationByCustomer(sessionID, customerNumber)
	if err == nil {
		return conv, nil
	}

	newConv := &bot.Conversation{
		SessionID:        sessionID,
		CustomerWANumber: customerNumber,
		Status:           bot.StatusBot,
		StartedAt:        time.Now(),
	}
	if err := s.repo.CreateConversation(newConv); err != nil {
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
	return msg, nil
}

func (s *ConversationService) AddMessage(convID uint, senderType, content string, tokenIn, tokenOut int) (*bot.Message, error) {
	return s.AddMessageWithConfig(convID, senderType, content, tokenIn, tokenOut, nil)
}

func (s *ConversationService) GetHistory(convID uint, limit int) ([]bot.Message, error) {
	conv, err := s.repo.FindConversationByIDWithMessages(convID)
	if err != nil {
		return nil, err
	}
	return conv.Messages, nil
}

func (s *ConversationService) Escalate(convID uint) error {
	conv, err := s.repo.FindConversationByID(convID)
	if err != nil {
		return err
	}
	conv.Status = bot.StatusEscalation
	return s.repo.UpdateConversation(conv)
}

func (s *ConversationService) ResetBot(convID uint) error {
	conv, err := s.repo.FindConversationByID(convID)
	if err != nil {
		return err
	}
	conv.Status = bot.StatusBot
	conv.AssignedAgentID = nil
	return s.repo.UpdateConversation(conv)
}

func (s *ConversationService) TakeOver(convID uint, agentID uint) error {
	conv, err := s.repo.FindConversationByID(convID)
	if err != nil {
		return err
	}
	conv.Status = bot.StatusEscalation
	conv.AssignedAgentID = &agentID
	return s.repo.UpdateConversation(conv)
}

func (s *ConversationService) CloseConversation(convID uint) error {
	conv, err := s.repo.FindConversationByID(convID)
	if err != nil {
		return err
	}
	conv.Status = bot.StatusDone
	return s.repo.UpdateConversation(conv)
}

func (s *ConversationService) GetConversation(id uint) (*bot.Conversation, error) {
	return s.repo.FindConversationByID(id)
}

func (s *ConversationService) GetConversationWithMessages(id uint) (*bot.Conversation, error) {
	return s.repo.FindConversationByIDWithMessages(id)
}

func (s *ConversationService) ListConversations(status string) ([]bot.Conversation, error) {
	if status != "" {
		return s.repo.FindConversationsByStatus(bot.ConversationStatus(status))
	}
	return s.repo.FindAllConversations()
}
