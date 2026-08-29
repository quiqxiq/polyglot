package conversation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/port"
)

type ConversationUseCase struct {
	repo      port.ConversationRepository
	publisher port.EventPublisher
}

// SetPublisher wires the realtime broadcaster (SSE hub) so conversation
// status changes (take-over, bot reset, close, escalation) are pushed to the
// frontend immediately.
func (s *ConversationUseCase) SetPublisher(p port.EventPublisher) {
	s.publisher = p
}

// ConversationStatusEvent is the SSE payload broadcast on every conversation
// status change.
type ConversationStatusEvent struct {
	ConversationID uint   `json:"conversation_id"`
	SessionID      uint   `json:"session_id"`
	CustomerNumber string `json:"customer_number"`
	Status         string `json:"status"`
}

func (s *ConversationUseCase) publishStatusChange(conv *bot.Conversation) {
	if s.publisher == nil || conv == nil {
		return
	}
	s.publisher.PublishEvent("conversation_status", ConversationStatusEvent{
		ConversationID: conv.ID,
		SessionID:      conv.SessionID,
		CustomerNumber: conv.CustomerWANumber,
		Status:         string(conv.Status),
	})
}

type Service = ConversationUseCase

func NewConversationUseCase(repo port.ConversationRepository) *ConversationUseCase {
	return &ConversationUseCase{repo: repo}
}

func NewService(repo port.ConversationRepository) *ConversationUseCase {
	return NewConversationUseCase(repo)
}

func (s *ConversationUseCase) GetActiveConversationByCustomer(ctx context.Context, sessionID uint, customerNumber string) (*bot.Conversation, error) {
	return s.repo.FindActiveConversationByCustomer(ctx, sessionID, customerNumber)
}

func (s *ConversationUseCase) GetOrCreateConversationWithTimeout(ctx context.Context, sessionID uint, customerNumber string, timeoutMinutes int) (*bot.Conversation, error) {
	conv, err := s.repo.FindActiveConversationByCustomer(ctx, sessionID, customerNumber)
	if err == nil {
		if timeoutMinutes > 0 && !conv.UpdatedAt.IsZero() && time.Since(conv.UpdatedAt) > time.Duration(timeoutMinutes)*time.Minute {
			conv.Status = bot.StatusDone
			_ = s.repo.UpdateConversation(ctx, conv)
		} else {
			return conv, nil
		}
	}

	newConv := &bot.Conversation{
		SessionID:        sessionID,
		CustomerWANumber: customerNumber,
		Status:           bot.StatusBot,
		StartedAt:        time.Now(),
	}
	if err := s.repo.CreateConversation(ctx, newConv); err != nil {
		return nil, err
	}
	return newConv, nil
}

func (s *ConversationUseCase) GetOrCreateConversation(ctx context.Context, sessionID uint, customerNumber string) (*bot.Conversation, error) {
	return s.GetOrCreateConversationWithTimeout(ctx, sessionID, customerNumber, 0)
}

func (s *ConversationUseCase) AddMessageWithConfig(ctx context.Context, convID uint, senderType string, content string, tokenIn, tokenOut int, llmConfigID *uint) (*bot.Message, error) {
	if convID == 0 {
		return nil, errors.New("conversation id is required")
	}
	msg := &bot.Message{
		ConversationID: convID,
		SenderType:     bot.SenderType(senderType),
		Content:        content,
		TokenIn:        tokenIn,
		TokenOut:       tokenOut,
		LLMConfigID:    llmConfigID,
		CreatedAt:      time.Now(),
	}
	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("persist message: %w", err)
	}
	return msg, nil
}

func (s *ConversationUseCase) AddMessage(ctx context.Context, convID uint, senderType, content string, tokenIn, tokenOut int) (*bot.Message, error) {
	return s.AddMessageWithConfig(ctx, convID, senderType, content, tokenIn, tokenOut, nil)
}

func (s *ConversationUseCase) GetHistory(ctx context.Context, convID uint, limit int) ([]bot.Message, error) {
	conv, err := s.repo.FindConversationByIDWithMessages(ctx, convID)
	if err != nil {
		return nil, err
	}
	return conv.Messages, nil
}

// GetRecentHistory returns the most recent N messages of a conversation in
// ascending (chronological) order — used to build the LLM prompt context.
func (s *ConversationUseCase) GetRecentHistory(ctx context.Context, convID uint, limit int) ([]bot.Message, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.FindRecentMessages(ctx, convID, limit)
}

func (s *ConversationUseCase) Escalate(ctx context.Context, convID uint) error {
	conv, err := s.repo.FindConversationByID(ctx, convID)
	if err != nil {
		return err
	}
	conv.Status = bot.StatusEscalation
	if err := s.repo.UpdateConversation(ctx, conv); err != nil {
		return err
	}
	s.publishStatusChange(conv)
	return nil
}

func (s *ConversationUseCase) ResetBot(ctx context.Context, convID uint) error {
	conv, err := s.repo.FindConversationByID(ctx, convID)
	if err != nil {
		return err
	}
	conv.Status = bot.StatusBot
	conv.AssignedAgentID = nil
	if err := s.repo.UpdateConversation(ctx, conv); err != nil {
		return err
	}
	s.publishStatusChange(conv)
	return nil
}

func (s *ConversationUseCase) TakeOver(ctx context.Context, convID uint, agentID uint) error {
	conv, err := s.repo.FindConversationByID(ctx, convID)
	if err != nil {
		return err
	}
	conv.Status = bot.StatusEscalation
	conv.AssignedAgentID = &agentID
	if err := s.repo.UpdateConversation(ctx, conv); err != nil {
		return err
	}
	s.publishStatusChange(conv)
	return nil
}

func (s *ConversationUseCase) CloseConversation(ctx context.Context, convID uint) error {
	conv, err := s.repo.FindConversationByID(ctx, convID)
	if err != nil {
		return err
	}
	conv.Status = bot.StatusDone
	if err := s.repo.UpdateConversation(ctx, conv); err != nil {
		return err
	}
	s.publishStatusChange(conv)
	return nil
}

func (s *ConversationUseCase) GetConversation(ctx context.Context, id uint) (*bot.Conversation, error) {
	return s.repo.FindConversationByID(ctx, id)
}

func (s *ConversationUseCase) GetConversationWithMessages(ctx context.Context, id uint) (*bot.Conversation, error) {
	return s.repo.FindConversationByIDWithMessages(ctx, id)
}

func (s *ConversationUseCase) ListConversations(ctx context.Context, status string) ([]bot.Conversation, error) {
	if status != "" {
		return s.repo.FindConversationsByStatus(ctx, bot.ConversationStatus(status))
	}
	return s.repo.FindAllConversations(ctx)
}

// ListConversationsBySession returns all conversations belonging to one WA
// session (perangkat), terbaru lebih dulu.
func (s *ConversationUseCase) ListConversationsBySession(ctx context.Context, sessionID uint) ([]bot.Conversation, error) {
	return s.repo.FindConversationsBySessionID(ctx, sessionID)
}
