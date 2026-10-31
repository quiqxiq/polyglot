package mocktest

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/domain/debt"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// FakePaymentMethodRepo is in-memory fake for port.PaymentMethodRepository.
type FakePaymentMethodRepo struct {
	mu      sync.Mutex
	methods map[string]billing.PaymentMethod
}

var _ port.PaymentMethodRepository = (*FakePaymentMethodRepo)(nil)

func NewFakePaymentMethodRepo() *FakePaymentMethodRepo {
	return &FakePaymentMethodRepo{methods: make(map[string]billing.PaymentMethod)}
}

func (f *FakePaymentMethodRepo) Save(_ context.Context, m billing.PaymentMethod) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if m.ID == "" {
		m.ID = idgen.New("pm")
	}
	f.methods[m.ID] = m
	return nil
}

func (f *FakePaymentMethodRepo) FindByID(_ context.Context, id string) (billing.PaymentMethod, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.methods[id]
	if !ok {
		return billing.PaymentMethod{}, billing.ErrNotFound
	}
	return m, nil
}

func (f *FakePaymentMethodRepo) List(_ context.Context, tenantID string, activeOnly bool) ([]billing.PaymentMethod, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []billing.PaymentMethod
	for _, m := range f.methods {
		if tenantID != "" && m.TenantID != tenantID {
			continue
		}
		if activeOnly && !m.IsActive {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

func (f *FakePaymentMethodRepo) SetActive(_ context.Context, id string, active bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.methods[id]
	if !ok {
		return billing.ErrNotFound
	}
	m.IsActive = active
	f.methods[id] = m
	return nil
}

// FakeDebtRepo is in-memory fake for port.DebtRepository.
type FakeDebtRepo struct {
	mu       sync.Mutex
	debts    map[string]debt.Debt
	payments map[string][]debt.Payment
}

var _ port.DebtRepository = (*FakeDebtRepo)(nil)

func NewFakeDebtRepo() *FakeDebtRepo {
	return &FakeDebtRepo{
		debts:    make(map[string]debt.Debt),
		payments: make(map[string][]debt.Payment),
	}
}

func (f *FakeDebtRepo) Save(_ context.Context, d debt.Debt) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if d.ID == "" {
		d.ID = idgen.New("debt")
	}
	f.debts[d.ID] = d
	return nil
}

func (f *FakeDebtRepo) FindByID(_ context.Context, id string) (debt.Debt, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	d, ok := f.debts[id]
	if !ok {
		return debt.Debt{}, debt.ErrNotFound
	}
	return d, nil
}

func (f *FakeDebtRepo) List(_ context.Context, filter port.DebtFilter) ([]debt.Debt, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []debt.Debt
	for _, d := range f.debts {
		if filter.CustomerID != "" && d.CustomerID != filter.CustomerID {
			continue
		}
		if filter.Status != "" && d.Status != filter.Status {
			continue
		}
		if filter.OverdueOnly {
			if d.Status != debt.StatusOpen && d.Status != debt.StatusPartial {
				continue
			}
			if d.DueDate == nil || !d.DueDate.Before(time.Now().UTC()) {
				continue
			}
		}
		out = append(out, d)
	}
	return out, len(out), nil
}

func (f *FakeDebtRepo) ListByCustomer(_ context.Context, customerID string) ([]debt.Debt, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []debt.Debt
	for _, d := range f.debts {
		if d.CustomerID == customerID {
			out = append(out, d)
		}
	}
	return out, nil
}

func (f *FakeDebtRepo) ApplyPayment(_ context.Context, cmd port.DebtPaymentCommand) (debt.Payment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	d, ok := f.debts[cmd.DebtID]
	if !ok {
		return debt.Payment{}, debt.ErrNotFound
	}
	if d.Status == debt.StatusPaid || d.RemainingAmount <= 0 {
		return debt.Payment{}, debt.ErrAlreadyPaid
	}
	if d.Status == debt.StatusCancelled {
		return debt.Payment{}, debt.ErrCancelled
	}
	if cmd.Amount > d.RemainingAmount+1e-9 {
		return debt.Payment{}, debt.ErrOverpayment
	}

	p := debt.Payment{
		ID:              idgen.New("dp"),
		TenantID:        cmd.TenantID,
		DebtID:          cmd.DebtID,
		Amount:          cmd.Amount,
		PaymentMethodID: cmd.PaymentMethodID,
		PaymentDate:     cmd.PaymentDate,
		Notes:           cmd.Notes,
		RecordedBy:      cmd.RecordedBy,
		CreatedAt:       time.Now().UTC(),
	}
	f.payments[cmd.DebtID] = append(f.payments[cmd.DebtID], p)

	d.RemainingAmount -= cmd.Amount
	if d.RemainingAmount <= 1e-9 {
		d.RemainingAmount = 0
		d.Status = debt.StatusPaid
		now := time.Now().UTC()
		d.PaidAt = &now
	} else {
		d.Status = debt.StatusPartial
	}
	f.debts[cmd.DebtID] = d
	return p, nil
}

func (f *FakeDebtRepo) ListPayments(_ context.Context, debtID string) ([]debt.Payment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.payments[debtID], nil
}

func (f *FakeDebtRepo) Summary(_ context.Context, tenantID, customerID string, asOf time.Time) (debt.Summary, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var s debt.Summary
	for _, d := range f.debts {
		if customerID != "" && d.CustomerID != customerID {
			continue
		}
		if d.Status == debt.StatusCancelled {
			continue
		}
		s.Count++
		s.TotalAmount += d.Amount
		s.Outstanding += d.RemainingAmount
		if (d.Status == debt.StatusOpen || d.Status == debt.StatusPartial) && d.DueDate != nil && d.DueDate.Before(asOf) {
			s.Overdue++
		}
	}
	return s, nil
}

// FakeBulkPaymentRepo is in-memory fake for port.BulkPaymentRepository.
type FakeBulkPaymentRepo struct {
	mu      sync.Mutex
	batches map[string]billing.BulkPaymentBatch
	items   map[string][]billing.BulkPaymentItem
}

var _ port.BulkPaymentRepository = (*FakeBulkPaymentRepo)(nil)

func NewFakeBulkPaymentRepo() *FakeBulkPaymentRepo {
	return &FakeBulkPaymentRepo{
		batches: make(map[string]billing.BulkPaymentBatch),
		items:   make(map[string][]billing.BulkPaymentItem),
	}
}

func (f *FakeBulkPaymentRepo) SaveBatch(_ context.Context, b billing.BulkPaymentBatch) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if b.ID == "" {
		b.ID = idgen.New("bpb")
	}
	f.batches[b.ID] = b
	return nil
}

func (f *FakeBulkPaymentRepo) FindBatchByID(_ context.Context, id string) (billing.BulkPaymentBatch, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	b, ok := f.batches[id]
	if !ok {
		return billing.BulkPaymentBatch{}, billing.ErrNotFound
	}
	return b, nil
}

func (f *FakeBulkPaymentRepo) ListBatches(_ context.Context, tenantID string, limit int) ([]billing.BulkPaymentBatch, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []billing.BulkPaymentBatch
	for _, b := range f.batches {
		if tenantID != "" && b.TenantID != tenantID {
			continue
		}
		out = append(out, b)
	}
	return out, nil
}

func (f *FakeBulkPaymentRepo) UpdateBatchStatus(_ context.Context, id, status string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	b, ok := f.batches[id]
	if !ok {
		return billing.ErrNotFound
	}
	b.Status = status
	f.batches[id] = b
	return nil
}

func (f *FakeBulkPaymentRepo) SaveItems(_ context.Context, items []billing.BulkPaymentItem) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, it := range items {
		if it.ID == "" {
			it.ID = idgen.New("bpi")
		}
		f.items[it.BatchID] = append(f.items[it.BatchID], it)
	}
	return nil
}

