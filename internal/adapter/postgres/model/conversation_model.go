package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/bot"
)

// ConversationModel is the GORM database model for Conversations.
// TableName eksplisit ke `conversations` (migrasi 000002) — tanpa ini GORM
// memakai `conversation_models`, divergen dari migrasi (prod).
type ConversationModel struct {
	ID               uint   `gorm:"primaryKey"`
	SessionID        uint   `gorm:"index"`
	CustomerWANumber string `gorm:"index"`
	Status           string `gorm:"default:bot"`
	AssignedAgentID  *uint
	StartedAt        time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time

	Messages []MessageModel `gorm:"foreignKey:ConversationID"`
}

// TableName maps ConversationModel ke tabel migrasi `conversations`.
func (ConversationModel) TableName() string { return "conversations" }

// ToDomain converts a conversation database model to its domain representation.
func (m *ConversationModel) ToDomain() *bot.Conversation {
	if m == nil {
		return nil
	}
	var messages []bot.Message
	for _, msg := range m.Messages {
		if d := msg.ToDomain(); d != nil {
			messages = append(messages, *d)
		}
	}
	return &bot.Conversation{
		ID:               m.ID,
		SessionID:        m.SessionID,
		CustomerWANumber: m.CustomerWANumber,
		Status:           bot.ConversationStatus(m.Status),
		AssignedAgentID:  m.AssignedAgentID,
		StartedAt:        m.StartedAt,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
		Messages:         messages,
	}
}

// ConversationModelFromDomain converts a conversation domain entity to a database model.
func ConversationModelFromDomain(c *bot.Conversation) *ConversationModel {
	if c == nil {
		return nil
	}
	return &ConversationModel{
		ID:               c.ID,
		SessionID:        c.SessionID,
		CustomerWANumber: c.CustomerWANumber,
		Status:           string(c.Status),
		AssignedAgentID:  c.AssignedAgentID,
		StartedAt:        c.StartedAt,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
	}
}

// MessageModel is the GORM database model for Messages.
// TableName eksplisit ke `messages` (migrasi 000002) — tanpa ini GORM memakai
// `message_models`, divergen dari migrasi (prod).
type MessageModel struct {
	ID             uint   `gorm:"primaryKey"`
	ConversationID uint   `gorm:"index"`
	SenderType     string `gorm:"not null"`
	Content        string `gorm:"type:text"`
	TokenIn        int    `gorm:"default:0"`
	TokenOut       int    `gorm:"default:0"`
	LLMConfigID    *uint
	CreatedAt      time.Time
}

// TableName maps MessageModel ke tabel migrasi `messages`.
func (MessageModel) TableName() string { return "messages" }

// ToDomain converts a message database model to its domain representation.
func (m *MessageModel) ToDomain() *bot.Message {
	if m == nil {
		return nil
	}
	return &bot.Message{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		SenderType:     bot.SenderType(m.SenderType),
		Content:        m.Content,
		TokenIn:        m.TokenIn,
		TokenOut:       m.TokenOut,
		LLMConfigID:    m.LLMConfigID,
		CreatedAt:      m.CreatedAt,
	}
}

// MessageModelFromDomain converts a message domain entity to a database model.
func MessageModelFromDomain(msg *bot.Message) *MessageModel {
	if msg == nil {
		return nil
	}
	return &MessageModel{
		ID:             msg.ID,
		ConversationID: msg.ConversationID,
		SenderType:     string(msg.SenderType),
		Content:        msg.Content,
		TokenIn:        msg.TokenIn,
		TokenOut:       msg.TokenOut,
		LLMConfigID:    msg.LLMConfigID,
		CreatedAt:      msg.CreatedAt,
	}
}
