package importer_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/importer"
)

type mockPlanRepo struct {
	port.ServicePlanRepository
	plans map[string]domainPlan.ServicePlan
}

func newMockPlanRepo() *mockPlanRepo {
	return &mockPlanRepo{plans: make(map[string]domainPlan.ServicePlan)}
}

func (m *mockPlanRepo) FindByName(_ context.Context, tenantID, name string) (domainPlan.ServicePlan, error) {
	if p, ok := m.plans[name]; ok {
		return p, nil
	}
	return domainPlan.ServicePlan{}, domainPlan.ErrNotFound
}

func (m *mockPlanRepo) Save(_ context.Context, plan domainPlan.ServicePlan) error {
	m.plans[plan.Name] = plan
	return nil
}

func TestPullRouterPlans_Unified(t *testing.T) {
	ctx := context.Background()

	pgw := &mockPPPGateway{
		listProfilesFn: func(_ context.Context, _ port.DeviceDriver, _ string) ([]port.PPPProfile, error) {
			return []port.PPPProfile{
				{Name: "default", RateLimit: "1M/2M", Comment: ""},
				{Name: "default-encryption", RateLimit: "1M/2M"}, // harus dilewati
				{Name: "Paket-Home-10M", RateLimit: "10M/5M", Comment: "Tagihan Rp 150.000", ParentQueue: "parent-all"},
			}, nil
		},
	}

	hgw := &mockHotspotGateway{
		getUserProfilesFn: func(_ context.Context, _ port.DeviceDriver) ([]port.HotspotUserProfile, error) {
			return []port.HotspotUserProfile{
				{Name: "Hotspot-Bulanan", RateLimit: "2M/5M", Comment: "Rp 100000", SharedUsers: "2", SessionTimeout: "30d"},
			}, nil
		},
	}

	repo := newMockPlanRepo()
	src := importer.NewRouterSource(pgw, hgw)
	src.SetPlanRepository(repo)

	rows, pppoeCount, hotspotCount, err := src.PullRouterPlans(ctx, nil, "ALL")
	require.NoError(t, err)
	assert.Equal(t, 2, pppoeCount)   // default + Paket-Home-10M (default-encryption dilewati)
	assert.Equal(t, 1, hotspotCount) // Hotspot-Bulanan
	assert.Equal(t, 3, len(rows))

	// Cari paket Home-10M
	var homePlan *importer.PlanRow
	for i := range rows {
		if rows[i].RouterProfile == "Paket-Home-10M" {
			homePlan = &rows[i]
			break
		}
	}
	require.NotNil(t, homePlan)
	assert.Equal(t, "Paket Home 10m", homePlan.Name)
	assert.Equal(t, "Paket-Home-10M", homePlan.RouterProfile)
	assert.Equal(t, "PPPOE", homePlan.ServiceType)
	assert.Equal(t, 10000, homePlan.BandwidthDownloadKbps)
	assert.Equal(t, 5000, homePlan.BandwidthUploadKbps)
	assert.Equal(t, float64(150000), homePlan.Price)
	assert.True(t, homePlan.IsNew)

	// Commit plans
	upsert := importer.NewUpsertUseCase(repo, nil, nil, nil, "dev-1")
	commitRes, err := upsert.CommitPlans(ctx, rows)
	require.NoError(t, err)
	assert.Equal(t, 3, commitRes.PlansCreated)
	assert.Equal(t, 0, commitRes.PlansUpdated)
	assert.Equal(t, 3, len(repo.plans))
}
