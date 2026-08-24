package portal_test

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	connectPortal "github.com/quixiq/polyglot/internal/adapter/connect/portal"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/portal"
)

func newPortalConnectFixture(t *testing.T) (*connectPortal.PortalConnectHandler, *mocktest.FakeNotificationSender) {
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
		Name: "Budi", Phone: "085606846141", Address: "Jl. Portal",
		Status: domainCustomer.StatusActive, PortalAccessCode: "77777777",
	}))

	return connectPortal.NewPortalConnectHandler(usecase), sender
}

func TestPortalConnectHandler_RequestOTP_Validation(t *testing.T) {
	handler, _ := newPortalConnectFixture(t)
	ctx := context.Background()

	// Identifier kosong -> error
	_, err := handler.RequestOTP(ctx, connect.NewRequest(&devicepb.RequestOTPRequest{Identifier: ""}))
	assert.Error(t, err)

	// Identifier valid -> sukses
	resp, err := handler.RequestOTP(ctx, connect.NewRequest(&devicepb.RequestOTPRequest{Identifier: "085606846141"}))
	require.NoError(t, err)
	assert.Contains(t, resp.Msg.Message, "OTP dikirim")
}

func TestPortalConnectHandler_Overview_Unauthenticated(t *testing.T) {
	handler, _ := newPortalConnectFixture(t)
	ctx := context.Background()

	// Token palsu -> 401 unauthenticated
	_, err := handler.Overview(ctx, connect.NewRequest(&devicepb.PortalOverviewRequest{Token: "invalid-token"}))
	assert.Error(t, err)

	var dateTmp time.Time
	_ = dateTmp
}
