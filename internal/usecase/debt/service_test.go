package debt_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainDebt "github.com/quixiq/polyglot/internal/domain/debt"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	debtUC "github.com/quixiq/polyglot/internal/usecase/debt"
)

func newService() (*debtUC.Service, *mocktest.FakeDebtRepo, *mocktest.FakeCustomerRepo) {
	repo := mocktest.NewFakeDebtRepo()
	customs := mocktest.NewFakeCustomerRepo()
	notif := mocktest.NewFakeNotificationRepo()
	settings := mocktest.NewFakeSettingReader(map[string]string{})
	return debtUC.NewService(repo, customs, notif, settings), repo, customs
}

func seedCustomer(t *testing.T, repo *mocktest.FakeCustomerRepo, id string) {
	t.Helper()
	if err := repo.Save(context.Background(), domainCustomer.Customer{
		ID: id, TenantID: "tenant-default", Name: "Budi", Phone: "08123456789", Status: "ACTIVE",
	}); err != nil {
		t.Fatalf("seed customer: %v", err)
	}
}

func TestCreateAndPayDebtPartialThenFull(t *testing.T) {
	svc, repo, customs := newService()
	ctx := context.Background()
	seedCustomer(t, customs, "cust-1")

	d, err := svc.CreateDebt(ctx, debtUC.CreateDebtInput{
		CustomerID:  "cust-1",
		Amount:      150000,
		Description: "pinjaman",
	})
	if err != nil {
		t.Fatalf("create debt: %v", err)
	}
	if d.Status != domainDebt.StatusOpen || d.RemainingAmount != 150000 {
		t.Fatalf("unexpected debt state: %+v", d)
	}

	pay, err := svc.PayDebt(ctx, port.DebtPaymentCommand{
		DebtID:           d.ID,
		Amount:           50000,
		CashAccountID:    "ca-1001-kas-kantor",
		IncomeCategoryID: "cc-debt-income",
	})
	if err != nil {
		t.Fatalf("pay partial: %v", err)
	}
	if pay.Amount != 50000 {
		t.Fatalf("unexpected payment amount: %v", pay.Amount)
	}

	got, err := repo.FindByID(ctx, d.ID)
	if err != nil {
		t.Fatalf("find debt: %v", err)
	}
	if got.Status != domainDebt.StatusPartial || got.RemainingAmount != 100000 {
		t.Fatalf("expected PARTIAL/100000, got %s/%v", got.Status, got.RemainingAmount)
	}

	// Overpay ditolak.
	_, err = svc.PayDebt(ctx, port.DebtPaymentCommand{
		DebtID:           d.ID,
		Amount:           200000,
		CashAccountID:    "ca-1001-kas-kantor",
		IncomeCategoryID: "cc-debt-income",
	})
	if !errors.Is(err, domainDebt.ErrOverpayment) {
		t.Fatalf("expected overpayment error, got %v", err)
	}

	// Pelunasan sisa.
	if _, err := svc.PayDebt(ctx, port.DebtPaymentCommand{
		DebtID:           d.ID,
		Amount:           100000,
		CashAccountID:    "ca-1001-kas-kantor",
		IncomeCategoryID: "cc-debt-income",
	}); err != nil {
		t.Fatalf("pay remaining: %v", err)
	}
	got, _ = repo.FindByID(ctx, d.ID)
	if got.Status != domainDebt.StatusPaid || got.RemainingAmount != 0 {
		t.Fatalf("expected PAID/0, got %s/%v", got.Status, got.RemainingAmount)
	}

	// Pembayaran setelah lunas ditolak.
	_, err = svc.PayDebt(ctx, port.DebtPaymentCommand{
		DebtID:           d.ID,
		Amount:           1000,
		CashAccountID:    "ca-1001-kas-kantor",
		IncomeCategoryID: "cc-debt-income",
	})
	if !errors.Is(err, domainDebt.ErrAlreadyPaid) {
		t.Fatalf("expected already paid error, got %v", err)
	}
}

func TestCancelDebtGuardsPaidAndPartiallyPaid(t *testing.T) {
	svc, _, customs := newService()
	ctx := context.Background()
	seedCustomer(t, customs, "cust-2")

	d, err := svc.CreateDebt(ctx, debtUC.CreateDebtInput{CustomerID: "cust-2", Amount: 100000})
	if err != nil {
		t.Fatalf("create debt: %v", err)
	}

	// Tanpa pembayaran → boleh dibatalkan.
	cancelled, err := svc.CancelDebt(ctx, d.ID)
	if err != nil {
		t.Fatalf("cancel unpaid debt: %v", err)
	}
	if cancelled.Status != domainDebt.StatusCancelled {
		t.Fatalf("expected CANCELLED, got %s", cancelled.Status)
	}

	// Hutang dengan pembayaran sebagian → tidak boleh dibatalkan.
	d2, err := svc.CreateDebt(ctx, debtUC.CreateDebtInput{CustomerID: "cust-2", Amount: 100000})
	if err != nil {
		t.Fatalf("create second debt: %v", err)
	}
	if _, err := svc.PayDebt(ctx, port.DebtPaymentCommand{
		DebtID: d2.ID, Amount: 40000,
		CashAccountID: "ca-1001-kas-kantor", IncomeCategoryID: "cc-debt-income",
	}); err != nil {
		t.Fatalf("pay partial: %v", err)
	}
	_, err = svc.CancelDebt(ctx, d2.ID)
	if !errors.Is(err, domainDebt.ErrHasPayments) {
		t.Fatalf("expected has payments error, got %v", err)
	}
}

func TestCreateDebtValidation(t *testing.T) {
	svc, _, _ := newService()
	ctx := context.Background()

	if _, err := svc.CreateDebt(ctx, debtUC.CreateDebtInput{Amount: 1000}); !errors.Is(err, domainDebt.ErrInvalidInput) {
		t.Fatalf("expected invalid input for missing customer, got %v", err)
	}
	if _, err := svc.CreateDebt(ctx, debtUC.CreateDebtInput{CustomerID: "x", Amount: 0}); !errors.Is(err, domainDebt.ErrInvalidInput) {
		t.Fatalf("expected invalid input for zero amount, got %v", err)
	}
}

func TestExportDebtsCSV(t *testing.T) {
	svc, _, customs := newService()
	ctx := context.Background()
	seedCustomer(t, customs, "cust-3")
	if _, err := svc.CreateDebt(ctx, debtUC.CreateDebtInput{
		CustomerID: "cust-3", Amount: 75000, Description: "instalasi",
	}); err != nil {
		t.Fatalf("create debt: %v", err)
	}

	payload, contentType, filename, err := svc.ExportDebts(ctx, port.DebtFilter{}, "csv")
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if !strings.Contains(contentType, "csv") || !strings.HasSuffix(filename, ".csv") {
		t.Fatalf("unexpected export metadata: %s %s", contentType, filename)
	}
	body := string(payload)
	if !strings.Contains(body, "Remaining") || !strings.Contains(body, "cust-3") {
		t.Fatalf("csv missing expected content: %s", body)
	}
}
