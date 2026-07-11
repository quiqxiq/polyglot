package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

// invoiceModel maps the `invoices` table to a GORM-friendly struct per
// migration 000014.
type invoiceModel struct {
	ID            string              `gorm:"column:id;primaryKey;type:uuid"`
	InvoiceNumber string              `gorm:"column:invoice_number;uniqueIndex;not null"`
	CustomerID    string              `gorm:"column:customer_id;not null;type:uuid"`
	BillingRunID  *string             `gorm:"column:billing_run_id;type:uuid"`
	PeriodStart   time.Time           `gorm:"column:period_start;not null"`
	PeriodEnd     time.Time           `gorm:"column:period_end;not null"`
	IssueDate     time.Time           `gorm:"column:issue_date;not null"`
	DueDate       time.Time           `gorm:"column:due_date;not null"`
	Status        string              `gorm:"column:status;not null;default:draft"`
	Subtotal      float64             `gorm:"column:subtotal;not null;default:0"`
	TaxAmount     float64             `gorm:"column:tax_amount;not null;default:0"`
	TotalAmount   float64             `gorm:"column:total_amount;not null;default:0"`
	Notes         string              `gorm:"column:notes"`
	CreatedAt     time.Time           `gorm:"column:created_at;not null;autoCreateTime"`
	Items         []invoiceItemModel  `gorm:"foreignKey:InvoiceID;references:ID"`
}

// TableName returns the explicit table name for the invoice model.
func (invoiceModel) TableName() string {
	return "invoices"
}

// invoiceItemModel maps the `invoice_items` table per migration 000014.
type invoiceItemModel struct {
	ID             string  `gorm:"column:id;primaryKey;type:uuid"`
	InvoiceID      string  `gorm:"column:invoice_id;not null;type:uuid"`
	SubscriptionID *string `gorm:"column:subscription_id;type:uuid"`
	ItemType       string  `gorm:"column:item_type;not null"`
	Description    string  `gorm:"column:description;not null"`
	Quantity       float64 `gorm:"column:quantity;not null;default:1"`
	UnitPrice      float64 `gorm:"column:unit_price;not null"`
	Amount         float64 `gorm:"column:amount;not null"`
}

// TableName returns the explicit table name for the invoice item model.
func (invoiceItemModel) TableName() string {
	return "invoice_items"
}

// toDomain maps an invoiceModel (with items) to the domain billing.Invoice.
func (m invoiceModel) toDomain() billing.Invoice {
	items := make([]billing.InvoiceItem, len(m.Items))
	for i, it := range m.Items {
		items[i] = billing.InvoiceItem{
			ID:             it.ID,
			InvoiceID:      it.InvoiceID,
			SubscriptionID: it.SubscriptionID,
			ItemType:       it.ItemType,
			Description:    it.Description,
			Quantity:       it.Quantity,
			UnitPrice:      it.UnitPrice,
			Amount:         it.Amount,
		}
	}

	return billing.Invoice{
		ID:            m.ID,
		InvoiceNumber: m.InvoiceNumber,
		CustomerID:    m.CustomerID,
		BillingRunID:  m.BillingRunID,
		PeriodStart:   m.PeriodStart,
		PeriodEnd:     m.PeriodEnd,
		IssueDate:     m.IssueDate,
		DueDate:       m.DueDate,
		Status:        m.Status,
		Subtotal:      m.Subtotal,
		TaxAmount:     m.TaxAmount,
		TotalAmount:   m.TotalAmount,
		Notes:         m.Notes,
		CreatedAt:     m.CreatedAt,
		Items:         items,
	}
}

