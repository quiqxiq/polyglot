package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
	"github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
)

type InvoiceRepository struct {
	db *gorm.DB
}

var _ port.InvoiceRepository = (*InvoiceRepository)(nil)

// NewInvoiceRepository returns a port.InvoiceRepository backed by GORM/Postgres.
func NewInvoiceRepository(db *gorm.DB) *InvoiceRepository {
	_ = db.AutoMigrate(&models.InvoiceModel{}, &models.PlanModel{})
	return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) Save(ctx context.Context, inv billing.Invoice) error {
	m := models.InvoiceModelFromDomain(inv)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *InvoiceRepository) FindByID(ctx context.Context, id string) (billing.Invoice, error) {
	var m models.InvoiceModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return billing.Invoice{}, ErrNotFound
	}
	return m.ToDomain(), err
}

func (r *InvoiceRepository) FindByCustomerID(ctx context.Context, customerID string) ([]billing.Invoice, error) {
	var mList []models.InvoiceModel
	err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Order("created_at desc").Find(&mList).Error
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
	var mList []models.InvoiceModel
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&mList).Error
	if err != nil {
		return nil, err
	}
	invoices := make([]billing.Invoice, len(mList))
	for i, m := range mList {
		invoices[i] = m.ToDomain()
	}
	return invoices, nil
}

func (r *InvoiceRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	res := r.db.WithContext(ctx).Model(&models.InvoiceModel{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

