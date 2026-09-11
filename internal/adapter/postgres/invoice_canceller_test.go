// F3-6: pembatalan invoice + jurnal koreksi kas harus atomik.
package postgres_test

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
)

func setupCancellerDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.InvoiceModel{}, &model.CashTransactionModel{},
		&model.CashAccountModel{}, &model.CashCategoryModel{},
	))
	require.NoError(t, db.Create(&model.CashAccountModel{
		ID: "ca-1", TenantID: "tenant-default", AccountCode: "1001", Name: "Kas", Type: "CASH", IsActive: true,
	}).Error)
	require.NoError(t, db.Create(&model.CashCategoryModel{
		ID: "cc-refund", TenantID: "tenant-default", Name: "Refund/Koreksi", Type: "EXPENSE", IsActive: true,
	}).Error)
	return db
}

func seedCancellableInvoice(t *testing.T, db *gorm.DB, id string, paid float64, status string) {
	t.Helper()
	subID := "sub-1"
	require.NoError(t, db.Create(&model.InvoiceModel{
		ID: id, TenantID: "tenant-default", InvoiceNumber: "INV-CAN-" + id,
		CustomerID: "cust-1", SubscriptionID: &subID, Period: "2026-09",
		Total: 110000, PaidAmount: paid, DueDate: timeDate("2026-09-20"),
		Status:            status,
		QRPayload:         "polyglot://invoice/" + id,
		ManualPaymentCode: "PAY-" + id,
	}).Error)
}

func TestInvoiceCanceller_PartialReversalJournal(t *testing.T) {
	db := setupCancellerDB(t)
	c := postgres.NewInvoiceCanceller(db)
	seedCancellableInvoice(t, db, "inv-partial", 50000, domainBilling.StatusPartial)

	inv, err := c.Cancel(context.Background(), port.CancelInvoiceCommand{
		InvoiceID: "inv-partial", Reason: "salah tagih",
		CashAccountID: "ca-1", ExpenseCategoryID: "cc-refund",
	})
	require.NoError(t, err)
	assert.Equal(t, domainBilling.StatusCancelled, inv.Status)
	assert.Equal(t, "salah tagih", inv.CancelReason)

	var trxCount int64
	require.NoError(t, db.Model(&model.CashTransactionModel{}).Count(&trxCount).Error)
	assert.Equal(t, int64(1), trxCount)

	var trx model.CashTransactionModel
	require.NoError(t, db.First(&trx).Error)
	assert.Equal(t, "OUT", trx.Direction)
	assert.Equal(t, "ca-1", trx.AccountID)
	assert.Equal(t, "cc-refund", trx.CategoryID)
	assert.InDelta(t, 50000, trx.Amount, 0.01)
}

func TestInvoiceCanceller_PaidInvoiceRejected(t *testing.T) {
	db := setupCancellerDB(t)
	c := postgres.NewInvoiceCanceller(db)
	seedCancellableInvoice(t, db, "inv-paid", 110000, domainBilling.StatusPaid)

	_, err := c.Cancel(context.Background(), port.CancelInvoiceCommand{
		InvoiceID: "inv-paid", Reason: "x", CashAccountID: "ca-1", ExpenseCategoryID: "cc-refund",
	})
	assert.ErrorIs(t, err, domainBilling.ErrInvoiceAlreadyPaid)

	var trxCount int64
	require.NoError(t, db.Model(&model.CashTransactionModel{}).Count(&trxCount).Error)
	assert.Equal(t, int64(0), trxCount)
}

func TestInvoiceCanceller_UnpaidNoReversal(t *testing.T) {
	db := setupCancellerDB(t)
	c := postgres.NewInvoiceCanceller(db)
	seedCancellableInvoice(t, db, "inv-unpaid", 0, domainBilling.StatusUnpaid)

	inv, err := c.Cancel(context.Background(), port.CancelInvoiceCommand{
		InvoiceID: "inv-unpaid", Reason: "pelanggan pindah",
		CashAccountID: "ca-1", ExpenseCategoryID: "cc-refund",
	})
	require.NoError(t, err)
	assert.Equal(t, domainBilling.StatusCancelled, inv.Status)

	var trxCount int64
	require.NoError(t, db.Model(&model.CashTransactionModel{}).Count(&trxCount).Error)
	assert.Equal(t, int64(0), trxCount)
}
