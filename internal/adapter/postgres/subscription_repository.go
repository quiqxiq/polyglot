package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

var _ port.SubscriptionRepository = (*SubscriptionRepository)(nil)

// NewSubscriptionRepository returns a port.SubscriptionRepository backed by GORM/Postgres.
func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	_ = db.AutoMigrate(&subscription.Subscription{})
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Save(ctx context.Context, sub subscription.Subscription) error {
	return r.db.WithContext(ctx).Save(&sub).Error
}

func (r *SubscriptionRepository) FindByID(ctx context.Context, id string) (subscription.Subscription, error) {
	var sub subscription.Subscription
	err := r.db.WithContext(ctx).First(&sub, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return subscription.Subscription{}, ErrNotFound
	}
	return sub, err
}

func (r *SubscriptionRepository) FindByCustomerID(ctx context.Context, customerID string) ([]subscription.Subscription, error) {
	var subs []subscription.Subscription
	err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Order("created_at desc").Find(&subs).Error
	return subs, err
}

func (r *SubscriptionRepository) FindAll(ctx context.Context) ([]subscription.Subscription, error) {
	var subs []subscription.Subscription
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&subs).Error
	return subs, err
}

func (r *SubscriptionRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	res := r.db.WithContext(ctx).Model(&subscription.Subscription{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
