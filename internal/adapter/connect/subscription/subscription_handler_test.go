package subscription_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	connectSub "github.com/quixiq/polyglot/internal/adapter/connect/subscription"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	subUC "github.com/quixiq/polyglot/internal/usecase/subscription"
)

func TestSubscriptionConnectHandler_CRUDAndEnrichment(t *testing.T) {
	planRepo := mocktest.NewFakeServicePlanRepo()
	subRepo := mocktest.NewFakeSubscriptionRepo()
	custRepo := mocktest.NewFakeCustomerRepo()
	invRepo := mocktest.NewFakeInvoiceRepo()
	manager := mocktest.NewFakeRouterAccountManager()
	audit := &mocktest.FakeAuditWriter{}

	ctx := context.Background()

	planRepo.Seed(domainPlan.ServicePlan{
		ID:                    "plan-50m",
		Name:                  "PAKET-50M",
		ServiceType:           "PPPOE",
		BandwidthDownloadKbps: 50000,
		BandwidthUploadKbps:   50000,
		Price:                 250000,
		IsActive:              true,
	})
	subRepo.Seed(domainSub.Subscription{
		ID:             "sub-100",
		CustomerID:     "cust-1",
		PlanID:         "plan-50m",
		ServiceType:    "PPPOE",
		RemoteUsername: "user_pppoe",
		Status:         domainSub.StatusActive,
	})

	manageUC := subUC.NewManageSubscriptionUseCase(subRepo, planRepo, custRepo, nil, manager, audit, invRepo)
	lifecycleUC := subUC.NewLifecycleUseCase(subRepo, planRepo, manager, audit)
	handler := connectSub.NewSubscriptionConnectHandler(manageUC, lifecycleUC)

	// 1. GetSubscription
	getResp, err := handler.GetSubscription(ctx, connect.NewRequest(&devicepb.GetSubscriptionRequest{Id: "sub-100"}))
	require.NoError(t, err)
	sub := getResp.Msg.Subscription
	assert.Equal(t, "sub-100", sub.Id)
	assert.Equal(t, "PAKET-50M", sub.PlanName)
	assert.Equal(t, "50M/50M", sub.RateLimit)
	require.NotNil(t, sub.PppoeConfig)
	assert.Equal(t, "50M/50M", sub.PppoeConfig.RateLimit)
	assert.Equal(t, "PAKET-50M", sub.PppoeConfig.RouterProfile)

	// 2. ListSubscriptions
	listResp, err := handler.ListSubscriptions(ctx, connect.NewRequest(&devicepb.ListSubscriptionsRequest{CustomerId: "cust-1"}))
	require.NoError(t, err)
	require.Len(t, listResp.Msg.Subscriptions, 1)
	assert.Equal(t, "PAKET-50M", listResp.Msg.Subscriptions[0].PlanName)
	assert.Equal(t, "50M/50M", listResp.Msg.Subscriptions[0].RateLimit)
}