func (f *FakeBulkPaymentRepo) ListItems(_ context.Context, batchID string) ([]billing.BulkPaymentItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.items[batchID], nil
}

func (f *FakeBulkPaymentRepo) ClaimNextQueuedItem(_ context.Context, batchID string) (billing.BulkPaymentItem, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	items := f.items[batchID]
	for i, it := range items {
		if it.Status == billing.BulkItemQueued {
			items[i].Status = billing.BulkItemProcessing
			f.items[batchID] = items
			return items[i], true, nil
		}
	}
	return billing.BulkPaymentItem{}, false, nil
}

func (f *FakeBulkPaymentRepo) UpdateItem(_ context.Context, item billing.BulkPaymentItem) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	items := f.items[item.BatchID]
	for i, it := range items {
		if it.ID == item.ID {
			items[i] = item
			f.items[item.BatchID] = items
			return nil
		}
	}
	return billing.ErrNotFound
}

func (f *FakeBulkPaymentRepo) CountItems(_ context.Context, batchID string) (queued, processing, succeeded, failed int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, it := range f.items[batchID] {
		switch it.Status {
		case billing.BulkItemQueued:
			queued++
		case billing.BulkItemProcessing:
			processing++
		case billing.BulkItemSucceeded:
			succeeded++
		case billing.BulkItemFailed:
			failed++
		}
	}
	return queued, processing, succeeded, failed, nil
}

