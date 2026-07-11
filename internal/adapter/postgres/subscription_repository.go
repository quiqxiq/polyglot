package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/domain/subscription"
)

// subscriptionModel maps the `subscriptions` table per migration 000009.
type subscriptionModel struct {
	ID                     string     `gorm:"column:id;primaryKey"`
	CustomerID             string     `gorm:"column:customer_id;not null"`
	PlanID                 string     `gorm:"column:plan_id;not null"`
	ServiceType            string     `gorm:"column:service_type;not null"`
	Status                 string     `gorm:"column:status;not null;default:'pending_install'"`
	DeviceID               string     `gorm:"column:device_id;not null"`
	ODPID                  *string    `gorm:"column:odp_id"`
	ODPPort                string     `gorm:"column:odp_port"`
	ONUSerialNumber        string     `gorm:"column:onu_serial_number"`
	IPPoolID               *string    `gorm:"column:ip_pool_id"`
	PPPoEUsername          string     `gorm:"column:pppoe_username"`
	PPPoEPasswordEncrypted string     `gorm:"column:pppoe_password_encrypted"`
	StaticIP               string     `gorm:"column:static_ip"`
	MACAddress             string     `gorm:"column:mac_address"`
	InstalledAt            *time.Time `gorm:"column:installed_at"`
	ActivatedAt            *time.Time `gorm:"column:activated_at"`
	SuspendedAt            *time.Time `gorm:"column:suspended_at"`
	TerminatedAt           *time.Time `gorm:"column:terminated_at"`
	SuspensionReason       string     `gorm:"column:suspension_reason"`
	CreatedAt              time.Time  `gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt              time.Time  `gorm:"column:updated_at;not null;autoUpdateTime"`
}

func (subscriptionModel) TableName() string {
	return "subscriptions"
}

func (m subscriptionModel) toDomain() subscription.Subscription {
	return subscription.Subscription{
		ID:                     m.ID,
		CustomerID:             m.CustomerID,
		PlanID:                 m.PlanID,
		ServiceType:            m.ServiceType,
		Status:                 m.Status,
		DeviceID:               m.DeviceID,
		ODPID:                  m.ODPID,
		ODPPort:                m.ODPPort,
		ONUSerialNumber:        m.ONUSerialNumber,
		IPPoolID:               m.IPPoolID,
		PPPoEUsername:          m.PPPoEUsername,
		PPPoEPasswordEncrypted: m.PPPoEPasswordEncrypted,
		StaticIP:               m.StaticIP,
		MACAddress:             m.MACAddress,
		InstalledAt:            m.InstalledAt,
		ActivatedAt:            m.ActivatedAt,
		SuspendedAt:            m.SuspendedAt,
		TerminatedAt:           m.TerminatedAt,
		SuspensionReason:       m.SuspensionReason,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
	}
}

func fromSubscriptionDomain(s subscription.Subscription) subscriptionModel {
	return subscriptionModel{
		ID:                     s.ID,
		CustomerID:             s.CustomerID,
		PlanID:                 s.PlanID,
		ServiceType:            s.ServiceType,
		Status:                 s.Status,
		DeviceID:               s.DeviceID,
		ODPID:                  s.ODPID,
		ODPPort:                s.ODPPort,
		ONUSerialNumber:        s.ONUSerialNumber,
		IPPoolID:               s.IPPoolID,
		PPPoEUsername:          s.PPPoEUsername,
		PPPoEPasswordEncrypted: s.PPPoEPasswordEncrypted,
		StaticIP:               s.StaticIP,
		MACAddress:             s.MACAddress,
		InstalledAt:            s.InstalledAt,
		ActivatedAt:            s.ActivatedAt,
		SuspendedAt:            s.SuspendedAt,
		TerminatedAt:           s.TerminatedAt,
		SuspensionReason:       s.SuspensionReason,
		CreatedAt:              s.CreatedAt,
		UpdatedAt:              s.UpdatedAt,
	}
}

// SubscriptionRepository implements port.SubscriptionRepository backed by PostgreSQL.
type SubscriptionRepository struct {
	db *gorm.DB
}

// NewSubscriptionRepository returns a port.SubscriptionRepository backed by GORM/Postgres.
func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// FindByID returns the subscription for the given id, or subscription.ErrNotFound.
func (r *SubscriptionRepository) FindByID(ctx context.Context, id string) (subscription.Subscription, error) {
	var m subscriptionModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return subscription.Subscription{}, fmt.Errorf("subscription %s: %w", id, subscription.ErrNotFound)
		}
		return subscription.Subscription{}, fmt.Errorf("subscription %s: %w", id, err)
	}
	return m.toDomain(), nil
}

// FindAll returns all subscriptions ordered by created_at desc.
func (r *SubscriptionRepository) FindAll(ctx context.Context) ([]subscription.Subscription, error) {
	var models []subscriptionModel
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	subs := make([]subscription.Subscription, len(models))
	for i, m := range models {
		subs[i] = m.toDomain()
	}
	return subs, nil
}

// FindByCustomer returns subscriptions for a specific customer.
func (r *SubscriptionRepository) FindByCustomer(ctx context.Context, customerID string) ([]subscription.Subscription, error) {
	var models []subscriptionModel
	if err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Order("created_at desc").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list subscriptions by customer: %w", err)
	}
	subs := make([]subscription.Subscription, len(models))
	for i, m := range models {
		subs[i] = m.toDomain()
	}
	return subs, nil
}

// FindByDevice returns subscriptions on a specific device.
func (r *SubscriptionRepository) FindByDevice(ctx context.Context, deviceID string) ([]subscription.Subscription, error) {
	var models []subscriptionModel
	if err := r.db.WithContext(ctx).Where("device_id = ?", deviceID).Order("created_at desc").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("list subscriptions by device: %w", err)
	}
	subs := make([]subscription.Subscription, len(models))
	for i, m := range models {
		subs[i] = m.toDomain()
	}
	return subs, nil
}

// Create inserts a new subscription.
func (r *SubscriptionRepository) Create(ctx context.Context, s subscription.Subscription) (subscription.Subscription, error) {
	m := fromSubscriptionDomain(s)
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return subscription.Subscription{}, fmt.Errorf("create subscription: %w", err)
	}
	return m.toDomain(), nil
}

// Update modifies an existing subscription.
func (r *SubscriptionRepository) Update(ctx context.Context, s subscription.Subscription) (subscription.Subscription, error) {
	m := fromSubscriptionDomain(s)
	if err := r.db.WithContext(ctx).Save(&m).Error; err != nil {
		return subscription.Subscription{}, fmt.Errorf("update subscription: %w", err)
	}
	return m.toDomain(), nil
}

// Delete removes a subscription by id.
func (r *SubscriptionRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&subscriptionModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}
	return nil
}
