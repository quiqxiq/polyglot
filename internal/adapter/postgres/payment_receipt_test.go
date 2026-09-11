// F3-4: bukti pembayaran WA dirender dari template PAYMENT_RECEIPT dengan
// fallback teks bawaan bila template tidak ada.
package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/port"
)

func TestPaymentProcessor_ReceiptUsesTemplate(t *testing.T) {
	db := setupISPDB(t)
	ctx := context.Background()

	require.NoError(t, db.Create(&model.CustomerModel{
		ID: "cust-rt", TenantID: "tenant-default", CustomerCode: "CUST-RT",
		Name: "Budi Santoso", Phone: "081200000000", Address: "Jl. RT",
		PortalAccessCode: "99998888", Status: domainCustomer.StatusActive,
	}).Error)
	subID := "sub-rt"
	require.NoError(t, db.Create(&model.InvoiceModel{
		ID: "inv-rt", TenantID: "tenant-default", InvoiceNumber: "INV-RT-1",
		CustomerID: "cust-rt", SubscriptionID: &subID, Period: "2026-09",
		Total: 110000, DueDate: timeDate("2026-09-20"), Status: domainBilling.StatusUnpaid,
		QRPayload: "polyglot://invoice/inv-rt", ManualPaymentCode: "PAY-RT1",
	}).Error)
	require.NoError(t, db.Create(&model.CashAccountModel{
		ID: "ca-1", TenantID: "tenant-default", AccountCode: "1001", Name: "Kas", Type: "CASH", IsActive: true,
	}).Error)
	require.NoError(t, db.Create(&model.CashCategoryModel{
		ID: "cc-tagihan", TenantID: "tenant-default", Name: "Tagihan", Type: "INCOME", IsActive: true,
	}).Error)
	require.NoError(t, db.Create(&model.NotificationTemplateModel{
		ID: "nt-receipt", TenantID: "tenant-default", TemplateKey: "PAYMENT_RECEIPT",
		Name: "Bukti", Content: "Halo {{customer_name}}, pembayaran Rp{{amount}} periode {{period}} diterima.",
		IsActive: true,
	}).Error)

	proc := postgres.NewPaymentProcessor(db)
	_, err := proc.ProcessCashPayment(ctx, port.CashPaymentCommand{
		TenantID: "tenant-default", InvoiceID: "inv-rt", Amount: 110000,
		CashAccountID: "ca-1", IncomeCategoryID: "cc-tagihan",
	})
	require.NoError(t, err)

	var wa model.WANotificationModel
	require.NoError(t, db.First(&wa, "invoice_id = ?", "inv-rt").Error)
	assert.Contains(t, wa.MessageContent, "Halo Budi Santoso")
	assert.Contains(t, wa.MessageContent, "Rp110000")
	assert.Contains(t, wa.MessageContent, "2026-09")
}

func TestPaymentProcessor_ReceiptFallsBackWithoutTemplate(t *testing.T) {
	db := setupISPDB(t)
	ctx := context.Background()

	require.NoError(t, db.Create(&model.CustomerModel{
		ID: "cust-fb", TenantID: "tenant-default", CustomerCode: "CUST-FB",
		Name: "Fallback", Phone: "081200000001", Address: "Jl. FB",
		PortalAccessCode: "77776666", Status: domainCustomer.StatusActive,
	}).Error)
	require.NoError(t, db.Create(&model.InvoiceModel{
		ID: "inv-fb", TenantID: "tenant-default", InvoiceNumber: "INV-FB-1",
		CustomerID: "cust-fb", Period: "2026-09", Total: 50000,
		DueDate: timeDate("2026-09-20"), Status: domainBilling.StatusUnpaid,
		QRPayload: "polyglot://invoice/inv-fb", ManualPaymentCode: "PAY-FB1",
	}).Error)
	require.NoError(t, db.Create(&model.CashAccountModel{
		ID: "ca-1", TenantID: "tenant-default", AccountCode: "1001", Name: "Kas", Type: "CASH", IsActive: true,
	}).Error)
	require.NoError(t, db.Create(&model.CashCategoryModel{
		ID: "cc-tagihan", TenantID: "tenant-default", Name: "Tagihan", Type: "INCOME", IsActive: true,
	}).Error)

	proc := postgres.NewPaymentProcessor(db)
	_, err := proc.ProcessCashPayment(ctx, port.CashPaymentCommand{
		TenantID: "tenant-default", InvoiceID: "inv-fb", Amount: 50000,
		CashAccountID: "ca-1", IncomeCategoryID: "cc-tagihan",
	})
	require.NoError(t, err)

	var wa model.WANotificationModel
	require.NoError(t, db.First(&wa, "invoice_id = ?", "inv-fb").Error)
	assert.Contains(t, wa.MessageContent, "Terima kasih")
}
