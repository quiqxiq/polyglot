package notification

import "time"

// Status pengiriman notifikasi WhatsApp.
const (
	StatusQueued = "QUEUED"
	StatusSent   = "SENT"
	StatusFailed = "FAILED"
)

// NotificationTemplate represents a customizable WhatsApp message template
// with placeholder variables ({{customer_name}}, {{amount}}, ...)
// (DATABASE-SCHEMA-ISP.md §2.10 — notification_templates).
type NotificationTemplate struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	TemplateKey   string    `json:"template_key"` // 'BILL_REMINDER', 'PAYMENT_RECEIPT', ...
	Name          string    `json:"name"`
	Content       string    `json:"content"`
	VariablesJSON string    `json:"variables_json"` // JSON array nama variabel
	IsActive      bool      `json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
