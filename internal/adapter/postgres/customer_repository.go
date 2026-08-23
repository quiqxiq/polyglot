package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
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
	_ = db.AutoMigrate(&model.CustomerModel{})
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) Save(ctx context.Context, c customer.Customer) error {
	m := model.CustomerModelFromDomain(c)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *CustomerRepository) FindByID(ctx context.Context, id string) (customer.Customer, error) {
	var m model.CustomerModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	return m.ToDomain(), err
}

func (r *CustomerRepository) FindAll(ctx context.Context) ([]customer.Customer, error) {
	var mList []model.CustomerModel
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
	return r.db.WithContext(ctx).Delete(&model.CustomerModel{}, "id = ?", id).Error
}

func (r *CustomerRepository) FindSubscriptions(ctx context.Context, customerID string) ([]subscription.Subscription, error) {
	var mList []model.SubscriptionModel
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

// FindByPortalAccessCode implements portal/quick-pay lookup (§4.2).
func (r *CustomerRepository) FindByPortalAccessCode(ctx context.Context, code string) (customer.Customer, error) {
	var m model.CustomerModel
	err := r.db.WithContext(ctx).First(&m, "portal_access_code = ? AND deleted_at IS NULL", code).Error
	if err != nil {
		return customer.Customer{}, mapNotFound(err)
	}
	return m.ToDomain(), nil
}

func (r *CustomerRepository) FindByPhone(ctx context.Context, phone string) (customer.Customer, error) {
	var m model.CustomerModel
	err := r.db.WithContext(ctx).First(&m, "phone = ? AND deleted_at IS NULL", phone).Error
	if err != nil {
		return customer.Customer{}, mapNotFound(err)
	}
	return m.ToDomain(), nil
}

func (r *CustomerRepository) FindByCustomerCode(ctx context.Context, code string) (customer.Customer, error) {
	var m model.CustomerModel
	err := r.db.WithContext(ctx).First(&m, "customer_code = ? AND deleted_at IS NULL", code).Error
	if err != nil {
		return customer.Customer{}, mapNotFound(err)
	}
	return m.ToDomain(), nil
}

// mapNotFound translates gorm not-found into the shared ErrNotFound.
func mapNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}
