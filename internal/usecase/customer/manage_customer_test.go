package customer_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	customerUC "github.com/quixiq/polyglot/internal/usecase/customer"
)

func TestManageCustomerUseCase_CreateCustomer(t *testing.T) {
	ctx := context.Background()
	repo := mocktest.NewFakeCustomerRepo()
	uc := customerUC.NewManageCustomerUseCase(repo, nil, nil, nil)

	t.Run("creates customer with auto-generated id, customer_code, and portal_code", func(t *testing.T) {
		lat := -6.2088
		lon := 106.8456
		c := domainCustomer.Customer{
			Name:      "Budi Santoso",
			Phone:     "081234567890",
			Email:     "budi@example.com",
			Address:   "Jl. Mawar No. 1",
			Latitude:  &lat,
			Longitude: &lon,
		}

		created, err := uc.CreateCustomer(ctx, c)
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.NotEmpty(t, created.CustomerCode)
		assert.NotEmpty(t, created.PortalAccessCode)
		assert.Equal(t, domainCustomer.StatusActive, created.Status)
		assert.Equal(t, "tenant-default", created.TenantID)
		assert.Equal(t, "Budi Santoso", created.Name)
		assert.NotNil(t, created.Latitude)
		assert.NotNil(t, created.Longitude)

		// Verify stored in repo
		stored, err := uc.GetCustomer(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, stored.Customer.ID)
		assert.Equal(t, created.CustomerCode, stored.Customer.CustomerCode)
	})

	t.Run("returns error when name is empty", func(t *testing.T) {
		c := domainCustomer.Customer{
			Phone: "081234567890",
		}
		_, err := uc.CreateCustomer(ctx, c)
		assert.ErrorIs(t, err, domainCustomer.ErrInvalidInput)
	})
}

func TestManageCustomerUseCase_UpdateCustomer(t *testing.T) {
	ctx := context.Background()
	repo := mocktest.NewFakeCustomerRepo()
	uc := customerUC.NewManageCustomerUseCase(repo, nil, nil, nil)

	c := domainCustomer.Customer{
		Name:    "Joko",
		Phone:   "0811111111",
		Address: "Jl. Melati",
	}
	created, err := uc.CreateCustomer(ctx, c)
	require.NoError(t, err)

	t.Run("updates customer details", func(t *testing.T) {
		created.Name = "Joko Widodo"
		created.Address = "Jl. Solo No. 10"

		updated, err := uc.UpdateCustomer(ctx, created)
		require.NoError(t, err)
		assert.Equal(t, "Joko Widodo", updated.Name)
		assert.Equal(t, "Jl. Solo No. 10", updated.Address)
	})

	t.Run("returns error when id is empty", func(t *testing.T) {
		_, err := uc.UpdateCustomer(ctx, domainCustomer.Customer{Name: "Tanpa ID"})
		assert.ErrorIs(t, err, domainCustomer.ErrInvalidInput)
	})
}

func TestManageCustomerUseCase_DeleteCustomer(t *testing.T) {
	ctx := context.Background()
	custRepo := mocktest.NewFakeCustomerRepo()
	subRepo := mocktest.NewFakeSubscriptionRepo()
	invRepo := mocktest.NewFakeInvoiceRepo()
	router := mocktest.NewFakeRouterAccountManager()

	uc := customerUC.NewManageCustomerUseCase(custRepo, subRepo, invRepo, router)

	created, err := uc.CreateCustomer(ctx, domainCustomer.Customer{Name: "Ahmad", Phone: "0822222222"})
	require.NoError(t, err)

	devID := "dev-1"
	err = subRepo.Save(ctx, domainSub.Subscription{
		ID:             "sub-1",
		CustomerID:     created.ID,
		DeviceID:       &devID,
		ServiceType:    "PPPOE",
		RemoteUsername: "ahmad_pppoe",
		Status:         domainSub.StatusActive,
	})
	require.NoError(t, err)

	err = invRepo.Save(ctx, domainBilling.Invoice{
		ID:         "inv-1",
		CustomerID: created.ID,
		Status:     domainBilling.StatusUnpaid,
	})
	require.NoError(t, err)

	// Delete customer should cascade delete subscriptions and invoices
	err = uc.DeleteCustomer(ctx, created.ID)
	require.NoError(t, err)

	_, err = uc.GetCustomer(ctx, created.ID)
	assert.Error(t, err)

	// Subscriptions and invoices must be deleted
	subs, err := subRepo.FindByCustomerID(ctx, created.ID)
	require.NoError(t, err)
	assert.Empty(t, subs)

	invoices, err := invRepo.FindByCustomerID(ctx, created.ID)
	require.NoError(t, err)
	assert.Empty(t, invoices)

	// Router account termination recorded
	assert.Equal(t, 1, router.Count("Terminate:ahmad_pppoe"))
}

func TestManageCustomerUseCase_ListAndGetEnriched(t *testing.T) {
	ctx := context.Background()
	custRepo := mocktest.NewFakeCustomerRepo()
	subRepo := mocktest.NewFakeSubscriptionRepo()
	invRepo := mocktest.NewFakeInvoiceRepo()

	uc := customerUC.NewManageCustomerUseCase(custRepo, subRepo, invRepo, nil)

	c1, err := uc.CreateCustomer(ctx, domainCustomer.Customer{Name: "Pelanggan 1", Phone: "081111111"})
	require.NoError(t, err)

	// Tambah 2 subscription aktif untuk c1
	err = subRepo.Save(ctx, domainSub.Subscription{
		ID:         "sub-1",
		CustomerID: c1.ID,
		Status:     domainSub.StatusActive,
	})
	require.NoError(t, err)
	err = subRepo.Save(ctx, domainSub.Subscription{
		ID:         "sub-2",
		CustomerID: c1.ID,
		Status:     domainSub.StatusIsolated,
	})
	require.NoError(t, err)

	// Tambah 1 invoice unpaid
	err = invRepo.Save(ctx, domainBilling.Invoice{
		ID:         "inv-1",
		CustomerID: c1.ID,
		Status:     domainBilling.StatusUnpaid,
	})
	require.NoError(t, err)

	t.Run("ListCustomers returns populated counters", func(t *testing.T) {
		list, err := uc.ListCustomers(ctx)
		require.NoError(t, err)
		require.Len(t, list, 1)
		assert.Equal(t, c1.ID, list[0].Customer.ID)
		assert.Equal(t, 2, list[0].ActiveSubscriptionsCount)
		assert.Equal(t, 1, list[0].UnpaidInvoicesCount)
	})

	t.Run("GetCustomer returns populated counters", func(t *testing.T) {
		detail, err := uc.GetCustomer(ctx, c1.ID)
		require.NoError(t, err)
		assert.Equal(t, c1.ID, detail.Customer.ID)
		assert.Equal(t, 2, detail.ActiveSubscriptionsCount)
		assert.Equal(t, 1, detail.UnpaidInvoicesCount)
	})
}
