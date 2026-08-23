package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/cashbook"
)

// CashAccountModel is the GORM model for cash/bank operational accounts
// (DATABASE-SCHEMA-ISP.md §2.8 — cash_accounts).
type CashAccountModel struct {
	ID string `gorm:"primaryKey"`

	TenantID    string `gorm:"type:text;not null;default:tenant-default"`
	AccountCode string `gorm:"type:varchar(30);unique;not null"` // '1001-KAS-KANTOR'
	Name        string `gorm:"type:varchar(100);not null"`
	Type        string `gorm:"type:varchar(30);not null;default:CASH"`
	IsActive    bool   `gorm:"not null;default:true"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (CashAccountModel) TableName() string {
	return "cash_accounts"
}

func (m *CashAccountModel) ToDomain() cashbook.CashAccount {
	if m == nil {
		return cashbook.CashAccount{}
	}
	return cashbook.CashAccount{
		ID:          m.ID,
		TenantID:    m.TenantID,
		AccountCode: m.AccountCode,
		Name:        m.Name,
		Type:        m.Type,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func CashAccountModelFromDomain(a cashbook.CashAccount) *CashAccountModel {
	return &CashAccountModel{
		ID:          a.ID,
		TenantID:    a.TenantID,
		AccountCode: a.AccountCode,
		Name:        a.Name,
		Type:        a.Type,
		IsActive:    a.IsActive,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

// CashCategoryModel is the GORM model for cash book categories
// (DATABASE-SCHEMA-ISP.md §2.8 — cash_categories).
type CashCategoryModel struct {
	ID       string `gorm:"primaryKey"`
	TenantID string `gorm:"type:text;not null;default:tenant-default;uniqueIndex:uq_cash_cat_tenant"`
	Name     string `gorm:"type:varchar(100);not null;uniqueIndex:uq_cash_cat_tenant"`
	Type     string `gorm:"type:varchar(20);not null"` // INCOME | EXPENSE
	IsActive bool   `gorm:"not null;default:true"`
}

func (CashCategoryModel) TableName() string {
	return "cash_categories"
}

func (m *CashCategoryModel) ToDomain() cashbook.CashCategory {
	if m == nil {
		return cashbook.CashCategory{}
	}
	return cashbook.CashCategory{
		ID:       m.ID,
		TenantID: m.TenantID,
		Name:     m.Name,
		Type:     m.Type,
		IsActive: m.IsActive,
	}
}

func CashCategoryModelFromDomain(c cashbook.CashCategory) *CashCategoryModel {
	return &CashCategoryModel{
		ID:       c.ID,
		TenantID: c.TenantID,
		Name:     c.Name,
		Type:     c.Type,
		IsActive: c.IsActive,
	}
}

// CashTransactionModel is the GORM model for the cash journal
// (DATABASE-SCHEMA-ISP.md §2.8 — cash_transactions).
type CashTransactionModel struct {
	ID string `gorm:"primaryKey"`

	TenantID      string `gorm:"type:text;not null;default:tenant-default"`
	TransactionNo string `gorm:"type:varchar(50);unique;not null"` // "TRX-202608-00125"
	AccountID     string `gorm:"type:text;not null;index"`
	CategoryID    string `gorm:"type:text;not null;index"`
	Direction     string `gorm:"type:varchar(10);not null"` // IN | OUT

	Amount  float64 `gorm:"type:numeric(15,2);not null"`
	TrxDate time.Time `gorm:"not null;index"`

	// Polymorphic reference tanpa FK — konsisten dengan dokumen.
	// Fix review: composite index untuk query rekonsiliasi.
	SourceType string `gorm:"type:varchar(30);index:idx_cash_trx_source,priority:1"`
	SourceID   string `gorm:"type:text;index:idx_cash_trx_source,priority:2"`

	Description string `gorm:"type:text;not null"`
	RecordedBy  *uint  `gorm:"column:recorded_by"`

	CreatedAt time.Time
}

func (CashTransactionModel) TableName() string {
	return "cash_transactions"
}

func (m *CashTransactionModel) ToDomain() cashbook.CashTransaction {
	if m == nil {
		return cashbook.CashTransaction{}
	}
	return cashbook.CashTransaction{
		ID:            m.ID,
		TenantID:      m.TenantID,
		TransactionNo: m.TransactionNo,
		AccountID:     m.AccountID,
		CategoryID:    m.CategoryID,
		Direction:     m.Direction,
		Amount:        m.Amount,
		TrxDate:       m.TrxDate,
		SourceType:    m.SourceType,
		SourceID:      m.SourceID,
		Description:   m.Description,
		RecordedBy:    m.RecordedBy,
		CreatedAt:     m.CreatedAt,
	}
}

func CashTransactionModelFromDomain(t cashbook.CashTransaction) *CashTransactionModel {
	return &CashTransactionModel{
		ID:            t.ID,
		TenantID:      t.TenantID,
		TransactionNo: t.TransactionNo,
		AccountID:     t.AccountID,
		CategoryID:    t.CategoryID,
		Direction:     t.Direction,
		Amount:        t.Amount,
		TrxDate:       t.TrxDate,
		SourceType:    t.SourceType,
		SourceID:      t.SourceID,
		Description:   t.Description,
		RecordedBy:    t.RecordedBy,
		CreatedAt:     t.CreatedAt,
	}
}
