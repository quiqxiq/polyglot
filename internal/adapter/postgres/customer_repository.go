package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/port"
)

type customerRepository struct {
	db *gorm.DB
}

var _ port.CustomerRepository = (*customerRepository)(nil)

// NewCustomerRepository returns a port.CustomerRepository backed by GORM/Postgres.
func NewCustomerRepository(db *gorm.DB) port.CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) Save(ctx context.Context, c customer.Customer) error {
	m := models.CustomerModelFromDomain(c)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *customerRepository) FindByID(ctx context.Context, id string) (customer.Customer, error) {
	var m models.CustomerModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if err != nil {
		return customer.Customer{}, err
	}
	return m.ToDomain(), nil
}

func (r *customerRepository) FindAll(ctx context.Context) ([]customer.Customer, error) {
	var list []models.CustomerModel
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&list).Error
	if err != nil {
		return nil, err
	}
	res := make([]customer.Customer, len(list))
	for i, m := range list {
		res[i] = m.ToDomain()
	}
	return res, nil
}

func (r *customerRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.CustomerModel{}, "id = ?", id).Error
}

func (r *customerRepository) FindSubscriptions(ctx context.Context, customerID string) ([]customer.Subscription, error) {
	var list []models.SubscriptionModel
	err := r.db.WithContext(ctx).Find(&list, "customer_id = ?", customerID).Error
	if err != nil {
		return nil, err
	}
	res := make([]customer.Subscription, len(list))
	for i, m := range list {
		res[i] = m.ToCustomerSubscription()
	}
	return res, nil
}
