package routerjob_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainRouterJob "github.com/quixiq/polyglot/internal/domain/routerjob"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/routerjob"
)

func exporterFixture(t *testing.T) (*uc.Exporter, *mocktest.FakeRouterJobRepo) {
	t.Helper()
	jobs := mocktest.NewFakeRouterJobRepo()
	subs := mocktest.NewFakeSubscriptionRepo()
	plans := mocktest.NewFakeServicePlanRepo()

	plans.Seed(domainPlan.ServicePlan{
		ID: "plan-1", Name: "HOME-20M", RouterProfileName: "HOME-20M", ServiceType: "PPPOE",
		Price: 150000, IsActive: true,
	})
	require.NoError(t, subs.Save(context.Background(), domainSubscription.Subscription{
		ID: "sub-1", TenantID: "tenant-default", CustomerID: "cust-1", PlanID: "plan-1",
		ServiceType: "PPPOE", RemoteUsername: "bs1234", RemotePassword: "secret99",
		RemoteAddress: "10.10.10.5", RouterProfile: "HOME-20M",
		Status: domainSubscription.StatusIsolated,
	}))
	jobs.Seed(domainRouterJob.Job{
		ID: "job-1", TenantID: "tenant-default", DeviceID: "dev-1",
		JobType: domainRouterJob.TypeProvisionPPPoE, Payload: []byte(`{"subscription_id":"sub-1"}`),
		IdempotencyKey: "provision:sub-1:1", Status: domainRouterJob.StatusDead,
		Attempts: 5, MaxAttempts: 5, LastError: "router unreachable",
	})
	jobs.Seed(domainRouterJob.Job{
		ID: "job-iso", TenantID: "tenant-default", DeviceID: "dev-1",
		JobType: domainRouterJob.TypeIsolate, Payload: []byte(`{"subscription_id":"sub-1"}`),
		IdempotencyKey: "isolate:sub-1", Status: domainRouterJob.StatusDead,
		Attempts: 5, MaxAttempts: 5,
	})
	return uc.NewExporter(jobs, subs, plans), jobs
}

func TestExporter_ProvisionScript(t *testing.T) {
	exp, _ := exporterFixture(t)
	script, err := exp.ExportScript(context.Background(), "job-1")
	require.NoError(t, err)
	assert.Contains(t, script, "/ppp secret add")
	assert.Contains(t, script, `name="bs1234"`)
	assert.Contains(t, script, `profile="HOME-20M"`)
	assert.Contains(t, script, "router unreachable")
}

func TestExporter_IsolateScript(t *testing.T) {
	exp, _ := exporterFixture(t)
	script, err := exp.ExportScript(context.Background(), "job-iso")
	require.NoError(t, err)
	assert.Contains(t, script, "/ip firewall address-list add")
	assert.Contains(t, script, `list="ISOLIR_USERS"`)
	assert.Contains(t, script, `profile="isolir"`)
}
