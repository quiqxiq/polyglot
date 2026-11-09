package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/cashbook"
	"github.com/quixiq/polyglot/internal/domain/debt"
	domainTenant "github.com/quixiq/polyglot/internal/domain/tenant"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// DebtRepository implements port.DebtRepository backed by PostgreSQL.
type DebtRepository struct {
	db *gorm.DB
}

var _ port.DebtRepository = (*DebtRepository)(nil)

// NewDebtRepository creates a new port.DebtRepository.
func NewDebtRepository(db *gorm.DB) *DebtRepository {
	return &DebtRepository{db: db}
}

// Save stores or updates a customer debt record.
func (r *DebtRepository) Save(ctx context.Context, d debt.Debt) error {
	if d.ID == "" {
		d.ID = idgen.New("debt")
	}
	if d.TenantID == "" {
		d.TenantID = domainTenant.DefaultID
	}
	now := time.Now().UTC()
	if d.CreatedAt.IsZero() {
		d.CreatedAt = now
	}
	d.UpdatedAt = now
	mdl := model.DebtModelFromDomain(d)
	return r.db.WithContext(ctx).Save(mdl).Error
}

// FindByID retrieves a debt record by ID.
func (r *DebtRepository) FindByID(ctx context.Context, id string) (debt.Debt, error) {
	var m model.DebtModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return debt.Debt{}, debt.ErrNotFound
		}
		return debt.Debt{}, err
	}
	return m.ToDomain(), nil
}

// List returns filtered debts along with the total count matching the filter.
func (r *DebtRepository) List(ctx context.Context, f port.DebtFilter) ([]debt.Debt, int, error) {
	tenantID := f.TenantID
	if tenantID == "" {
		tenantID = domainTenant.DefaultID
	}
	q := r.db.WithContext(ctx).Model(&model.DebtModel{}).Where("customer_debts.tenant_id = ?", tenantID)

	if f.CustomerID != "" {
		q = q.Where("customer_debts.customer_id = ?", f.CustomerID)
	}
	if f.Status != "" {
		q = q.Where("customer_debts.status = ?", f.Status)
	}
	if f.OverdueOnly {
		asOf := time.Now().UTC()
		if f.AsOf != nil {
			asOf = *f.AsOf
		}
		q = q.Where("customer_debts.status IN ('OPEN','PARTIAL') AND customer_debts.due_date IS NOT NULL AND customer_debts.due_date < ?", asOf)
	}
	if strings.TrimSpace(f.Search) != "" {
		search := "%" + strings.TrimSpace(f.Search) + "%"
		q = q.Joins("LEFT JOIN customers ON customers.id = customer_debts.customer_id").
			Where("customer_debts.description ILIKE ? OR customer_debts.id ILIKE ? OR customers.name ILIKE ?", search, search, search)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if f.Limit > 0 {
		q = q.Limit(f.Limit).Offset(f.Offset)
	}
	var list []model.DebtModel
	if err := q.Order("customer_debts.created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	out := make([]debt.Debt, len(list))
	for i, m := range list {
		out[i] = m.ToDomain()
	}
	return out, int(total), nil
}

// ListByCustomer returns all debts for a specific customer.
func (r *DebtRepository) ListByCustomer(ctx context.Context, customerID string) ([]debt.Debt, error) {
	var list []model.DebtModel
	if err := r.db.WithContext(ctx).
		Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]debt.Debt, len(list))
	for i, m := range list {
		out[i] = m.ToDomain()
	}
	return out, nil
}

