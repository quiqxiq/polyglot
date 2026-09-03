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

func fixture(t *testing.T, phone string) (*uc.UseCase, *mocktest.FakePortalRepo, *mocktest.FakeCustomerRepo, *mocktest.FakeNotificationSender) {
	t.Helper()
	portals := mocktest.NewFakePortalRepo()
	customers := mocktest.NewFakeCustomerRepo()
	sender := &mocktest.FakeNotificationSender{}
	settings := mocktest.NewFakeSettingReader(map[string]string{
		"isp.otp_ttl_minutes":      "5",
		"isp.otp_max_attempts":     "3",
		"isp.portal_session_hours": "12",
	})
	usecase := uc.NewUseCase(portals, customers,
		mocktest.NewFakeSubscriptionRepo(), mocktest.NewFakeInvoiceRepo(),
		mocktest.NewFakePaymentReader(), sender, settings)

	require.NoError(t, customers.Save(context.Background(), domainCustomer.Customer{
		ID: "c1", TenantID: "tenant-default", CustomerCode: "CUST-1",
		Name: "Budi", Phone: phone, Address: "Jl. Portal",
		Status: domainCustomer.StatusIsolated, PortalAccessCode: "77777777",
	}))
	return usecase, portals, customers, sender
}

func TestRequestOTP_SendsViaWA_MaskedPhone(t *testing.T) {
	usecase, _, _, sender := fixture(t, "085606846141")
	masked, err := usecase.RequestOTP(context.Background(), "085606846141")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(masked, "085"))
	assert.True(t, strings.Contains(masked, "****"))
	require.Len(t, sender.Messages, 1)
	assert.Contains(t, sender.Messages[0].Content, "Kode login portal")
}

func TestRequestOTP_UnknownIdentifier_GenericError(t *testing.T) {
	usecase, _, _, sender := fixture(t, "085606846141")
	_, err := usecase.RequestOTP(context.Background(), "089900000000")
	assert.ErrorIs(t, err, domainCustomer.ErrPortalBadCredentials)
	assert.Empty(t, sender.Messages)
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
	usecase, portals, customers, sender := fixture(t, "085606846141")
	ctx := context.Background()

	_, err := usecase.RequestOTP(ctx, "085606846141")
	require.NoError(t, err)

	code := lastCodeFromSender(sender)
	require.NotEmpty(t, code, "OTP harus terkirim ke fake sender")
	t.Logf("DEBUG extracted=%q storedHashCount=%d", code, portals.OTPCount())

	token, cust, err := usecase.Login(ctx, "085606846141", code)
	t.Logf("DEBUG login err=%v tokenLen=%d", err, len(token))
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, "Budi", cust.Name)

	got, err := usecase.Authenticate(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, "c1", got.ID)

	require.NoError(t, usecase.Logout(ctx, token))
	_, err = usecase.Authenticate(ctx, token)
	assert.ErrorIs(t, err, domainCustomer.ErrPortalBadCredentials)

	// OTP salah pada permintaan baru → login gagal.
	_, err = usecase.RequestOTP(ctx, "085606846141")
	require.NoError(t, err)
	_, _, err = usecase.Login(ctx, "085606846141", "000000")
	assert.Error(t, err)

	_ = customers
}

func lastCodeFromSender(sender *mocktest.FakeNotificationSender) string {
	msgs := sender.Sent()
	if len(msgs) == 0 {
		return ""
	}
	content := msgs[len(msgs)-1].Content
	for _, f := range strings.Fields(content) {
		trimmed := strings.TrimSuffix(f, ".")
		if len(trimmed) == 6 && isDigits(trimmed) {
			return trimmed
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
		mocktest.NewFakePaymentReader(), sender, settings)

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
	assert.Equal(t, "CUST-100", bill.CustomerCode)
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
