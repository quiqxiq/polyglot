package cashbook

import "time"

// Tipe rekening kas.
const (
	AccountTypeCash = "CASH"
	AccountTypeBank = "BANK"
)

// CashAccount represents a cash/bank operational account
// (DATABASE-SCHEMA-ISP.md §2.8 — cash_accounts).
type CashAccount struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	AccountCode string    `json:"account_code"` // '1001-KAS-KANTOR', '1002-BANK-BCA'
	Name        string    `json:"name"`         // 'Kas Kasir Utama', 'Rekening BCA Operasional'
	Type        string    `json:"type"`
	IsActive    bool      `json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
