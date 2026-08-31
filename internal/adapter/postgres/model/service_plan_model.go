package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/plan"
)

// ServicePlanModel is the GORM model for ISP service plans with full
// MikroTik profile parameters (DATABASE-SCHEMA-ISP.md §2.2).
type ServicePlanModel struct {
	ID                    string `gorm:"primaryKey"`
	TenantID              string `gorm:"type:text;not null;default:tenant-default;uniqueIndex:uq_service_plans_tenant_name"`
	Name                  string `gorm:"type:varchar(100);not null;uniqueIndex:uq_service_plans_tenant_name"`
	ServiceType           string `gorm:"type:varchar(20);not null;index"`
	BandwidthDownloadKbps int    `gorm:"not null"`
	BandwidthUploadKbps   int    `gorm:"not null"`
	BurstDownloadKbps     int
	BurstUploadKbps       int
	BurstThresholdKbps    int
	BurstTimeSeconds      int
	Price                 float64 `gorm:"type:numeric(15,2);not null"`
	InstallationFee       float64 `gorm:"type:numeric(15,2)"`
	TaxPercent            float64 `gorm:"type:numeric(5,2)"`

	IPPoolName        string `gorm:"type:varchar(50)"`
	RemoteAddressPool string `gorm:"type:varchar(50)"`
	ParentQueue       string `gorm:"type:varchar(50);default:none"`
	AddressList       string `gorm:"type:varchar(50)"`
	SharedUsers       int    `gorm:"default:1"`
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
	sharedUsers := m.SharedUsers
	if sharedUsers <= 0 {
		sharedUsers = 1
	}
	p := plan.ServicePlan{
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
		InstallationFee:       m.InstallationFee,
		TaxPercent:            m.TaxPercent,
		IPPoolName:            m.IPPoolName,
		RemoteAddressPool:     m.RemoteAddressPool,
		ParentQueue:           m.ParentQueue,
		AddressList:           m.AddressList,
		SharedUsers:           sharedUsers,
		IsActive:              m.IsActive,
		Description:           m.Description,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
	}

	if m.ServiceType == plan.TypeHotspot {
		p.Hotspot = &plan.HotspotPlanConfig{
			IPPoolName:  m.IPPoolName,
			AddressList: m.AddressList,
			SharedUsers: sharedUsers,
		}
	} else {
		p.PPPoE = &plan.PPPoEPlanConfig{
			RemoteAddressPool: m.RemoteAddressPool,
			AddressList:       m.AddressList,
		}
	}
	return p
}

// ServicePlanModelFromDomain converts a service plan domain entity to a database model.
func ServicePlanModelFromDomain(p plan.ServicePlan) *ServicePlanModel {
	sharedUsers := p.SharedUsers
	if sharedUsers <= 0 {
		sharedUsers = 1
	}
	m := &ServicePlanModel{
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
		InstallationFee:       p.InstallationFee,
		TaxPercent:            p.TaxPercent,
		IPPoolName:            p.IPPoolName,
		RemoteAddressPool:     p.RemoteAddressPool,
		ParentQueue:           p.ParentQueue,
		AddressList:           p.AddressList,
		SharedUsers:           sharedUsers,
		IsActive:              p.IsActive,
		Description:           p.Description,
		CreatedAt:             p.CreatedAt,
		UpdatedAt:             p.UpdatedAt,
	}

	if p.PPPoE != nil {
		if p.PPPoE.RemoteAddressPool != "" {
			m.RemoteAddressPool = p.PPPoE.RemoteAddressPool
		}
		if p.PPPoE.AddressList != "" {
			m.AddressList = p.PPPoE.AddressList
		}
	}
	if p.Hotspot != nil {
		if p.Hotspot.IPPoolName != "" {
			m.IPPoolName = p.Hotspot.IPPoolName
		}
		if p.Hotspot.AddressList != "" {
			m.AddressList = p.Hotspot.AddressList
		}
		if p.Hotspot.SharedUsers > 0 {
			m.SharedUsers = p.Hotspot.SharedUsers
		}
	}
	return m
}
