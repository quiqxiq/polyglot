package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/subscription"
)

// SubscriptionModel is the GORM model for customer plan subscriptions
// mapped to a MikroTik router (DATABASE-SCHEMA-ISP.md §2.5).
//
// DEVIASI vs dokumen skema: kolom `remote_password` VARCHAR(100) plaintext
// diganti `remote_password_cipher` — kredensial disimpan terenkripsi AES-GCM
// via vault (prinsip arsitektur migrasi 000001). Enkripsi/dekripsi dilakukan
// layer repository pada fase berikutnya; mapper di sini passthrough.
type SubscriptionModel struct {
	ID string `gorm:"primaryKey"`
	// device_id UUID nullable → devices(id); tipe eksplisit agar cocok
	// dengan kolom UUID tabel devices.
	TenantID    string  `gorm:"type:text;not null;default:tenant-default"`
	CustomerID  string  `gorm:"type:text;not null;index"`
	PlanID      string  `gorm:"type:text;not null;index"`
	DeviceID    *string `gorm:"column:device_id;type:uuid;index"`
	ServiceType string  `gorm:"type:varchar(20);not null;default:PPPOE"`

	RemoteUsername string `gorm:"type:varchar(100);not null"`
	RemotePassword string `gorm:"column:remote_password_cipher;type:text;not null"`
	LocalAddress   string `gorm:"type:varchar(45)"`
	RemoteAddress  string `gorm:"type:varchar(45)"`
	ParentQueue    string `gorm:"type:varchar(50);default:none"`
	RateLimit      string `gorm:"type:varchar(100)"`

	RouterProfile   string `gorm:"column:router_profile;type:varchar(100)"`
	ProvisionStatus string `gorm:"type:varchar(20);not null;default:NONE"`

	BillingCycle       string     `gorm:"type:varchar(20);not null;default:MONTHLY"`
	BillingDay         int        `gorm:"not null;default:1"`
	AutoIsolate        bool       `gorm:"not null;default:true"`
	IsolationGraceDays int        `gorm:"not null;default:3"`
	Status             string     `gorm:"type:varchar(20);not null;default:ACTIVE;index"`
	StartDate          time.Time  `gorm:"type:date;not null;default:CURRENT_DATE"`
	EndDate            *time.Time // nullable — fix review dari TIMESTAMPTZ NOT NULL lama

	CustomPrice        *float64 `gorm:"type:numeric(15,2)"`
	CurrentPeriodStart *time.Time
	CurrentPeriodEnd   *time.Time
	Notes              string `gorm:"type:text"`
	DeletedAt          *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName returns the database table name for subscriptions.
func (SubscriptionModel) TableName() string {
	return "subscriptions"
}

// ToDomain converts a subscription database model to its domain representation.
func (m *SubscriptionModel) ToDomain() subscription.Subscription {
	if m == nil {
		return subscription.Subscription{}
	}
	return subscription.Subscription{
		ID:                 m.ID,
		TenantID:           m.TenantID,
		CustomerID:         m.CustomerID,
		PlanID:             m.PlanID,
		DeviceID:           m.DeviceID,
		ServiceType:        m.ServiceType,
		RemoteUsername:     m.RemoteUsername,
		RemotePassword:     m.RemotePassword, // ciphertext sampai vault wiring di repository
		LocalAddress:       m.LocalAddress,
		RemoteAddress:      m.RemoteAddress,
		ParentQueue:        m.ParentQueue,
		RateLimit:          m.RateLimit,
		RouterProfile:      m.RouterProfile,
		ProvisionStatus:    m.ProvisionStatus,
		BillingCycle:       m.BillingCycle,
		BillingDay:         m.BillingDay,
		AutoIsolate:        m.AutoIsolate,
		IsolationGraceDays: m.IsolationGraceDays,
		Status:             m.Status,
		StartDate:          m.StartDate,
		EndDate:            m.EndDate,
		CustomPrice:        m.CustomPrice,
		CurrentPeriodStart: m.CurrentPeriodStart,
		CurrentPeriodEnd:   m.CurrentPeriodEnd,
		Notes:              m.Notes,
		DeletedAt:          m.DeletedAt,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}

func SubscriptionModelFromDomain(s subscription.Subscription) *SubscriptionModel {
	return &SubscriptionModel{
		ID:                 s.ID,
		TenantID:           s.TenantID,
		CustomerID:         s.CustomerID,
		PlanID:             s.PlanID,
		DeviceID:           s.DeviceID,
		ServiceType:        s.ServiceType,
		RemoteUsername:     s.RemoteUsername,
		RemotePassword:     s.RemotePassword,
		LocalAddress:       s.LocalAddress,
		RemoteAddress:      s.RemoteAddress,
		ParentQueue:        s.ParentQueue,
		RateLimit:          s.RateLimit,
		RouterProfile:      s.RouterProfile,
		ProvisionStatus:    s.ProvisionStatus,
		BillingCycle:       s.BillingCycle,
		BillingDay:         s.BillingDay,
		AutoIsolate:        s.AutoIsolate,
		IsolationGraceDays: s.IsolationGraceDays,
		Status:             s.Status,
		StartDate:          s.StartDate,
		EndDate:            s.EndDate,
		CustomPrice:        s.CustomPrice,
		CurrentPeriodStart: s.CurrentPeriodStart,
		CurrentPeriodEnd:   s.CurrentPeriodEnd,
		Notes:              s.Notes,
		DeletedAt:          s.DeletedAt,
		CreatedAt:          s.CreatedAt,
		UpdatedAt:          s.UpdatedAt,
	}
}
