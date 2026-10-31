// Package collection mengelola efektivitas penagihan piutang dan penanganan pelanggan overdue (W6).
package collection

import (
	"context"
	"fmt"
	"time"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/phone"
)

// Service provides debt collection and overdue management operations.
type Service struct {
	invoices   port.InvoiceRepository
	debts      port.DebtRepository
	customs    port.CustomerRepository
	subs       port.SubscriptionRepository
	accountMgr port.RouterAccountManager
	notif      port.NotificationRepository
	settings   port.SettingReader
}

// NewService constructs a collection service.
func NewService(
	invoices port.InvoiceRepository,
	debts port.DebtRepository,
	customs port.CustomerRepository,
	subs port.SubscriptionRepository,
	accountMgr port.RouterAccountManager,
	notif port.NotificationRepository,
	settings port.SettingReader,
) *Service {
	return &Service{
		invoices:   invoices,
		debts:      debts,
		customs:    customs,
		subs:       subs,
		accountMgr: accountMgr,
		notif:      notif,
		settings:   settings,
	}
}

// Summary merangkum metrik penagihan piutang.
type Summary struct {
	TotalOutstanding float64 `json:"total_outstanding"`
	OverdueInvoices  int     `json:"overdue_invoices"`
	OverdueDebts     int     `json:"overdue_debts"`
	OverdueCustomers int     `json:"overdue_customers"`
}

// GetSummary calculates collection summary statistics.
func (s *Service) GetSummary(ctx context.Context, tenantID string) (Summary, error) {
	if tenantID == "" {
		tenantID = "tenant-default"
	}
	var res Summary

	if fRepo, ok := s.invoices.(port.InvoiceFilteredRepository); ok {
		summary, err := fRepo.Summary(ctx, port.InvoiceFilter{
			TenantID:    tenantID,
			OverdueOnly: true,
		})
		if err == nil {
			res.TotalOutstanding += summary.TotalOutstanding
			res.OverdueInvoices = summary.Count
		}
	}

	if s.debts != nil {
		debtSum, err := s.debts.Summary(ctx, tenantID, "", time.Now().UTC())
		if err == nil {
			res.TotalOutstanding += debtSum.Outstanding
			res.OverdueDebts = debtSum.Overdue
		}
	}

	return res, nil
}

