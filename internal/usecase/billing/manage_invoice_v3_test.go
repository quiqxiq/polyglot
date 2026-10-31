package billing_test

import (
	"context"
	"strings"
	"testing"
	"time"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
)

func newInvoiceUCWithSettings(t *testing.T) (*billingUC.InvoiceUseCase, *mocktest.FakeInvoiceRepo) {
	t.Helper()
	repo := mocktest.NewFakeInvoiceRepo()
	customs := mocktest.NewFakeCustomerRepo()
	notif := mocktest.NewFakeNotificationRepo()
	settings := mocktest.NewFakeSettingReader(map[string]string{
		"isp.tax_percent": "0",
	})
	if err := customs.Save(context.Background(), domainCustomer.Customer{
		ID: "cust-1", TenantID: "tenant-default", Name: "Siti", Phone: "0812000000", Status: "ACTIVE",
	}); err != nil {
		t.Fatalf("seed customer: %v", err)
	}
	uc := billingUC.NewInvoiceUseCase(repo).
		WithSettings(settings).
		WithNotifications(notif, customs)
	return uc, repo
}

func TestCreateManualInvoiceAppliesTax(t *testing.T) {
	uc, repo := newInvoiceUCWithSettings(t)
	due := time.Now().UTC().AddDate(0, 0, 7)
	inv, err := uc.CreateManualInvoice(context.Background(), billingUC.CreateManualInvoiceInput{
		CustomerID: "cust-1",
		DueDate:    due,
		TaxPercent: 11,
		Items: []billingUC.CreateInvoiceItemInput{
			{Description: "Instalasi", Quantity: 1, UnitPrice: 100000},
			{Description: "Kabel", Quantity: 2, UnitPrice: 25000},
		},
		Notes: "PSB baru",
	})
	if err != nil {
		t.Fatalf("create manual invoice: %v", err)
	}
	if inv.InvoiceType != domainBilling.InvoiceTypeManual {
		t.Fatalf("expected MANUAL invoice type, got %s", inv.InvoiceType)
	}
	if inv.Subtotal != 150000 {
		t.Fatalf("expected subtotal 150000, got %v", inv.Subtotal)
	}
	if inv.TaxPercent != 11 || inv.TaxAmount != 16500 {
		t.Fatalf("expected tax 11%%/16500, got %v/%v", inv.TaxPercent, inv.TaxAmount)
	}
	if inv.Total != 166500 {
		t.Fatalf("expected total 166500, got %v", inv.Total)
	}
	if inv.Status != domainBilling.StatusUnpaid || inv.InvoiceNumber == "" || inv.ManualPaymentCode == "" {
		t.Fatalf("unexpected invoice: %+v", inv)
	}
	if inv.Outstanding() != 166500 {
		t.Fatalf("expected outstanding 166500, got %v", inv.Outstanding())
	}

	stored, err := repo.FindByID(context.Background(), inv.ID)
	if err != nil {
		t.Fatalf("find stored invoice: %v", err)
	}
	if stored.Total != inv.Total {
		t.Fatalf("stored total mismatch: %v", stored.Total)
	}
}

func TestCreateManualInvoiceValidation(t *testing.T) {
	uc, _ := newInvoiceUCWithSettings(t)
	ctx := context.Background()

	if _, err := uc.CreateManualInvoice(ctx, billingUC.CreateManualInvoiceInput{
		Items: []billingUC.CreateInvoiceItemInput{{Description: "x", Quantity: 1, UnitPrice: 1000}},
	}); err == nil {
		t.Fatal("expected error for missing customer")
	}
	if _, err := uc.CreateManualInvoice(ctx, billingUC.CreateManualInvoiceInput{
		CustomerID: "cust-1",
	}); err == nil {
		t.Fatal("expected error for empty items")
	}
}

func TestExportInvoicesCSV(t *testing.T) {
	uc, _ := newInvoiceUCWithSettings(t)
	ctx := context.Background()
	if _, err := uc.CreateManualInvoice(ctx, billingUC.CreateManualInvoiceInput{
		CustomerID: "cust-1",
		Items:      []billingUC.CreateInvoiceItemInput{{Description: "Add-on", Quantity: 1, UnitPrice: 50000}},
	}); err != nil {
		t.Fatalf("create invoice: %v", err)
	}

	payload, contentType, filename, err := uc.ExportInvoices(ctx, port.InvoiceFilter{TenantID: "tenant-default", CustomerID: "cust-1"}, "csv")
	if err != nil {
		t.Fatalf("export invoices: %v", err)
	}
	if !strings.Contains(contentType, "csv") || !strings.HasSuffix(filename, ".csv") {
		t.Fatalf("unexpected export metadata: %s %s", contentType, filename)
	}
	body := string(payload)
	if !strings.Contains(body, "No. Invoice") || !strings.Contains(body, "cust-1") {
		t.Fatalf("csv missing expected content: %s", body)
	}
}
