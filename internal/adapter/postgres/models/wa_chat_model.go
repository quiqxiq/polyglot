package models

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/bot"
)

// WAChatModel is the GORM database model for the WhatsApp chat mirror.
type WAChatModel struct {
	ID                 uint   `gorm:"primaryKey"`
	SessionID          uint   `gorm:"uniqueIndex:uq_wa_chats_session_jid"`
	ChatJID            string `gorm:"uniqueIndex:uq_wa_chats_session_jid"`
	DisplayName        string
	IsGroup            bool
	LastMessageID      string
	LastMessagePreview string `gorm:"type:text"`
	LastMessageTime    time.Time
	UnreadCount        int `gorm:"default:0"`
	BotEnabled         bool `gorm:"default:true"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// TableName maps WAChatModel ke tabel migrasi `wa_chats`.
func (WAChatModel) TableName() string { return "wa_chats" }

func (m *WAChatModel) ToDomain() *bot.WAChat {
	if m == nil {
		return nil
	}
	return &bot.WAChat{
		ID:                 m.ID,
		SessionID:          m.SessionID,
		ChatJID:            m.ChatJID,
		DisplayName:        m.DisplayName,
		IsGroup:            m.IsGroup,
		LastMessageID:      m.LastMessageID,
		LastMessagePreview: m.LastMessagePreview,
		LastMessageTime:    m.LastMessageTime,
		UnreadCount:        m.UnreadCount,
		BotEnabled:         m.BotEnabled,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}

func WAChatModelFromDomain(c *bot.WAChat) *WAChatModel {
	if c == nil {
		return nil
	}
	return &WAChatModel{
		ID:                 c.ID,
		SessionID:          c.SessionID,
		ChatJID:            c.ChatJID,
		DisplayName:        c.DisplayName,
		IsGroup:            c.IsGroup,
		LastMessageID:      c.LastMessageID,
		LastMessagePreview: c.LastMessagePreview,
		LastMessageTime:    c.LastMessageTime,
		UnreadCount:        c.UnreadCount,
		BotEnabled:         c.BotEnabled,
		CreatedAt:          c.CreatedAt,
		UpdatedAt:          c.UpdatedAt,
	}
}

// WAMessageModel is the GORM database model for the WhatsApp message mirror.
type WAMessageModel struct {
	ID          uint   `gorm:"primaryKey"`
	SessionID   uint   `gorm:"uniqueIndex:uq_wa_messages_session_wa_id"`
	ChatJID     string `gorm:"index:idx_wa_messages_session_chat_time"`
	WAMessageID string `gorm:"uniqueIndex:uq_wa_messages_session_wa_id"`
	SenderJID   string
	SenderName  string
	Content     string `gorm:"type:text"`
	MediaType   string `gorm:"default:text"`
	IsFromMe    bool
	IsRead      bool
	Timestamp   time.Time `gorm:"index:idx_wa_messages_session_chat_time,priority:2"`
	CreatedAt   time.Time
}

// TableName maps WAMessageModel ke tabel migrasi `wa_messages`.
func (WAMessageModel) TableName() string { return "wa_messages" }

func (m *WAMessageModel) ToDomain() *bot.WAMessage {
	if m == nil {
		return nil
	}
	return &bot.WAMessage{
		ID:          m.ID,
		SessionID:   m.SessionID,
		ChatJID:     m.ChatJID,
		WAMessageID: m.WAMessageID,
		SenderJID:   m.SenderJID,
		SenderName:  m.SenderName,
		Content:     m.Content,
		MediaType:   m.MediaType,
		IsFromMe:    m.IsFromMe,
		IsRead:      m.IsRead,
		Timestamp:   m.Timestamp,
		CreatedAt:   m.CreatedAt,
	}
}

func WAMessageModelFromDomain(m *bot.WAMessage) *WAMessageModel {
	if m == nil {
		return nil
	}
	return &WAMessageModel{
		ID:          m.ID,
		SessionID:   m.SessionID,
		ChatJID:     m.ChatJID,
		WAMessageID: m.WAMessageID,
		SenderJID:   m.SenderJID,
		SenderName:  m.SenderName,
		Content:     m.Content,
		MediaType:   m.MediaType,
		IsFromMe:    m.IsFromMe,
		IsRead:      m.IsRead,
		Timestamp:   m.Timestamp,
		CreatedAt:   m.CreatedAt,
	}
}
