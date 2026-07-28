package billing

import "time"

// Invoice represents a billing invoice mapped to the `invoices` table per
// migration 000014. Items are stored separately in the `invoice_items` table.
type Invoice struct {
	ID            string
	InvoiceNumber string
	CustomerID    string
	BillingRunID  *string
	PeriodStart   time.Time
	PeriodEnd     time.Time
	IssueDate     time.Time
	DueDate       time.Time
	Status        string // "draft", "issued", "partially_paid", "paid", "overdue", "void"
	Subtotal      float64
	TaxAmount     float64
	TotalAmount   float64
	Notes         string
	CreatedAt     time.Time

	// Items holds the line items for this invoice. Loaded eagerly via
	// GORM Preload or manually by the repository.
	Items []InvoiceItem
}

// InvoiceItem represents a single line item within an invoice, mapped to
// the `invoice_items` table per migration 000014.
type InvoiceItem struct {
	ID             string
	InvoiceID      string
	SubscriptionID *string
	ItemType       string // "subscription_fee", "installation_fee", etc.
	Description    string
	Quantity       float64
	UnitPrice      float64
	Amount         float64
}
