package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

// GatewaySettlementModel is the GORM model for gateway_settlements table.
type GatewaySettlementModel struct {
	ID            string     `gorm:"primaryKey"`
	TenantID      string     `gorm:"type:text;not null;default:tenant-default"`
	Gateway       string     `gorm:"type:varchar(30);not null;index:idx_gateway_settlements_period"`
	SettlementRef string     `gorm:"type:varchar(100);not null"`
	PeriodFrom    *time.Time `gorm:"type:date;index:idx_gateway_settlements_period"`
	PeriodTo      *time.Time `gorm:"type:date;index:idx_gateway_settlements_period"`
	GrossAmount   float64    `gorm:"type:numeric(15,2);not null;default:0"`
	FeeAmount     float64    `gorm:"type:numeric(15,2);not null;default:0"`
	NetAmount     float64    `gorm:"type:numeric(15,2);not null;default:0"`
	BankAccountID *string    `gorm:"column:bank_account_id;type:text"`
	Status        string     `gorm:"type:varchar(20);not null;default:PENDING"`
	Notes         string     `gorm:"type:text"`
	CreatedBy     *uint      `gorm:"column:created_by"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// TableName returns the table name for gateway_settlements.
func (GatewaySettlementModel) TableName() string {
	return "gateway_settlements"
}

// ToDomain converts GatewaySettlementModel to domain billing.GatewaySettlement.
func (m *GatewaySettlementModel) ToDomain() billing.GatewaySettlement {
	if m == nil {
		return billing.GatewaySettlement{}
	}
	return billing.GatewaySettlement{
		ID:            m.ID,
		TenantID:      m.TenantID,
		Gateway:       m.Gateway,
		SettlementRef: m.SettlementRef,
		PeriodFrom:    m.PeriodFrom,
		PeriodTo:      m.PeriodTo,
		GrossAmount:   m.GrossAmount,
		FeeAmount:     m.FeeAmount,
		NetAmount:     m.NetAmount,
		BankAccountID: m.BankAccountID,
		Status:        m.Status,
		Notes:         m.Notes,
		CreatedBy:     m.CreatedBy,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

// GatewaySettlementModelFromDomain converts domain billing.GatewaySettlement to GatewaySettlementModel.
func GatewaySettlementModelFromDomain(s billing.GatewaySettlement) *GatewaySettlementModel {
	return &GatewaySettlementModel{
		ID:            s.ID,
		TenantID:      s.TenantID,
		Gateway:       s.Gateway,
		SettlementRef: s.SettlementRef,
		PeriodFrom:    s.PeriodFrom,
		PeriodTo:      s.PeriodTo,
		GrossAmount:   s.GrossAmount,
		FeeAmount:     s.FeeAmount,
		NetAmount:     s.NetAmount,
		BankAccountID: s.BankAccountID,
		Status:        s.Status,
		Notes:         s.Notes,
		CreatedBy:     s.CreatedBy,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}
}
