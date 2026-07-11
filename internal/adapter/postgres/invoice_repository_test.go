package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/testutil"
)

func TestInvoiceRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&invoiceModel{}, &invoiceItemModel{}))

	repo := NewInvoiceRepository(db)

	inv := billing.Invoice{
		ID:            "inv-001",
		InvoiceNumber: "INV-2024001",
		CustomerID:    "cust-001",
		PeriodStart:   time.Now(),
		PeriodEnd:     time.Now().AddDate(0, 1, 0),
		IssueDate:     time.Now(),
		DueDate:       time.Now().AddDate(0, 0, 14),
		Status:        "draft",
		Subtotal:      100000,
		TaxAmount:     10000,
		TotalAmount:   110000,
		Items: []billing.InvoiceItem{
			{
				ID:          "item-001",
				ItemType:    "subscription_fee",
				Description: "Monthly internet",
				Quantity:    1,
				UnitPrice:   100000,
				Amount:      100000,
			},
		},
	}

	created, err := repo.Create(ctx, inv)
	require.NoError(t, err)
	assert.Equal(t, inv.InvoiceNumber, created.InvoiceNumber)
	assert.Len(t, created.Items, 1)
	assert.NotEmpty(t, created.Items[0].InvoiceID)

	found, err := repo.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.InvoiceNumber, found.InvoiceNumber)
	assert.Len(t, found.Items, 1)

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	byCustomer, err := repo.FindByCustomer(ctx, inv.CustomerID)
	require.NoError(t, err)
	assert.Len(t, byCustomer, 1)

	updated := found
	updated.Status = "issued"
	updated.Items = append(updated.Items, billing.InvoiceItem{
		ID:          "item-002",
		ItemType:    "late_fee",
		Description: "Late fee",
		Quantity:    1,
		UnitPrice:   10000,
		Amount:      10000,
	})
	_, err = repo.Update(ctx, updated)
	require.NoError(t, err)

	found, err = repo.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "issued", found.Status)
	assert.Len(t, found.Items, 2)

	require.NoError(t, repo.Delete(ctx, created.ID))
	_, err = repo.FindByID(ctx, created.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, billing.ErrNotFound)
}

func TestInvoiceRepository_FindByID_NotFound(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&invoiceModel{}, &invoiceItemModel{}))

	repo := NewInvoiceRepository(db)

	_, err = repo.FindByID(ctx, "missing")
	require.Error(t, err)
	assert.ErrorIs(t, err, billing.ErrNotFound)
}
