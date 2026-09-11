// ConversionWriter tests (F1-12) — membuktikan SaveConversion benar-benar
// atomik: seluruh artefak tersimpan bersama, dan kegagalan di tengah
// me-rollback semuanya.
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
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

func setupConversionDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.CustomerModel{}, &model.SubscriptionModel{},
		&model.InvoiceModel{}, &model.InvoiceItemModel{},
		&model.RegistrationModel{},
	))
	return db
}

func conversionArtifacts() port.ConversionArtifacts {
	subID := "sub-conv"
	return port.ConversionArtifacts{
		Registration: domainRegistration.Registration{
			ID: "reg-conv", RegistrationNo: "REG-202609-9001", PlanID: "plan-1",
			FullName: "Budi", Phone: "085606846141", Address: "Jl. Test",
			Status:     domainRegistration.StatusActive,
			CustomerID: "cust-conv", SubscriptionID: "sub-conv", InvoiceID: "inv-conv",
		},
		Customer: domainCustomer.Customer{
			ID: "cust-conv", TenantID: "tenant-default", CustomerCode: "CUST-90001",
			Name: "Budi", Phone: "085606846141", Address: "Jl. Test",
			PortalAccessCode: "90000001", Status: domainCustomer.StatusActive,
		},
		Subscription: domainSubscription.Subscription{
			ID: "sub-conv", TenantID: "tenant-default", CustomerID: "cust-conv",
			PlanID: "plan-1", ServiceType: "PPPOE", RemoteUsername: "bs9001",
			RemotePassword: "secret99", Status: domainSubscription.StatusActive,
		},
		Invoice: domainBilling.Invoice{
			ID: "inv-conv", TenantID: "tenant-default", InvoiceNumber: "INV-202609-9001",
			CustomerID: "cust-conv", SubscriptionID: &subID, Period: "2026-09",
			Total: 110000, Status: domainBilling.StatusUnpaid,
			QRPayload: "polyglot://invoice/inv-conv", ManualPaymentCode: "PAY-900001",
		},
		Items: []domainBilling.InvoiceItem{{
			ID: "itm-conv", InvoiceID: "inv-conv", Description: "Paket",
			Quantity: 1, UnitPrice: 100000, Amount: 100000,
			ItemType: domainBilling.ItemTypeSubscriptionFee,
		}},
	}
}

var conversionTables = []string{
	"customers", "subscriptions", "invoices", "invoice_items", "registrations",
}

func TestConversionWriter_SavesAllArtifactsAtomically(t *testing.T) {
	db := setupConversionDB(t)
	w := postgres.NewConversionWriter(db, fakeVaultForTest{})

	require.NoError(t, w.SaveConversion(context.Background(), conversionArtifacts()))

	for _, table := range conversionTables {
		var n int64
		require.NoError(t, db.Table(table).Count(&n).Error, table)
		assert.Equal(t, int64(1), n, "tabel %s harus berisi 1 baris", table)
	}

	// Password subscription harus tersimpan terenkripsi.
	var stored model.SubscriptionModel
	require.NoError(t, db.First(&stored, "id = ?", "sub-conv").Error)
	assert.Equal(t, "enc:secret99", stored.RemotePassword)
}

func TestConversionWriter_RollsBackOnFailure(t *testing.T) {
	db := setupConversionDB(t)
	w := postgres.NewConversionWriter(db, fakeVaultForTest{})

	// Trigger SQLite menolak INSERT invoice → transaksi harus rollback penuh.
	require.NoError(t, db.Exec(
		`CREATE TRIGGER fail_invoice BEFORE INSERT ON invoices BEGIN SELECT RAISE(ABORT, 'boom'); END;`,
	).Error)

	err := w.SaveConversion(context.Background(), conversionArtifacts())
	require.Error(t, err)

	for _, table := range conversionTables {
		var n int64
		require.NoError(t, db.Table(table).Count(&n).Error, table)
		assert.Equal(t, int64(0), n, "tabel %s harus kosong setelah rollback", table)
	}
}
