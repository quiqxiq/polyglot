package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/registration"
)

// RegistrationModel is the GORM model for the new-customer signup flow
// (DATABASE-SCHEMA-ISP.md §2.3 — registrations).
// reviewed_by / assigned_technician_id BIGINT → users.id (uint).
type RegistrationModel struct {
	ID string `gorm:"primaryKey"`

	TenantID       string `gorm:"type:text;not null;default:tenant-default"`
	RegistrationNo string `gorm:"type:varchar(30);unique;not null"` // "REG-202608-0001"
	PlanID         string `gorm:"type:text;not null;index"`

	FullName  string   `gorm:"type:varchar(100);not null"`
	Phone     string   `gorm:"type:varchar(20);not null;index"`
	Email     string   `gorm:"type:varchar(100)"`
	Address   string   `gorm:"type:text;not null"`
	Latitude  *float64 `gorm:"type:double precision"`
	Longitude *float64 `gorm:"type:double precision"`
	Notes     string   `gorm:"type:text"`

	Status string `gorm:"type:varchar(20);not null;default:PENDING;index"`

	ReviewedBy           *uint `gorm:"column:reviewed_by"`
	ReviewedAt           *time.Time
	AdminNotes           string     `gorm:"type:text"`
	ScheduledInstallDate *time.Time `gorm:"type:date"`
	// Fix review: TIME eksplisit, bukan VARCHAR(20).
	ScheduledInstallTime *time.Time `gorm:"type:time"`
	AssignedTechnicianID *uint      `gorm:"column:assigned_technician_id;index"`

	InstalledAt     *time.Time
	TechnicianNotes string `gorm:"type:text"`

	CustomerID     string `gorm:"type:text;index"`
	SubscriptionID string `gorm:"type:text"`
	InvoiceID      string `gorm:"type:text"`

	RejectedAt     *time.Time
	RejectedReason string `gorm:"type:text"`
	CancelledAt    *time.Time
	CancelReason   string `gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (RegistrationModel) TableName() string {
	return "registrations"
}

func (m *RegistrationModel) ToDomain() registration.Registration {
	if m == nil {
		return registration.Registration{}
	}
	return registration.Registration{
		ID:                   m.ID,
		TenantID:             m.TenantID,
		RegistrationNo:       m.RegistrationNo,
		PlanID:               m.PlanID,
		FullName:             m.FullName,
		Phone:                m.Phone,
		Email:                m.Email,
		Address:              m.Address,
		Latitude:             m.Latitude,
		Longitude:            m.Longitude,
		Notes:                m.Notes,
		Status:               m.Status,
		ReviewedBy:           m.ReviewedBy,
		ReviewedAt:           m.ReviewedAt,
		AdminNotes:           m.AdminNotes,
		ScheduledInstallDate: m.ScheduledInstallDate,
		ScheduledInstallTime: m.ScheduledInstallTime,
		AssignedTechnicianID: m.AssignedTechnicianID,
		InstalledAt:          m.InstalledAt,
		TechnicianNotes:      m.TechnicianNotes,
		CustomerID:           m.CustomerID,
		SubscriptionID:       m.SubscriptionID,
		InvoiceID:            m.InvoiceID,
		RejectedAt:           m.RejectedAt,
		RejectedReason:       m.RejectedReason,
		CancelledAt:          m.CancelledAt,
		CancelReason:         m.CancelReason,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
	}
}

func RegistrationModelFromDomain(r registration.Registration) *RegistrationModel {
	return &RegistrationModel{
		ID:                   r.ID,
		TenantID:             r.TenantID,
		RegistrationNo:       r.RegistrationNo,
		PlanID:               r.PlanID,
		FullName:             r.FullName,
		Phone:                r.Phone,
		Email:                r.Email,
		Address:              r.Address,
		Latitude:             r.Latitude,
		Longitude:            r.Longitude,
		Notes:                r.Notes,
		Status:               r.Status,
		ReviewedBy:           r.ReviewedBy,
		ReviewedAt:           r.ReviewedAt,
		AdminNotes:           r.AdminNotes,
		ScheduledInstallDate: r.ScheduledInstallDate,
		ScheduledInstallTime: r.ScheduledInstallTime,
		AssignedTechnicianID: r.AssignedTechnicianID,
		InstalledAt:          r.InstalledAt,
		TechnicianNotes:      r.TechnicianNotes,
		CustomerID:           r.CustomerID,
		SubscriptionID:       r.SubscriptionID,
		InvoiceID:            r.InvoiceID,
		RejectedAt:           r.RejectedAt,
		RejectedReason:       r.RejectedReason,
		CancelledAt:          r.CancelledAt,
		CancelReason:         r.CancelReason,
		CreatedAt:            r.CreatedAt,
		UpdatedAt:            r.UpdatedAt,
	}
}
