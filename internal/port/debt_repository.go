package port

import (
	"context"
	"time"

	"github.com/quixiq/polyglot/internal/domain/debt"
)

// DebtFilter menyaring daftar hutang.
type DebtFilter struct {
	TenantID    string
	CustomerID  string
	Status      string
	Search      string
	OverdueOnly bool
	AsOf        *time.Time
	Limit       int
	Offset      int
}

// DebtPaymentCommand adalah perintah pelunasan hutang (atomik di adapter).
type DebtPaymentCommand struct {
	TenantID         string
	DebtID           string
	Amount           float64
	PaymentMethodID  *string
	PaymentDate      time.Time
	Notes            string
	RecordedBy       *uint
	CashAccountID    string
	IncomeCategoryID string
}

// DebtRepository adalah persistensi hutang-piutang + pembayaran & jurnal kas.
type DebtRepository interface {
	Save(ctx context.Context, d debt.Debt) error
	FindByID(ctx context.Context, id string) (debt.Debt, error)
	List(ctx context.Context, f DebtFilter) ([]debt.Debt, int, error)
	ListByCustomer(ctx context.Context, customerID string) ([]debt.Debt, error)
	ApplyPayment(ctx context.Context, cmd DebtPaymentCommand) (debt.Payment, error)
	ListPayments(ctx context.Context, debtID string) ([]debt.Payment, error)
	Summary(ctx context.Context, tenantID, customerID string, asOf time.Time) (debt.Summary, error)
}
