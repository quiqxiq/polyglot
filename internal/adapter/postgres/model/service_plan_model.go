package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/plan"
)

// ServicePlanModel is the GORM model for master paket layanan.
// Router-specific fields are NAME REFERENCES, not router state snapshots.
type ServicePlanModel struct {
	ID           string  `gorm:"primaryKey"`
	TenantID     string  `gorm:"not null;index;default:tenant-default"`
	Name         string  `gorm:"not null"`
	ServiceType  string  `gorm:"column:service_type;not null;index"`
	RateDownKbps int     `gorm:"column:rate_down_kbps;not null"`
	RateUpKbps   int     `gorm:"column:rate_up_kbps;not null"`
	Price        float64 `gorm:"not null"`
	IPPoolName   string  `gorm:"column:ip_pool_name"`
	ParentQueue  string  `gorm:"column:parent_queue"`
	AddressList  string  `gorm:"column:address_list"`
	SharedUsers  int     `gorm:"not null;default:1"`
	IsActive     bool    `gorm:"not null;default:true"`
	Description  string  `gorm:"type:text"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (ServicePlanModel) TableName() string {
	return "service_plans"
}

func (m *ServicePlanModel) ToDomain() plan.Plan {
	if m == nil {
		return plan.Plan{}
	}
	return plan.Plan{
		ID:           m.ID,
		TenantID:     m.TenantID,
		Name:         m.Name,
		ServiceType:  m.ServiceType,
		RateDownKbps: m.RateDownKbps,
		RateUpKbps:   m.RateUpKbps,
		Price:        m.Price,
		IPPoolName:   m.IPPoolName,
		ParentQueue:  m.ParentQueue,
		AddressList:  m.AddressList,
		SharedUsers:  m.SharedUsers,
		IsActive:     m.IsActive,
		Description:  m.Description,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func ServicePlanModelFromDomain(p plan.Plan) *ServicePlanModel {
	tenantID := p.TenantID
	if tenantID == "" {
		tenantID = "tenant-default"
	}
	shared := p.SharedUsers
	if shared <= 0 {
		shared = 1
	}
	return &ServicePlanModel{
		ID:           p.ID,
		TenantID:     tenantID,
		Name:         p.Name,
		ServiceType:  p.ServiceType,
		RateDownKbps: p.RateDownKbps,
		RateUpKbps:   p.RateUpKbps,
		Price:        p.Price,
		IPPoolName:   p.IPPoolName,
		ParentQueue:  p.ParentQueue,
		AddressList:  p.AddressList,
		SharedUsers:  shared,
		IsActive:     p.IsActive,
		Description:  p.Description,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}
