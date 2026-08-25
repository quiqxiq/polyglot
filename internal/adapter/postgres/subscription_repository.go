package postgres

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

var _ port.SubscriptionRepository = (*SubscriptionRepository)(nil)

// SubscriptionRepository persists the MAPPING-ONLY subscriptions table.
type SubscriptionRepository struct {
	db *gorm.DB
}

// NewSubscriptionRepository returns a port.SubscriptionRepository backed by GORM/Postgres.
func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Save(ctx context.Context, sub subscription.Subscription) error {
	m := model.SubscriptionModelFromDomain(sub)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *SubscriptionRepository) FindByID(ctx context.Context, id string) (subscription.Subscription, error) {
	var m model.SubscriptionModel
	err := r.db.WithContext(ctx).First(&m, "id = ? AND deleted_at IS NULL", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return subscription.Subscription{}, subscription.ErrNotFound
		}
		return subscription.Subscription{}, err
	}
	return m.ToDomain(), nil
}

func (r *SubscriptionRepository) FindByCustomerID(ctx context.Context, customerID string) ([]subscription.Subscription, error) {
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

func (r *SubscriptionRepository) FindAll(ctx context.Context) ([]subscription.Subscription, error) {
	var mList []model.SubscriptionModel
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
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

func (r *SubscriptionRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	res := r.db.WithContext(ctx).Model(&model.SubscriptionModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return subscription.ErrNotFound
	}
	return nil
}

// UpdateMapping persists provisioning results after the device-side account
// exists. Empty values are written verbatim so callers can clear mappings.
func (r *SubscriptionRepository) UpdateMapping(ctx context.Context, id, remoteUsername, remoteID, deviceID string) error {
	updates := map[string]any{}
	var deviceIDPtr *string
	if deviceID != "" {
		d := deviceID
		deviceIDPtr = &d
	}
	updates["device_id"] = deviceIDPtr
	updates["remote_username"] = remoteUsername
	updates["remote_id"] = remoteID

	res := r.db.WithContext(ctx).Model(&model.SubscriptionModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return subscription.ErrNotFound
	}
	return nil
}

func (r *SubscriptionRepository) FindByDeviceAndUsername(ctx context.Context, deviceID, username string) (subscription.Subscription, error) {
	var m model.SubscriptionModel
	err := r.db.WithContext(ctx).
		Where("device_id = ? AND remote_username = ? AND deleted_at IS NULL", deviceID, username).
		First(&m).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return subscription.Subscription{}, subscription.ErrNotFound
		}
		return subscription.Subscription{}, err
	}
	return m.ToDomain(), nil
}

// SetIsolation records isolation lifecycle fields atomically with the
// status change: isolated_at/isolation_reason when isolating, NULLs when
// returning to service.
func (r *SubscriptionRepository) SetIsolation(ctx context.Context, id string, status string, isolatedAt *time.Time, reason string) error {
	updates := map[string]any{
		"status":           status,
		"isolated_at":      isolatedAt,
		"isolation_reason": reason,
	}
	res := r.db.WithContext(ctx).Model(&model.SubscriptionModel{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return subscription.ErrNotFound
	}
	return nil
}
