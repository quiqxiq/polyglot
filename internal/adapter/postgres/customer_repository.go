package postgres

import (
	"context"
	"errors"
	"time"

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
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) Save(ctx context.Context, c customer.Customer) error {
	m := model.CustomerModelFromDomain(c)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *CustomerRepository) FindByID(ctx context.Context, id string) (customer.Customer, error) {
	var m model.CustomerModel
	err := r.db.WithContext(ctx).First(&m, "id = ? AND deleted_at IS NULL", id).Error
	return m.ToDomain(), mapNotFound(err)
}

func (r *CustomerRepository) FindAll(ctx context.Context) ([]customer.Customer, error) {
	var mList []model.CustomerModel
	err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Find(&mList).Error
	if err != nil {
		return nil, err
	}
	customers := make([]customer.Customer, len(mList))
	for i, m := range mList {
		customers[i] = m.ToDomain()
	}
	return customers, nil
}

// Delete soft-deletes a customer (F3-8): baris dipertahankan untuk audit.
func (r *CustomerRepository) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Model(&model.CustomerModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", time.Now())
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CustomerRepository) FindSubscriptions(ctx context.Context, customerID string) ([]subscription.Subscription, error) {
	var mList []model.SubscriptionModel
	err := r.db.WithContext(ctx).Find(&mList, "customer_id = ? AND deleted_at IS NULL", customerID).Error
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
