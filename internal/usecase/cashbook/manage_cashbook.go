package cashbook

import (
	"context"
	"fmt"
	"time"

	domainCashbook "github.com/quixiq/polyglot/internal/domain/cashbook"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// ManageCashbookUseCase orchestrates cashbook accounts, categories, and ledger entries.
type ManageCashbookUseCase struct {
	repo port.CashbookRepository
}

// NewManageCashbookUseCase constructs a ManageCashbookUseCase.
func NewManageCashbookUseCase(repo port.CashbookRepository) *ManageCashbookUseCase {
	return &ManageCashbookUseCase{repo: repo}
}

// ListAccounts returns cash accounts, optionally filtered by active status.
func (uc *ManageCashbookUseCase) ListAccounts(ctx context.Context, activeOnly bool) ([]domainCashbook.CashAccount, error) {
	if uc.repo == nil {
		return nil, domainCashbook.ErrNotFound
	}
	accounts, err := uc.repo.FindAccounts(ctx, activeOnly)
	if err != nil {
		return nil, fmt.Errorf("find accounts: %w", err)
	}
	return accounts, nil
}

// SaveAccount creates or updates a cash account, ensuring defaults and IDs.
func (uc *ManageCashbookUseCase) SaveAccount(ctx context.Context, a domainCashbook.CashAccount) (*domainCashbook.CashAccount, error) {
	if uc.repo == nil {
		return nil, domainCashbook.ErrInvalidInput
	}
	if a.ID == "" {
		a.ID = idgen.New("ca")
	}
	if a.TenantID == "" {
		a.TenantID = "tenant-default"
	}
	if a.Type == "" {
		a.Type = domainCashbook.AccountTypeCash
	}
	if err := uc.repo.SaveAccount(ctx, a); err != nil {
		return nil, fmt.Errorf("save account: %w", err)
	}
	return &a, nil
}

// ListCategories returns cash categories.
func (uc *ManageCashbookUseCase) ListCategories(ctx context.Context, activeOnly bool) ([]domainCashbook.CashCategory, error) {
	if uc.repo == nil {
		return nil, domainCashbook.ErrNotFound
	}
	categories, err := uc.repo.FindCategories(ctx, activeOnly)
	if err != nil {
		return nil, fmt.Errorf("find categories: %w", err)
	}
	return categories, nil
}

// SaveCategory creates or updates a category.
func (uc *ManageCashbookUseCase) SaveCategory(ctx context.Context, c domainCashbook.CashCategory) (*domainCashbook.CashCategory, error) {
	if uc.repo == nil {
		return nil, domainCashbook.ErrInvalidInput
	}
	if c.ID == "" {
		c.ID = idgen.New("cc")
	}
	if c.TenantID == "" {
		c.TenantID = "tenant-default"
	}
	if c.Type == "" {
		c.Type = domainCashbook.CategoryTypeExpense
	}
	if err := uc.repo.SaveCategory(ctx, c); err != nil {
		return nil, fmt.Errorf("save category: %w", err)
	}
	return &c, nil
}

// AddTransactionInput represents parameters for creating a cash transaction.
type AddTransactionInput struct {
	AccountID   string
	CategoryID  string
	Direction   string
	Amount      float64
	Description string
}

// AddTransaction records a cash ledger transaction with generated transaction number.
func (uc *ManageCashbookUseCase) AddTransaction(ctx context.Context, input AddTransactionInput) (*domainCashbook.CashTransaction, error) {
	if uc.repo == nil {
		return nil, domainCashbook.ErrInvalidInput
	}
	if input.AccountID == "" {
		return nil, fmt.Errorf("%w: account id is required", domainCashbook.ErrInvalidInput)
	}
	now := time.Now()
	dir := input.Direction
	if dir == "" {
		dir = domainCashbook.DirectionOut
	}
	t := domainCashbook.CashTransaction{
		ID:            idgen.New("trx"),
		TenantID:      "tenant-default",
		TransactionNo: fmt.Sprintf("TRX-%s-%06d", now.Format("200601"), now.UnixNano()%1000000),
		AccountID:     input.AccountID,
		CategoryID:    input.CategoryID,
		Direction:     dir,
		Amount:        input.Amount,
		TrxDate:       now,
		SourceType:    domainCashbook.SourceExpense,
		Description:   input.Description,
		CreatedAt:     now,
	}
	if err := uc.repo.SaveTransaction(ctx, t); err != nil {
		return nil, fmt.Errorf("save transaction: %w", err)
	}
	return &t, nil
}

// ListTransactions retrieves transactions based on query filter.
func (uc *ManageCashbookUseCase) ListTransactions(ctx context.Context, filter port.CashTransactionFilter) ([]domainCashbook.CashTransaction, error) {
	if uc.repo == nil {
		return nil, domainCashbook.ErrNotFound
	}
	transactions, err := uc.repo.FindTransactions(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("find transactions: %w", err)
	}
	return transactions, nil
}

// Balances calculates account balances for a time window.
func (uc *ManageCashbookUseCase) Balances(ctx context.Context, filter port.CashTransactionFilter) (map[string]float64, error) {
	if uc.repo == nil {
		return nil, domainCashbook.ErrNotFound
	}
	balances, err := uc.repo.BalanceByAccounts(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("balance by accounts: %w", err)
	}
	return balances, nil
}
