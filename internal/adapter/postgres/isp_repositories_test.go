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
	domainCashbook "github.com/quixiq/polyglot/internal/domain/cashbook"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainReporting "github.com/quixiq/polyglot/internal/domain/reporting"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

// setupISPDB menyiapkan skema ISP penuh via AutoMigrate (konvensi
// device_repository_test.go: sqlite in-memory untuk repo CRUD).
func setupISPDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(
		&model.ServicePlanModel{},
		&model.CustomerModel{},
		&model.SubscriptionModel{},
		&model.InvoiceModel{},
		&model.InvoiceItemModel{},
		&model.PaymentMethodModel{},
		&model.PaymentModel{},
		&model.GatewayTransactionModel{},
		&model.CashAccountModel{},
		&model.CashCategoryModel{},
		&model.CashTransactionModel{},
		&model.NotificationTemplateModel{},
		&model.WANotificationModel{},
		&model.DailySnapshotModel{},
		&model.PortalSessionModel{},
		&model.PortalOTPModel{},
		&model.AuditLogModel{},
	)
	require.NoError(t, err)
	return db
}

// fakeVaultForTest dideklarasikan di isp_testutil_test.go.

func TestServicePlanRepository_CRUD(t *testing.T) {
	db := setupISPDB(t)
	repo := postgres.NewServicePlanRepository(db)
	ctx := context.Background()

	p := domainPlan.ServicePlan{
		ID: "plan-1", TenantID: "tenant-default", Name: "100-RB-100",
		ServiceType:           domainPlan.TypePPPoE,
		BandwidthDownloadKbps: 5120, BandwidthUploadKbps: 5120,
		Price: 100000, TaxPercent: 10, IsActive: true,
		SharedUsers: 1,
		ParentQueue: "none",
	}
	require.NoError(t, repo.Save(ctx, p))

	got, err := repo.FindByID(ctx, "plan-1")
	require.NoError(t, err)
	assert.Equal(t, "100-RB-100", got.Name)
	assert.InDelta(t, 100000, got.Price, 0.01)

	byName, err := repo.FindByName(ctx, "tenant-default", "100-RB-100")
	require.NoError(t, err)
	assert.Equal(t, p.ID, byName.ID)

	list, err := repo.List(ctx, true)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	require.NoError(t, repo.Delete(ctx, p.ID))
	_, err = repo.FindByID(ctx, p.ID)
	assert.ErrorIs(t, err, postgres.ErrNotFound)
}

func TestCustomerLookups_AndSoftFilter(t *testing.T) {
	db := setupISPDB(t)
	repo := postgres.NewCustomerRepository(db)
	ctx := context.Background()

	c := domainCustomer.Customer{
		ID: "cust-1", TenantID: "tenant-default",
		CustomerCode: "CUST-01075", Name: "MATRAJI-KT",
		Phone: "085606846141", Address: "KATAPANG",
		PortalAccessCode: "12345678", Status: domainCustomer.StatusActive,
	}
	require.NoError(t, repo.Save(ctx, c))

	got, err := repo.FindByPortalAccessCode(ctx, "12345678")
	require.NoError(t, err)
	assert.Equal(t, c.ID, got.ID)

	byCode, err := repo.FindByCustomerCode(ctx, "CUST-01075")
	require.NoError(t, err)
	assert.Equal(t, c.ID, byCode.ID)

	byPhone, err := repo.FindByPhone(ctx, "085606846141")
	require.NoError(t, err)
	assert.Equal(t, c.ID, byPhone.ID)

	_, err = repo.FindByPortalAccessCode(ctx, "99999999")
	assert.ErrorIs(t, err, postgres.ErrNotFound)
}

