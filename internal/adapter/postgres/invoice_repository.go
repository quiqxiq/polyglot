package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
)

type InvoiceRepository struct {
	db *gorm.DB
}

var _ port.InvoiceRepository = (*InvoiceRepository)(nil)
var _ port.InvoiceFilteredRepository = (*InvoiceRepository)(nil)

// NewInvoiceRepository returns a port.InvoiceRepository backed by GORM/Postgres.
func NewInvoiceRepository(db *gorm.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) Save(ctx context.Context, inv billing.Invoice) error {
	m := model.InvoiceModelFromDomain(inv)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *InvoiceRepository) FindByID(ctx context.Context, id string) (billing.Invoice, error) {
	var m model.InvoiceModel
	err := r.db.WithContext(ctx).First(&m, "id = ? AND deleted_at IS NULL", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return billing.Invoice{}, ErrNotFound
	}
	return m.ToDomain(), err
}

func (r *InvoiceRepository) FindByCustomerID(ctx context.Context, customerID string) ([]billing.Invoice, error) {
	var mList []model.InvoiceModel
	err := r.db.WithContext(ctx).Where("customer_id = ? AND deleted_at IS NULL", customerID).Order("created_at desc").Find(&mList).Error
	if err != nil {
		return nil, err
	}
	invoices := make([]billing.Invoice, len(mList))
	for i, m := range mList {
		invoices[i] = m.ToDomain()
	}
	return invoices, nil
}

func (r *InvoiceRepository) FindAll(ctx context.Context) ([]billing.Invoice, error) {
	var mList []model.InvoiceModel
	err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Order("created_at desc").Find(&mList).Error
	if err != nil {
		return nil, err
	}
	invoices := make([]billing.Invoice, len(mList))
	for i, m := range mList {
		invoices[i] = m.ToDomain()
	}
	return invoices, nil
}

// FindPaged implements tenant + limit/offset filtering (F6-8).
func (r *InvoiceRepository) FindPaged(ctx context.Context, f port.PageFilter) ([]billing.Invoice, error) {
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL")
	if f.TenantID != "" {
		q = q.Where("tenant_id = ?", f.TenantID)
	}
	if f.Limit > 0 {
		q = q.Limit(f.Limit)
	}
	if f.Offset > 0 {
		q = q.Offset(f.Offset)
	}
	var mList []model.InvoiceModel
	if err := q.Order("created_at desc").Find(&mList).Error; err != nil {
		return nil, err
	}
	invoices := make([]billing.Invoice, len(mList))
	for i, m := range mList {
		invoices[i] = m.ToDomain()
	}
	return invoices, nil
}

func (r *InvoiceRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	res := r.db.WithContext(ctx).Model(&model.InvoiceModel{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// FindByPaymentCode implements the cashier quick-pay lookup (§4.2).
func (r *InvoiceRepository) FindByPaymentCode(ctx context.Context, code string) (billing.Invoice, error) {
	var m model.InvoiceModel
	err := r.db.WithContext(ctx).First(&m, "manual_payment_code = ? AND deleted_at IS NULL", code).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return billing.Invoice{}, ErrNotFound
	}
	return m.ToDomain(), err
}

// FindByQRPayload implements the QR-scan lookup (§4.2).
func (r *InvoiceRepository) FindByQRPayload(ctx context.Context, qr string) (billing.Invoice, error) {
	var m model.InvoiceModel
	err := r.db.WithContext(ctx).First(&m, "qr_payload = ? AND deleted_at IS NULL", qr).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return billing.Invoice{}, ErrNotFound
	}
	return m.ToDomain(), err
}

// FindBySubscriptionPeriod implements billing-run idempotency check.
func (r *InvoiceRepository) FindBySubscriptionPeriod(ctx context.Context, subscriptionID, period string) (billing.Invoice, error) {
	var m model.InvoiceModel
	err := r.db.WithContext(ctx).
		Where("subscription_id = ? AND period = ? AND deleted_at IS NULL", subscriptionID, period).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return billing.Invoice{}, ErrNotFound
	}
	return m.ToDomain(), err
}