// invoiceFromDomain maps a domain billing.Invoice to an invoiceModel.
func invoiceFromDomain(inv billing.Invoice) invoiceModel {
	items := make([]invoiceItemModel, len(inv.Items))
	for i, it := range inv.Items {
		items[i] = invoiceItemModel{
			ID:             it.ID,
			InvoiceID:      it.InvoiceID,
			SubscriptionID: it.SubscriptionID,
			ItemType:       it.ItemType,
			Description:    it.Description,
			Quantity:       it.Quantity,
			UnitPrice:      it.UnitPrice,
			Amount:         it.Amount,
		}
	}

	return invoiceModel{
		ID:            inv.ID,
		InvoiceNumber: inv.InvoiceNumber,
		CustomerID:    inv.CustomerID,
		BillingRunID:  inv.BillingRunID,
		PeriodStart:   inv.PeriodStart,
		PeriodEnd:     inv.PeriodEnd,
		IssueDate:     inv.IssueDate,
		DueDate:       inv.DueDate,
		Status:        inv.Status,
		Subtotal:      inv.Subtotal,
		TaxAmount:     inv.TaxAmount,
		TotalAmount:   inv.TotalAmount,
		Notes:         inv.Notes,
		Items:         items,
	}
}

// InvoiceRepository implements port.InvoiceRepository backed by PostgreSQL.
type InvoiceRepository struct {
	db *gorm.DB
}

// NewInvoiceRepository returns a port.InvoiceRepository backed by GORM/Postgres.
func NewInvoiceRepository(db *gorm.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

// FindByID returns the invoice (with items) for the given id, or
// billing.ErrNotFound.
func (r *InvoiceRepository) FindByID(ctx context.Context, id string) (billing.Invoice, error) {
	var m invoiceModel
	if err := r.db.WithContext(ctx).Preload("Items").First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return billing.Invoice{}, fmt.Errorf("invoice %s: %w", id, billing.ErrNotFound)
		}
		return billing.Invoice{}, fmt.Errorf("invoice %s: %w", id, err)
	}
	return m.toDomain(), nil
}

// FindAll returns all invoices (with items) ordered by created_at desc.
func (r *InvoiceRepository) FindAll(ctx context.Context) ([]billing.Invoice, error) {
	var models []invoiceModel
	if err := r.db.WithContext(ctx).Preload("Items").Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list invoices: %w", err)
	}

	invoices := make([]billing.Invoice, len(models))
	for i, m := range models {
		invoices[i] = m.toDomain()
	}
	return invoices, nil
}

// FindByCustomer returns invoices for a specific customer.
func (r *InvoiceRepository) FindByCustomer(ctx context.Context, customerID string) ([]billing.Invoice, error) {
	var models []invoiceModel
	if err := r.db.WithContext(ctx).Preload("Items").Where("customer_id = ?", customerID).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("find invoices by customer %s: %w", customerID, err)
	}

	invoices := make([]billing.Invoice, len(models))
	for i, m := range models {
		invoices[i] = m.toDomain()
	}
	return invoices, nil
}

// Create inserts a new invoice with its items.
func (r *InvoiceRepository) Create(ctx context.Context, inv billing.Invoice) (billing.Invoice, error) {
	m := invoiceFromDomain(inv)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return billing.Invoice{}, fmt.Errorf("create invoice: %w", err)
	}
	return m.toDomain(), nil
}

// Update modifies an existing invoice.
func (r *InvoiceRepository) Update(ctx context.Context, inv billing.Invoice) (billing.Invoice, error) {
	m := invoiceFromDomain(inv)
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return billing.Invoice{}, fmt.Errorf("update invoice: %w", err)
	}
	return m.toDomain(), nil
}

// Delete removes an invoice (and its items via CASCADE) by id.
func (r *InvoiceRepository) Delete(ctx context.Context, id string) error {
	// Delete items first (SQLite doesn't support CASCADE), then the invoice.
	if err := r.db.WithContext(ctx).Where("invoice_id = ?", id).Delete(&invoiceItemModel{}).Error; err != nil {
		return fmt.Errorf("delete invoice items: %w", err)
	}
	if err := r.db.WithContext(ctx).Delete(&invoiceModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete invoice: %w", err)
	}
	return nil
}
