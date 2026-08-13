package models

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/bot"
)

// WAChatModel is the GORM database model for the WhatsApp chat mirror.
// Catatan: kolom yang memuat akronim diberi tag `column:` eksplisit karena
// NamingStrategy GORM mengubah `ChatJID` menjadi `chat_j_id` (akronim J + ID
// dipecah jadi dua kata), sementara migrasi 000004 dan query raw repo
// memakai `chat_jid`. Tanpa tag ini skema AutoMigrate (dev) divergen dari
// migrasi (prod) dan query memakai kolom `chat_jid` akan gagal.
type WAChatModel struct {
	ID                 uint   `gorm:"primaryKey"`
	SessionID          uint   `gorm:"uniqueIndex:uq_wa_chats_session_jid"`
	ChatJID            string `gorm:"column:chat_jid;uniqueIndex:uq_wa_chats_session_jid"`
	DisplayName        string
	IsGroup            bool
	LastMessageID      string
	LastMessagePreview string `gorm:"type:text"`
	LastMessageTime    time.Time
	UnreadCount        int  `gorm:"default:0"`
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
// Sama seperti WAChatModel: `ChatJID`/`SenderJID` perlu tag `column:` eksplisit
// agar sesuai dengan migrasi 000004 (`chat_jid`, `sender_jid`).
type WAMessageModel struct {
	ID          uint   `gorm:"primaryKey"`
	SessionID   uint   `gorm:"uniqueIndex:uq_wa_messages_session_wa_id"`
	ChatJID     string `gorm:"column:chat_jid;index:idx_wa_messages_session_chat_time"`
	WAMessageID string `gorm:"uniqueIndex:uq_wa_messages_session_wa_id"`
	SenderJID   string `gorm:"column:sender_jid"`
	SenderName  string
	Content     string `gorm:"type:text"`
	MediaType   string `gorm:"default:text"`
	IsFromMe    bool
	IsRead      bool
	Status      string    `gorm:"type:varchar(20);default:sent;not null"`
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
		Status:      m.Status,
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
		Status:      m.Status,
		Timestamp:   m.Timestamp,
		CreatedAt:   m.CreatedAt,
	}
}
