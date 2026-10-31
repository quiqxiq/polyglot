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

	// TargetDeviceID — router BRAS yang dipilih teknisi saat pemasangan.
	TargetDeviceID *string `gorm:"column:target_device_id;type:uuid;index"`

	InstalledAt     *time.Time
	TechnicianNotes string `gorm:"type:text"`

	CustomerID     string `gorm:"type:text;index"`
	SubscriptionID string `gorm:"type:text"`
	InvoiceID      string `gorm:"type:text"`

	RejectedAt     *time.Time
	RejectedReason string `gorm:"type:text"`
	CancelledAt    *time.Time
	CancelReason   string `gorm:"type:text"`

	// Kolom PSB v2 (migrasi 000032).
	NIK                       string     `gorm:"column:nik;type:varchar(20)"`
	IDCardPhotoURL            string     `gorm:"column:id_card_photo_url;type:text"`
	HousePhotoURL             string     `gorm:"column:house_photo_url;type:text"`
	SurveyDate                *time.Time `gorm:"column:survey_date;type:date"`
	SurveyTime                *time.Time `gorm:"column:survey_time;type:time"`
	SurveyorID                *uint      `gorm:"column:surveyor_id"`
	SurveyStatus              string     `gorm:"column:survey_status;type:varchar(20)"`
	SurveyNotes               string     `gorm:"column:survey_notes;type:text"`
	EstimatedCableLengthMeter int        `gorm:"column:estimated_cable_length_meter"`
	PublicToken               string     `gorm:"column:public_token;type:varchar(64);unique"`
	InstallationStartedAt     *time.Time `gorm:"column:installation_started_at"`
	InstallationPhotoURL      string     `gorm:"column:installation_photo_url;type:text"`
	ONTMACAddress             string     `gorm:"column:ont_mac_address;type:varchar(30)"`
	ONTSerialNumber           string     `gorm:"column:ont_serial_number;type:varchar(50)"`
	OLTDeviceID               *string    `gorm:"column:olt_device_id"`
	PONIndex                  *int       `gorm:"column:pon_index"`
	ONUIndex                  *int       `gorm:"column:onu_index"`
	DiscountAmount            float64    `gorm:"column:discount_amount;type:numeric(15,2);not null;default:0"`
	SalesAgentID              *uint      `gorm:"column:sales_agent_id"`

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
		ID:                        m.ID,
		TenantID:                  m.TenantID,
		RegistrationNo:            m.RegistrationNo,
		PlanID:                    m.PlanID,
		FullName:                  m.FullName,
		Phone:                     m.Phone,
		Email:                     m.Email,
		Address:                   m.Address,
		Latitude:                  m.Latitude,
		Longitude:                 m.Longitude,
		Notes:                     m.Notes,
		Status:                    m.Status,
		ReviewedBy:                m.ReviewedBy,
		ReviewedAt:                m.ReviewedAt,
		AdminNotes:                m.AdminNotes,
		ScheduledInstallDate:      m.ScheduledInstallDate,
		ScheduledInstallTime:      m.ScheduledInstallTime,
		AssignedTechnicianID:      m.AssignedTechnicianID,
		TargetDeviceID:            derefOrEmpty(m.TargetDeviceID),
		InstalledAt:               m.InstalledAt,
		TechnicianNotes:           m.TechnicianNotes,
		CustomerID:                m.CustomerID,
		SubscriptionID:            m.SubscriptionID,
		InvoiceID:                 m.InvoiceID,
		RejectedAt:                m.RejectedAt,
		RejectedReason:            m.RejectedReason,
		CancelledAt:               m.CancelledAt,
		CancelReason:              m.CancelReason,
		NIK:                       m.NIK,
		IDCardPhotoURL:            m.IDCardPhotoURL,
		HousePhotoURL:             m.HousePhotoURL,
		SurveyDate:                m.SurveyDate,
		SurveyTime:                m.SurveyTime,
		SurveyorID:                m.SurveyorID,
		SurveyStatus:              m.SurveyStatus,
		SurveyNotes:               m.SurveyNotes,
		EstimatedCableLengthMeter: m.EstimatedCableLengthMeter,
		PublicToken:               m.PublicToken,
		InstallationStartedAt:     m.InstallationStartedAt,
		InstallationPhotoURL:      m.InstallationPhotoURL,
		ONTMACAddress:             m.ONTMACAddress,
		ONTSerialNumber:           m.ONTSerialNumber,
		OLTDeviceID:               derefOrEmpty(m.OLTDeviceID),
		PONIndex:                  m.PONIndex,
		ONUIndex:                  m.ONUIndex,
		DiscountAmount:            m.DiscountAmount,
		SalesAgentID:              m.SalesAgentID,
		CreatedAt:                 m.CreatedAt,
		UpdatedAt:                 m.UpdatedAt,
	}
}

func RegistrationModelFromDomain(r registration.Registration) *RegistrationModel {
	return &RegistrationModel{
		ID:                        r.ID,
		TenantID:                  r.TenantID,
		RegistrationNo:            r.RegistrationNo,
		PlanID:                    r.PlanID,
		FullName:                  r.FullName,
		Phone:                     r.Phone,
		Email:                     r.Email,
		Address:                   r.Address,
		Latitude:                  r.Latitude,
		Longitude:                 r.Longitude,
		Notes:                     r.Notes,
		Status:                    r.Status,
		ReviewedBy:                r.ReviewedBy,
		ReviewedAt:                r.ReviewedAt,
		AdminNotes:                r.AdminNotes,
		ScheduledInstallDate:      r.ScheduledInstallDate,
		ScheduledInstallTime:      r.ScheduledInstallTime,
		AssignedTechnicianID:      r.AssignedTechnicianID,
		TargetDeviceID:            ptrIfNotEmpty(r.TargetDeviceID),
		InstalledAt:               r.InstalledAt,
		TechnicianNotes:           r.TechnicianNotes,
		CustomerID:                r.CustomerID,
		SubscriptionID:            r.SubscriptionID,
		InvoiceID:                 r.InvoiceID,
		RejectedAt:                r.RejectedAt,
		RejectedReason:            r.RejectedReason,
		CancelledAt:               r.CancelledAt,
		CancelReason:              r.CancelReason,
		NIK:                       r.NIK,
		IDCardPhotoURL:            r.IDCardPhotoURL,
		HousePhotoURL:             r.HousePhotoURL,
		SurveyDate:                r.SurveyDate,
		SurveyTime:                r.SurveyTime,
		SurveyorID:                r.SurveyorID,
		SurveyStatus:              r.SurveyStatus,
		SurveyNotes:               r.SurveyNotes,
		EstimatedCableLengthMeter: r.EstimatedCableLengthMeter,
		PublicToken:               r.PublicToken,
		InstallationStartedAt:     r.InstallationStartedAt,
		InstallationPhotoURL:      r.InstallationPhotoURL,
		ONTMACAddress:             r.ONTMACAddress,
		ONTSerialNumber:           r.ONTSerialNumber,
		OLTDeviceID:               ptrIfNotEmpty(r.OLTDeviceID),
		PONIndex:                  r.PONIndex,
		ONUIndex:                  r.ONUIndex,
		DiscountAmount:            r.DiscountAmount,
		SalesAgentID:              r.SalesAgentID,
		CreatedAt:                 r.CreatedAt,
		UpdatedAt:                 r.UpdatedAt,
	}
}

func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
