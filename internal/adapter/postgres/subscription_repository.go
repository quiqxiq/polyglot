package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
	"github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

var _ port.SubscriptionRepository = (*SubscriptionRepository)(nil)

// NewSubscriptionRepository returns a port.SubscriptionRepository backed by GORM/Postgres.
func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	_ = db.AutoMigrate(&models.SubscriptionModel{})
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Save(ctx context.Context, sub subscription.Subscription) error {
	m := models.SubscriptionModelFromDomain(sub)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *SubscriptionRepository) FindByID(ctx context.Context, id string) (subscription.Subscription, error) {
	var m models.SubscriptionModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return subscription.Subscription{}, ErrNotFound
	}
	return m.ToDomain(), err
}

func (r *SubscriptionRepository) FindByCustomerID(ctx context.Context, customerID string) ([]subscription.Subscription, error) {
	var mList []models.SubscriptionModel
	err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Order("created_at desc").Find(&mList).Error
	if err != nil {
		return nil, err
	}
	subs := make([]subscription.Subscription, len(mList))
	for i, m := range mList {
		subs[i] = m.ToDomain()
	}
	return subs, nil
}

func (r *SubscriptionRepository) FindAll(ctx context.Context) ([]subscription.Subscription, error) {
	var mList []models.SubscriptionModel
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&mList).Error
	if err != nil {
		return nil, err
	}
	subs := make([]subscription.Subscription, len(mList))
	for i, m := range mList {
		subs[i] = m.ToDomain()
	}
	return subs, nil
}

func (r *SubscriptionRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	res := r.db.WithContext(ctx).Model(&models.SubscriptionModel{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

