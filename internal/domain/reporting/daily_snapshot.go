package reporting

import (
	"encoding/json"
	"time"
)

// DailyFinancialSnapshot is a pre-aggregated daily recap powering instant
// daily/monthly/yearly financial reports
// (DATABASE-SCHEMA-ISP.md §2.9 — daily_financial_snapshots).
// Unik per (tenant_id, snapshot_date); di-refresh oleh job agregasi harian.
type DailyFinancialSnapshot struct {
	ID                  uint            `json:"id"`
	TenantID            string          `json:"tenant_id"`
	SnapshotDate        time.Time       `json:"snapshot_date"`
	InvoiceCount        int             `json:"invoice_count"`
	InvoiceTotal        float64         `json:"invoice_total"`
	PaymentCount        int             `json:"payment_count"`
	PaymentTotal        float64         `json:"payment_total"`
	OutstandingTotal    float64         `json:"outstanding_total"`
	ExpenseTotal        float64         `json:"expense_total"`
	ActiveSubscriptions int             `json:"active_subscriptions"`
	CashBalanceJSON     json.RawMessage `json:"cash_balance_json"`

	CreatedAt time.Time `json:"created_at"`
}
