package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/cashbook"
	"github.com/quixiq/polyglot/internal/port"
)

type CashbookRepository struct {
	db *gorm.DB
}

var _ port.CashbookRepository = (*CashbookRepository)(nil)

// NewCashbookRepository returns a port.CashbookRepository backed by GORM/Postgres.
func NewCashbookRepository(db *gorm.DB) *CashbookRepository {
	return &CashbookRepository{db: db}
}

func (r *CashbookRepository) SaveAccount(ctx context.Context, a cashbook.CashAccount) error {
	return r.db.WithContext(ctx).Save(model.CashAccountModelFromDomain(a)).Error
}

func (r *CashbookRepository) FindAccounts(ctx context.Context, activeOnly bool) ([]cashbook.CashAccount, error) {
	var mList []model.CashAccountModel
	q := r.db.WithContext(ctx)
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	if err := q.Order("account_code").Find(&mList).Error; err != nil {
		return nil, err
	}
	out := make([]cashbook.CashAccount, len(mList))
	for i := range mList {
		out[i] = mList[i].ToDomain()
	}
	return out, nil
}

func (r *CashbookRepository) FindAccountByID(ctx context.Context, id string) (cashbook.CashAccount, error) {
	var m model.CashAccountModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cashbook.CashAccount{}, ErrNotFound
	}
	return m.ToDomain(), err
}

func (r *CashbookRepository) SaveCategory(ctx context.Context, c cashbook.CashCategory) error {
	return r.db.WithContext(ctx).Save(model.CashCategoryModelFromDomain(c)).Error
}

func (r *CashbookRepository) FindCategories(ctx context.Context, activeOnly bool) ([]cashbook.CashCategory, error) {
	var mList []model.CashCategoryModel
	q := r.db.WithContext(ctx)
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	if err := q.Order("name").Find(&mList).Error; err != nil {
		return nil, err
	}
	out := make([]cashbook.CashCategory, len(mList))
	for i := range mList {
		out[i] = mList[i].ToDomain()
	}
	return out, nil
}

func (r *CashbookRepository) FindCategoryByID(ctx context.Context, id string) (cashbook.CashCategory, error) {
	var m model.CashCategoryModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cashbook.CashCategory{}, ErrNotFound
	}
	return m.ToDomain(), err
}

func (r *CashbookRepository) SaveTransaction(ctx context.Context, t cashbook.CashTransaction) error {
	m := model.CashTransactionModelFromDomain(t)
	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		return err
	}
	t.CreatedAt = m.CreatedAt
	return nil
}

func (r *CashbookRepository) FindTransactions(ctx context.Context, f port.CashTransactionFilter) ([]cashbook.CashTransaction, error) {
	q := r.applyFilter(ctx, f)
	var mList []model.CashTransactionModel
	if f.Limit > 0 {
		q = q.Limit(f.Limit)
	}
	if err := q.Order("trx_date desc").Find(&mList).Error; err != nil {
		return nil, err
	}
	out := make([]cashbook.CashTransaction, len(mList))
	for i := range mList {
		out[i] = mList[i].ToDomain()
	}
	return out, nil
}

func (r *CashbookRepository) BalanceByAccounts(ctx context.Context, f port.CashTransactionFilter) (map[string]float64, error) {
	type row struct {
		AccountID string  `gorm:"column:account_id"`
		Signed    float64 `gorm:"column:signed"`
	}
	q := r.applyFilter(ctx, f)
	var rows []row
	err := q.Select("account_id, SUM(CASE WHEN direction = 'IN' THEN amount ELSE -amount END) AS signed").
		Group("account_id").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]float64, len(rows))
	for _, rrow := range rows {
		out[rrow.AccountID] = rrow.Signed
	}
	return out, nil
}

func (r *CashbookRepository) applyFilter(ctx context.Context, f port.CashTransactionFilter) *gorm.DB {
	q := r.db.WithContext(ctx).Model(&model.CashTransactionModel{})
	if f.AccountID != "" {
		q = q.Where("account_id = ?", f.AccountID)
	}
	if f.CategoryID != "" {
		q = q.Where("category_id = ?", f.CategoryID)
	}
	if f.Direction != "" {
		q = q.Where("direction = ?", f.Direction)
	}
	if f.SourceType != "" {
		q = q.Where("source_type = ?", f.SourceType)
	}
	if f.SourceID != "" {
		q = q.Where("source_id = ?", f.SourceID)
	}
	if !f.From.IsZero() {
		q = q.Where("trx_date >= ?", f.From)
	}
	if !f.To.IsZero() {
		q = q.Where("trx_date <= ?", f.To)
	}
	return q
}
