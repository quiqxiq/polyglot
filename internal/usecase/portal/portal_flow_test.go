package portal_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/portal"
)

func fixture(t *testing.T, phone string) (*uc.UseCase, *mocktest.FakePortalRepo, *mocktest.FakeCustomerRepo, *mocktest.FakeInvoiceRepo, *mocktest.FakeNotificationRepo) {
	t.Helper()
	portals := mocktest.NewFakePortalRepo()
	customers := mocktest.NewFakeCustomerRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	notif := mocktest.NewFakeNotificationRepo()
	sender := &mocktest.FakeNotificationSender{}
	settings := mocktest.NewFakeSettingReader(map[string]string{
		"isp.otp_ttl_minutes":      "5",
		"isp.otp_max_attempts":     "3",
		"isp.portal_session_hours": "12",
	})
	usecase := uc.NewUseCase(portals, customers,
		mocktest.NewFakeSubscriptionRepo(), invoices,
		mocktest.NewFakePaymentReader(), sender, notif, settings)

	require.NoError(t, customers.Save(context.Background(), domainCustomer.Customer{
		ID: "c1", TenantID: "tenant-default", CustomerCode: "CUST-1",
		Name: "Budi", Phone: phone, Address: "Jl. Portal",
		Status: domainCustomer.StatusIsolated, PortalAccessCode: "77777777",
	}))
	return usecase, portals, customers, invoices, notif
}

func TestRequestOTP_QueuesViaWA_MaskedPhone(t *testing.T) {
	usecase, _, _, _, notif := fixture(t, "085606846141")
	masked, err := usecase.RequestOTP(context.Background(), "085606846141")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(masked, "085"))
	assert.True(t, strings.Contains(masked, "****"))

	// F5-7: OTP masuk antrean worker WhatsApp, bukan kirim langsung.
	queued := notif.Queued()
	require.Len(t, queued, 1)
	assert.Equal(t, "PORTAL_OTP", queued[0].MessageType)
	assert.Contains(t, queued[0].MessageContent, "Kode login portal")
}

func TestRequestOTP_UnknownIdentifier_GenericError(t *testing.T) {
	usecase, _, _, _, notif := fixture(t, "085606846141")
	_, err := usecase.RequestOTP(context.Background(), "089900000000")
	assert.ErrorIs(t, err, domainCustomer.ErrPortalBadCredentials)
	assert.Empty(t, notif.Queued())
}

func isDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func TestLogin_FullFlow_AndWrongOTP(t *testing.T) {
	usecase, _, _, _, notif := fixture(t, "085606846141")
	ctx := context.Background()

	_, err := usecase.RequestOTP(ctx, "085606846141")
	require.NoError(t, err)

	code := lastCodeFromNotifications(notif)
	require.NotEmpty(t, code, "OTP harus masuk antrean notifikasi")

	token, cust, expiresAt, err := usecase.Login(ctx, "085606846141", code)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, "Budi", cust.Name)
	// F5-3: expiry sesi mengikuti setting (12 jam).
	assert.Greater(t, expiresAt.Unix(), int64(0))

	got, err := usecase.Authenticate(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, "c1", got.ID)

	require.NoError(t, usecase.Logout(ctx, token))
	_, err = usecase.Authenticate(ctx, token)
	assert.ErrorIs(t, err, domainCustomer.ErrPortalBadCredentials)

	// OTP salah pada permintaan baru → login gagal.
	_, err = usecase.RequestOTP(ctx, "085606846141")
	require.NoError(t, err)
	_, _, _, err = usecase.Login(ctx, "085606846141", "000000")
	assert.Error(t, err)
}

func lastCodeFromNotifications(notif *mocktest.FakeNotificationRepo) string {
	queued := notif.Queued()
	for i := len(queued) - 1; i >= 0; i-- {
		for _, f := range strings.Fields(queued[i].MessageContent) {
			trimmed := strings.TrimSuffix(f, ".")
			if len(trimmed) == 6 && isDigits(trimmed) {
				return trimmed
			}
		}
	}
	return ""
}

