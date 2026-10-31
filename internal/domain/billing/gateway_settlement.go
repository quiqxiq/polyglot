package billing

import "time"

// Status settlement payment gateway.
const (
	SettlementStatusPending     = "PENDING"
	SettlementStatusMatched     = "MATCHED"
	SettlementStatusDiscrepancy = "DISCREPANCY"
)

// GatewaySettlement mencatat settlement/pencairan dana dari payment gateway
// untuk rekonsiliasi terhadap transaksi & pembayaran yang tercatat.
type GatewaySettlement struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	Gateway       string     `json:"gateway"`
	SettlementRef string     `json:"settlement_ref"`
	PeriodFrom    *time.Time `json:"period_from,omitempty"`
	PeriodTo      *time.Time `json:"period_to,omitempty"`
	GrossAmount   float64    `json:"gross_amount"`
	FeeAmount     float64    `json:"fee_amount"`
	NetAmount     float64    `json:"net_amount"`
	BankAccountID *string    `json:"bank_account_id,omitempty"`
	Status        string     `json:"status"`
	Notes         string     `json:"notes,omitempty"`
	CreatedBy     *uint      `json:"created_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
