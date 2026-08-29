package cashbook

import "time"

// Arah mutasi kas.
const (
	DirectionIn  = "IN"  // kas masuk
	DirectionOut = "OUT" // kas keluar
)

// Sumber transaksi kas.
const (
	SourcePayment  = "PAYMENT" // otomatis saat invoice lunas
	SourceExpense  = "EXPENSE" // pencatatan manual operasional
	SourceTransfer = "TRANSFER"
)

// CashTransaction represents one cash-in/cash-out journal entry
// (DATABASE-SCHEMA-ISP.md §2.8 — cash_transactions).
type CashTransaction struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	TransactionNo string    `json:"transaction_no"` // "TRX-202608-00125"
	AccountID     string    `json:"account_id"`
	CategoryID    string    `json:"category_id"`
	Direction     string    `json:"direction"` // IN | OUT
	Amount        float64   `json:"amount"`
	TrxDate       time.Time `json:"trx_date"`
	SourceType    string    `json:"source_type,omitempty"`
	SourceID      string    `json:"source_id,omitempty"` // payments.id bila SourcePayment
	Description   string    `json:"description"`
	RecordedBy    *uint     `json:"recorded_by,omitempty"` // users.id

	CreatedAt time.Time `json:"created_at"`
}
