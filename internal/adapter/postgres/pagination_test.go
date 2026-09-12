// F6-8: FindPaged memfilter tenant + limit/offset di level query.
package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/port"
)

func TestFindPaged_TenantAndPagination(t *testing.T) {
	db := setupISPDB(t)
	ctx := context.Background()
	custRepo := postgres.NewCustomerRepository(db)

	for _, tc := range []struct{ id, tenant string }{
		{"cust-p1", "tenant-a"}, {"cust-p2", "tenant-a"}, {"cust-p3", "tenant-b"},
	} {
		require.NoError(t, custRepo.Save(ctx, domainCustomer.Customer{
			ID: tc.id, TenantID: tc.tenant, CustomerCode: "C-" + tc.id,
			Name: "Nama", Phone: "081200000000", Address: "Jl. Test",
			PortalAccessCode: "P" + tc.id, Status: domainCustomer.StatusActive,
		}))
	}

	// Filter tenant.
	a, err := custRepo.FindPaged(ctx, port.PageFilter{TenantID: "tenant-a"})
	require.NoError(t, err)
	require.Len(t, a, 2)
	for _, c := range a {
		assert.Equal(t, "tenant-a", c.TenantID)
	}

	// Limit + offset.
	first, err := custRepo.FindPaged(ctx, port.PageFilter{TenantID: "tenant-a", Limit: 1})
	require.NoError(t, err)
	require.Len(t, first, 1)

	second, err := custRepo.FindPaged(ctx, port.PageFilter{TenantID: "tenant-a", Limit: 1, Offset: 1})
	require.NoError(t, err)
	require.Len(t, second, 1)
	assert.NotEqual(t, first[0].ID, second[0].ID)

	// Offset melebihi jumlah baris → kosong, bukan error.
	empty, err := custRepo.FindPaged(ctx, port.PageFilter{TenantID: "tenant-a", Offset: 10})
	require.NoError(t, err)
	assert.Empty(t, empty)
}