func TestInvoiceRepo_SaveWithItems_AndCashierLookups(t *testing.T) {
	db := setupISPDB(t)
	repo := postgres.NewInvoiceRepository(db)
	ctx := context.Background()

	subID := "sub-1"
	inv := domainBilling.Invoice{
		ID: "inv-1", TenantID: "tenant-default", InvoiceNumber: "INV-202608-00001",
		CustomerID: "cust-1", SubscriptionID: &subID, Period: "2026-08",
		Subtotal: 100000, TaxAmount: 10000, Total: 110000,
		DueDate: timeDate("2026-08-25"), Status: domainBilling.StatusUnpaid,
		QRPayload: "polyglot://invoice/inv-1", ManualPaymentCode: "PAY-892147",
	}
	items := []domainBilling.InvoiceItem{{
		ID: "itm-1", InvoiceID: inv.ID, Description: "Paket 100-RB-100",
		Quantity: 1, UnitPrice: 100000, Amount: 100000,
		ItemType: domainBilling.ItemTypeSubscriptionFee,
	}}
	require.NoError(t, repo.SaveWithItems(ctx, inv, items))

	byItems, err := repo.FindByID(ctx, "inv-1")
	require.NoError(t, err)
	assert.InDelta(t, 110000, byItems.Total, 0.01)

	byPeriod, err := repo.FindBySubscriptionPeriod(ctx, subID, "2026-08")
	require.NoError(t, err)
	assert.Equal(t, inv.ID, byPeriod.ID)

	_, err = repo.FindBySubscriptionPeriod(ctx, subID, "2026-09")
	assert.ErrorIs(t, err, postgres.ErrNotFound)

	byPayCode, err := repo.FindByPaymentCode(ctx, "PAY-892147")
	require.NoError(t, err)
	assert.Equal(t, inv.ID, byPayCode.ID)

	byQR, err := repo.FindByQRPayload(ctx, "polyglot://invoice/inv-1")
	require.NoError(t, err)
	assert.Equal(t, inv.ID, byQR.ID)
}

func TestSubscriptionRepo_VaultRoundTrip_AndLookups(t *testing.T) {
	db := setupISPDB(t)
	vault := fakeVaultForTest{}
	repo := postgres.NewSubscriptionRepository(db, vault)
	ctx := context.Background()

	deviceID := "dev-router-uuid"
	sub := domainSubscription.Subscription{
		ID: "sub-1", TenantID: "tenant-default",
		CustomerID: "cust-1", PlanID: "plan-1",
		DeviceID: &deviceID, ServiceType: "PPPOE",
		RemoteUsername: "MATRAJI-KT", RemotePassword: "rahasia123",
		BillingCycle: domainSubscription.CycleMonthly, BillingDay: 25,
		Status: domainSubscription.StatusActive,
	}
	require.NoError(t, repo.Save(ctx, sub))

	// Round-trip: plaintext kembali setelah disimpan sebagai ciphertext.
	got, err := repo.FindByID(ctx, "sub-1")
	require.NoError(t, err)
	assert.Equal(t, "rahasia123", got.RemotePassword)

	// Kolom DB berisi ciphertext, bukan plaintext.
	var raw model.SubscriptionModel
	require.NoError(t, db.First(&raw, "id = ?", "sub-1").Error)
	assert.NotEqual(t, "rahasia123", raw.RemotePassword)
	assert.Contains(t, raw.RemotePassword, "enc:")

	byRouter, err := repo.FindByDeviceAndUsername(ctx, deviceID, "MATRAJI-KT")
	require.NoError(t, err)
	assert.Equal(t, sub.ID, byRouter.ID)

	active, err := repo.ListActive(ctx)
	require.NoError(t, err)
	assert.Len(t, active, 1)
}

