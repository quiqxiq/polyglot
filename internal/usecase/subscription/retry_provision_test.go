package subscription_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainRouterJob "github.com/quixiq/polyglot/internal/domain/routerjob"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/subscription"
)

type fakeRetrier struct {
	enqueued bool
	job      domainRouterJob.Job
	err      error
}

func (f *fakeRetrier) EnqueueProvision(_ context.Context, tenantID, deviceID, subscriptionID, serviceType string) (domainRouterJob.Job, error) {
	f.enqueued = true
	if f.err != nil {
		return domainRouterJob.Job{}, f.err
	}
	return f.job, nil
}

func TestRetryProvision_EnqueuesJob(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	deviceID := "dev-1"
	require.NoError(t, subs.Save(context.Background(), domainSubscription.Subscription{
		ID: "sub-1", TenantID: "tenant-default", PlanID: "plan-1", ServiceType: "PPPOE",
		DeviceID: &deviceID, Status: domainSubscription.StatusActive,
		ProvisionStatus: domainSubscription.ProvisionFailed,
	}))
	retrier := &fakeRetrier{job: domainRouterJob.Job{ID: "job-9", Status: domainRouterJob.StatusQueued}}
	lifecycle := uc.NewLifecycleUseCase(subs, mocktest.NewFakeServicePlanRepo(), mocktest.NewFakeRouterAccountManager(), &mocktest.FakeAuditWriter{}).
		WithProvisionRetrier(retrier)

	job, err := lifecycle.RetryProvision(context.Background(), "sub-1")
	require.NoError(t, err)
	assert.Equal(t, "job-9", job.ID)
	assert.True(t, retrier.enqueued)

	stored, err := subs.FindByID(context.Background(), "sub-1")
	require.NoError(t, err)
	assert.Equal(t, domainSubscription.ProvisionPending, stored.ProvisionStatus)
}

func TestRetryProvision_RequiresDevice(t *testing.T) {
	subs := mocktest.NewFakeSubscriptionRepo()
	require.NoError(t, subs.Save(context.Background(), domainSubscription.Subscription{
		ID: "sub-2", PlanID: "plan-1", ServiceType: "PPPOE",
		Status: domainSubscription.StatusActive,
	}))
	lifecycle := uc.NewLifecycleUseCase(subs, mocktest.NewFakeServicePlanRepo(), mocktest.NewFakeRouterAccountManager(), &mocktest.FakeAuditWriter{})

	_, err := lifecycle.RetryProvision(context.Background(), "sub-2")
	require.Error(t, err)
	assert.ErrorIs(t, err, domainSubscription.ErrInvalidInput)
}
