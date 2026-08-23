package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainDevice "github.com/quixiq/polyglot/internal/domain/device"
)

// fakeVaultForTest implements port.CredentialVault penuh dengan kripto
// prefix-based deterministik (cukup untuk round-trip test).
type fakeVaultForTest struct{}

func (fakeVaultForTest) EncryptString(_ context.Context, p string) (string, error) {
	return "enc:" + p, nil
}

func (fakeVaultForTest) DecryptString(_ context.Context, c string) (string, error) {
	if len(c) > 4 && c[:4] == "enc:" {
		return c[4:], nil
	}
	return "", errors.New("not encrypted")
}

func (fakeVaultForTest) Get(_ context.Context, _ string) (domainDevice.Credentials, error) {
	return domainDevice.Credentials{}, domainDevice.ErrNotFound
}

func (fakeVaultForTest) Save(_ context.Context, _ string, _ domainDevice.Credentials) error {
	return nil
}

func timeDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func timeNow() time.Time { return time.Now() }

func dayStamp(offset int) string {
	return time.Now().AddDate(0, 0, offset).Format("2006-01-02")
}

// seedPaidFixture menyiapkan invoice UNPAID + rekening/kategori untuk
// pengujian PaymentProcessor.
func seedPaidFixture(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.AutoMigrate(&model.CustomerModel{}))

	subID := "sub-pay"
	cust := model.CustomerModel{
		ID: "cust-pay", TenantID: "tenant-default", CustomerCode: "CUST-PAY",
		Name: "Pelanggan Pay", Phone: "081234567890", Address: "Jl. Bayar",
		Status: domainBilling.StatusUnpaid,
	}
	require.NoError(t, db.Create(&cust).Error)

	inv := model.InvoiceModel{
		ID: "inv-pay", TenantID: "tenant-default", InvoiceNumber: "INV-X-1",
		CustomerID: cust.ID, SubscriptionID: &subID, Period: "2026-08",
		Total: 110000, DueDate: timeDate("2026-08-20"),
		Status:    domainBilling.StatusUnpaid,
		QRPayload: "polyglot://invoice/inv-pay", ManualPaymentCode: "PAY-000001",
	}
	require.NoError(t, db.Create(&inv).Error)

	partial := inv
	partial.ID = "inv-partial"
	partial.InvoiceNumber = "INV-X-2"
	partial.Total = 110000
	partial.QRPayload = "polyglot://invoice/inv-partial"
	partial.ManualPaymentCode = "PAY-000002"
	require.NoError(t, db.Create(&partial).Error)

	require.NoError(t, db.Create(&model.CashAccountModel{ID: "ca-1", AccountCode: "1001", Name: "Kas", Type: "CASH", IsActive: true}).Error)
	require.NoError(t, db.Create(&model.CashCategoryModel{ID: "cc-1", Name: "Tagihan", Type: "INCOME"}).Error)

	_ = sqlite.Dialector{} // keep import referenced if setup changes
}