func TestPaymentProcessor_HappyPath_AllFourArtifacts(t *testing.T) {
	db := setupISPDB(t)
	ctx := context.Background()

	seedPaidFixture(t, db)
	proc := postgres.NewPaymentProcessor(db)

	cashier := uint(5)
	pay, err := proc.ProcessCashPayment(ctx, port.CashPaymentCommand{
		TenantID:         "tenant-default",
		InvoiceID:        "inv-pay",
		Amount:           110000,
		CashAccountID:    "ca-1",
		IncomeCategoryID: "cc-1",
		ReceivedBy:       &cashier,
		ScanMethod:       domainBilling.ScanCodeInput,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, pay.ID)

	// 1. invoice PAID.
	var inv model.InvoiceModel
	require.NoError(t, db.First(&inv, "id = ?", "inv-pay").Error)
	assert.Equal(t, domainBilling.StatusPaid, inv.Status)
	assert.NotNil(t, inv.PaidAt)
	assert.InDelta(t, 110000, inv.PaidAmount, 0.01)

	// 2. payment row.
	var payCount int64
	require.NoError(t, db.Model(&model.PaymentModel{}).Where("invoice_id = ?", "inv-pay").Count(&payCount).Error)
	assert.Equal(t, int64(1), payCount)

	// 3. jurnal kas IN.
	var cash model.CashTransactionModel
	require.NoError(t, db.First(&cash, "source_type = ? AND source_id = ?", domainCashbook.SourcePayment, pay.ID).Error)
	assert.Equal(t, domainCashbook.DirectionIn, cash.Direction)
	assert.InDelta(t, 110000, cash.Amount, 0.01)

	// 4. WA receipt queued.
	var wa model.WANotificationModel
	require.NoError(t, db.First(&wa, "invoice_id = ?", "inv-pay").Error)
	assert.Equal(t, domainNotification.StatusQueued, wa.Status)
	assert.Equal(t, "081234567890", wa.RecipientPhone)

	// Guard idempoten: bayar lagi ditolak.
	_, err = proc.ProcessCashPayment(ctx, port.CashPaymentCommand{
		InvoiceID: "inv-pay", Amount: 1000, CashAccountID: "ca-1", IncomeCategoryID: "cc-1",
	})
	assert.ErrorIs(t, err, domainBilling.ErrInvoiceAlreadyPaid)

	// Guard overpayment pada invoice lain.
	_, err = proc.ProcessCashPayment(ctx, port.CashPaymentCommand{
		TenantID: "tenant-default", InvoiceID: "inv-partial", Amount: 500000,
		CashAccountID: "ca-1", IncomeCategoryID: "cc-1",
	})
	assert.ErrorIs(t, err, domainBilling.ErrOverpayment)
}

func TestCashbookRepository_Filters_And_Balance(t *testing.T) {
	db := setupISPDB(t)
	repo := postgres.NewCashbookRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.SaveAccount(ctx, domainCashbook.CashAccount{ID: "ca-1", Name: "Kas Kantor", Type: domainCashbook.AccountTypeCash, IsActive: true}))
	require.NoError(t, repo.SaveCategory(ctx, domainCashbook.CashCategory{ID: "cc-in", Name: "Tagihan", Type: domainCashbook.CategoryTypeIncome, IsActive: true}))
	require.NoError(t, repo.SaveCategory(ctx, domainCashbook.CashCategory{ID: "cc-out", Name: "Listrik", Type: domainCashbook.CategoryTypeExpense, IsActive: true}))

	mk := func(id, dir, cat string, amount float64, day int) domainCashbook.CashTransaction {
		return domainCashbook.CashTransaction{
			ID: id, TransactionNo: "TRX-" + id, AccountID: "ca-1", CategoryID: cat,
			Direction: dir, Amount: amount, TrxDate: timeDate(dayStamp(day)),
			Description: "trx " + id,
		}
	}
	require.NoError(t, repo.SaveTransaction(ctx, mk("t1", domainCashbook.DirectionIn, "cc-in", 200000, 1)))
	require.NoError(t, repo.SaveTransaction(ctx, mk("t2", domainCashbook.DirectionOut, "cc-out", 50000, 2)))
	require.NoError(t, repo.SaveTransaction(ctx, mk("t3", domainCashbook.DirectionIn, "cc-in", 300000, 3)))

	balances, err := repo.BalanceByAccounts(ctx, port.CashTransactionFilter{})
	require.NoError(t, err)
	assert.InDelta(t, 450000, balances["ca-1"], 0.01)

	inOnly, err := repo.FindTransactions(ctx, port.CashTransactionFilter{Direction: domainCashbook.DirectionIn})
	require.NoError(t, err)
	assert.Len(t, inOnly, 2)

	from := timeDate(dayStamp(2))
	to := timeDate(dayStamp(3))
	ranged, err := repo.FindTransactions(ctx, port.CashTransactionFilter{From: from, To: to})
	require.NoError(t, err)
	assert.Len(t, ranged, 2)
}

