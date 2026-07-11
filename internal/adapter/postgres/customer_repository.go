package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/domain/customer"
)

// customerModel maps the `customers` table per migration 000008.
type customerModel struct {
	ID             string     `gorm:"column:id;primaryKey"`
	FullName       string     `gorm:"column:full_name;not null"`
	IDNumber       string     `gorm:"column:id_number"`
	Phone          string     `gorm:"column:phone;not null"`
	WhatsApp       string     `gorm:"column:whatsapp"`
	Email          string     `gorm:"column:email"`
	Address        string     `gorm:"column:address;not null"`
	LocationLat    *float64   `gorm:"column:location_lat"`
	LocationLng    *float64   `gorm:"column:location_lng"`
	CustomerType   string     `gorm:"column:customer_type;not null;default:'residential'"`
	Status         string     `gorm:"column:status;not null;default:'prospect'"`
	ReferralSource string     `gorm:"column:referral_source"`
	Notes          string     `gorm:"column:notes"`
	RegisteredAt   time.Time  `gorm:"column:registered_at;not null;autoCreateTime"`
	TerminatedAt   *time.Time `gorm:"column:terminated_at"`
}

func (customerModel) TableName() string {
	return "customers"
}

func (m customerModel) toDomain() customer.Customer {
	return customer.Customer{
		ID:             m.ID,
		FullName:       m.FullName,
		IDNumber:       m.IDNumber,
		Phone:          m.Phone,
		WhatsApp:       m.WhatsApp,
		Email:          m.Email,
		Address:        m.Address,
		LocationLat:    m.LocationLat,
		LocationLng:    m.LocationLng,
		CustomerType:   m.CustomerType,
		Status:         m.Status,
		ReferralSource: m.ReferralSource,
		Notes:          m.Notes,
		RegisteredAt:   m.RegisteredAt,
		TerminatedAt:   m.TerminatedAt,
	}
}

func fromCustomerDomain(c customer.Customer) customerModel {
	return customerModel{
		ID:             c.ID,
		FullName:       c.FullName,
		IDNumber:       c.IDNumber,
		Phone:          c.Phone,
		WhatsApp:       c.WhatsApp,
		Email:          c.Email,
		Address:        c.Address,
		LocationLat:    c.LocationLat,
		LocationLng:    c.LocationLng,
		CustomerType:   c.CustomerType,
		Status:         c.Status,
		ReferralSource: c.ReferralSource,
		Notes:          c.Notes,
		RegisteredAt:   c.RegisteredAt,
		TerminatedAt:   c.TerminatedAt,
	}
}

// CustomerRepository implements port.CustomerRepository backed by PostgreSQL.
type CustomerRepository struct {
	db *gorm.DB
}

// NewCustomerRepository returns a port.CustomerRepository backed by GORM/Postgres.
func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

// FindByID returns the customer for the given id, or customer.ErrNotFound.
func (r *CustomerRepository) FindByID(ctx context.Context, id string) (customer.Customer, error) {
	var m customerModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customer.Customer{}, fmt.Errorf("customer %s: %w", id, customer.ErrNotFound)
		}
		return customer.Customer{}, fmt.Errorf("customer %s: %w", id, err)
	}
	return m.toDomain(), nil
}

// FindAll returns all customers ordered by full_name.
func (r *CustomerRepository) FindAll(ctx context.Context) ([]customer.Customer, error) {
	var models []customerModel
	if err := r.db.WithContext(ctx).Order("full_name").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list customers: %w", err)
	}
	customers := make([]customer.Customer, len(models))
	for i, m := range models {
		customers[i] = m.toDomain()
	}
	return customers, nil
}

// Create inserts a new customer.
func (r *CustomerRepository) Create(ctx context.Context, c customer.Customer) (customer.Customer, error) {
	m := fromCustomerDomain(c)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return customer.Customer{}, fmt.Errorf("create customer: %w", err)
	}
	return m.toDomain(), nil
}

// Update modifies an existing customer.
func (r *CustomerRepository) Update(ctx context.Context, c customer.Customer) (customer.Customer, error) {
	m := fromCustomerDomain(c)
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return customer.Customer{}, fmt.Errorf("update customer: %w", err)
	}
	return m.toDomain(), nil
}

// Delete removes a customer by id.
func (r *CustomerRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&customerModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete customer: %w", err)
	}
	return nil
}
