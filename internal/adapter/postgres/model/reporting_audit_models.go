package model

import (
	"encoding/json"
	"time"

	"github.com/quixiq/polyglot/internal/domain/audit"
	"github.com/quixiq/polyglot/internal/domain/reporting"
)

// DailySnapshotModel is the GORM model for the pre-aggregated daily
// financial recap (DATABASE-SCHEMA-ISP.md §2.9 — daily_financial_snapshots).
type DailySnapshotModel struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	TenantID     string    `gorm:"type:text;not null;default:tenant-default;uniqueIndex:uq_daily_snapshot"`
	SnapshotDate time.Time `gorm:"type:date;not null;uniqueIndex:uq_daily_snapshot;index"`

	InvoiceCount int     `gorm:"not null;default:0"`
	InvoiceTotal float64 `gorm:"type:numeric(15,2);not null;default:0.00"`
	PaymentCount int     `gorm:"not null;default:0"`
	PaymentTotal float64 `gorm:"type:numeric(15,2);not null;default:0.00"`
	Outstanding  float64 `gorm:"column:outstanding_total;type:numeric(15,2);not null;default:0.00"`
	ExpenseTotal float64 `gorm:"type:numeric(15,2);not null;default:0.00"`

	ActiveSubscriptions int `gorm:"not null;default:0"`

	CashBalanceJSON json.RawMessage `gorm:"column:cash_balance_json;type:jsonb;not null;default:'{}'"`

	CreatedAt time.Time
}

func (DailySnapshotModel) TableName() string {
	return "daily_financial_snapshots"
}

func (m *DailySnapshotModel) ToDomain() reporting.DailyFinancialSnapshot {
	if m == nil {
		return reporting.DailyFinancialSnapshot{}
	}
	return reporting.DailyFinancialSnapshot{
		ID:                  m.ID,
		TenantID:            m.TenantID,
		SnapshotDate:        m.SnapshotDate,
		InvoiceCount:        m.InvoiceCount,
		InvoiceTotal:        m.InvoiceTotal,
		PaymentCount:        m.PaymentCount,
		PaymentTotal:        m.PaymentTotal,
		OutstandingTotal:    m.Outstanding,
		ExpenseTotal:        m.ExpenseTotal,
		ActiveSubscriptions: m.ActiveSubscriptions,
		CashBalanceJSON:     m.CashBalanceJSON,
		CreatedAt:           m.CreatedAt,
	}
}

func DailySnapshotModelFromDomain(s reporting.DailyFinancialSnapshot) *DailySnapshotModel {
	return &DailySnapshotModel{
		ID:                  s.ID,
		TenantID:            s.TenantID,
		SnapshotDate:        s.SnapshotDate,
		InvoiceCount:        s.InvoiceCount,
		InvoiceTotal:        s.InvoiceTotal,
		PaymentCount:        s.PaymentCount,
		PaymentTotal:        s.PaymentTotal,
		Outstanding:         s.OutstandingTotal,
		ExpenseTotal:        s.ExpenseTotal,
		ActiveSubscriptions: s.ActiveSubscriptions,
		CashBalanceJSON:     s.CashBalanceJSON,
		CreatedAt:           s.CreatedAt,
	}
}

// AuditLogModel is the GORM model for the audit trail
// (DATABASE-SCHEMA-ISP.md §2.10 — audit_logs).
type AuditLogModel struct {
	ID       uint   `gorm:"primaryKey;autoIncrement"`
	TenantID string `gorm:"type:text;not null;default:tenant-default"`

	ActorType   string `gorm:"type:varchar(20);not null;default:USER"` // USER | SYSTEM | PORTAL
	ActorID     string `gorm:"column:actor_id;type:text"`
	Action      string `gorm:"type:varchar(50);not null"`
	EntityType  string `gorm:"type:varchar(50);not null;index:idx_audit_logs_entity,priority:1"`
	EntityID    string `gorm:"type:text;not null;index:idx_audit_logs_entity,priority:2"`
	Description string `gorm:"type:text"`
	IPAddress   string `gorm:"type:varchar(45)"`

	CreatedAt time.Time `gorm:"index"`
}

func (AuditLogModel) TableName() string {
	return "audit_logs"
}

func (m *AuditLogModel) ToDomain() audit.AuditLog {
	if m == nil {
		return audit.AuditLog{}
	}
	return audit.AuditLog{
		ID:          m.ID,
		TenantID:    m.TenantID,
		ActorType:   m.ActorType,
		ActorID:     m.ActorID,
		Action:      m.Action,
		EntityType:  m.EntityType,
		EntityID:    m.EntityID,
		Description: m.Description,
		IPAddress:   m.IPAddress,
		CreatedAt:   m.CreatedAt,
	}
}

func AuditLogModelFromDomain(l audit.AuditLog) *AuditLogModel {
	return &AuditLogModel{
		ID:          l.ID,
		TenantID:    l.TenantID,
		ActorType:   l.ActorType,
		ActorID:     l.ActorID,
		Action:      l.Action,
		EntityType:  l.EntityType,
		EntityID:    l.EntityID,
		Description: l.Description,
		IPAddress:   l.IPAddress,
		CreatedAt:   l.CreatedAt,
	}
}
