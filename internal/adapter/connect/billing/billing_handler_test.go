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
	manager := mocktest.NewFakeRouterAccountManager()
	audit := &mocktest.FakeAuditWriter{}

	invUC := billingUC.NewInvoiceUseCase(invRepo)
	checkoutUC := billingUC.NewCheckoutUseCase(invRepo, custRepo, paymentProc)
	subUC := billingUC.NewSubscriptionUseCase(subRepo)
	lifecycleUC := billingUC.NewSubscriptionLifecycleUseCase(subRepo, planRepo, manager, audit)
	planUC := billingUC.NewPlanUseCase(planRepo, subRepo)
	runBillingUC := billingUC.NewRunBillingUseCase(subRepo, planRepo, invRepo)

	handler := connectBilling.NewBillingConnectHandler(
		invUC, checkoutUC, subUC, lifecycleUC, planUC, runBillingUC,
	)
	return handler, invRepo, subRepo, planRepo
}

func TestBillingConnectHandler_PlanCRUD(t *testing.T) {
	handler, _, _, _ := newBillingConnectFixture(t)
	ctx := context.Background()

	// 1. Create Plan
	createResp, err := handler.CreatePlan(ctx, connect.NewRequest(&devicepb.CreatePlanRequest{
		Plan: &devicepb.Plan{
			Name:                  "PAKET-50M",
			ServiceType:           "PPPOE",
			BandwidthDownloadKbps: 50000,
			BandwidthUploadKbps:   50000,
			Price:                 250000,
		},
	}))
	require.NoError(t, err)
	planID := createResp.Msg.Plan.Id
	assert.NotEmpty(t, planID)
	assert.Equal(t, "PAKET-50M", createResp.Msg.Plan.Name)

	// 2. Get Plan
	getResp, err := handler.GetPlan(ctx, connect.NewRequest(&devicepb.GetPlanRequest{Id: planID}))
	require.NoError(t, err)
	assert.Equal(t, "PAKET-50M", getResp.Msg.Plan.Name)

	// 3. List Plans
	listResp, err := handler.ListPlans(ctx, connect.NewRequest(&devicepb.ListPlansRequest{ActiveOnly: true}))
	require.NoError(t, err)
	assert.Len(t, listResp.Msg.Plans, 1)

	// 4. Update Plan
	updateResp, err := handler.UpdatePlan(ctx, connect.NewRequest(&devicepb.UpdatePlanRequest{
		Plan: &devicepb.Plan{
			Id:                    planID,
			Name:                  "PAKET-50M-PROMO",
			ServiceType:           "PPPOE",
			BandwidthDownloadKbps: 50000,
			BandwidthUploadKbps:   50000,
			Price:                 225000,
		},
	}))
	require.NoError(t, err)
	assert.Equal(t, "PAKET-50M-PROMO", updateResp.Msg.Plan.Name)
	assert.InDelta(t, 225000, updateResp.Msg.Plan.Price, 0.01)

	// 5. Delete Plan
	delResp, err := handler.DeletePlan(ctx, connect.NewRequest(&devicepb.DeletePlanRequest{Id: planID}))
	require.NoError(t, err)
	assert.Contains(t, delResp.Msg.Message, "deleted")
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

	// CashierPay via RPC
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
