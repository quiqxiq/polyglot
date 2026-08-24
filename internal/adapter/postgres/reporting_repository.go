package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/reporting"
	"github.com/quixiq/polyglot/internal/port"
)

type ReportingRepository struct {
	db *gorm.DB
}

var _ port.ReportingRepository = (*ReportingRepository)(nil)

// NewReportingRepository returns a port.ReportingRepository backed by GORM/Postgres.
func NewReportingRepository(db *gorm.DB) *ReportingRepository {
	return &ReportingRepository{db: db}
}

func (r *ReportingRepository) UpsertSnapshot(ctx context.Context, s reporting.DailyFinancialSnapshot) error {
	m := model.DailySnapshotModelFromDomain(s)
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND snapshot_date = ?", s.TenantID, s.SnapshotDate).
		Assign(*m).
		FirstOrCreate(m).Error
	if err != nil {
		return err
	}
	s.ID = m.ID
	return nil
}

func (r *ReportingRepository) GetByDate(ctx context.Context, tenantID string, date time.Time) (reporting.DailyFinancialSnapshot, error) {
	var m model.DailySnapshotModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND snapshot_date = ?", tenantID, date).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return reporting.DailyFinancialSnapshot{}, ErrNotFound
	}
	return m.ToDomain(), err
}

func (r *ReportingRepository) ListRange(ctx context.Context, tenantID string, from, to time.Time) ([]reporting.DailyFinancialSnapshot, error) {
	var mList []model.DailySnapshotModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND snapshot_date >= ? AND snapshot_date <= ?", tenantID, from, to).
		Order("snapshot_date asc").
		Find(&mList).Error
	if err != nil {
		return nil, err
	}
	out := make([]reporting.DailyFinancialSnapshot, len(mList))
	for i := range mList {
		out[i] = mList[i].ToDomain()
	}
	return out, nil
}

// RecomputeDaily implements port.SnapshotComputer: agregasi SQL harian →
// upsert daily_financial_snapshots (idempoten per tanggal).
func (r *ReportingRepository) RecomputeDaily(ctx context.Context, tenantID string, date time.Time) error {
	day := date.Format("2006-01-02")

	var invCount int
	var invTotal float64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*), COALESCE(SUM(total),0) FROM invoices
		WHERE tenant_id = ? AND created_at::date = ? AND deleted_at IS NULL`,
		tenantID, day).Row().Scan(&invCount, &invTotal)
	if err != nil {
		return fmt.Errorf("aggregate invoices: %w", err)
	}

	var payCount int
	var payTotal float64
	err = r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*), COALESCE(SUM(amount),0) FROM payments
		WHERE tenant_id = ? AND payment_date::date = ?`,
		tenantID, day).Row().Scan(&payCount, &payTotal)
	if err != nil {
		return fmt.Errorf("aggregate payments: %w", err)
	}

	var expenseTotal float64
	err = r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(amount),0) FROM cash_transactions
		WHERE tenant_id = ? AND direction = 'OUT' AND trx_date::date = ?`,
		tenantID, day).Scan(&expenseTotal).Error
	if err != nil {
		return fmt.Errorf("aggregate expenses: %w", err)
	}

	var outstanding float64
	err = r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(total - paid_amount),0) FROM invoices
		WHERE tenant_id = ? AND status IN ('UNPAID','PARTIAL','OVERDUE') AND deleted_at IS NULL`,
		tenantID).Scan(&outstanding).Error
	if err != nil {
		return fmt.Errorf("aggregate outstanding: %w", err)
	}

	var activeSubs int
	err = r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM subscriptions
		WHERE tenant_id = ? AND status = 'ACTIVE' AND deleted_at IS NULL`,
		tenantID).Scan(&activeSubs).Error
	if err != nil {
		return fmt.Errorf("count active subscriptions: %w", err)
	}

	balances, err := r.balanceJSON(tenantID)
	if err != nil {
		return err
	}

	snap := reporting.DailyFinancialSnapshot{
		TenantID:            tenantID,
		SnapshotDate:        date,
		InvoiceCount:        invCount,
		InvoiceTotal:        invTotal,
		PaymentCount:        payCount,
		PaymentTotal:        payTotal,
		OutstandingTotal:    outstanding,
		ExpenseTotal:        expenseTotal,
		ActiveSubscriptions: activeSubs,
		CashBalanceJSON:     balances,
	}
	return r.UpsertSnapshot(ctx, snap)
}

// balanceJSON menghitung saldo berjalan per rekening hingga tanggal tsb.
func (r *ReportingRepository) balanceJSON(tenantID string) ([]byte, error) {
	type row struct {
		AccountID string  `gorm:"column:account_id"`
		Signed    float64 `gorm:"column:signed"`
	}
	var rows []row
	err := r.db.Raw(`
		SELECT account_id,
		       SUM(CASE WHEN direction='IN' THEN amount ELSE -amount END) AS signed
		FROM cash_transactions WHERE tenant_id = ?
		GROUP BY account_id`, tenantID).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("aggregate balances: %w", err)
	}
	out := map[string]float64{}
	for _, rw := range rows {
		out[rw.AccountID] = rw.Signed
	}
	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	return b, nil
}
