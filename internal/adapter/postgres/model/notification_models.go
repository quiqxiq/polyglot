package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/notification"
)

// NotificationTemplateModel is the GORM model for WhatsApp message templates
// (DATABASE-SCHEMA-ISP.md §2.10 — notification_templates).
type NotificationTemplateModel struct {
	ID       string `gorm:"primaryKey"`
	TenantID string `gorm:"type:text;not null;default:tenant-default;uniqueIndex:uq_notif_template"`
	// template_key VARCHAR(50) — 'BILL_REMINDER', 'PAYMENT_RECEIPT', ...
	TemplateKey   string `gorm:"type:varchar(50);not null;uniqueIndex:uq_notif_template"`
	Name          string `gorm:"type:varchar(100);not null"`
	Content       string `gorm:"type:text;not null"`
	VariablesJSON string `gorm:"column:variables_json;type:jsonb;not null;default:'[]'"`
	IsActive      bool   `gorm:"not null;default:true"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName returns the database table name for notification templates.
func (NotificationTemplateModel) TableName() string {
	return "notification_templates"
}

// ToDomain converts a notification template database model to its domain representation.
func (m *NotificationTemplateModel) ToDomain() notification.NotificationTemplate {
	if m == nil {
		return notification.NotificationTemplate{}
	}
	return notification.NotificationTemplate{
		ID:            m.ID,
		TenantID:      m.TenantID,
		TemplateKey:   m.TemplateKey,
		Name:          m.Name,
		Content:       m.Content,
		VariablesJSON: m.VariablesJSON,
		IsActive:      m.IsActive,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

// NotificationTemplateModelFromDomain converts a notification template domain entity to a database model.
func NotificationTemplateModelFromDomain(t notification.NotificationTemplate) *NotificationTemplateModel {
	return &NotificationTemplateModel{
		ID:            t.ID,
		TenantID:      t.TenantID,
		TemplateKey:   t.TemplateKey,
		Name:          t.Name,
		Content:       t.Content,
		VariablesJSON: t.VariablesJSON,
		IsActive:      t.IsActive,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
}

// WANotificationModel is the GORM model for the WhatsApp notification
// queue/log (DATABASE-SCHEMA-ISP.md §2.10 — wa_notifications).
type WANotificationModel struct {
	ID         string  `gorm:"primaryKey"`
	TenantID   string  `gorm:"type:text;not null;default:tenant-default"`
	TemplateID *string `gorm:"column:template_id;type:text"`
	CustomerID *string `gorm:"column:customer_id;type:text;index"`
	InvoiceID  *string `gorm:"column:invoice_id;type:text"`

	RecipientPhone string `gorm:"type:varchar(20);not null"`
	MessageType    string `gorm:"type:varchar(50);not null"`
	MessageContent string `gorm:"type:text;not null"`

	Status       string `gorm:"type:varchar(20);not null;default:QUEUED;index"`
	ErrorMessage string `gorm:"type:text"`
	SentAt       *time.Time
	Attempts     int `gorm:"not null;default:0"`

	CreatedAt time.Time
}

// TableName returns the database table name for WhatsApp notifications.
func (WANotificationModel) TableName() string {
	return "wa_notifications"
}

// ToDomain converts a WhatsApp notification database model to its domain representation.
func (m *WANotificationModel) ToDomain() notification.WANotification {
	if m == nil {
		return notification.WANotification{}
	}
	return notification.WANotification{
		ID:             m.ID,
		TenantID:       m.TenantID,
		TemplateID:     m.TemplateID,
		CustomerID:     m.CustomerID,
		InvoiceID:      m.InvoiceID,
		RecipientPhone: m.RecipientPhone,
		MessageType:    m.MessageType,
		MessageContent: m.MessageContent,
		Status:         m.Status,
		ErrorMessage:   m.ErrorMessage,
		SentAt:         m.SentAt,
		CreatedAt:      m.CreatedAt,
	}
}

// WANotificationModelFromDomain converts a WhatsApp notification domain entity to a database model.
func WANotificationModelFromDomain(n notification.WANotification) *WANotificationModel {
	return &WANotificationModel{
		ID:             n.ID,
		TenantID:       n.TenantID,
		TemplateID:     n.TemplateID,
		CustomerID:     n.CustomerID,
		InvoiceID:      n.InvoiceID,
		RecipientPhone: n.RecipientPhone,
		MessageType:    n.MessageType,
		MessageContent: n.MessageContent,
		Status:         n.Status,
		ErrorMessage:   n.ErrorMessage,
		SentAt:         n.SentAt,
		CreatedAt:      n.CreatedAt,
	}
}
