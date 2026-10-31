package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
)

// GatewayTransactionRepository persists payment gateway transactions in PostgreSQL.
type GatewayTransactionRepository struct {
	db *gorm.DB
}

var _ port.GatewayTransactionRepository = (*GatewayTransactionRepository)(nil)
var _ port.GatewayTransactionFilteredRepository = (*GatewayTransactionRepository)(nil)

// NewGatewayTransactionRepository constructs a gateway transaction repository.
func NewGatewayTransactionRepository(db *gorm.DB) *GatewayTransactionRepository {
	return &GatewayTransactionRepository{db: db}
}

// Save creates or updates a gateway transaction.
func (r *GatewayTransactionRepository) Save(ctx context.Context, t domainBilling.GatewayTransaction) error {
	return r.db.WithContext(ctx).Save(model.GatewayTransactionModelFromDomain(t)).Error
}

// FindByExternalID returns a transaction by provider and external identifier.
func (r *GatewayTransactionRepository) FindByExternalID(ctx context.Context, gateway, externalID string) (domainBilling.GatewayTransaction, error) {
	var m model.GatewayTransactionModel
	err := r.db.WithContext(ctx).
		Where("gateway = ? AND external_id = ?", gateway, externalID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domainBilling.GatewayTransaction{}, ErrNotFound
	}
	if err != nil {
		return domainBilling.GatewayTransaction{}, err
	}
	return m.ToDomain(), nil
}

// FindByInvoice returns all transactions associated with an invoice.
func (r *GatewayTransactionRepository) FindByInvoice(ctx context.Context, invoiceID string) ([]domainBilling.GatewayTransaction, error) {
	var mList []model.GatewayTransactionModel
	err := r.db.WithContext(ctx).Where("invoice_id = ?", invoiceID).Find(&mList).Error
	if err != nil {
		return nil, err
	}
	out := make([]domainBilling.GatewayTransaction, len(mList))
	for i := range mList {
		out[i] = mList[i].ToDomain()
	}
	return out, nil
}

func (r *GatewayTransactionRepository) UpdateStatus(ctx context.Context, id, status string) error {
	res := r.db.WithContext(ctx).Model(&model.GatewayTransactionModel{}).
		Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GatewayTransactionRepository) LinkPayment(ctx context.Context, id, paymentID string, feeAmount float64) error {
	return r.db.WithContext(ctx).Model(&model.GatewayTransactionModel{}).
		Where("id = ?", id).
		Updates(map[string]any{"payment_id": paymentID, "fee_amount": feeAmount}).Error
}

// ListFiltered returns gateway transactions matching filter criteria.
func (r *GatewayTransactionRepository) ListFiltered(ctx context.Context, f port.GatewayTransactionFilter) ([]domainBilling.GatewayTransaction, int, error) {
	tenantID := f.TenantID
	if tenantID == "" {
		tenantID = "tenant-default"
	}
	q := r.db.WithContext(ctx).Model(&model.GatewayTransactionModel{}).
		Where("tenant_id = ?", tenantID)

	if f.Gateway != "" {
		q = q.Where("gateway = ?", f.Gateway)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.InvoiceID != "" {
		q = q.Where("invoice_id = ?", f.InvoiceID)
	}
	if f.From != nil {
		q = q.Where("created_at >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where("created_at <= ?", *f.To)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if f.Limit > 0 {
		q = q.Limit(f.Limit).Offset(f.Offset)
	}
	var list []model.GatewayTransactionModel
	if err := q.Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	out := make([]domainBilling.GatewayTransaction, len(list))
	for i, m := range list {
		out[i] = m.ToDomain()
	}
	return out, int(total), nil
}

// ListPendingBefore returns transactions stuck in PENDING status before a given timestamp.
func (r *GatewayTransactionRepository) ListPendingBefore(ctx context.Context, tenantID string, before time.Time, limit int) ([]domainBilling.GatewayTransaction, error) {
	if tenantID == "" {
		tenantID = "tenant-default"
	}
	if limit <= 0 {
		limit = 100
	}
	var list []model.GatewayTransactionModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND status = ? AND updated_at < ?", tenantID, domainBilling.GatewayStatusPending, before).
		Order("updated_at ASC").
		Limit(limit).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	out := make([]domainBilling.GatewayTransaction, len(list))
	for i, m := range list {
		out[i] = m.ToDomain()
	}
	return out, nil
}
