package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/registration"
)

// RegistrationModel is the GORM model for new-subscriber applications.
type RegistrationModel struct {
	ID                    string     `gorm:"primaryKey"`
	TenantID              string     `gorm:"not null;index;default:tenant-default"`
	RegistrationNo        string     `gorm:"column:registration_no;uniqueIndex"`
	PlanID                string     `gorm:"column:plan_id;not null"`
	FullName              string     `gorm:"column:full_name;not null"`
	Phone                 string     `gorm:"not null;index"`
	Address               string     `gorm:"type:text;not null"`
	Latitude              float64    `gorm:"column:latitude"`
	Longitude             float64    `gorm:"column:longitude"`
	Notes                 string     `gorm:"type:text"`
	Status                string     `gorm:"not null;default:PENDING;index"`
	ReviewedBy            *int64     `gorm:"column:reviewed_by"`
	ReviewedAt            *time.Time `gorm:"column:reviewed_at"`
	AdminNotes            string     `gorm:"column:admin_notes;type:text"`
	ScheduledInstallDate  *time.Time `gorm:"column:scheduled_install_date"`
	AssignedTechnicianID  *int64     `gorm:"column:assigned_technician_id"`
	DeviceID              *string    `gorm:"column:device_id"`
	InstalledAt           *time.Time `gorm:"column:installed_at"`
	TechnicianNotes       string     `gorm:"column:technician_notes;type:text"`
	CustomerID            *string    `gorm:"column:customer_id"`
	SubscriptionID        *string    `gorm:"column:subscription_id"`
	RejectedAt            *time.Time `gorm:"column:rejected_at"`
	RejectedReason        string     `gorm:"column:rejected_reason;type:text"`
	CancelledAt           *time.Time `gorm:"column:cancelled_at"`
	CancelReason          string     `gorm:"column:cancel_reason;type:text"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (RegistrationModel) TableName() string {
	return "registrations"
}

func (m *RegistrationModel) ToDomain() registration.Registration {
	if m == nil {
		return registration.Registration{}
	}
	deviceID := ""
	if m.DeviceID != nil {
		deviceID = *m.DeviceID
	}
	customerID := ""
	if m.CustomerID != nil {
		customerID = *m.CustomerID
	}
	subscriptionID := ""
	if m.SubscriptionID != nil {
		subscriptionID = *m.SubscriptionID
	}
	return registration.Registration{
		ID:                   m.ID,
		TenantID:             m.TenantID,
		RegistrationNo:       m.RegistrationNo,
		PlanID:               m.PlanID,
		FullName:             m.FullName,
		Phone:                m.Phone,
		Address:              m.Address,
		Latitude:             m.Latitude,
		Longitude:            m.Longitude,
		Notes:                m.Notes,
		Status:               m.Status,
		ReviewedBy:           m.ReviewedBy,
		ReviewedAt:           m.ReviewedAt,
		AdminNotes:           m.AdminNotes,
		ScheduledInstallDate: m.ScheduledInstallDate,
		AssignedTechnicianID: m.AssignedTechnicianID,
		DeviceID:             deviceID,
		InstalledAt:          m.InstalledAt,
		TechnicianNotes:      m.TechnicianNotes,
		CustomerID:           customerID,
		SubscriptionID:       subscriptionID,
		RejectedAt:           m.RejectedAt,
		RejectedReason:       m.RejectedReason,
		CancelledAt:          m.CancelledAt,
		CancelReason:         m.CancelReason,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
	}
}

func RegistrationModelFromDomain(r registration.Registration) *RegistrationModel {
	tenantID := r.TenantID
	if tenantID == "" {
		tenantID = "tenant-default"
	}
	status := r.Status
	if status == "" {
		status = registration.StatusPending
	}
	var deviceID *string
	if r.DeviceID != "" {
		d := r.DeviceID
		deviceID = &d
	}
	var customerID, subscriptionID *string
	if r.CustomerID != "" {
		c := r.CustomerID
		customerID = &c
	}
	if r.SubscriptionID != "" {
		s := r.SubscriptionID
		subscriptionID = &s
	}
	return &RegistrationModel{
		ID:                   r.ID,
		TenantID:             tenantID,
		RegistrationNo:       r.RegistrationNo,
		PlanID:               r.PlanID,
		FullName:             r.FullName,
		Phone:                r.Phone,
		Address:              r.Address,
		Latitude:             r.Latitude,
		Longitude:            r.Longitude,
		Notes:                r.Notes,
		Status:               status,
		ReviewedBy:           r.ReviewedBy,
		ReviewedAt:           r.ReviewedAt,
		AdminNotes:           r.AdminNotes,
		ScheduledInstallDate: r.ScheduledInstallDate,
		AssignedTechnicianID: r.AssignedTechnicianID,
		DeviceID:             deviceID,
		InstalledAt:          r.InstalledAt,
		TechnicianNotes:      r.TechnicianNotes,
		CustomerID:           customerID,
		SubscriptionID:       subscriptionID,
		RejectedAt:           r.RejectedAt,
		RejectedReason:       r.RejectedReason,
		CancelledAt:          r.CancelledAt,
		CancelReason:         r.CancelReason,
		CreatedAt:            r.CreatedAt,
		UpdatedAt:            r.UpdatedAt,
	}
}
