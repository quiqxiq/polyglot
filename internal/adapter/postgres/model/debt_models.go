package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/debt"
)

// DebtModel is the GORM model for customer_debts table.
type DebtModel struct {
	ID              string     `gorm:"primaryKey"`
	TenantID        string     `gorm:"type:text;not null;default:tenant-default"`
	CustomerID      string     `gorm:"type:text;not null;index"`
	SubscriptionID  *string    `gorm:"column:subscription_id;type:text"`
	Amount          float64    `gorm:"type:numeric(15,2);not null"`
	RemainingAmount float64    `gorm:"type:numeric(15,2);not null"`
	Description     string     `gorm:"type:text"`
	DebtDate        time.Time  `gorm:"type:date;not null"`
	DueDate         *time.Time `gorm:"type:date;index"`
	Status          string     `gorm:"type:varchar(20);not null;default:OPEN;index"`
	PaidAt          *time.Time
	CreatedBy       *uint `gorm:"column:created_by"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TableName returns the table name for customer_debts.
func (DebtModel) TableName() string {
	return "customer_debts"
}

// ToDomain converts DebtModel to domain debt.Debt.
func (m *DebtModel) ToDomain() debt.Debt {
	if m == nil {
		return debt.Debt{}
	}
	return debt.Debt{
		ID:              m.ID,
		TenantID:        m.TenantID,
		CustomerID:      m.CustomerID,
		SubscriptionID:  m.SubscriptionID,
		Amount:          m.Amount,
		RemainingAmount: m.RemainingAmount,
		Description:     m.Description,
		DebtDate:        m.DebtDate,
		DueDate:         m.DueDate,
		Status:          m.Status,
		PaidAt:          m.PaidAt,
		CreatedBy:       m.CreatedBy,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

// DebtModelFromDomain converts domain debt.Debt to DebtModel.
func DebtModelFromDomain(d debt.Debt) *DebtModel {
	return &DebtModel{
		ID:              d.ID,
		TenantID:        d.TenantID,
		CustomerID:      d.CustomerID,
		SubscriptionID:  d.SubscriptionID,
		Amount:          d.Amount,
		RemainingAmount: d.RemainingAmount,
		Description:     d.Description,
		DebtDate:        d.DebtDate,
		DueDate:         d.DueDate,
		Status:          d.Status,
		PaidAt:          d.PaidAt,
		CreatedBy:       d.CreatedBy,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
}

// DebtPaymentModel is the GORM model for debt_payments table.
type DebtPaymentModel struct {
	ID                string  `gorm:"primaryKey"`
	TenantID          string  `gorm:"type:text;not null;default:tenant-default"`
	DebtID            string  `gorm:"type:text;not null;index"`
	Amount            float64 `gorm:"type:numeric(15,2);not null"`
	PaymentMethodID   *string `gorm:"column:payment_method_id;type:text"`
	CashTransactionID *string `gorm:"column:cash_transaction_id;type:text"`
	PaymentDate       time.Time
	Notes             string `gorm:"type:text"`
	RecordedBy        *uint  `gorm:"column:recorded_by"`
	CreatedAt         time.Time
}

// TableName returns the table name for debt_payments.
func (DebtPaymentModel) TableName() string {
	return "debt_payments"
}

// ToDomain converts DebtPaymentModel to domain debt.Payment.
func (m *DebtPaymentModel) ToDomain() debt.Payment {
	if m == nil {
		return debt.Payment{}
	}
	return debt.Payment{
		ID:                m.ID,
		TenantID:          m.TenantID,
		DebtID:            m.DebtID,
		Amount:            m.Amount,
		PaymentMethodID:   m.PaymentMethodID,
		CashTransactionID: m.CashTransactionID,
		PaymentDate:       m.PaymentDate,
		Notes:             m.Notes,
		RecordedBy:        m.RecordedBy,
		CreatedAt:         m.CreatedAt,
	}
}

// DebtPaymentModelFromDomain converts domain debt.Payment to DebtPaymentModel.
func DebtPaymentModelFromDomain(p debt.Payment) *DebtPaymentModel {
	return &DebtPaymentModel{
		ID:                p.ID,
		TenantID:          p.TenantID,
		DebtID:            p.DebtID,
		Amount:            p.Amount,
		PaymentMethodID:   p.PaymentMethodID,
		CashTransactionID: p.CashTransactionID,
		PaymentDate:       p.PaymentDate,
		Notes:             p.Notes,
		RecordedBy:        p.RecordedBy,
		CreatedAt:         p.CreatedAt,
	}
}
