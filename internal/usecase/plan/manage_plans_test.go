package plan

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
)

type mockPlanRepo struct {
	plans map[string]domainPlan.Plan
}

func newMockPlanRepo() *mockPlanRepo {
	return &mockPlanRepo{plans: make(map[string]domainPlan.Plan)}
}

func (m *mockPlanRepo) Save(ctx context.Context, p domainPlan.Plan) error {
	m.plans[p.ID] = p
	return nil
}

func (m *mockPlanRepo) FindByID(ctx context.Context, id string) (domainPlan.Plan, error) {
	p, ok := m.plans[id]
	if !ok {
		return domainPlan.Plan{}, domainPlan.ErrNotFound
	}
	return p, nil
}

func (m *mockPlanRepo) List(ctx context.Context, serviceType string, activeOnly bool, limit int) ([]domainPlan.Plan, error) {
	out := make([]domainPlan.Plan, 0)
	for _, p := range m.plans {
		if serviceType != "" && p.ServiceType != serviceType {
			continue
		}
		if activeOnly && !p.IsActive {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

func (m *mockPlanRepo) Delete(ctx context.Context, id string) error {
	delete(m.plans, id)
	return nil
}

func validPlan(name string) domainPlan.Plan {
	return domainPlan.Plan{
		Name:         name,
		ServiceType:  domainPlan.ServiceTypePPPoE,
		RateDownKbps: 10240,
		RateUpKbps:   10240,
		Price:        150000,
		IPPoolName:   "PPPOE-POOL",
		IsActive:     true,
	}
}

func TestManagePlans_Create(t *testing.T) {
	uc := NewManagePlansUseCase(newMockPlanRepo())

	t.Run("sukses generate ID dan persist", func(t *testing.T) {
		repo := newMockPlanRepo()
		uc := NewManagePlansUseCase(repo)
		created, err := uc.Create(context.Background(), validPlan("100-RB"))
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, "tenant-default", created.TenantID)
		assert.Equal(t, 1, created.SharedUsers)
		assert.Len(t, repo.plans, 1)
	})

	t.Run("nama duplikat ditolak", func(t *testing.T) {
		uc := NewManagePlansUseCase(newMockPlanRepo())
		_, err := uc.Create(context.Background(), validPlan("100-RB"))
		require.NoError(t, err)
		_, err = uc.Create(context.Background(), validPlan("100-RB"))
		assert.ErrorIs(t, err, domainPlan.ErrAlreadyExists)
	})

	t.Run("service type tidak valid ditolak", func(t *testing.T) {
		p := validPlan("BAD")
		p.ServiceType = "FIBER"
		_, err := uc.Create(context.Background(), p)
		assert.ErrorIs(t, err, domainPlan.ErrInvalidServiceType)
	})

	t.Run("rate nol ditolak", func(t *testing.T) {
		p := validPlan("ZERO")
		p.RateDownKbps = 0
		_, err := uc.Create(context.Background(), p)
		assert.ErrorIs(t, err, domainPlan.ErrInvalidRate)
	})
}

func TestManagePlans_UpdateDelete(t *testing.T) {
	ctx := context.Background()
	repo := newMockPlanRepo()
	uc := NewManagePlansUseCase(repo)

	created, err := uc.Create(ctx, validPlan("100-RB"))
	require.NoError(t, err)

	t.Run("update mengubah field", func(t *testing.T) {
		created.Price = 175000
		updated, err := uc.Update(ctx, created)
		require.NoError(t, err)
		assert.Equal(t, 175000.0, updated.Price)
		got, err := repo.FindByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, 175000.0, got.Price)
	})

	t.Run("rename ke nama yang sudah dipakai ditolak", func(t *testing.T) {
		_, err := uc.Create(ctx, validPlan("LAIN"))
		require.NoError(t, err)
		dup := created
		dup.Name = "LAIN"
		_, err = uc.Update(ctx, dup)
		assert.ErrorIs(t, err, domainPlan.ErrAlreadyExists)
	})

	t.Run("delete lalu not found", func(t *testing.T) {
		err := uc.Delete(ctx, created.ID)
		require.NoError(t, err)
		_, err = uc.Get(ctx, created.ID)
		assert.ErrorIs(t, err, domainPlan.ErrNotFound)
	})
}
