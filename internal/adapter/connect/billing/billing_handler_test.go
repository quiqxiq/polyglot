package billing_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	connectBilling "github.com/quixiq/polyglot/internal/adapter/connect/billing"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
)

func newBillingConnectFixture(t *testing.T) (*connectBilling.BillingConnectHandler, *mocktest.FakeInvoiceRepo, *mocktest.FakeSubscriptionRepo, *mocktest.FakeServicePlanRepo) {
	t.Helper()
	invRepo := mocktest.NewFakeInvoiceRepo()
	subRepo := mocktest.NewFakeSubscriptionRepo()
	planRepo := mocktest.NewFakeServicePlanRepo()
	custRepo := mocktest.NewFakeCustomerRepo()
	paymentProc := &mocktest.FakePaymentProcessor{
		Pay: domainBilling.Payment{ID: "pay-1", PaymentNo: "PAY-001"},
	}

	invUC := billingUC.NewInvoiceUseCase(invRepo)
	checkoutUC := billingUC.NewCheckoutUseCase(invRepo, custRepo, paymentProc)
	runBillingUC := billingUC.NewRunBillingUseCase(subRepo, planRepo, invRepo)

	handler := connectBilling.NewBillingConnectHandler(invUC, checkoutUC, runBillingUC)
	return handler, invRepo, subRepo, planRepo
}

func TestBillingConnectHandler_CashierPay(t *testing.T) {
	handler, invRepo, _, _ := newBillingConnectFixture(t)
	ctx := context.Background()

	require.NoError(t, invRepo.Save(ctx, domainBilling.Invoice{
		ID:                "inv-10",
		CustomerID:        "cust-1",
		Total:             100000,
		PaidAmount:        0,
		Status:            domainBilling.StatusUnpaid,
		ManualPaymentCode: "PAY-101010",
	}))

	payResp, err := handler.CashierPay(ctx, connect.NewRequest(&devicepb.CashierPayRequest{
		InvoiceId:        "inv-10",
		Amount:           100000,
		CashAccountId:    "ca-1001-kas-kantor",
		IncomeCategoryId: "cc-tagihan",
	}))
	require.NoError(t, err)
	assert.Equal(t, "pay-1", payResp.Msg.PaymentId)
	assert.Equal(t, "PAY-001", payResp.Msg.PaymentNo)
}

func TestBillingConnectHandler_GenerateInvoices(t *testing.T) {
	handler, _, subRepo, planRepo := newBillingConnectFixture(t)
	ctx := context.Background()

	planRepo.Seed(domainPlan.ServicePlan{
		ID: "plan-a", Name: "HOME-10M", ServiceType: "PPPOE",
		Price: 100000, IsActive: true,
	})
	subRepo.Seed(domainSub.Subscription{
		ID: "sub-1", CustomerID: "c1", PlanID: "plan-a",
		Status: domainSub.StatusActive,
	})

	genResp, err := handler.GenerateInvoices(ctx, connect.NewRequest(&devicepb.GenerateInvoicesRequest{
		Period: "2026-08",
	}))
	require.NoError(t, err)
	assert.Equal(t, int32(1), genResp.Msg.Created)
	assert.Equal(t, int32(0), genResp.Msg.Skipped)
}