// FakeGatewaySettlementRepo is in-memory fake for port.GatewaySettlementRepository.
type FakeGatewaySettlementRepo struct {
	mu          sync.Mutex
	settlements map[string]billing.GatewaySettlement
}

var _ port.GatewaySettlementRepository = (*FakeGatewaySettlementRepo)(nil)

func NewFakeGatewaySettlementRepo() *FakeGatewaySettlementRepo {
	return &FakeGatewaySettlementRepo{settlements: make(map[string]billing.GatewaySettlement)}
}

func (f *FakeGatewaySettlementRepo) Save(_ context.Context, s billing.GatewaySettlement) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if s.ID == "" {
		s.ID = idgen.New("gws")
	}
	f.settlements[s.ID] = s
	return nil
}

func (f *FakeGatewaySettlementRepo) FindByID(_ context.Context, id string) (billing.GatewaySettlement, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.settlements[id]
	if !ok {
		return billing.GatewaySettlement{}, billing.ErrNotFound
	}
	return s, nil
}

func (f *FakeGatewaySettlementRepo) FindByRef(_ context.Context, tenantID, gateway, ref string) (billing.GatewaySettlement, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, s := range f.settlements {
		if s.TenantID == tenantID && s.Gateway == gateway && s.SettlementRef == ref {
			return s, nil
		}
	}
	return billing.GatewaySettlement{}, billing.ErrNotFound
}

func (f *FakeGatewaySettlementRepo) List(_ context.Context, filter port.GatewaySettlementFilter) ([]billing.GatewaySettlement, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []billing.GatewaySettlement
	for _, s := range f.settlements {
		if filter.TenantID != "" && s.TenantID != filter.TenantID {
			continue
		}
		if filter.Gateway != "" && s.Gateway != filter.Gateway {
			continue
		}
		if filter.Status != "" && s.Status != filter.Status {
			continue
		}
		out = append(out, s)
	}
	return out, nil
}

// FakeInvoiceFilteredRepo is in-memory fake for port.InvoiceFilteredRepository.
type FakeInvoiceFilteredRepo struct {
	mu       sync.Mutex
	invoices map[string]billing.Invoice
}

var _ port.InvoiceFilteredRepository = (*FakeInvoiceFilteredRepo)(nil)

func NewFakeInvoiceFilteredRepo() *FakeInvoiceFilteredRepo {
	return &FakeInvoiceFilteredRepo{invoices: make(map[string]billing.Invoice)}
}

func (f *FakeInvoiceFilteredRepo) Seed(inv billing.Invoice) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.invoices[inv.ID] = inv
}

