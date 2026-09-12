package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
)

type InvoiceRepository struct {
	db *gorm.DB
}

var _ port.InvoiceRepository = (*InvoiceRepository)(nil)

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
