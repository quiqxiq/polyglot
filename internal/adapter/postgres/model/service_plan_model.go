package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/plan"
)

// ServicePlanModel is the GORM model for ISP service plans with full
// MikroTik profile parameters (DATABASE-SCHEMA-ISP.md §2.2).
// Tabel berganti nama dari `plans` (migrasi 000006) menjadi `service_plans`;
// migrasi SQL prod untuk rename + copy data menyusul di fase berikutnya.
type ServicePlanModel struct {
	ID string `gorm:"primaryKey"`
	// Nama paket unik per tenant (mis. "100-RB-100", "HOME-20M").
	TenantID              string `gorm:"type:text;not null;default:tenant-default;uniqueIndex:uq_service_plans_tenant_name"`
	Name                  string `gorm:"type:varchar(100);not null;uniqueIndex:uq_service_plans_tenant_name"`
	ServiceType           string `gorm:"type:varchar(20);not null;index"`
	BandwidthDownloadKbps int    `gorm:"not null"`
	BandwidthUploadKbps   int    `gorm:"not null"`
	BurstDownloadKbps     int
	BurstUploadKbps       int
	BurstThresholdKbps    int
	BurstTimeSeconds      int

	Price           float64 `gorm:"type:numeric(15,2);not null"`
	SellingPrice    float64 `gorm:"type:numeric(15,2)"`
	InstallationFee float64 `gorm:"type:numeric(15,2)"`
	TaxPercent      float64 `gorm:"type:numeric(5,2)"`

	Validity        string `gorm:"type:varchar(20);default:30d"`
	ValidityMode    string `gorm:"type:varchar(20);default:CALENDAR"`
	SimultaneousUse int    `gorm:"default:1"`

	IPPoolName  string `gorm:"type:varchar(50)"`
	ParentQueue string `gorm:"type:varchar(50);default:none"`
	AddressList string `gorm:"type:varchar(50)"`
	SharedUsers int    `gorm:"default:1"`
	ExpireMode  string `gorm:"type:varchar(10);default:ntf"`
	LockUser    bool   `gorm:"default:false"`
	LockServer  bool   `gorm:"default:false"`
	// RemoteAddressPool: pool IP sumber alamat pelanggan (kolom remote-address
	// pada /ppp/profile RouterOS) — relevan untuk PPPOE/DEDICATED.
	RemoteAddressPool string `gorm:"type:varchar(50)"`
	LimitUptime       string `gorm:"type:varchar(20)"`
	LimitBytes        string `gorm:"type:varchar(20)"` // NULL/empty = unlimited flat rate
	IsActive          bool   `gorm:"not null;default:true;index"`
	Description       string `gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName returns the database table name for service plans.
func (ServicePlanModel) TableName() string {
	return "service_plans"
}

// ToDomain converts a service plan database model to its domain representation.
func (m *ServicePlanModel) ToDomain() plan.ServicePlan {
	if m == nil {
		return plan.ServicePlan{}
	}
	return plan.ServicePlan{
		ID:                    m.ID,
		TenantID:              m.TenantID,
		Name:                  m.Name,
		ServiceType:           m.ServiceType,
		BandwidthDownloadKbps: m.BandwidthDownloadKbps,
		BandwidthUploadKbps:   m.BandwidthUploadKbps,
		BurstDownloadKbps:     m.BurstDownloadKbps,
		BurstUploadKbps:       m.BurstUploadKbps,
		BurstThresholdKbps:    m.BurstThresholdKbps,
		BurstTimeSeconds:      m.BurstTimeSeconds,
		Price:                 m.Price,
		SellingPrice:          m.SellingPrice,
		InstallationFee:       m.InstallationFee,
		TaxPercent:            m.TaxPercent,
		Validity:              m.Validity,
		ValidityMode:          m.ValidityMode,
		SimultaneousUse:       m.SimultaneousUse,
		IPPoolName:            m.IPPoolName,
		RemoteAddressPool:     m.RemoteAddressPool,
		ParentQueue:           m.ParentQueue,
		AddressList:           m.AddressList,
		SharedUsers:           m.SharedUsers,
		ExpireMode:            m.ExpireMode,
		LockUser:              m.LockUser,
		LockServer:            m.LockServer,
		LimitUptime:           m.LimitUptime,
		LimitBytes:            m.LimitBytes,
		IsActive:              m.IsActive,
		Description:           m.Description,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
	}
}

// ServicePlanModelFromDomain converts a service plan domain entity to a database model.
func ServicePlanModelFromDomain(p plan.ServicePlan) *ServicePlanModel {
	return &ServicePlanModel{
		ID:                    p.ID,
		TenantID:              p.TenantID,
		Name:                  p.Name,
		ServiceType:           p.ServiceType,
		BandwidthDownloadKbps: p.BandwidthDownloadKbps,
		BandwidthUploadKbps:   p.BandwidthUploadKbps,
		BurstDownloadKbps:     p.BurstDownloadKbps,
		BurstUploadKbps:       p.BurstUploadKbps,
		BurstThresholdKbps:    p.BurstThresholdKbps,
		BurstTimeSeconds:      p.BurstTimeSeconds,
		Price:                 p.Price,
		SellingPrice:          p.SellingPrice,
		InstallationFee:       p.InstallationFee,
		TaxPercent:            p.TaxPercent,
		Validity:              p.Validity,
		ValidityMode:          p.ValidityMode,
		SimultaneousUse:       p.SimultaneousUse,
		IPPoolName:            p.IPPoolName,
		RemoteAddressPool:     p.RemoteAddressPool,
		ParentQueue:           p.ParentQueue,
		AddressList:           p.AddressList,
		SharedUsers:           p.SharedUsers,
		ExpireMode:            p.ExpireMode,
		LockUser:              p.LockUser,
		LockServer:            p.LockServer,
		LimitUptime:           p.LimitUptime,
		LimitBytes:            p.LimitBytes,
		IsActive:              p.IsActive,
		Description:           p.Description,
		CreatedAt:             p.CreatedAt,
		UpdatedAt:             p.UpdatedAt,
	}
}
