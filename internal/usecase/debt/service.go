// Package debt mengorkestrasi hutang-piutang manual pelanggan (W5 PLAN-BILLING-OPS-V3).
package debt

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	domainDebt "github.com/quixiq/polyglot/internal/domain/debt"
	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
	domainTenant "github.com/quixiq/polyglot/internal/domain/tenant"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/phone"
	"github.com/xuri/excelize/v2"
)

// Service provides debt management operations.
type Service struct {
	repo     port.DebtRepository
	customs  port.CustomerRepository
	notif    port.NotificationRepository
	settings port.SettingReader
}

// NewService constructs a debt management service.
func NewService(
	repo port.DebtRepository,
	customs port.CustomerRepository,
	notif port.NotificationRepository,
	settings port.SettingReader,
) *Service {
	return &Service{
		repo:     repo,
		customs:  customs,
		notif:    notif,
		settings: settings,
	}
}

// CreateDebtInput is the input payload for creating a debt record.
type CreateDebtInput struct {
	TenantID       string
	CustomerID     string
	SubscriptionID *string
	Amount         float64
	Description    string
	DebtDate       time.Time
	DueDate        *time.Time
	CreatedBy      *uint
}

// CreateDebt creates a new debt for a customer.
func (s *Service) CreateDebt(ctx context.Context, in CreateDebtInput) (domainDebt.Debt, error) {
	if in.CustomerID == "" || in.Amount <= 0 {
		return domainDebt.Debt{}, fmt.Errorf("%w: customer_id and positive amount are required", domainDebt.ErrInvalidInput)
	}
	if in.DebtDate.IsZero() {
		in.DebtDate = time.Now().UTC()
	}
	if in.TenantID == "" {
		in.TenantID = domainTenant.DefaultID
	}
	if s.customs != nil {
		if _, err := s.customs.FindByID(ctx, in.CustomerID); err != nil {
			return domainDebt.Debt{}, fmt.Errorf("customer %s: %w", in.CustomerID, err)
		}
	}

	d := domainDebt.Debt{
		ID:              idgen.New("debt"),
		TenantID:        in.TenantID,
		CustomerID:      in.CustomerID,
		SubscriptionID:  in.SubscriptionID,
		Amount:          in.Amount,
		RemainingAmount: in.Amount,
		Description:     in.Description,
		DebtDate:        in.DebtDate,
		DueDate:         in.DueDate,
		Status:          domainDebt.StatusOpen,
		CreatedBy:       in.CreatedBy,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	if err := s.repo.Save(ctx, d); err != nil {
		return domainDebt.Debt{}, err
	}
	return d, nil
}

// UpdateDebt updates description or due date of an outstanding debt.
func (s *Service) UpdateDebt(ctx context.Context, id, description string, dueDate *time.Time) (domainDebt.Debt, error) {
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domainDebt.Debt{}, err
	}
	if d.Status == domainDebt.StatusPaid {
		return domainDebt.Debt{}, domainDebt.ErrAlreadyPaid
	}
	if d.Status == domainDebt.StatusCancelled {
		return domainDebt.Debt{}, domainDebt.ErrCancelled
	}

	d.Description = description
	d.DueDate = dueDate
	d.UpdatedAt = time.Now().UTC()
	if err := s.repo.Save(ctx, d); err != nil {
		return domainDebt.Debt{}, err
	}
	return d, nil
}

// CancelDebt marks an unpaid debt as cancelled.
func (s *Service) CancelDebt(ctx context.Context, id string) (domainDebt.Debt, error) {
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domainDebt.Debt{}, err
	}
	if d.Status == domainDebt.StatusPaid {
		return domainDebt.Debt{}, domainDebt.ErrAlreadyPaid
	}
	if d.RemainingAmount < d.Amount {
		return domainDebt.Debt{}, domainDebt.ErrHasPayments
	}

	d.Status = domainDebt.StatusCancelled
	d.UpdatedAt = time.Now().UTC()
	if err := s.repo.Save(ctx, d); err != nil {
		return domainDebt.Debt{}, err
	}
	return d, nil
}

