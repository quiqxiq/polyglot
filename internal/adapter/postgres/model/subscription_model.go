package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/subscription"
)

// SubscriptionModel is the GORM model for the MAPPING-ONLY subscriptions
// table: identity + relational mapping to a Mikrotik account. No password,
// no router-state snapshots.
type SubscriptionModel struct {
	ID              string     `gorm:"primaryKey"`
	TenantID        string     `gorm:"not null;index;default:tenant-default"`
	CustomerID      string     `gorm:"not null;index"`
	PlanID          string     `gorm:"not null;index"`
	DeviceID        *string    `gorm:"column:device_id;index"`
	ServiceType     string     `gorm:"column:service_type;not null"`
	RemoteUsername  string     `gorm:"column:remote_username"`
	RemoteID        string     `gorm:"column:remote_id"`
	BillingDay      int        `gorm:"column:billing_day;not null;default:1"`
	Status          string     `gorm:"not null;default:PENDING_PROVISION;index"`
	StartDate       time.Time  `gorm:"column:start_date;not null"`
	IsolatedAt      *time.Time `gorm:"column:isolated_at"`
	IsolationReason string     `gorm:"column:isolation_reason;type:text"`
	Notes           string     `gorm:"type:text"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time `gorm:"index"`
}

func (SubscriptionModel) TableName() string {
	return "subscriptions"
}

func (m *SubscriptionModel) ToDomain() subscription.Subscription {
	if m == nil {
		return subscription.Subscription{}
	}
	deviceID := ""
	if m.DeviceID != nil {
		deviceID = *m.DeviceID
	}
	billingDay := m.BillingDay
	if billingDay == 0 {
		billingDay = 1
	}
	return subscription.Subscription{
		ID:              m.ID,
		TenantID:        m.TenantID,
		CustomerID:      m.CustomerID,
		PlanID:          m.PlanID,
		DeviceID:        deviceID,
		ServiceType:     m.ServiceType,
		RemoteUsername:  m.RemoteUsername,
		RemoteID:        m.RemoteID,
		BillingDay:      billingDay,
		Status:          m.Status,
		StartDate:       m.StartDate,
		IsolatedAt:      m.IsolatedAt,
		IsolationReason: m.IsolationReason,
		Notes:           m.Notes,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func SubscriptionModelFromDomain(s subscription.Subscription) *SubscriptionModel {
	tenantID := s.TenantID
	if tenantID == "" {
		tenantID = "tenant-default"
	}
	status := s.Status
	if status == "" {
		status = subscription.StatusPendingProvision
	}
	billingDay := s.BillingDay
	if billingDay == 0 {
		billingDay = 1
	}
	var deviceID *string
	if s.DeviceID != "" {
		d := s.DeviceID
		deviceID = &d
	}
	return &SubscriptionModel{
		ID:              s.ID,
		TenantID:        tenantID,
		CustomerID:      s.CustomerID,
		PlanID:          s.PlanID,
		DeviceID:        deviceID,
		ServiceType:     s.ServiceType,
		RemoteUsername:  s.RemoteUsername,
		RemoteID:        s.RemoteID,
		BillingDay:      billingDay,
		Status:          status,
		StartDate:       s.StartDate,
		IsolatedAt:      s.IsolatedAt,
		IsolationReason: s.IsolationReason,
		Notes:           s.Notes,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}
