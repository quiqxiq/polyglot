package postgres

import (
	"context"
	"errors"
	"fmt"
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
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return customer.Customer{}, customer.ErrNotFound
	}
	if err != nil {
		return customer.Customer{}, err
	}
	return m.ToDomain(), nil
}

func (r *CustomerRepository) FindAll(ctx context.Context) ([]customer.Customer, error) {
	var mList []model.CustomerModel
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("created_at desc").
		Find(&mList).Error
	if err != nil {
		return nil, err
	}
	customers := make([]customer.Customer, len(mList))
	for i, m := range mList {
		customers[i] = m.ToDomain()
	}
	return customers, nil
}

// Delete hard-deletes the row. Prefer SoftDelete for production flows so
// billing history stays referentially intact.
func (r *CustomerRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.CustomerModel{}, "id = ?", id).Error
}

func (r *CustomerRepository) SoftDelete(ctx context.Context, id string, at time.Time) error {
	res := r.db.WithContext(ctx).Model(&model.CustomerModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", at)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return customer.ErrNotFound
	}
	return nil
}

func (r *CustomerRepository) FindByPhone(ctx context.Context, phone string) (customer.Customer, error) {
	var m model.CustomerModel
	err := r.db.WithContext(ctx).
		Where("phone = ? AND deleted_at IS NULL", phone).
		Order("created_at desc").
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return customer.Customer{}, customer.ErrNotFound
	}
	if err != nil {
		return customer.Customer{}, err
	}
	return m.ToDomain(), nil
}

// NextCustomerCode generates the next sequential customer code ("00001",
// "00002", ...). Uniqueness is enforced by the partial unique index; on a
// race the caller retries with the next Save error.
func (r *CustomerRepository) NextCustomerCode(ctx context.Context) (string, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.CustomerModel{}).
		Where("deleted_at IS NULL").
		Count(&count).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("%05d", count+1), nil
}

func (r *CustomerRepository) FindSubscriptions(ctx context.Context, customerID string) ([]subscription.Subscription, error) {
	var mList []model.SubscriptionModel
	err := r.db.WithContext(ctx).
		Where("customer_id = ? AND deleted_at IS NULL", customerID).
		Order("created_at desc").
		Find(&mList).Error
	if err != nil {
		return nil, err
	}
	subs := make([]subscription.Subscription, len(mList))
	for i, m := range mList {
		subs[i] = m.ToDomain()
	}
	return subs, nil
}
