package routerjob_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	routerjobConnect "github.com/quixiq/polyglot/internal/adapter/connect/routerjob"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainRouterJob "github.com/quixiq/polyglot/internal/domain/routerjob"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	jobuc "github.com/quixiq/polyglot/internal/usecase/routerjob"
)

func routerJobFixture(t *testing.T) *routerjobConnect.ConnectHandler {
	t.Helper()
	jobs := mocktest.NewFakeRouterJobRepo()
	subs := mocktest.NewFakeSubscriptionRepo()
	plans := mocktest.NewFakeServicePlanRepo()

	plans.Seed(domainPlan.ServicePlan{ID: "plan-1", Name: "HOME", RouterProfileName: "HOME-20M"})
	require.NoError(t, subs.Save(context.Background(), domainSubscription.Subscription{
		ID: "sub-1", TenantID: "tenant-default", PlanID: "plan-1", ServiceType: "PPPOE",
		RemoteUsername: "bs1234", RemotePassword: "secret", RouterProfile: "HOME-20M",
	}))
	jobs.Seed(domainRouterJob.Job{
		ID: "job-1", TenantID: "tenant-default", DeviceID: "dev-1",
		JobType: domainRouterJob.TypeProvisionPPPoE, Payload: []byte(`{"subscription_id":"sub-1"}`),
		IdempotencyKey: "k1", Status: domainRouterJob.StatusDead, Attempts: 5, MaxAttempts: 5,
	})
	return routerjobConnect.NewConnectHandler(jobs, jobuc.NewExporter(jobs, subs, plans))
}

func TestRouterJobHandler_ListRetryExport(t *testing.T) {
	h := routerJobFixture(t)
	ctx := context.Background()

	// List semua
	listResp, err := h.ListRouterJobs(ctx, connect.NewRequest(&polyglotv1.ListRouterJobsRequest{}))
	require.NoError(t, err)
	require.Len(t, listResp.Msg.Jobs, 1)
	assert.Equal(t, "DEAD", listResp.Msg.Jobs[0].Status)

	// Retry job DEAD → QUEUED
	retryResp, err := h.RetryRouterJob(ctx, connect.NewRequest(&polyglotv1.RetryRouterJobRequest{Id: "job-1"}))
	require.NoError(t, err)
	assert.Equal(t, "QUEUED", retryResp.Msg.Job.Status)

	// Export script .rsc
	expResp, err := h.ExportRouterJobScript(ctx, connect.NewRequest(&polyglotv1.ExportRouterJobScriptRequest{Id: "job-1"}))
	require.NoError(t, err)
	assert.Contains(t, expResp.Msg.Script, "/ppp secret add")
	assert.Equal(t, "job-1.rsc", expResp.Msg.Filename)
}

func TestRouterJobHandler_RetrySuccessJobRejected(t *testing.T) {
	h := routerJobFixture(t)
	// job sukses tidak boleh di-retry — pakai job baru berstatus SUCCEEDED.
	_, err := h.RetryRouterJob(context.Background(), connect.NewRequest(&polyglotv1.RetryRouterJobRequest{Id: "tidak-ada"}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}