// SaveWithItems persists the invoice aggregate (header + line items)
// atomically.
func (r *InvoiceRepository) SaveWithItems(ctx context.Context, inv billing.Invoice, items []billing.InvoiceItem) error {
	m := model.InvoiceModelFromDomain(inv)
	itemModels := make([]model.InvoiceItemModel, len(items))
	for i, it := range items {
		itemModels[i] = *model.InvoiceItemModelFromDomain(it)
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(m).Error; err != nil {
			return err
		}
		if len(itemModels) == 0 {
			return nil
		}
		return tx.Create(&itemModels).Error
	})
}

// AddGatewayFeeItem menambahkan biaya layanan gateway ke invoice secara
// idempoten (item AD_HOC dengan prefix deskripsi tetap). Bila item sudah ada,
// invoice dikembalikan apa adanya.
func (r *InvoiceRepository) AddGatewayFeeItem(ctx context.Context, invoiceID string, amount float64) (billing.Invoice, error) {
	var out billing.Invoice
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var inv model.InvoiceModel
		q := tx
		if r.db.Dialector.Name() == "postgres" {
			q = q.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := q.First(&inv, "id = ? AND deleted_at IS NULL", invoiceID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		var existing int64
		if err := tx.Model(&model.InvoiceItemModel{}).
			Where("invoice_id = ? AND item_type = ? AND description LIKE ?", invoiceID, billing.ItemTypeAdHoc, "Biaya layanan gateway%").
			Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			out = inv.ToDomain()
			return nil
		}
		item := &model.InvoiceItemModel{
			ID:          newID("itm"),
			InvoiceID:   inv.ID,
			Description: "Biaya layanan gateway",
			Quantity:    1,
			UnitPrice:   amount,
			Amount:      amount,
			ItemType:    billing.ItemTypeAdHoc,
			CreatedAt:   time.Now(),
		}
		if err := tx.Create(item).Error; err != nil {
			return err
		}
		updates := map[string]any{
			"subtotal": inv.Subtotal + amount,
			"total":    inv.Total + amount,
		}
		if err := tx.Model(&model.InvoiceModel{}).Where("id = ?", inv.ID).Updates(updates).Error; err != nil {
			return err
		}
		inv.Subtotal += amount
		inv.Total += amount
		out = inv.ToDomain()
		return nil
	})
	if err != nil {
		return billing.Invoice{}, err
	}
	return out, nil
}

// HasForSubscription reports whether any invoice points at the given
// subscription — delete-guard for manage_subscription.
func (r *InvoiceRepository) HasForSubscription(ctx context.Context, subID string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.InvoiceModel{}).
		Where("subscription_id = ? AND deleted_at IS NULL", subID).
		Count(&n).Error
	return n > 0, err
}

