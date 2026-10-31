package customer_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/customer"
)

func TestGetCustomer360_AggregatesDomains(t *testing.T) {
	ctx := context.Background()
	customerRepo := mocktest.NewFakeCustomerRepo()
	subRepo := mocktest.NewFakeSubscriptionRepo()
	invRepo := mocktest.NewFakeInvoiceRepo()
	regRepo := mocktest.NewFakeRegistrationRepo()
	payments := mocktest.NewFakePaymentReader()

	require.NoError(t, customerRepo.Save(ctx, domainCustomer.Customer{
		ID: "cust-1", Name: "Budi", Phone: "085606846141", Status: "ACTIVE",
	}))
	require.NoError(t, subRepo.Save(ctx, domainSubscription.Subscription{
		ID: "sub-1", CustomerID: "cust-1", PlanID: "plan-1", ServiceType: "PPPOE",
		Status: domainSubscription.StatusActive, ProvisionStatus: domainSubscription.ProvisionOK,
		RouterProfile: "HOME-20M",
	}))
	require.NoError(t, invRepo.Save(ctx, domainBilling.Invoice{
		ID: "inv-1", CustomerID: "cust-1", InvoiceNumber: "INV-1", Total: 100000,
		Status: domainBilling.StatusUnpaid, DueDate: time.Now(),
	}))
	payments.Payments = []domainBilling.Payment{{ID: "pay-1", InvoiceID: "inv-1", Amount: 100000}}
	require.NoError(t, regRepo.Save(ctx, domainRegistration.Registration{
		ID: "reg-1", Phone: "085606846141", Status: domainRegistration.StatusActive,
		RegistrationNo: "REG-202609-0001",
	}))

	useCase := uc.NewManageCustomerUseCase(customerRepo, subRepo, invRepo, nil).
		With360(payments, regRepo)
	data, err := useCase.GetCustomer360(ctx, "cust-1")
	require.NoError(t, err)
	assert.Equal(t, "cust-1", data.Customer.ID)
	require.Len(t, data.Subscriptions, 1)
	assert.Equal(t, "HOME-20M", data.Subscriptions[0].RouterProfile)
	require.Len(t, data.Invoices, 1)
	require.Len(t, data.Payments, 1)
	require.Len(t, data.Registrations, 1)
	assert.Equal(t, "REG-202609-0001", data.Registrations[0].RegistrationNo)
}

func TestGetCustomer360_NotFound(t *testing.T) {
	customerRepo := mocktest.NewFakeCustomerRepo()
	useCase := uc.NewManageCustomerUseCase(customerRepo, nil, nil, nil)
	_, err := useCase.GetCustomer360(context.Background(), "tidak-ada")
	require.Error(t, err)
	assert.ErrorIs(t, err, domainCustomer.ErrCustomerNotFound)
}
