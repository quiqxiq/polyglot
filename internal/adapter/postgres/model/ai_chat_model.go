package model

import (
	"encoding/json"
	"time"

	"github.com/quixiq/polyglot/internal/domain/aichat"
)

type AIConversationModel struct {
	ID            string  `gorm:"primaryKey"`
	TenantID      string  `gorm:"column:tenant_id;type:text;not null;default:tenant-default"`
	UserID        *uint   `gorm:"column:user_id"`
	DeviceID      *string `gorm:"column:device_id;type:uuid"`
	ModelConfigID string  `gorm:"column:model_config_id;type:text"`
	Title         string  `gorm:"type:text"`
	Status        string  `gorm:"type:varchar(20);not null;default:ACTIVE"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ArchivedAt    *time.Time `gorm:"column:archived_at"`
}

func (AIConversationModel) TableName() string { return "ai_conversations" }

func (m *AIConversationModel) ToDomain() aichat.Conversation {
	if m == nil {
		return aichat.Conversation{}
	}
	var uid uint
	if m.UserID != nil {
		uid = *m.UserID
	}
	var devID string
	if m.DeviceID != nil {
		devID = *m.DeviceID
	}
	return aichat.Conversation{
		ID: m.ID, TenantID: m.TenantID, UserID: uid, DeviceID: devID,
		ModelConfigID: m.ModelConfigID, Title: m.Title, Status: m.Status,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt, ArchivedAt: m.ArchivedAt,
	}
}

func AIConversationModelFromDomain(c aichat.Conversation) *AIConversationModel {
	m := &AIConversationModel{
		ID: c.ID, TenantID: c.TenantID, ModelConfigID: c.ModelConfigID,
		Title: c.Title, Status: c.Status, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt, ArchivedAt: c.ArchivedAt,
	}
	if c.UserID > 0 {
		uid := c.UserID
		m.UserID = &uid
	}
	if c.DeviceID != "" {
		d := c.DeviceID
		m.DeviceID = &d
	}
	return m
}

type AIChatEventModel struct {
	ID             int64           `gorm:"primaryKey;autoIncrement"`
	ConversationID string          `gorm:"column:conversation_id;type:text;not null;index"`
	Seq            int64           `gorm:"column:seq;not null"`
	EventType      string          `gorm:"column:event_type;type:varchar(30);not null"`
	Payload        json.RawMessage `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt      time.Time
}

func (AIChatEventModel) TableName() string { return "ai_chat_events" }

func (m *AIChatEventModel) ToDomain() aichat.Event {
	if m == nil {
		return aichat.Event{}
	}
	return aichat.Event{ID: m.ID, ConversationID: m.ConversationID, Seq: m.Seq, Type: m.EventType, Payload: m.Payload, CreatedAt: m.CreatedAt}
}