func (f *FakeInvoiceFilteredRepo) FindFiltered(_ context.Context, filter port.InvoiceFilter) ([]billing.Invoice, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []billing.Invoice
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	for _, inv := range f.invoices {
		if filter.CustomerID != "" && inv.CustomerID != filter.CustomerID {
			continue
		}
		if filter.Status != "" && inv.Status != filter.Status {
			continue
		}
		if filter.Period != "" && inv.Period != filter.Period {
			continue
		}
		if filter.OverdueOnly {
			if (inv.Status == billing.StatusUnpaid || inv.Status == billing.StatusPartial || inv.Status == billing.StatusOverdue) && inv.DueDate.Before(today) {
				// matched
			} else {
				continue
			}
		}
		if filter.UnpaidOnly {
			if inv.Status != billing.StatusUnpaid && inv.Status != billing.StatusPartial && inv.Status != billing.StatusOverdue {
				continue
			}
		}
		if filter.Search != "" {
			if !strings.Contains(strings.ToLower(inv.InvoiceNumber), strings.ToLower(filter.Search)) {
				continue
			}
		}
		out = append(out, inv)
	}
	return out, len(out), nil
}

func (f *FakeInvoiceFilteredRepo) Summary(_ context.Context, filter port.InvoiceFilter) (port.InvoiceSummary, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var s port.InvoiceSummary
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	for _, inv := range f.invoices {
		if filter.Status != "" && inv.Status != filter.Status {
			continue
		}
		s.Count++
		s.TotalBilled += inv.Total
		s.TotalPaid += inv.PaidAmount
		outstanding := inv.Total - inv.PaidAmount
		s.TotalOutstanding += outstanding
		if (inv.Status == billing.StatusUnpaid || inv.Status == billing.StatusPartial || inv.Status == billing.StatusOverdue) && inv.DueDate.Before(today) {
			s.TotalOverdue += outstanding
		}
	}
	return s, nil
}

func (f *FakeInvoiceFilteredRepo) FindOverdueCandidates(_ context.Context, tenantID string, now time.Time, limit int) ([]billing.Invoice, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []billing.Invoice
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	for _, inv := range f.invoices {
		if (inv.Status == billing.StatusUnpaid || inv.Status == billing.StatusPartial) && inv.DueDate.Before(today) {
			out = append(out, inv)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

// FakeGatewayTransactionFilteredRepo is in-memory fake for port.GatewayTransactionFilteredRepository.
type FakeGatewayTransactionFilteredRepo struct {
	mu  sync.Mutex
	txs map[string]billing.GatewayTransaction
}

var _ port.GatewayTransactionFilteredRepository = (*FakeGatewayTransactionFilteredRepo)(nil)

func NewFakeGatewayTransactionFilteredRepo() *FakeGatewayTransactionFilteredRepo {
	return &FakeGatewayTransactionFilteredRepo{txs: make(map[string]billing.GatewayTransaction)}
}

func (f *FakeGatewayTransactionFilteredRepo) Seed(tx billing.GatewayTransaction) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.txs[tx.ID] = tx
}

func (f *FakeGatewayTransactionFilteredRepo) ListFiltered(_ context.Context, filter port.GatewayTransactionFilter) ([]billing.GatewayTransaction, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []billing.GatewayTransaction
	for _, tx := range f.txs {
		if filter.Gateway != "" && tx.Gateway != filter.Gateway {
			continue
		}
		if filter.Status != "" && tx.Status != filter.Status {
			continue
		}
		if filter.InvoiceID != "" && (tx.InvoiceID == nil || *tx.InvoiceID != filter.InvoiceID) {
			continue
		}
		out = append(out, tx)
	}
	return out, len(out), nil
}

func (f *FakeGatewayTransactionFilteredRepo) ListPendingBefore(_ context.Context, tenantID string, before time.Time, limit int) ([]billing.GatewayTransaction, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []billing.GatewayTransaction
	for _, tx := range f.txs {
		if tx.Status == billing.GatewayStatusPending && tx.UpdatedAt.Before(before) {
			out = append(out, tx)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}