// Delete soft-deletes an invoice by id (F3-8) — baris item dipertahankan.
func (r *InvoiceRepository) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Model(&model.InvoiceModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", time.Now())
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteByCustomerID soft-deletes all invoices belonging to customerID (F3-8).
func (r *InvoiceRepository) DeleteByCustomerID(ctx context.Context, customerID string) error {
	return r.db.WithContext(ctx).Model(&model.InvoiceModel{}).
		Where("customer_id = ? AND deleted_at IS NULL", customerID).
		Update("deleted_at", time.Now()).Error
}

func (r *InvoiceRepository) applyFilter(ctx context.Context, f port.InvoiceFilter, now time.Time) *gorm.DB {
	tenantID := f.TenantID
	if tenantID == "" {
		tenantID = "tenant-default"
	}
	q := r.db.WithContext(ctx).Model(&model.InvoiceModel{}).
		Where("invoices.tenant_id = ? AND invoices.deleted_at IS NULL", tenantID)

	if f.CustomerID != "" {
		q = q.Where("invoices.customer_id = ?", f.CustomerID)
	}
	if f.SubscriptionID != "" {
		q = q.Where("invoices.subscription_id = ?", f.SubscriptionID)
	}
	if f.Status != "" {
		q = q.Where("invoices.status = ?", f.Status)
	}
	if f.Period != "" {
		q = q.Where("invoices.period = ?", f.Period)
	}
	if f.InvoiceType != "" {
		q = q.Where("invoices.invoice_type = ?", f.InvoiceType)
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if f.OverdueOnly {
		q = q.Where("invoices.status IN ('UNPAID','PARTIAL','OVERDUE') AND invoices.due_date < ?", today)
	} else if f.UnpaidOnly {
		if f.UnpaidDaysDisplay > 0 {
			cutoff := today.AddDate(0, 0, f.UnpaidDaysDisplay)
			q = q.Where("invoices.status IN ('UNPAID','PARTIAL','OVERDUE') AND invoices.due_date <= ?", cutoff)
		} else {
			q = q.Where("invoices.status IN ('UNPAID','PARTIAL','OVERDUE')")
		}
	}
	if f.DueFrom != nil {
		q = q.Where("invoices.due_date >= ?", *f.DueFrom)
	}
	if f.DueTo != nil {
		q = q.Where("invoices.due_date <= ?", *f.DueTo)
	}
	if f.CreatedFrom != nil {
		q = q.Where("invoices.created_at >= ?", *f.CreatedFrom)
	}
	if f.CreatedTo != nil {
		q = q.Where("invoices.created_at <= ?", *f.CreatedTo)
	}
	if search := strings.TrimSpace(f.Search); search != "" {
		like := "%" + search + "%"
		q = q.Joins("LEFT JOIN customers ON customers.id = invoices.customer_id").
			Where("invoices.invoice_number ILIKE ? OR invoices.manual_payment_code ILIKE ? OR customers.name ILIKE ?", like, like, like)
	}
	return q
}

// FindFiltered returns invoices matching filter criteria and total count.
func (r *InvoiceRepository) FindFiltered(ctx context.Context, f port.InvoiceFilter) ([]billing.Invoice, int, error) {
	now := time.Now().UTC()
	q := r.applyFilter(ctx, f, now)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if f.Limit > 0 {
		q = q.Limit(f.Limit).Offset(f.Offset)
	}
	var mList []model.InvoiceModel
	if err := q.Order("invoices.due_date ASC, invoices.created_at DESC").Find(&mList).Error; err != nil {
		return nil, 0, err
	}

	invoices := make([]billing.Invoice, len(mList))
	for i, m := range mList {
		invoices[i] = m.ToDomain()
	}
	return invoices, int(total), nil
}

// Summary returns summary aggregates for invoices matching filter criteria.
func (r *InvoiceRepository) Summary(ctx context.Context, f port.InvoiceFilter) (port.InvoiceSummary, error) {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	q := r.applyFilter(ctx, f, now)

	type summaryRow struct {
		Count            int
		TotalBilled      float64
		TotalPaid        float64
		TotalOutstanding float64
		TotalOverdue     float64
	}
	var row summaryRow
	err := q.Select(`
		COUNT(*) AS count,
		COALESCE(SUM(invoices.total), 0) AS total_billed,
		COALESCE(SUM(invoices.paid_amount), 0) AS total_paid,
		COALESCE(SUM(invoices.total - invoices.paid_amount), 0) AS total_outstanding,
		COALESCE(SUM(CASE WHEN invoices.status IN ('UNPAID','PARTIAL','OVERDUE') AND invoices.due_date < ? THEN (invoices.total - invoices.paid_amount) ELSE 0 END), 0) AS total_overdue
	`, today).Scan(&row).Error
	if err != nil {
		return port.InvoiceSummary{}, err
	}

	return port.InvoiceSummary{
		Count:            row.Count,
		TotalBilled:      row.TotalBilled,
		TotalPaid:        row.TotalPaid,
		TotalOutstanding: row.TotalOutstanding,
		TotalOverdue:     row.TotalOverdue,
	}, nil
}

// FindOverdueCandidates returns outstanding invoices where due_date < now.
func (r *InvoiceRepository) FindOverdueCandidates(ctx context.Context, tenantID string, now time.Time, limit int) ([]billing.Invoice, error) {
	if tenantID == "" {
		tenantID = "tenant-default"
	}
	if limit <= 0 {
		limit = 100
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	var mList []model.InvoiceModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL AND status IN ('UNPAID','PARTIAL') AND due_date < ?", tenantID, today).
		Order("due_date ASC").
		Limit(limit).
		Find(&mList).Error
	if err != nil {
		return nil, err
	}
	invoices := make([]billing.Invoice, len(mList))
	for i, m := range mList {
		invoices[i] = m.ToDomain()
	}
	return invoices, nil
}
