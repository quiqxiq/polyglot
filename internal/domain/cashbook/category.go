package cashbook

import "time"

// Kategori pos arus kas.
const (
	CategoryTypeIncome  = "INCOME"
	CategoryTypeExpense = "EXPENSE"
)

// CashCategory represents an income/expense category of the cash book
// (DATABASE-SCHEMA-ISP.md §2.8 — cash_categories).
type CashCategory struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
	Name     string `json:"name"` // 'Tagihan Pelanggan', 'Biaya Listrik', 'Bandwidth Uplink', 'Gaji'
	Type     string `json:"type"`
	IsActive bool   `json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