// PayDebt applies a payment against a customer debt.
func (s *Service) PayDebt(ctx context.Context, cmd port.DebtPaymentCommand) (domainDebt.Payment, error) {
	if cmd.CashAccountID == "" {
		cmd.CashAccountID = "ca-1001-kas-kantor"
		if s.settings != nil {
			if v := s.settings.GetValue(ctx, "isp.manual_payment_cash_account_id", ""); v != "" {
				cmd.CashAccountID = v
			}
		}
	}
	if cmd.IncomeCategoryID == "" {
		cmd.IncomeCategoryID = "cc-debt-income"
		if s.settings != nil {
			if v := s.settings.GetValue(ctx, "isp.debt_income_category_id", "cc-debt-income"); v != "" {
				cmd.IncomeCategoryID = v
			}
		}
	}

	pay, err := s.repo.ApplyPayment(ctx, cmd)
	if err != nil {
		return domainDebt.Payment{}, err
	}

	// Cek jika lunas, kirim notifikasi DEBT_PAID
	if d, errDebt := s.repo.FindByID(ctx, cmd.DebtID); errDebt == nil && d.Status == domainDebt.StatusPaid && s.notif != nil && s.customs != nil {
		if cust, errCust := s.customs.FindByID(ctx, d.CustomerID); errCust == nil && cust.Phone != "" {
			_ = s.notif.Queue(ctx, domainNotification.WANotification{
				ID:             idgen.New("wa"),
				TenantID:       d.TenantID,
				CustomerID:     &cust.ID,
				RecipientPhone: phone.Normalize(cust.Phone),
				MessageType:    "DEBT_PAID",
				MessageContent: fmt.Sprintf("Terima kasih %s, pembayaran hutang sebesar Rp%.0f telah kami terima.", cust.Name, cmd.Amount),
				Status:         domainNotification.StatusQueued,
				CreatedAt:      time.Now().UTC(),
			})
		}
	}

	return pay, nil
}

// ListDebts returns debts matching filter.
func (s *Service) ListDebts(ctx context.Context, f port.DebtFilter) ([]domainDebt.Debt, int, error) {
	return s.repo.List(ctx, f)
}

// GetDebt retrieves a debt by ID.
func (s *Service) GetDebt(ctx context.Context, id string) (domainDebt.Debt, error) {
	return s.repo.FindByID(ctx, id)
}

// ListDebtPayments returns payments made for a debt.
func (s *Service) ListDebtPayments(ctx context.Context, debtID string) ([]domainDebt.Payment, error) {
	return s.repo.ListPayments(ctx, debtID)
}

// GetDebtSummary returns aggregate summary of debts for a customer.
func (s *Service) GetDebtSummary(ctx context.Context, tenantID, customerID string) (domainDebt.Summary, error) {
	return s.repo.Summary(ctx, tenantID, customerID, time.Now().UTC())
}

// ExportDebts exports filtered debts into CSV or XLSX format.
func (s *Service) ExportDebts(ctx context.Context, f port.DebtFilter, format string) ([]byte, string, string, error) {
	f.Limit = 10000
	debts, _, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, "", "", err
	}

	headers := []string{"ID", "Customer ID", "Amount", "Remaining", "Status", "Debt Date", "Due Date", "Description"}
	rows := make([][]string, len(debts))
	for i, d := range debts {
		dueStr := "-"
		if d.DueDate != nil {
			dueStr = d.DueDate.Format("2006-01-02")
		}
		rows[i] = []string{
			d.ID,
			d.CustomerID,
			fmt.Sprintf("%.2f", d.Amount),
			fmt.Sprintf("%.2f", d.RemainingAmount),
			d.Status,
			d.DebtDate.Format("2006-01-02"),
			dueStr,
			d.Description,
		}
	}

	filename := fmt.Sprintf("debts-%s", time.Now().Format("20060102150405"))
	if strings.ToLower(format) == "xlsx" {
		f := excelize.NewFile()
		defer func() { _ = f.Close() }()
		sheet := "Debts"
		f.SetSheetName("Sheet1", sheet)
		for colIdx, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
			_ = f.SetCellValue(sheet, cell, h)
		}
		for rowIdx, r := range rows {
			for colIdx, val := range r {
				cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
				_ = f.SetCellValue(sheet, cell, val)
			}
		}
		var buf bytes.Buffer
		if err := f.Write(&buf); err != nil {
			return nil, "", "", err
		}
		return buf.Bytes(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", filename + ".xlsx", nil
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write(headers)
	_ = w.WriteAll(rows)
	w.Flush()
	return buf.Bytes(), "text/csv", filename + ".csv", nil
}
