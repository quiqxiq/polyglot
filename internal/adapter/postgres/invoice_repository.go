package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

// invoiceModel maps the `invoices` table per migration 000014.
type invoiceModel struct {
	ID            string    `gorm:"column:id;primaryKey"`
	InvoiceNumber string    `gorm:"column:invoice_number;not null;unique"`
	CustomerID    string    `gorm:"column:customer_id;not null"`
	BillingRunID  *string   `gorm:"column:billing_run_id"`
	PeriodStart   time.Time `gorm:"column:period_start;not null"`
	PeriodEnd     time.Time `gorm:"column:period_end;not null"`
	IssueDate     time.Time `gorm:"column:issue_date;not null"`
	DueDate       time.Time `gorm:"column:due_date;not null"`
	Status        string    `gorm:"column:status;not null;default:'draft'"`
	Subtotal      float64   `gorm:"column:subtotal;not null;default:0"`
	TaxAmount     float64   `gorm:"column:tax_amount;not null;default:0"`
	TotalAmount   float64   `gorm:"column:total_amount;not null;default:0"`
	Notes         string    `gorm:"column:notes"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;autoCreateTime"`
}

// invoiceItemModel maps the `invoice_items` table per migration 000014.
type invoiceItemModel struct {
	ID             string  `gorm:"column:id;primaryKey"`
	InvoiceID      string  `gorm:"column:invoice_id;not null"`
	SubscriptionID *string `gorm:"column:subscription_id"`
	ItemType       string  `gorm:"column:item_type;not null"`
	Description    string  `gorm:"column:description;not null"`
	Quantity       float64 `gorm:"column:quantity;not null;default:1"`
	UnitPrice      float64 `gorm:"column:unit_price;not null"`
	Amount         float64 `gorm:"column:amount;not null"`
}

func (invoiceModel) TableName() string     { return "invoices" }
func (invoiceItemModel) TableName() string { return "invoice_items" }

func (m invoiceModel) toDomain(items []invoiceItemModel) billing.Invoice {
	invItems := make([]billing.InvoiceItem, len(items))
	for i, it := range items {
		invItems[i] = it.toDomain()
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
		Items:         invItems,
	}
}

func fromInvoiceDomain(inv billing.Invoice) invoiceModel {
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
		CreatedAt:     inv.CreatedAt,
	}
}

func (m invoiceItemModel) toDomain() billing.InvoiceItem {
	return billing.InvoiceItem{
		ID:             m.ID,
		InvoiceID:      m.InvoiceID,
		SubscriptionID: m.SubscriptionID,
		ItemType:       m.ItemType,
		Description:    m.Description,
		Quantity:       m.Quantity,
		UnitPrice:      m.UnitPrice,
		Amount:         m.Amount,
	}
}

func fromInvoiceItemDomain(item billing.InvoiceItem) invoiceItemModel {
	return invoiceItemModel{
		ID:             item.ID,
		InvoiceID:      item.InvoiceID,
		SubscriptionID: item.SubscriptionID,
		ItemType:       item.ItemType,
		Description:    item.Description,
		Quantity:       item.Quantity,
		UnitPrice:      item.UnitPrice,
		Amount:         item.Amount,
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

// FindByID returns the invoice (with items) for the given id, or billing.ErrNotFound.
func (r *InvoiceRepository) FindByID(ctx context.Context, id string) (billing.Invoice, error) {
	var m invoiceModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return billing.Invoice{}, fmt.Errorf("invoice %s: %w", id, billing.ErrNotFound)
		}
		return billing.Invoice{}, fmt.Errorf("invoice %s: %w", id, err)
	}

	var items []invoiceItemModel
	if err := r.db.WithContext(ctx).Where("invoice_id = ?", id).Find(&items).Error; err != nil {
		return billing.Invoice{}, fmt.Errorf("invoice %s: load items: %w", id, err)
	}
	return m.toDomain(items), nil
}

// FindAll returns all invoices ordered by created_at desc.
func (r *InvoiceRepository) FindAll(ctx context.Context) ([]billing.Invoice, error) {
	var models []invoiceModel
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list invoices: %w", err)
	}
	invoices := make([]billing.Invoice, len(models))
	for i, m := range models {
		invoices[i] = m.toDomain(nil)
	}
	return invoices, nil
}

// FindByCustomer returns invoices for a specific customer.
func (r *InvoiceRepository) FindByCustomer(ctx context.Context, customerID string) ([]billing.Invoice, error) {
	var models []invoiceModel
	if err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Order("created_at desc").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list invoices by customer: %w", err)
	}
	invoices := make([]billing.Invoice, len(models))
	for i, m := range models {
		invoices[i] = m.toDomain(nil)
	}
	return invoices, nil
}

// Create inserts a new invoice with its items.
func (r *InvoiceRepository) Create(ctx context.Context, inv billing.Invoice) (billing.Invoice, error) {
	var created invoiceModel
	var items []invoiceItemModel

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m := fromInvoiceDomain(inv)
		if err := tx.Create(&m).Error; err != nil {
			return err
		}
		created = m

		modelItems := make([]invoiceItemModel, len(inv.Items))
		for i, item := range inv.Items {
			modelItems[i] = fromInvoiceItemDomain(item)
			modelItems[i].InvoiceID = m.ID
		}
		if len(modelItems) > 0 {
			if err := tx.Create(&modelItems).Error; err != nil {
				return err
			}
		}
		items = modelItems
		return nil
	})
	if err != nil {
		return billing.Invoice{}, fmt.Errorf("create invoice: %w", err)
	}
	return created.toDomain(items), nil
}

// Update modifies an existing invoice.
func (r *InvoiceRepository) Update(ctx context.Context, inv billing.Invoice) (billing.Invoice, error) {
	var updated invoiceModel
	var items []invoiceItemModel

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m := fromInvoiceDomain(inv)
		if err := tx.Save(&m).Error; err != nil {
			return err
		}
		updated = m

		if err := tx.Where("invoice_id = ?", m.ID).Delete(&invoiceItemModel{}).Error; err != nil {
			return err
		}
		modelItems := make([]invoiceItemModel, len(inv.Items))
		for i, item := range inv.Items {
			modelItems[i] = fromInvoiceItemDomain(item)
			modelItems[i].InvoiceID = m.ID
		}
		if len(modelItems) > 0 {
			if err := tx.Create(&modelItems).Error; err != nil {
				return err
			}
		}
		items = modelItems
		return nil
	})
	if err != nil {
		return billing.Invoice{}, fmt.Errorf("update invoice: %w", err)
	}
	return updated.toDomain(items), nil
}

// Delete removes an invoice (and its items via CASCADE) by id.
func (r *InvoiceRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&invoiceModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete invoice: %w", err)
	}
	return nil
}
