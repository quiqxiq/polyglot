package portal_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	assert.ErrorIs(t, err, uc.ErrInvalidCredentials)
	assert.Empty(t, sender.Messages)
}

func extractCode(content string) string {
	fields := strings.Fields(content)
	for i, f := range fields {
		if f == ":" && i > 0 {
			return fields[i-1]
		}
	}
	for _, f := range fields {
		if len(f) == 6 && isDigits(f) {
			return f
		}
	}
	return ""
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
	assert.ErrorIs(t, err, uc.ErrInvalidCredentials)

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
