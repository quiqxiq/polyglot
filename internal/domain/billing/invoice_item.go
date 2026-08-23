package billing

import "time"

// Tipe item baris faktur.
const (
	ItemTypeSubscriptionFee = "SUBSCRIPTION_FEE"
	ItemTypeInstallationFee = "INSTALLATION_FEE"
	ItemTypeAdHoc           = "AD_HOC"
)

// InvoiceItem represents one line item of an invoice
// (DATABASE-SCHEMA-ISP.md §2.6 — invoice_items).
type InvoiceItem struct {
	ID          string    `json:"id"`
	InvoiceID   string    `json:"invoice_id"`
	Description string    `json:"description"`
	Quantity    float64   `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	Amount      float64   `json:"amount"` // quantity * unit_price
	ItemType    string    `json:"item_type"`
	CreatedAt   time.Time `json:"created_at"`
}
