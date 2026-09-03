package plan_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	connectPlan "github.com/quixiq/polyglot/internal/adapter/connect/plan"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	planUC "github.com/quixiq/polyglot/internal/usecase/plan"
)

func TestPlanConnectHandler_PlanCRUD(t *testing.T) {
	planRepo := mocktest.NewFakeServicePlanRepo()
	subRepo := mocktest.NewFakeSubscriptionRepo()
	manager := mocktest.NewFakeRouterAccountManager()

	uc := planUC.NewManagePlanUseCase(planRepo, subRepo, manager)
	handler := connectPlan.NewPlanConnectHandler(uc)
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