// ApplyPayment atomik melunasi/mencicil hutang dan mencatat mutasi kas masuk.
func (r *DebtRepository) ApplyPayment(ctx context.Context, cmd port.DebtPaymentCommand) (debt.Payment, error) {
	if cmd.DebtID == "" || cmd.Amount <= 0 {
		return debt.Payment{}, fmt.Errorf("%w: debt_id and positive amount are required", debt.ErrInvalidInput)
	}
	if cmd.CashAccountID == "" || cmd.IncomeCategoryID == "" {
		return debt.Payment{}, fmt.Errorf("%w: cash_account_id and income_category_id are required", debt.ErrInvalidInput)
	}

	tenantID := cmd.TenantID
	if tenantID == "" {
		tenantID = domainTenant.DefaultID
	}

	var outPayment debt.Payment
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := tx
		if r.db.Dialector.Name() == "postgres" {
			q = q.Clauses(lockingClause())
		}

		var d model.DebtModel
		if err := q.First(&d, "id = ?", cmd.DebtID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return debt.ErrNotFound
			}
			return err
		}

		switch d.Status {
		case debt.StatusPaid:
			return debt.ErrAlreadyPaid
		case debt.StatusCancelled:
			return debt.ErrCancelled
		}
		if d.RemainingAmount <= 0 {
			return debt.ErrAlreadyPaid
		}
		if cmd.Amount > d.RemainingAmount+1e-9 {
			return fmt.Errorf("%w: sisa hutang %.2f, dibayar %.2f", debt.ErrOverpayment, d.RemainingAmount, cmd.Amount)
		}

		now := time.Now().UTC()
		payDate := cmd.PaymentDate
		if payDate.IsZero() {
			payDate = now
		}

		// 1. Catat jurnal kas masuk
		trxID := newID("trx")
		cashTrx := &model.CashTransactionModel{
			ID:            trxID,
			TenantID:      tenantID,
			TransactionNo: fmt.Sprintf("TRX-%s-%06d", now.Format("200601"), nextDocumentSeq(tx, "cash_transactions_no_seq")),
			AccountID:     cmd.CashAccountID,
			CategoryID:    cmd.IncomeCategoryID,
			Direction:     cashbook.DirectionIn,
			Amount:        cmd.Amount,
			TrxDate:       payDate,
			SourceType:    "DEBT_PAYMENT",
			SourceID:      d.ID,
			Description:   fmt.Sprintf("Pembayaran hutang %s", d.ID),
			RecordedBy:    cmd.RecordedBy,
		}
		if err := tx.Create(cashTrx).Error; err != nil {
			return fmt.Errorf("failed to create cash transaction: %w", err)
		}

		// 2. Buat debt payment
		payID := newID("dp")
		debtPay := &model.DebtPaymentModel{
			ID:                payID,
			TenantID:          tenantID,
			DebtID:            d.ID,
			Amount:            cmd.Amount,
			PaymentMethodID:   cmd.PaymentMethodID,
			CashTransactionID: &trxID,
			PaymentDate:       payDate,
			Notes:             cmd.Notes,
			RecordedBy:        cmd.RecordedBy,
			CreatedAt:         now,
		}
		if err := tx.Create(debtPay).Error; err != nil {
			return fmt.Errorf("failed to create debt payment: %w", err)
		}

		// 3. Update debt remaining & status
		newRemaining := d.RemainingAmount - cmd.Amount
		if newRemaining < 0 {
			newRemaining = 0
		}
		newStatus := debt.StatusPartial
		var paidAt *time.Time
		if newRemaining <= 1e-9 {
			newRemaining = 0
			newStatus = debt.StatusPaid
			paidAt = &now
		}

		updates := map[string]any{
			"remaining_amount": newRemaining,
			"status":           newStatus,
			"updated_at":       now,
		}
		if paidAt != nil {
			updates["paid_at"] = paidAt
		}
		if err := tx.Model(&model.DebtModel{}).Where("id = ?", d.ID).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update debt: %w", err)
		}

		outPayment = debtPay.ToDomain()
		return nil
	})
	if err != nil {
		return debt.Payment{}, err
	}
	return outPayment, nil
}

// ListPayments returns payment history for a debt record.
func (r *DebtRepository) ListPayments(ctx context.Context, debtID string) ([]debt.Payment, error) {
	var list []model.DebtPaymentModel
	if err := r.db.WithContext(ctx).
		Where("debt_id = ?", debtID).
		Order("payment_date DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]debt.Payment, len(list))
	for i, m := range list {
		out[i] = m.ToDomain()
	}
	return out, nil
}

// Summary returns aggregated stats for customer debts.
func (r *DebtRepository) Summary(ctx context.Context, tenantID, customerID string, asOf time.Time) (debt.Summary, error) {
	if tenantID == "" {
		tenantID = domainTenant.DefaultID
	}
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}

	q := r.db.WithContext(ctx).Model(&model.DebtModel{}).
		Where("tenant_id = ? AND status != ?", tenantID, debt.StatusCancelled)
	if customerID != "" {
		q = q.Where("customer_id = ?", customerID)
	}

	type aggResult struct {
		Count       int
		TotalAmount float64
		Outstanding float64
	}
	var agg aggResult
	if err := q.Select("COUNT(*) as count, COALESCE(SUM(amount), 0) as total_amount, COALESCE(SUM(remaining_amount), 0) as outstanding").
		Scan(&agg).Error; err != nil {
		return debt.Summary{}, err
	}

	var overdueCount int64
	qOverdue := r.db.WithContext(ctx).Model(&model.DebtModel{}).
		Where("tenant_id = ? AND status IN ('OPEN','PARTIAL') AND due_date IS NOT NULL AND due_date < ?", tenantID, asOf)
	if customerID != "" {
		qOverdue = qOverdue.Where("customer_id = ?", customerID)
	}
	if err := qOverdue.Count(&overdueCount).Error; err != nil {
		return debt.Summary{}, err
	}

	return debt.Summary{
		Count:       agg.Count,
		TotalAmount: agg.TotalAmount,
		Outstanding: agg.Outstanding,
		Overdue:     int(overdueCount),
	}, nil
}