func TestNotificationAndReporting_Repos(t *testing.T) {
	db := setupISPDB(t)
	notif := postgres.NewNotificationRepository(db)
	reporting := postgres.NewReportingRepository(db)
	ctx := context.Background()

	tpl := domainNotification.NotificationTemplate{
		ID: "nt-1", TemplateKey: "BILL_REMINDER", Name: "Tagihan",
		Content: "Yth {{customer_name}}", VariablesJSON: `["customer_name"]`, IsActive: true,
	}
	require.NoError(t, notif.SaveTemplate(ctx, tpl))
	gotTpl, err := notif.FindTemplateByKey(ctx, "tenant-default", "BILL_REMINDER")
	require.NoError(t, err)
	assert.Equal(t, tpl.ID, gotTpl.ID)

	custID := "cust-wa"
	invID := "inv-wa"
	require.NoError(t, notif.Queue(ctx, domainNotification.WANotification{
		ID: "wa-1", TenantID: "tenant-default", CustomerID: &custID, InvoiceID: &invID,
		RecipientPhone: "08111", MessageType: "BILL_REMINDER", MessageContent: "x",
	}))
	pending, err := notif.Pending(ctx, 10)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	require.NoError(t, notif.MarkSent(ctx, "wa-1", timeNow()))
	pending2, _ := notif.Pending(ctx, 10)
	assert.Empty(t, pending2)

	// Snapshot upsert: dua kali tanggal sama = satu baris ter-update.
	snap := domainReporting.DailyFinancialSnapshot{
		TenantID: "tenant-default", SnapshotDate: timeDate("2026-08-23"),
		PaymentCount: 3, PaymentTotal: 330000, CashBalanceJSON: []byte(`{"ca-1":450000}`),
	}
	require.NoError(t, reporting.UpsertSnapshot(ctx, snap))
	snap.PaymentTotal = 340000
	require.NoError(t, reporting.UpsertSnapshot(ctx, snap))

	rangeOut, err := reporting.ListRange(ctx, "tenant-default", timeDate("2026-08-01"), timeDate("2026-08-31"))
	require.NoError(t, err)
	require.Len(t, rangeOut, 1)
	assert.InDelta(t, 340000, rangeOut[0].PaymentTotal, 0.01)
}

func TestSubscription_ProvisionFields_RoundTrip(t *testing.T) {
	db := setupISPDB(t)
	repo := postgres.NewSubscriptionRepository(db, fakeVaultForTest{})
	ctx := context.Background()

	deviceID := "dev-x"
	sub := domainSubscription.Subscription{
		ID: "sub-prov", TenantID: "tenant-default",
		CustomerID: "c1", PlanID: "p1",
		DeviceID: &deviceID, ServiceType: "PPPOE",
		RemoteUsername: "U", RemotePassword: "pw",
		RouterProfile:   "HOME-20M",
		ProvisionStatus: domainSubscription.ProvisionPending,
		Status:          domainSubscription.StatusActive,
	}
	require.NoError(t, repo.Save(ctx, sub))

	got, err := repo.FindByID(ctx, sub.ID)
	require.NoError(t, err)
	assert.Equal(t, "HOME-20M", got.RouterProfile)
	assert.Equal(t, domainSubscription.ProvisionPending, got.ProvisionStatus)

	got.ProvisionStatus = domainSubscription.ProvisionOK
	require.NoError(t, repo.Save(ctx, got))
	reloaded, _ := repo.FindByID(ctx, sub.ID)
	assert.Equal(t, domainSubscription.ProvisionOK, reloaded.ProvisionStatus)
}

func TestPaymentProcessor_OnPaidHook_Fired(t *testing.T) {
	db := setupISPDB(t)
	ctx := context.Background()
	seedPaidFixture(t, db)

	var fired int
	proc := postgres.NewPaymentProcessor(db)
	proc.OnPaid = func(_ context.Context, inv domainBilling.Invoice, pay domainBilling.Payment) {
		fired++
		assert.Equal(t, "inv-pay", inv.ID)
		assert.NotEmpty(t, pay.ID)
	}

	_, err := proc.ProcessCashPayment(ctx, port.CashPaymentCommand{
		TenantID: "tenant-default", InvoiceID: "inv-pay", Amount: 110000,
		CashAccountID: "ca-1", IncomeCategoryID: "cc-1",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, fired)

	// Gagal pembayaran → hook tidak menyala.
	_, err = proc.ProcessCashPayment(ctx, port.CashPaymentCommand{
		TenantID: "tenant-default", InvoiceID: "inv-partial", Amount: 999999,
		CashAccountID: "ca-1", IncomeCategoryID: "cc-1",
	})
	require.Error(t, err)
	assert.Equal(t, 1, fired)
}
