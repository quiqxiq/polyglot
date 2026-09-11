// F3-8: delete customer/subscription/invoice harus soft delete — baris tetap
// ada untuk audit tetapi tidak lagi terlihat oleh query normal.
package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
)

func TestSoftDelete_CustomerSubscriptionInvoice(t *testing.T) {
	db := setupISPDB(t)
	ctx := context.Background()
	custRepo := postgres.NewCustomerRepository(db)
	subRepo := postgres.NewSubscriptionRepository(db, fakeVaultForTest{})
	invRepo := postgres.NewInvoiceRepository(db)

	cust := domainCustomer.Customer{
		ID: "cust-sd", TenantID: "tenant-default", CustomerCode: "CUST-SD",
		Name: "Soft", Phone: "081200000000", Address: "Jl. Soft",
		PortalAccessCode: "11112222", Status: domainCustomer.StatusActive,
	}
	require.NoError(t, custRepo.Save(ctx, cust))

	subID := "sub-sd"
	require.NoError(t, subRepo.Save(ctx, domainSubscription.Subscription{
		ID: subID, TenantID: "tenant-default", CustomerID: cust.ID, PlanID: "plan-1",
		RemoteUsername: "sd1", RemotePassword: "pw", Status: domainSubscription.StatusActive,
	}))
	require.NoError(t, invRepo.Save(ctx, domainBilling.Invoice{
		ID: "inv-sd", TenantID: "tenant-default", InvoiceNumber: "INV-SD-1",
		CustomerID: cust.ID, SubscriptionID: &subID, Period: "2026-09", Total: 100000,
		Status: domainBilling.StatusUnpaid, QRPayload: "polyglot://invoice/inv-sd",
		ManualPaymentCode: "PAY-SD1",
	}))

	require.NoError(t, invRepo.Delete(ctx, "inv-sd"))
	require.NoError(t, subRepo.Delete(ctx, subID))
	require.NoError(t, custRepo.Delete(ctx, cust.ID))

	_, err := invRepo.FindByID(ctx, "inv-sd")
	assert.ErrorIs(t, err, postgres.ErrNotFound)
	_, err = subRepo.FindByID(ctx, subID)
	assert.ErrorIs(t, err, postgres.ErrNotFound)
	_, err = custRepo.FindByID(ctx, cust.ID)
	assert.ErrorIs(t, err, postgres.ErrNotFound)

	// Baris tetap ada untuk audit.
	var custCount, subCount, invCount int64
	require.NoError(t, db.Model(&model.CustomerModel{}).Count(&custCount).Error)
	require.NoError(t, db.Model(&model.SubscriptionModel{}).Count(&subCount).Error)
	require.NoError(t, db.Model(&model.InvoiceModel{}).Count(&invCount).Error)
	assert.Equal(t, int64(1), custCount, "customer soft-deleted tetap ada")
	assert.Equal(t, int64(1), subCount, "subscription soft-deleted tetap ada")
	assert.Equal(t, int64(1), invCount, "invoice soft-deleted tetap ada")
}