// OverdueCustomer represents a customer with past due invoices.
type OverdueCustomer struct {
	CustomerID   string    `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	Phone        string    `json:"phone"`
	Address      string    `json:"address"`
	InvoiceCount int       `json:"invoice_count"`
	Outstanding  float64   `json:"outstanding"`
	OldestDue    time.Time `json:"oldest_due"`
	DaysOverdue  int       `json:"days_overdue"`
	Status       string    `json:"status"`
}

// ListOverdueCustomers returns customers with overdue balance.
func (s *Service) ListOverdueCustomers(ctx context.Context, tenantID string, limit int) ([]OverdueCustomer, error) {
	if tenantID == "" {
		tenantID = "tenant-default"
	}
	fRepo, ok := s.invoices.(port.InvoiceFilteredRepository)
	if !ok {
		return nil, nil
	}

	now := time.Now().UTC()
	candidates, err := fRepo.FindOverdueCandidates(ctx, tenantID, now, limit*5)
	if err != nil {
		return nil, err
	}

	custMap := make(map[string]*OverdueCustomer)
	for _, inv := range candidates {
		custID := inv.CustomerID
		entry, exists := custMap[custID]
		if !exists {
			custName := custID
			var custPhone, custAddr, custStatus string
			if s.customs != nil {
				if c, err := s.customs.FindByID(ctx, custID); err == nil {
					custName = c.Name
					custPhone = c.Phone
					custAddr = c.Address
					custStatus = c.Status
				}
			}
			entry = &OverdueCustomer{
				CustomerID:   custID,
				CustomerName: custName,
				Phone:        custPhone,
				Address:      custAddr,
				OldestDue:    inv.DueDate,
				Status:       custStatus,
			}
			custMap[custID] = entry
		}

		entry.InvoiceCount++
		entry.Outstanding += (inv.Total - inv.PaidAmount)
		if inv.DueDate.Before(entry.OldestDue) {
			entry.OldestDue = inv.DueDate
		}
	}

	out := make([]OverdueCustomer, 0, len(custMap))
	for _, entry := range custMap {
		entry.DaysOverdue = int(now.Sub(entry.OldestDue).Hours() / 24)
		out = append(out, *entry)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

// SendReminderBulk sends overdue reminders to selected customers.
func (s *Service) SendReminderBulk(ctx context.Context, customerIDs []string) (sent int, failed int, err error) {
	if s.notif == nil || s.customs == nil {
		return 0, 0, nil
	}
	for _, cid := range customerIDs {
		cust, err := s.customs.FindByID(ctx, cid)
		if err != nil || cust.Phone == "" {
			failed++
			continue
		}
		invoices, _ := s.invoices.FindByCustomerID(ctx, cid)
		var overdueTotal float64
		var oldestPeriod string
		for _, inv := range invoices {
			if inv.IsOutstanding() && inv.DueDate.Before(time.Now().UTC()) {
				overdueTotal += inv.Outstanding()
				if oldestPeriod == "" {
					oldestPeriod = inv.Period
				}
			}
		}
		if overdueTotal <= 0 {
			continue
		}

		content := fmt.Sprintf("Yth %s, Anda memiliki tagihan tertunggak periode %s sebesar Rp%.0f. Mohon segera melakukan pembayaran agar layanan tetap aktif.", cust.Name, oldestPeriod, overdueTotal)
		qerr := s.notif.Queue(ctx, domainNotification.WANotification{
			ID:             idgen.New("wa"),
			TenantID:       cust.TenantID,
			CustomerID:     &cust.ID,
			RecipientPhone: phone.Normalize(cust.Phone),
			MessageType:    "BILL_OVERDUE",
			MessageContent: content,
			Status:         domainNotification.StatusQueued,
			CreatedAt:      time.Now().UTC(),
		})
		if qerr != nil {
			failed++
		} else {
			sent++
		}
	}
	return sent, failed, nil
}

// IsolateBulk triggers isolation for active subscriptions of given customers.
func (s *Service) IsolateBulk(ctx context.Context, customerIDs []string) (isolated int, failed int, err error) {
	if s.accountMgr == nil || s.subs == nil {
		return 0, 0, nil
	}
	cfg := port.LoadISPSettings(ctx, s.settings)
	opt := port.IsolationOptions{
		AddressList:   cfg.IsolirAddressList,
		IsolirProfile: cfg.PPPoEIsolirProfile,
	}

	for _, cid := range customerIDs {
		subs, _ := s.subs.FindByCustomerID(ctx, cid)
		for _, sub := range subs {
			if sub.Status != domainSubscription.StatusActive {
				continue
			}
			devID := ""
			if sub.DeviceID != nil {
				devID = *sub.DeviceID
			}
			if devID == "" {
				continue
			}
			if err := s.accountMgr.Isolate(ctx, devID, sub.ServiceType, sub.RemoteUsername, opt); err != nil {
				failed++
			} else {
				_ = s.subs.UpdateStatus(ctx, sub.ID, domainSubscription.StatusIsolated)
				isolated++
			}
		}
	}
	return isolated, failed, nil
}

// RestoreBulk un-isolates subscriptions of given customers.
func (s *Service) RestoreBulk(ctx context.Context, customerIDs []string) (restored int, failed int, err error) {
	if s.accountMgr == nil || s.subs == nil {
		return 0, 0, nil
	}
	cfg := port.LoadISPSettings(ctx, s.settings)

	for _, cid := range customerIDs {
		subs, _ := s.subs.FindByCustomerID(ctx, cid)
		for _, sub := range subs {
			if sub.Status != domainSubscription.StatusIsolated {
				continue
			}
			devID := ""
			if sub.DeviceID != nil {
				devID = *sub.DeviceID
			}
			if devID == "" {
				continue
			}
			profile := sub.RouterProfile
			if profile == "" {
				profile = "default"
			}
			if err := s.accountMgr.Restore(ctx, devID, sub.ServiceType, sub.RemoteUsername, profile, cfg.IsolirAddressList); err != nil {
				failed++
			} else {
				_ = s.subs.UpdateStatus(ctx, sub.ID, domainSubscription.StatusActive)
				restored++
			}
		}
	}
	return restored, failed, nil
}

var _ = domainBilling.InvoiceTypeManual
var _ = domainCustomer.Customer{}
