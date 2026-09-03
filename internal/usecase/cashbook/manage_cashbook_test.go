package cashbook

import (
	"context"
	"testing"

	domainCashbook "github.com/quixiq/polyglot/internal/domain/cashbook"
	"github.com/quixiq/polyglot/internal/port"
)

type mockCashbookRepo struct {
	accounts     []domainCashbook.CashAccount
	categories   []domainCashbook.CashCategory
	transactions []domainCashbook.CashTransaction
}

func (m *mockCashbookRepo) FindAccounts(ctx context.Context, activeOnly bool) ([]domainCashbook.CashAccount, error) {
	return m.accounts, nil
}

func (m *mockCashbookRepo) FindAccountByID(ctx context.Context, id string) (domainCashbook.CashAccount, error) {
	for _, a := range m.accounts {
		if a.ID == id {
			return a, nil
		}
	}
	return domainCashbook.CashAccount{}, domainCashbook.ErrNotFound
}

func (m *mockCashbookRepo) SaveAccount(ctx context.Context, a domainCashbook.CashAccount) error {
	m.accounts = append(m.accounts, a)
	return nil
}

func (m *mockCashbookRepo) FindCategories(ctx context.Context, activeOnly bool) ([]domainCashbook.CashCategory, error) {
	return m.categories, nil
}

func (m *mockCashbookRepo) FindCategoryByID(ctx context.Context, id string) (domainCashbook.CashCategory, error) {
	for _, c := range m.categories {
		if c.ID == id {
			return c, nil
		}
	}
	return domainCashbook.CashCategory{}, domainCashbook.ErrNotFound
}

func (m *mockCashbookRepo) SaveCategory(ctx context.Context, c domainCashbook.CashCategory) error {
	m.categories = append(m.categories, c)
	return nil
}

func (m *mockCashbookRepo) SaveTransaction(ctx context.Context, t domainCashbook.CashTransaction) error {
	m.transactions = append(m.transactions, t)
	return nil
}

func (m *mockCashbookRepo) FindTransactions(ctx context.Context, filter port.CashTransactionFilter) ([]domainCashbook.CashTransaction, error) {
	return m.transactions, nil
}

func (m *mockCashbookRepo) BalanceByAccounts(ctx context.Context, filter port.CashTransactionFilter) (map[string]float64, error) {
	return map[string]float64{"ca-1": 150000}, nil
}

func TestManageCashbookUseCase(t *testing.T) {
	ctx := context.Background()
	repo := &mockCashbookRepo{}
	uc := NewManageCashbookUseCase(repo)

	// Test SaveAccount with default generation
	acc, err := uc.SaveAccount(ctx, domainCashbook.CashAccount{
		Name: "Kas Utama",
	})
	if err != nil {
		t.Fatalf("unexpected error saving account: %v", err)
	}
	if acc.ID == "" || acc.Type != domainCashbook.AccountTypeCash || acc.TenantID != "tenant-default" {
		t.Errorf("expected defaults applied to account, got %+v", acc)
	}

	// Test ListAccounts
	accounts, err := uc.ListAccounts(ctx, false)
	if err != nil || len(accounts) != 1 {
		t.Fatalf("expected 1 account, got %d (err: %v)", len(accounts), err)
	}

	// Test SaveCategory
	cat, err := uc.SaveCategory(ctx, domainCashbook.CashCategory{
		Name: "Listrik",
	})
	if err != nil {
		t.Fatalf("unexpected error saving category: %v", err)
	}
	if cat.ID == "" || cat.Type != domainCashbook.CategoryTypeExpense {
		t.Errorf("expected defaults applied to category, got %+v", cat)
	}

	// Test ListCategories
	categories, err := uc.ListCategories(ctx, false)
	if err != nil || len(categories) != 1 {
		t.Fatalf("expected 1 category, got %d (err: %v)", len(categories), err)
	}

	// Test AddTransaction
	trx, err := uc.AddTransaction(ctx, AddTransactionInput{
		AccountID:   acc.ID,
		CategoryID:  cat.ID,
		Amount:      50000,
		Description: "Bayar token listrik",
	})
	if err != nil {
		t.Fatalf("unexpected error adding transaction: %v", err)
	}
	if trx.TransactionNo == "" || trx.Direction != domainCashbook.DirectionOut {
		t.Errorf("expected generated transaction with default DirectionOut, got %+v", trx)
	}

	// Test ListTransactions
	trxs, err := uc.ListTransactions(ctx, port.CashTransactionFilter{})
	if err != nil || len(trxs) != 1 {
		t.Fatalf("expected 1 transaction, got %d (err: %v)", len(trxs), err)
	}

	// Test Balances
	balances, err := uc.Balances(ctx, port.CashTransactionFilter{})
	if err != nil || balances["ca-1"] != 150000 {
		t.Fatalf("unexpected balances: %v", balances)
	}
}
