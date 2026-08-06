package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/port"
)

type customerRepository struct {
	db *gorm.DB
}

// NewCustomerRepository returns a port.CustomerRepository backed by GORM/Postgres.
func NewCustomerRepository(db *gorm.DB) port.CustomerRepository {
	_ = db.AutoMigrate(&customer.Customer{}, &customer.Subscription{})
	return &customerRepository{db: db}
}

func (r *customerRepository) Save(ctx context.Context, c customer.Customer) error {
	return r.db.WithContext(ctx).Save(&c).Error
}

func (r *customerRepository) FindByID(ctx context.Context, id string) (customer.Customer, error) {
	var c customer.Customer
	err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error
	return c, err
}

func (r *customerRepository) FindAll(ctx context.Context) ([]customer.Customer, error) {
	var customers []customer.Customer
	err := r.db.WithContext(ctx).Find(&customers).Error
	return customers, err
}

func (r *customerRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&customer.Customer{}, "id = ?", id).Error
}

func (r *customerRepository) FindSubscriptions(ctx context.Context, customerID string) ([]customer.Subscription, error) {
	var subs []customer.Subscription
	err := r.db.WithContext(ctx).Find(&subs, "customer_id = ?", customerID).Error
	return subs, err
}
