package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

type SubscriptionRepository struct {
	db    *gorm.DB
	vault port.CredentialVault
}

var _ port.SubscriptionRepository = (*SubscriptionRepository)(nil)

// NewSubscriptionRepository returns a port.SubscriptionRepository backed by
// GORM/Postgres. Vault dipakai untuk menyandikan RemotePassword:
// plaintext hanya hidup di memori, kolom remote_password_cipher berisi
// ciphertext AES-GCM (prinsip kredensial migrasi 000001).
func NewSubscriptionRepository(db *gorm.DB, vault port.CredentialVault) *SubscriptionRepository {
	return &SubscriptionRepository{db: db, vault: vault}
}

func (r *SubscriptionRepository) Save(ctx context.Context, sub subscription.Subscription) error {
	m := model.SubscriptionModelFromDomain(sub)
	if r.vault != nil && sub.RemotePassword != "" {
		cipher, err := r.vault.EncryptString(ctx, sub.RemotePassword)
		if err != nil {
			return err
		}
		m.RemotePassword = cipher
	}
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *SubscriptionRepository) FindByID(ctx context.Context, id string) (subscription.Subscription, error) {
	var m model.SubscriptionModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return subscription.Subscription{}, ErrNotFound
		}
		return subscription.Subscription{}, err
	}
	sub := m.ToDomain()
	r.decryptPassword(ctx, &sub)
	return sub, nil
}

func (r *SubscriptionRepository) FindByCustomerID(ctx context.Context, customerID string) ([]subscription.Subscription, error) {
	var mList []model.SubscriptionModel
	err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Order("created_at desc").Find(&mList).Error
	if err != nil {
		return nil, err
	}
	subs := make([]subscription.Subscription, len(mList))
	for i := range mList {
		subs[i] = mList[i].ToDomain()
		r.decryptPassword(ctx, &subs[i])
	}
	return subs, nil
}

func (r *SubscriptionRepository) FindAll(ctx context.Context) ([]subscription.Subscription, error) {
	var mList []model.SubscriptionModel
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&mList).Error
	if err != nil {
		return nil, err
	}
	subs := make([]subscription.Subscription, len(mList))
	for i := range mList {
		subs[i] = mList[i].ToDomain()
		r.decryptPassword(ctx, &subs[i])
	}
	return subs, nil
}

func (r *SubscriptionRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	res := r.db.WithContext(ctx).Model(&model.SubscriptionModel{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// FindByDeviceAndUsername implements router sync/isolation lookups: cari
// langganan berdasarkan router (devices.id) + username PPP/hotspot.
func (r *SubscriptionRepository) FindByDeviceAndUsername(ctx context.Context, deviceID, username string) (subscription.Subscription, error) {
	var m model.SubscriptionModel
	err := r.db.WithContext(ctx).
		Where("device_id = ? AND remote_username = ? AND deleted_at IS NULL", deviceID, username).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return subscription.Subscription{}, ErrNotFound
		}
		return subscription.Subscription{}, err
	}
	sub := m.ToDomain()
	r.decryptPassword(ctx, &sub)
	return sub, nil
}

// ListActive returns all non-deleted ACTIVE subscriptions.
func (r *SubscriptionRepository) ListActive(ctx context.Context) ([]subscription.Subscription, error) {
	var mList []model.SubscriptionModel
	err := r.db.WithContext(ctx).
		Where("status = ? AND deleted_at IS NULL", subscription.StatusActive).
		Order("created_at asc").
		Find(&mList).Error
	if err != nil {
		return nil, err
	}
	subs := make([]subscription.Subscription, len(mList))
	for i := range mList {
		subs[i] = mList[i].ToDomain()
		r.decryptPassword(ctx, &subs[i])
	}
	return subs, nil
}

// ListLifecycle returns ACTIVE + ISOLATED non-deleted subscriptions.
func (r *SubscriptionRepository) ListLifecycle(ctx context.Context) ([]subscription.Subscription, error) {
	var mList []model.SubscriptionModel
	err := r.db.WithContext(ctx).
		Where("status IN ? AND deleted_at IS NULL",
			[]string{subscription.StatusActive, subscription.StatusIsolated}).
		Order("created_at asc").
		Find(&mList).Error
	if err != nil {
		return nil, err
	}
	subs := make([]subscription.Subscription, len(mList))
	for i := range mList {
		subs[i] = mList[i].ToDomain()
		r.decryptPassword(ctx, &subs[i])
	}
	return subs, nil
}

// decryptPassword mengembalikan plaintext RemotePassword dari kolom
// remote_password_cipher. Bila vault nil / gagal dekripsi, biarkan nilai apa
// adanya agar pemanggil bisa membedakan "belum terenkripsi" (dev lama).
func (r *SubscriptionRepository) decryptPassword(ctx context.Context, sub *subscription.Subscription) {
	if r.vault == nil || sub.RemotePassword == "" {
		return
	}
	plain, err := r.vault.DecryptString(ctx, sub.RemotePassword)
	if err != nil {
		return
	}
	sub.RemotePassword = plain
}

// HasActiveForPlan implements the delete-guard lookup for service plans.
func (r *SubscriptionRepository) HasActiveForPlan(ctx context.Context, planID string) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.SubscriptionModel{}).
		Where("plan_id = ? AND deleted_at IS NULL AND status IN ?",
			planID,
			[]string{subscription.StatusActive, subscription.StatusIsolated}).
		Count(&n).Error
	return n > 0, err
}

// Delete implements hard-delete for the manage-subscription flow.
func (r *SubscriptionRepository) Delete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.SubscriptionModel{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
