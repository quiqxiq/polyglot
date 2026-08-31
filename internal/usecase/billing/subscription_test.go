package billing_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
)

func TestSubscriptionUseCase_Enrichment(t *testing.T) {
	subRepo := mocktest.NewFakeSubscriptionRepo()
	planRepo := mocktest.NewFakeServicePlanRepo()
	custRepo := mocktest.NewFakeCustomerRepo()
	ctx := context.Background()

	// Seed customer
	require.NoError(t, custRepo.Save(ctx, domainCustomer.Customer{
		ID:           "cust-001",
		Name:         "John Doe",
		Phone:        "081234567890",
		CustomerCode: "CUST-001",
	}))

	// Seed plan with rate limit
	planRepo.Seed(domainPlan.ServicePlan{
		ID:                    "plan-50m",
		Name:                  "PAKET-50M",
		ServiceType:           "PPPOE",
		BandwidthDownloadKbps: 50000,
		BandwidthUploadKbps:   50000,
		Price:                 250000,
		IsActive:              true,
	})

	// Seed subscription
	subRepo.Seed(domainSub.Subscription{
		ID:             "sub-001",
		CustomerID:     "cust-001",
		PlanID:         "plan-50m",
		ServiceType:    "PPPOE",
		RemoteUsername: "john_pppoe",
		Status:         domainSub.StatusActive,
	})

	subUC := billingUC.NewSubscriptionUseCase(subRepo, planRepo, custRepo, nil)

	// Test GetSubscription (enriched)
	detail, err := subUC.GetSubscription(ctx, "sub-001")
	require.NoError(t, err)
	assert.Equal(t, "sub-001", detail.Subscription.ID)
	require.NotNil(t, detail.Plan)
	assert.Equal(t, "PAKET-50M", detail.Plan.Name)
	assert.Equal(t, "50M/50M", detail.Plan.RateLimit())
	require.NotNil(t, detail.Customer)
	assert.Equal(t, "John Doe", detail.Customer.Name)
	assert.Equal(t, "081234567890", detail.Customer.Phone)

	// Test ListSubscriptions
	list, err := subUC.ListSubscriptions(ctx, "cust-001")
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "PAKET-50M", list[0].Plan.Name)
	assert.Equal(t, "John Doe", list[0].Customer.Name)
}

func TestSubscriptionUseCase_CRUD(t *testing.T) {
	subRepo := mocktest.NewFakeSubscriptionRepo()
	subUC := billingUC.NewSubscriptionUseCase(subRepo, nil, nil, nil)
	ctx := context.Background()

	created, err := subUC.CreateSubscription(ctx, domainSub.Subscription{
		ID:         "sub-new",
		CustomerID: "cust-1",
		PlanID:     "plan-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "sub-new", created.ID)
	assert.Equal(t, domainSub.StatusActive, created.Status)

	cancelled, err := subUC.CancelSubscription(ctx, "sub-new")
	require.NoError(t, err)
	assert.Equal(t, domainSub.StatusCancelled, cancelled.Status)
}
