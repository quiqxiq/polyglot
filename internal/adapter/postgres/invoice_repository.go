package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
)

type InvoiceRepository struct {
	db *gorm.DB
}

var _ port.InvoiceRepository = (*InvoiceRepository)(nil)

// NewInvoiceRepository returns a port.InvoiceRepository backed by GORM/Postgres.
func NewInvoiceRepository(db *gorm.DB) *InvoiceRepository {
	_ = db.AutoMigrate(&billing.Invoice{}, &billing.Plan{})
	return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) Save(ctx context.Context, inv billing.Invoice) error {
	return r.db.WithContext(ctx).Save(&inv).Error
}

func (r *InvoiceRepository) FindByID(ctx context.Context, id string) (billing.Invoice, error) {
	var inv billing.Invoice
	err := r.db.WithContext(ctx).First(&inv, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return billing.Invoice{}, ErrNotFound
	}
	return inv, err
}

func (r *InvoiceRepository) FindByCustomerID(ctx context.Context, customerID string) ([]billing.Invoice, error) {
	var invoices []billing.Invoice
	err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Order("created_at desc").Find(&invoices).Error
	return invoices, err
}

func (r *InvoiceRepository) FindAll(ctx context.Context) ([]billing.Invoice, error) {
	var invoices []billing.Invoice
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&invoices).Error
	return invoices, err
}

func (r *InvoiceRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	res := r.db.WithContext(ctx).Model(&billing.Invoice{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
