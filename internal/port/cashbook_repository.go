package port

import (
	"context"
	"time"

	"github.com/quixiq/polyglot/internal/domain/cashbook"
)

// CashTransactionFilter narrows cash journal queries.
// Zero values are ignored; From/To bound trx_date inclusively.
type CashTransactionFilter struct {
	AccountID  string
	CategoryID string
	Direction  string // 'IN' | 'OUT' | ''
	From       time.Time
	To         time.Time
	SourceType string
	SourceID   string
	Limit      int
}

// CashbookRepository defines persistence operations for the simple cash
// book: accounts, categories and the IN/OUT journal
// (DATABASE-SCHEMA-ISP.md §2.8). Digabung per area sesuai keputusan desain.
type CashbookRepository interface {
	// Accounts
	SaveAccount(ctx context.Context, a cashbook.CashAccount) error
	FindAccounts(ctx context.Context, activeOnly bool) ([]cashbook.CashAccount, error)
	FindAccountByID(ctx context.Context, id string) (cashbook.CashAccount, error)

	// Categories
	SaveCategory(ctx context.Context, c cashbook.CashCategory) error
	FindCategories(ctx context.Context, activeOnly bool) ([]cashbook.CashCategory, error)
	FindCategoryByID(ctx context.Context, id string) (cashbook.CashCategory, error)

	// Journal
	SaveTransaction(ctx context.Context, t cashbook.CashTransaction) error
	FindTransactions(ctx context.Context, f CashTransactionFilter) ([]cashbook.CashTransaction, error)

	// Balance returns signed net balance per account id over the filter
	// window (IN positive, OUT negative).
	BalanceByAccounts(ctx context.Context, f CashTransactionFilter) (map[string]float64, error)
}
