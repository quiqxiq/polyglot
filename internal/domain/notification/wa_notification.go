package notification

import "time"

// WANotification represents one queued/sent WhatsApp notification record
// (DATABASE-SCHEMA-ISP.md §2.10 — wa_notifications).
type WANotification struct {
	ID             string     `json:"id"`
	TenantID       string     `json:"tenant_id"`
	TemplateID     *string    `json:"template_id,omitempty"`
	CustomerID     *string    `json:"customer_id,omitempty"`
	InvoiceID      *string    `json:"invoice_id,omitempty"`
	RecipientPhone string     `json:"recipient_phone"`
	MessageType    string     `json:"message_type"`
	MessageContent string     `json:"message_content"` // hasil render akhir
	Status         string     `json:"status"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	SentAt         *time.Time `json:"sent_at,omitempty"`
	Attempts       int        `json:"attempts"`

	CreatedAt time.Time `json:"created_at"`
}
