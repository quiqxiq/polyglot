package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainDebt "github.com/quixiq/polyglot/internal/domain/debt"
	"github.com/quixiq/polyglot/internal/port"
)

func seedBillingOpsCustomer(t *testing.T, db *gorm.DB) string {
	t.Helper()
	const id = "cust-billing-ops"
	err := db.Exec(`INSERT INTO customers (id, tenant_id, customer_code, name, phone, address, portal_access_code)
		VALUES (?, 'tenant-default', 'CUST-BO-1', 'Pelanggan Billing Ops', '081200000099', 'Jl. Uji', '99001122')`, id).Error
	require.NoError(t, err)
	return id
}

func TestBillingOpsRepositories(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	_, _, dsn := startMigratedPostgres(t)
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	ctx := context.Background()
	customerID := seedBillingOpsCustomer(t, db)

	t.Run("payment_method_crud", func(t *testing.T) {
		repo := postgres.NewPaymentMethodRepository(db)
		m := domainBilling.PaymentMethod{
			ID: "pm-test-bca", TenantID: "tenant-default", Name: "Transfer BCA",
			Type: "BANK", IsActive: true, BankName: "BCA",
			AccountNumber: "1234567890", AccountHolder: "PT Polyglot", SortOrder: 1,
		}
		require.NoError(t, repo.Save(ctx, m))

		got, err := repo.FindByID(ctx, "pm-test-bca")
		require.NoError(t, err)
		assert.Equal(t, "BCA", got.BankName)
		assert.Equal(t, "1234567890", got.AccountNumber)

		list, err := repo.List(ctx, "tenant-default", true)
		require.NoError(t, err)
		assert.NotEmpty(t, list)

		require.NoError(t, repo.SetActive(ctx, "pm-test-bca", false))
		got, err = repo.FindByID(ctx, "pm-test-bca")
		require.NoError(t, err)
		assert.False(t, got.IsActive)
	})

	t.Run("debt_payment_and_journal", func(t *testing.T) {
		repo := postgres.NewDebtRepository(db)
		d := domainDebt.Debt{
			ID: "debt-test-1", TenantID: "tenant-default", CustomerID: customerID,
			Amount: 150000, RemainingAmount: 150000, DebtDate: time.Now().UTC(),
			Status: domainDebt.StatusOpen, Description: "uji hutang",
		}
		require.NoError(t, repo.Save(ctx, d))

		pay, err := repo.ApplyPayment(ctx, port.DebtPaymentCommand{
			TenantID: "tenant-default", DebtID: d.ID, Amount: 50000,
			CashAccountID: "ca-1001-kas-kantor", IncomeCategoryID: "cc-debt-income",
		})
		require.NoError(t, err)
		assert.Equal(t, 50000.0, pay.Amount)
		require.NotNil(t, pay.CashTransactionID)

		var journalCount int64
		require.NoError(t, db.Table("cash_transactions").
			Where("source_type = ? AND source_id = ?", "DEBT_PAYMENT", d.ID).Count(&journalCount).Error)
		assert.Equal(t, int64(1), journalCount)

		got, err := repo.FindByID(ctx, d.ID)
		require.NoError(t, err)
		assert.Equal(t, domainDebt.StatusPartial, got.Status)
		assert.InDelta(t, 100000, got.RemainingAmount, 0.01)

		// Overpay ditolak dan tidak menambah jurnal.
		_, err = repo.ApplyPayment(ctx, port.DebtPaymentCommand{
			TenantID: "tenant-default", DebtID: d.ID, Amount: 200000,
			CashAccountID: "ca-1001-kas-kantor", IncomeCategoryID: "cc-debt-income",
		})
		require.ErrorIs(t, err, domainDebt.ErrOverpayment)
		require.NoError(t, db.Table("cash_transactions").
			Where("source_type = ? AND source_id = ?", "DEBT_PAYMENT", d.ID).Count(&journalCount).Error)
		assert.Equal(t, int64(1), journalCount)

		// Pelunasan.
		_, err = repo.ApplyPayment(ctx, port.DebtPaymentCommand{
			TenantID: "tenant-default", DebtID: d.ID, Amount: 100000,
			CashAccountID: "ca-1001-kas-kantor", IncomeCategoryID: "cc-debt-income",
		})
		require.NoError(t, err)
		got, err = repo.FindByID(ctx, d.ID)
		require.NoError(t, err)
		assert.Equal(t, domainDebt.StatusPaid, got.Status)
		assert.InDelta(t, 0, got.RemainingAmount, 0.01)
		assert.NotNil(t, got.PaidAt)

		summary, err := repo.Summary(ctx, "tenant-default", customerID, time.Now().UTC())
		require.NoError(t, err)
		assert.Equal(t, 1, summary.Count)
	})

	t.Run("bulk_payment_claim", func(t *testing.T) {
		invRepo := postgres.NewInvoiceRepository(db)
		due := time.Now().UTC().AddDate(0, 0, 7)
		for i, id := range []string{"inv-bulk-x", "inv-bulk-y"} {
			require.NoError(t, invRepo.Save(ctx, domainBilling.Invoice{
				ID: id, TenantID: "tenant-default", InvoiceNumber: "INV-BULK-" + id,
				CustomerID: customerID, Period: "2026-09", Subtotal: 1000, Total: 1000,
				DueDate: due, Status: domainBilling.StatusUnpaid,
				QRPayload: "qr-" + id, ManualPaymentCode: "PAY-" + id,
			}))
			_ = i
		}
		repo := postgres.NewBulkPaymentRepository(db)
		batch := domainBilling.BulkPaymentBatch{
			ID: "batch-test-1", TenantID: "tenant-default",
			InvoiceCount: 2, Status: domainBilling.BulkStatusQueued,
		}
		require.NoError(t, repo.SaveBatch(ctx, batch))
		require.NoError(t, repo.SaveItems(ctx, []domainBilling.BulkPaymentItem{
			{ID: "bpi-1", BatchID: batch.ID, InvoiceID: "inv-bulk-x", Amount: 1000, Status: domainBilling.BulkItemQueued},
			{ID: "bpi-2", BatchID: batch.ID, InvoiceID: "inv-bulk-y", Amount: 2000, Status: domainBilling.BulkItemQueued},
		}))

		item, found, err := repo.ClaimNextQueuedItem(ctx, batch.ID)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, domainBilling.BulkItemProcessing, item.Status)

		item.Status = domainBilling.BulkItemSucceeded
		require.NoError(t, repo.UpdateItem(ctx, item))

		queued, processing, succeeded, failed, err := repo.CountItems(ctx, batch.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, queued)
		assert.Equal(t, 0, processing)
		assert.Equal(t, 1, succeeded)
		assert.Equal(t, 0, failed)

		require.NoError(t, repo.UpdateBatchStatus(ctx, batch.ID, domainBilling.BulkStatusDone))
		got, err := repo.FindBatchByID(ctx, batch.ID)
		require.NoError(t, err)
		assert.Equal(t, domainBilling.BulkStatusDone, got.Status)
	})

	t.Run("gateway_settlement", func(t *testing.T) {
		repo := postgres.NewGatewaySettlementRepository(db)
		from := time.Now().UTC().AddDate(0, 0, -7)
		to := time.Now().UTC()
		s := domainBilling.GatewaySettlement{
			ID: "gws-test-1", TenantID: "tenant-default", Gateway: "TRIPAY",
			SettlementRef: "SETTLE-001", PeriodFrom: &from, PeriodTo: &to,
			GrossAmount: 1000000, FeeAmount: 15000, NetAmount: 985000,
			Status: domainBilling.SettlementStatusPending,
		}
		require.NoError(t, repo.Save(ctx, s))

		got, err := repo.FindByRef(ctx, "tenant-default", "TRIPAY", "SETTLE-001")
		require.NoError(t, err)
		assert.InDelta(t, 985000, got.NetAmount, 0.01)

		list, err := repo.List(ctx, port.GatewaySettlementFilter{TenantID: "tenant-default", Gateway: "TRIPAY"})
		require.NoError(t, err)
		assert.Len(t, list, 1)
	})
}

func TestInvoiceFilteredRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	_, _, dsn := startMigratedPostgres(t)
	gdb, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	ctx := context.Background()
	customerID := seedBillingOpsCustomer(t, gdb)

	repo := postgres.NewInvoiceRepository(gdb)
	now := time.Now().UTC()
	dueSoon := now.AddDate(0, 0, 3)
	overdue := now.AddDate(0, 0, -10)
	invoices := []domainBilling.Invoice{
		{ID: "inv-filter-1", TenantID: "tenant-default", InvoiceNumber: "INV-F-1", CustomerID: customerID,
			Period: "2026-09", InvoiceType: domainBilling.InvoiceTypeRecurring, Subtotal: 100000, Total: 100000,
			DueDate: dueSoon, Status: domainBilling.StatusUnpaid, QRPayload: "qr-f-1", ManualPaymentCode: "PAY-F-1"},
		{ID: "inv-filter-2", TenantID: "tenant-default", InvoiceNumber: "INV-F-2", CustomerID: customerID,
			Period: "2026-08", InvoiceType: domainBilling.InvoiceTypeRecurring, Subtotal: 110000, Total: 110000,
			DueDate: overdue, Status: domainBilling.StatusUnpaid, QRPayload: "qr-f-2", ManualPaymentCode: "PAY-F-2"},
		{ID: "inv-filter-3", TenantID: "tenant-default", InvoiceNumber: "INV-F-3", CustomerID: customerID,
			Period: "2026-08", InvoiceType: domainBilling.InvoiceTypeRecurring, Subtotal: 90000, Total: 90000,
			PaidAmount: 90000, DueDate: overdue, Status: domainBilling.StatusPaid, QRPayload: "qr-f-3", ManualPaymentCode: "PAY-F-3"},
	}
	for _, inv := range invoices {
		require.NoError(t, repo.Save(ctx, inv))
	}

	got, total, err := repo.FindFiltered(ctx, port.InvoiceFilter{TenantID: "tenant-default", OverdueOnly: true})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, got, 1)
	assert.Equal(t, "inv-filter-2", got[0].ID)

	got, total, err = repo.FindFiltered(ctx, port.InvoiceFilter{TenantID: "tenant-default", Status: domainBilling.StatusPaid})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, got, 1)

	summary, err := repo.Summary(ctx, port.InvoiceFilter{TenantID: "tenant-default"})
	require.NoError(t, err)
	assert.Equal(t, 3, summary.Count)
	assert.InDelta(t, 300000, summary.TotalBilled, 0.01)
	assert.InDelta(t, 90000, summary.TotalPaid, 0.01)
	assert.InDelta(t, 210000, summary.TotalOutstanding, 0.01)
	assert.InDelta(t, 110000, summary.TotalOverdue, 0.01)

	candidates, err := repo.FindOverdueCandidates(ctx, "tenant-default", now, 10)
	require.NoError(t, err)
	require.Len(t, candidates, 1)
	assert.Equal(t, "inv-filter-2", candidates[0].ID)
}
