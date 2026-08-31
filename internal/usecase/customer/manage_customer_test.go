package customer_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	customerUC "github.com/quixiq/polyglot/internal/usecase/customer"
)

func TestManageCustomerUseCase_CreateCustomer(t *testing.T) {
	ctx := context.Background()
	repo := mocktest.NewFakeCustomerRepo()
	uc := customerUC.NewManageCustomerUseCase(repo)

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
		assert.Equal(t, created.ID, stored.ID)
		assert.Equal(t, created.CustomerCode, stored.CustomerCode)
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
	uc := customerUC.NewManageCustomerUseCase(repo)

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
	repo := mocktest.NewFakeCustomerRepo()
	uc := customerUC.NewManageCustomerUseCase(repo)

	created, err := uc.CreateCustomer(ctx, domainCustomer.Customer{Name: "Ahmad", Phone: "0822222222"})
	require.NoError(t, err)

	err = uc.DeleteCustomer(ctx, created.ID)
	require.NoError(t, err)

	_, err = uc.GetCustomer(ctx, created.ID)
	assert.Error(t, err)
}