func TestLookupBill_ByPhoneAndPaymentCode(t *testing.T) {
	ctx := context.Background()
	portals := mocktest.NewFakePortalRepo()
	customers := mocktest.NewFakeCustomerRepo()
	invoices := mocktest.NewFakeInvoiceRepo()
	sender := &mocktest.FakeNotificationSender{}
	settings := mocktest.NewFakeSettingReader(map[string]string{})
	usecase := uc.NewUseCase(portals, customers,
		mocktest.NewFakeSubscriptionRepo(), invoices,
		mocktest.NewFakePaymentReader(), sender, mocktest.NewFakeNotificationRepo(), settings)

	require.NoError(t, customers.Save(ctx, domainCustomer.Customer{
		ID: "c1", TenantID: "tenant-default", CustomerCode: "CUST-100",
		Name: "Budi Santoso", Phone: "081234567890", Address: "Jl. Merdeka",
		Status: domainCustomer.StatusIsolated, PortalAccessCode: "88888888",
	}))

	// Save an unpaid invoice
	require.NoError(t, invoices.Save(ctx, domainBilling.Invoice{
		ID: "inv-1", CustomerID: "c1", InvoiceNumber: "INV-2026-001",
		Period: "2026-09", Total: 150000, PaidAmount: 0,
		Status: "UNPAID", ManualPaymentCode: "PAY-1234",
	}))

	// 1. Lookup by phone
	bill, err := usecase.LookupBill(ctx, "081234567890")
	require.NoError(t, err)
	assert.Equal(t, "B**i S*****o", bill.CustomerName)
	assert.Equal(t, "inv-1", bill.InvoiceID)
	assert.Equal(t, float64(150000), bill.Outstanding)

	// 2. Lookup by payment code
	billByCode, err := usecase.LookupBill(ctx, "PAY-1234")
	require.NoError(t, err)
	assert.Equal(t, "inv-1", billByCode.InvoiceID)

	// 3. Lookup unknown
	_, err = usecase.LookupBill(ctx, "089999999999")
	assert.Error(t, err)
}

// F5-4: invoice hanya bisa diakses pelanggan pemiliknya.
func TestInvoiceForCustomer_Ownership(t *testing.T) {
	usecase, _, customers, invoices, _ := fixture(t, "085606846141")
	ctx := context.Background()

	require.NoError(t, invoices.Save(ctx, domainBilling.Invoice{
		ID: "inv-own", CustomerID: "c1", InvoiceNumber: "INV-OWN",
		Total: 100000, Status: domainBilling.StatusUnpaid,
	}))
	require.NoError(t, customers.Save(ctx, domainCustomer.Customer{
		ID: "c2", TenantID: "tenant-default", CustomerCode: "CUST-2",
		Name: "Lain", Phone: "081200000002", Address: "Jl. Lain",
		PortalAccessCode: "66666666",
	}))
	require.NoError(t, invoices.Save(ctx, domainBilling.Invoice{
		ID: "inv-other", CustomerID: "c2", InvoiceNumber: "INV-OTHER",
		Total: 100000, Status: domainBilling.StatusUnpaid,
	}))

	// Milik sendiri → OK.
	inv, err := usecase.InvoiceForCustomer(ctx, "c1", "inv-own")
	require.NoError(t, err)
	assert.Equal(t, "c1", inv.CustomerID)

	// Milik pelanggan lain → forbidden (bukan not-found) agar tidak bocor.
	_, err = usecase.InvoiceForCustomer(ctx, "c1", "inv-other")
	assert.ErrorIs(t, err, domainCustomer.ErrPortalForbidden)

	// Tidak ada → forbidden juga.
	_, err = usecase.InvoiceForCustomer(ctx, "c1", "inv-missing")
	assert.ErrorIs(t, err, domainCustomer.ErrPortalForbidden)
}
