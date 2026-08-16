package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

// Ensure CustomerRepository satisfies port.CustomerRepository at compile time.
var _ port.CustomerRepository = (*CustomerRepository)(nil)

type CustomerRepository struct {
	db *gorm.DB
}

// NewCustomerRepository returns a port.CustomerRepository backed by GORM/Postgres.
func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	_ = db.AutoMigrate(&models.CustomerModel{})
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) Save(ctx context.Context, c customer.Customer) error {
	m := models.CustomerModelFromDomain(c)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *CustomerRepository) FindByID(ctx context.Context, id string) (customer.Customer, error) {
	var m models.CustomerModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	return m.ToDomain(), err
}

func (r *CustomerRepository) FindAll(ctx context.Context) ([]customer.Customer, error) {
	var mList []models.CustomerModel
	err := r.db.WithContext(ctx).Find(&mList).Error
	if err != nil {
		return nil, err
	}
	customers := make([]customer.Customer, len(mList))
	for i, m := range mList {
		customers[i] = m.ToDomain()
	}
	return customers, nil
}

func (r *CustomerRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.CustomerModel{}, "id = ?", id).Error
}

func (r *CustomerRepository) FindSubscriptions(ctx context.Context, customerID string) ([]subscription.Subscription, error) {
	var mList []models.SubscriptionModel
	err := r.db.WithContext(ctx).Find(&mList, "customer_id = ?", customerID).Error
	if err != nil {
		return nil, err
	}
	subs := make([]subscription.Subscription, len(mList))
	for i, m := range mList {
		subs[i] = m.ToDomain()
	}
	return subs, nil
}

