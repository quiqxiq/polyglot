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

	now := time.Now().Truncate(time.Second)
	inv := billing.Invoice{
		ID:            "inv-001",
		InvoiceNumber: "INV-2026-0001",
		CustomerID:    "cust-001",
		PeriodStart:   now,
		PeriodEnd:     now.AddDate(0, 1, 0),
		IssueDate:     now,
		DueDate:       now.AddDate(0, 0, 14),
		Status:        "draft",
		Subtotal:      150000,
		TaxAmount:     16500,
		TotalAmount:   166500,
		Items: []billing.InvoiceItem{
			{
				ID:          "item-001",
				InvoiceID:   "inv-001",
				ItemType:    "subscription_fee",
				Description: "Home 10M - July 2026",
				Quantity:    1,
				UnitPrice:   150000,
				Amount:      150000,
			},
		},
	}

	created, err := repo.Create(ctx, inv)
	require.NoError(t, err)
	assert.Equal(t, inv.InvoiceNumber, created.InvoiceNumber)
	assert.Len(t, created.Items, 1)

	found, err := repo.FindByID(ctx, inv.ID)
	require.NoError(t, err)
	assert.Equal(t, inv.InvoiceNumber, found.InvoiceNumber)
	assert.Equal(t, float64(166500), found.TotalAmount)
	assert.Len(t, found.Items, 1)
	assert.Equal(t, "subscription_fee", found.Items[0].ItemType)

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	updated := inv
	updated.Status = "issued"
	_, err = repo.Update(ctx, updated)
	require.NoError(t, err)

	found, err = repo.FindByID(ctx, inv.ID)
	require.NoError(t, err)
	assert.Equal(t, "issued", found.Status)

	require.NoError(t, repo.Delete(ctx, inv.ID))
	_, err = repo.FindByID(ctx, inv.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, billing.ErrNotFound)
}

func TestInvoiceRepository_FindByCustomer(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&invoiceModel{}, &invoiceItemModel{}))

	repo := NewInvoiceRepository(db)

	now := time.Now().Truncate(time.Second)
	inv1 := billing.Invoice{
		ID:            "inv-101",
		InvoiceNumber: "INV-2026-0101",
		CustomerID:    "cust-A",
		PeriodStart:   now,
		PeriodEnd:     now.AddDate(0, 1, 0),
		IssueDate:     now,
		DueDate:       now.AddDate(0, 0, 14),
		Status:        "draft",
		Subtotal:      100000,
		TotalAmount:   100000,
	}
	inv2 := billing.Invoice{
		ID:            "inv-102",
		InvoiceNumber: "INV-2026-0102",
		CustomerID:    "cust-A",
		PeriodStart:   now.AddDate(0, 1, 0),
		PeriodEnd:     now.AddDate(0, 2, 0),
		IssueDate:     now.AddDate(0, 1, 0),
		DueDate:       now.AddDate(0, 1, 14),
		Status:        "draft",
		Subtotal:      100000,
		TotalAmount:   100000,
	}
	inv3 := billing.Invoice{
		ID:            "inv-103",
		InvoiceNumber: "INV-2026-0103",
		CustomerID:    "cust-B",
		PeriodStart:   now,
		PeriodEnd:     now.AddDate(0, 1, 0),
		IssueDate:     now,
		DueDate:       now.AddDate(0, 0, 14),
		Status:        "draft",
		Subtotal:      200000,
		TotalAmount:   200000,
	}

	_, err = repo.Create(ctx, inv1)
	require.NoError(t, err)
	_, err = repo.Create(ctx, inv2)
	require.NoError(t, err)
	_, err = repo.Create(ctx, inv3)
	require.NoError(t, err)

	custA, err := repo.FindByCustomer(ctx, "cust-A")
	require.NoError(t, err)
	assert.Len(t, custA, 2)

	custB, err := repo.FindByCustomer(ctx, "cust-B")
	require.NoError(t, err)
	assert.Len(t, custB, 1)
}

func TestInvoiceRepository_FindByID_NotFound(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&invoiceModel{}, &invoiceItemModel{}))

	repo := NewInvoiceRepository(db)

	_, err = repo.FindByID(ctx, "non-existent")
	require.Error(t, err)
	assert.ErrorIs(t, err, billing.ErrNotFound)
}
