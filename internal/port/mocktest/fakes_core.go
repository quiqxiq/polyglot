package mocktest

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	domainAudit "github.com/quixiq/polyglot/internal/domain/audit"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

// ─── CustomerRepository ─────────────────────────────────────────────────

type FakeCustomerRepo struct {
	mu   sync.Mutex
	byID map[string]domainCustomer.Customer
}

func NewFakeCustomerRepo() *FakeCustomerRepo {
	return &FakeCustomerRepo{byID: map[string]domainCustomer.Customer{}}
}

func (f *FakeCustomerRepo) Save(_ context.Context, c domainCustomer.Customer) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[c.ID] = c
	return nil
}

func (f *FakeCustomerRepo) FindByID(_ context.Context, id string) (domainCustomer.Customer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byID[id]
	if !ok {
		return domainCustomer.Customer{}, ErrFakeNotFound
	}
	return c, nil
}

func (f *FakeCustomerRepo) FindAll(_ context.Context) ([]domainCustomer.Customer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domainCustomer.Customer, 0, len(f.byID))
	for _, c := range f.byID {
		out = append(out, c)
	}
	return out, nil
}

func (f *FakeCustomerRepo) Delete(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.byID, id)
	return nil
}

func (f *FakeCustomerRepo) FindSubscriptions(_ context.Context, customerID string) ([]domainSubscription.Subscription, error) {
	return nil, nil
}

func (f *FakeCustomerRepo) FindByPortalAccessCode(_ context.Context, code string) (domainCustomer.Customer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.byID {
		if c.PortalAccessCode == code {
			return c, nil
		}
	}
	return domainCustomer.Customer{}, ErrFakeNotFound
}

func (f *FakeCustomerRepo) FindByPhone(_ context.Context, phone string) (domainCustomer.Customer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.byID {
		if c.Phone == phone {
			return c, nil
		}
	}
	return domainCustomer.Customer{}, ErrFakeNotFound
}

func (f *FakeCustomerRepo) FindByCustomerCode(_ context.Context, code string) (domainCustomer.Customer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.byID {
		if c.CustomerCode == code {
			return c, nil
		}
	}
	return domainCustomer.Customer{}, ErrFakeNotFound
}

// ─── SubscriptionRepository ─────────────────────────────────────────────

type FakeSubscriptionRepo struct {
	mu   sync.Mutex
	byID map[string]domainSubscription.Subscription
}

func NewFakeSubscriptionRepo() *FakeSubscriptionRepo {
	return &FakeSubscriptionRepo{byID: map[string]domainSubscription.Subscription{}}
}

func (f *FakeSubscriptionRepo) Seed(s domainSubscription.Subscription) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[s.ID] = s
}

func (f *FakeSubscriptionRepo) Save(_ context.Context, s domainSubscription.Subscription) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[s.ID] = s
	return nil
}

func (f *FakeSubscriptionRepo) FindByID(_ context.Context, id string) (domainSubscription.Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.byID[id]
	if !ok {
		return domainSubscription.Subscription{}, ErrFakeNotFound
	}
	return s, nil
}

func (f *FakeSubscriptionRepo) FindByCustomerID(_ context.Context, cid string) ([]domainSubscription.Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domainSubscription.Subscription
	for _, s := range f.byID {
		if s.CustomerID == cid {
			out = append(out, s)
		}
	}
	return out, nil
}

func (f *FakeSubscriptionRepo) FindAll(_ context.Context) ([]domainSubscription.Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domainSubscription.Subscription, 0, len(f.byID))
	for _, s := range f.byID {
		out = append(out, s)
	}
	return out, nil
}

func (f *FakeSubscriptionRepo) UpdateStatus(_ context.Context, id, status string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.byID[id]
	if !ok {
		return ErrFakeNotFound
	}
	s.Status = status
	f.byID[id] = s
	return nil
}

func (f *FakeSubscriptionRepo) FindByDeviceAndUsername(_ context.Context, deviceID, username string) (domainSubscription.Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, s := range f.byID {
		if s.DeviceID != nil && *s.DeviceID == deviceID && s.RemoteUsername == username {
			return s, nil
		}
	}
	return domainSubscription.Subscription{}, ErrFakeNotFound
}

func (f *FakeSubscriptionRepo) ListActive(_ context.Context) ([]domainSubscription.Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domainSubscription.Subscription
	for _, s := range f.byID {
		if s.Status == domainSubscription.StatusActive && s.DeletedAt == nil {
			out = append(out, s)
		}
	}
	return out, nil
}

// ─── InvoiceRepository ──────────────────────────────────────────────────

type FakeInvoiceRepo struct {
	mu    sync.Mutex
	byID  map[string]domainBilling.Invoice
	items map[string][]domainBilling.InvoiceItem // invoice_id → items
}

func NewFakeInvoiceRepo() *FakeInvoiceRepo {
	return &FakeInvoiceRepo{
		byID:  map[string]domainBilling.Invoice{},
		items: map[string][]domainBilling.InvoiceItem{},
	}
}

func (f *FakeInvoiceRepo) Save(_ context.Context, inv domainBilling.Invoice) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[inv.ID] = inv
	return nil
}

func (f *FakeInvoiceRepo) SaveWithItems(_ context.Context, inv domainBilling.Invoice, items []domainBilling.InvoiceItem) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[inv.ID] = inv
	f.items[inv.ID] = items
	return nil
}

func (f *FakeInvoiceRepo) FindByID(_ context.Context, id string) (domainBilling.Invoice, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	inv, ok := f.byID[id]
	if !ok {
		return domainBilling.Invoice{}, ErrFakeNotFound
	}
	return inv, nil
}

func (f *FakeInvoiceRepo) FindByCustomerID(_ context.Context, cid string) ([]domainBilling.Invoice, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domainBilling.Invoice
	for _, inv := range f.byID {
		if inv.CustomerID == cid {
			out = append(out, inv)
		}
	}
	return out, nil
}

func (f *FakeInvoiceRepo) FindAll(_ context.Context) ([]domainBilling.Invoice, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domainBilling.Invoice, 0, len(f.byID))
	for _, inv := range f.byID {
		out = append(out, inv)
	}
	return out, nil
}

func (f *FakeInvoiceRepo) UpdateStatus(_ context.Context, id, status string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	inv, ok := f.byID[id]
	if !ok {
		return ErrFakeNotFound
	}
	inv.Status = status
	f.byID[id] = inv
	return nil
}

func (f *FakeInvoiceRepo) FindByPaymentCode(_ context.Context, code string) (domainBilling.Invoice, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, inv := range f.byID {
		if inv.ManualPaymentCode == code {
			return inv, nil
		}
	}
	return domainBilling.Invoice{}, ErrFakeNotFound
}

func (f *FakeInvoiceRepo) FindByQRPayload(_ context.Context, qr string) (domainBilling.Invoice, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, inv := range f.byID {
		if inv.QRPayload == qr {
			return inv, nil
		}
	}
	return domainBilling.Invoice{}, ErrFakeNotFound
}

func (f *FakeInvoiceRepo) FindBySubscriptionPeriod(_ context.Context, subID, period string) (domainBilling.Invoice, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, inv := range f.byID {
		if inv.SubscriptionID != nil && *inv.SubscriptionID == subID && inv.Period == period {
			return inv, nil
		}
	}
	return domainBilling.Invoice{}, ErrFakeNotFound
}

// ItemsOf exposes stored line items for assertions.
func (f *FakeInvoiceRepo) ItemsOf(invoiceID string) []domainBilling.InvoiceItem {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.items[invoiceID]
}

// ─── NotificationRepository ─────────────────────────────────────────────

type FakeNotificationRepo struct {
	mu        sync.Mutex
	queued    []domainNotification.WANotification
	templates map[string]domainNotification.NotificationTemplate // key → tpl
}

func NewFakeNotificationRepo() *FakeNotificationRepo {
	return &FakeNotificationRepo{templates: map[string]domainNotification.NotificationTemplate{}}
}

func (f *FakeNotificationRepo) SeedTemplate(t domainNotification.NotificationTemplate) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.templates[t.TemplateKey] = t
}

func (f *FakeNotificationRepo) SaveTemplate(_ context.Context, t domainNotification.NotificationTemplate) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.templates[t.TemplateKey] = t
	return nil
}

func (f *FakeNotificationRepo) FindTemplateByKey(_ context.Context, _, key string) (domainNotification.NotificationTemplate, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.templates[key]
	if !ok {
		return domainNotification.NotificationTemplate{}, ErrFakeNotFound
	}
	return t, nil
}

func (f *FakeNotificationRepo) ListTemplates(_ context.Context, _ bool) ([]domainNotification.NotificationTemplate, error) {
	return nil, nil
}

func (f *FakeNotificationRepo) Queue(_ context.Context, n domainNotification.WANotification) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.queued = append(f.queued, n)
	return nil
}

func (f *FakeNotificationRepo) Queued() []domainNotification.WANotification {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domainNotification.WANotification, len(f.queued))
	copy(out, f.queued)
	return out
}

func (f *FakeNotificationRepo) FindByID(_ context.Context, id string) (domainNotification.WANotification, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, n := range f.queued {
		if n.ID == id {
			return n, nil
		}
	}
	return domainNotification.WANotification{}, ErrFakeNotFound
}

func (f *FakeNotificationRepo) Pending(_ context.Context, limit int) ([]domainNotification.WANotification, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domainNotification.WANotification
	for _, n := range f.queued {
		if n.Status == domainNotification.StatusQueued {
			out = append(out, n)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (f *FakeNotificationRepo) MarkSent(_ context.Context, id string, sentAt time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := range f.queued {
		if f.queued[i].ID == id {
			f.queued[i].Status = domainNotification.StatusSent
			f.queued[i].SentAt = &sentAt
			return nil
		}
	}
	return ErrFakeNotFound
}

func (f *FakeNotificationRepo) MarkFailed(_ context.Context, id, errMsg string) error { return nil }

func (f *FakeNotificationRepo) ListByCustomer(_ context.Context, cid string, _ int) ([]domainNotification.WANotification, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domainNotification.WANotification
	for _, n := range f.queued {
		if n.CustomerID != nil && *n.CustomerID == cid {
			out = append(out, n)
		}
	}
	return out, nil
}

// ─── AuditLogWriter ─────────────────────────────────────────────────────

type FakeAuditWriter struct {
	mu      sync.Mutex
	Entries []domainAudit.AuditLog
}

func (f *FakeAuditWriter) Write(_ context.Context, e domainAudit.AuditLog) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Entries = append(f.Entries, e)
	return nil
}

func (f *FakeAuditWriter) Count(action string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, e := range f.Entries {
		if e.Action == action {
			n++
		}
	}
	return n
}

// ─── PaymentProcessor ───────────────────────────────────────────────────

type FakePaymentProcessor struct {
	mu   sync.Mutex
	Cmds []port.CashPaymentCommand
	Err  error
	Pay  domainBilling.Payment
}

func (f *FakePaymentProcessor) ProcessCashPayment(_ context.Context, cmd port.CashPaymentCommand) (domainBilling.Payment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Cmds = append(f.Cmds, cmd)
	if f.Err != nil {
		return domainBilling.Payment{}, f.Err
	}
	return f.Pay, nil
}

// ─── RouterAccountManager ───────────────────────────────────────────────

// FakeRouterAccountManager merekam semua call lifecycle akun router.
type FakeRouterAccountManager struct {
	mu sync.Mutex

	Calls []string // "Provision", "Isolate", ... dengan detail
	Fail  map[string]error
}

func (f *FakeRouterAccountManager) record(op string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, op)
	for prefix, err := range f.Fail {
		if strings.HasPrefix(op, prefix) {
			return err
		}
	}
	return nil
}

func (f *FakeRouterAccountManager) Provision(_ context.Context, _, _ string, a port.SubscriberAccount) error {
	return f.record("Provision:" + a.Username + "@" + a.Profile)
}

func (f *FakeRouterAccountManager) UpdateAccount(_ context.Context, _, _, username, newProfile string) error {
	return f.record("UpdateAccount:" + username + "->" + newProfile)
}

func (f *FakeRouterAccountManager) EnsureProfile(_ context.Context, _, _, profileName, rateLimit string) error {
	return f.record("EnsureProfile:" + profileName + "@" + rateLimit)
}

func (f *FakeRouterAccountManager) Isolate(_ context.Context, _, _, username string, opt port.IsolationOptions) error {
	return f.record(fmt.Sprintf("Isolate:%s->%s(list=%s)", username, opt.IsolirProfile, opt.AddressList))
}

func (f *FakeRouterAccountManager) Restore(_ context.Context, _, _, username, normalProfile, addressList string) error {
	return f.record("Restore:" + username + "->" + normalProfile)
}

func (f *FakeRouterAccountManager) Suspend(_ context.Context, _, _, username string) error {
	return f.record("Suspend:" + username)
}

func (f *FakeRouterAccountManager) Terminate(_ context.Context, _, _, username string) error {
	return f.record("Terminate:" + username)
}

// Count returns how many calls start with the given prefix.
func (f *FakeRouterAccountManager) Count(prefix string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, c := range f.Calls {
		if strings.HasPrefix(c, prefix) {
			n++
		}
	}
	return n
}

// ─── SettingReader ──────────────────────────────────────────────────────

type FakeSettingReader struct {
	mu   sync.Mutex
	vals map[string]string
}

func NewFakeSettingReader(vals map[string]string) *FakeSettingReader {
	out := map[string]string{}
	for k, v := range vals {
		out[k] = v
	}
	return &FakeSettingReader{vals: out}
}

func (f *FakeSettingReader) GetValue(_ context.Context, key, fallback string) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if v, ok := f.vals[key]; ok && v != "" {
		return v
	}
	return fallback
}

// NewFakeRouterAccountManager constructs an empty recorder.
func NewFakeRouterAccountManager() *FakeRouterAccountManager {
	return &FakeRouterAccountManager{}
}

// ListLifecycle implements the lifecycle view (ACTIVE + ISOLATED).
func (f *FakeSubscriptionRepo) ListLifecycle(_ context.Context) ([]domainSubscription.Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domainSubscription.Subscription
	for _, s := range f.byID {
		if s.DeletedAt != nil {
			continue
		}
		if s.Status == domainSubscription.StatusActive || s.Status == domainSubscription.StatusIsolated {
			out = append(out, s)
		}
	}
	return out, nil
}

// HasActiveForPlan implements the delete-guard lookup.
func (f *FakeSubscriptionRepo) HasActiveForPlan(_ context.Context, planID string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, s := range f.byID {
		if s.PlanID == planID && s.DeletedAt == nil &&
			(s.Status == domainSubscription.StatusActive || s.Status == domainSubscription.StatusIsolated) {
			return true, nil
		}
	}
	return false, nil
}

// ─── NotificationSender ─────────────────────────────────────────────────

type FakeNotificationSender struct {
	mu       sync.Mutex
	Messages []struct{ Phone, Content string }
	Err      error
}

// Sent returns a copy of sent messages.
func (f *FakeNotificationSender) Sent() []struct{ Phone, Content string } {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]struct{ Phone, Content string }, len(f.Messages))
	copy(out, f.Messages)
	return out
}

func (f *FakeNotificationSender) Send(_ context.Context, phone, content string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Err != nil {
		return f.Err
	}
	f.Messages = append(f.Messages, struct{ Phone, Content string }{phone, content})
	return nil
}

// ─── PortalRepository (in-memory) ───────────────────────────────────────

type FakePortalRepo struct {
	mu       sync.Mutex
	sessions map[string]domainCustomer.PortalSession
	otps     []domainCustomer.PortalOTP
}

func NewFakePortalRepo() *FakePortalRepo {
	return &FakePortalRepo{sessions: map[string]domainCustomer.PortalSession{}}
}

func (f *FakePortalRepo) SaveSession(_ context.Context, s domainCustomer.PortalSession) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sessions[s.SessionToken] = s
	return nil
}

func (f *FakePortalRepo) FindValidSession(_ context.Context, token string) (domainCustomer.PortalSession, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.sessions[token]
	if !ok || time.Now().After(s.ExpiresAt) {
		return domainCustomer.PortalSession{}, ErrFakeNotFound
	}
	return s, nil
}

func (f *FakePortalRepo) DeleteSession(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for tok, s := range f.sessions {
		if s.ID == id {
			delete(f.sessions, tok)
		}
	}
	return nil
}

func (f *FakePortalRepo) SaveOTP(_ context.Context, o domainCustomer.PortalOTP) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.otps = append(f.otps, o)
	return nil
}

func (f *FakePortalRepo) ConsumeOTP(_ context.Context, phoneN, codeHash string, maxAttempts int) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := len(f.otps) - 1; i >= 0; i-- {
		o := &f.otps[i]
		if o.Phone != phoneN || time.Now().After(o.ExpiresAt) {
			continue
		}
		if o.Attempts >= maxAttempts {
			return false, port.ErrOTPLocked
		}
		if o.CodeHash == codeHash {
			now := time.Now()
			o.ConsumedAt = &now
			return true, nil
		}
		o.Attempts++
		if o.Attempts >= maxAttempts {
			now := time.Now()
			o.ConsumedAt = &now
			return false, port.ErrOTPLocked
		}
		return false, nil
	}
	return false, port.ErrOTPNotFound
}

// ─── GatewayTransactionRepository ───────────────────────────────────────

type FakeGatewayTxRepo struct {
	mu   sync.Mutex
	byID map[string]domainBilling.GatewayTransaction
}

func NewFakeGatewayTxRepo() *FakeGatewayTxRepo {
	return &FakeGatewayTxRepo{byID: map[string]domainBilling.GatewayTransaction{}}
}

func (f *FakeGatewayTxRepo) Save(_ context.Context, t domainBilling.GatewayTransaction) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[t.ID] = t
	return nil
}

func (f *FakeGatewayTxRepo) FindByExternalID(_ context.Context, gateway, ext string) (domainBilling.GatewayTransaction, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, t := range f.byID {
		if t.Gateway == gateway && t.ExternalID == ext {
			return t, nil
		}
	}
	return domainBilling.GatewayTransaction{}, ErrFakeNotFound
}

func (f *FakeGatewayTxRepo) FindByInvoice(_ context.Context, invoiceID string) ([]domainBilling.GatewayTransaction, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domainBilling.GatewayTransaction
	for _, t := range f.byID {
		if t.InvoiceID != nil && *t.InvoiceID == invoiceID {
			out = append(out, t)
		}
	}
	return out, nil
}

func (f *FakeGatewayTxRepo) UpdateStatus(_ context.Context, id, status string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.byID[id]
	if !ok {
		return ErrFakeNotFound
	}
	t.Status = status
	f.byID[id] = t
	return nil
}

func (f *FakeGatewayTxRepo) LinkPayment(_ context.Context, id, paymentID string, fee float64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.byID[id]
	if !ok {
		return ErrFakeNotFound
	}
	t.PaymentID = &paymentID
	t.FeeAmount = fee
	f.byID[id] = t
	return nil
}

// Get exposes stored tx for assertions.
func (f *FakeGatewayTxRepo) Get(id string) domainBilling.GatewayTransaction {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.byID[id]
}

// ─── PaymentGateway ─────────────────────────────────────────────────────

type FakePaymentGateway struct {
	Charge    port.ChargeResult
	Err       error
	Event     port.WebhookEvent
	ParseErr  error
	IsEnabled bool
}

func (f *FakePaymentGateway) Name() string                 { return "FAKE" }
func (f *FakePaymentGateway) Enabled(context.Context) bool { return f.IsEnabled }
func (f *FakePaymentGateway) CreateCharge(context.Context, port.ChargeRequest) (port.ChargeResult, error) {
	if f.Err != nil {
		return port.ChargeResult{}, f.Err
	}
	return f.Charge, nil
}
func (f *FakePaymentGateway) ParseWebhook(context.Context, []byte, string) (port.WebhookEvent, error) {
	if f.ParseErr != nil {
		return port.WebhookEvent{}, f.ParseErr
	}
	return f.Event, nil
}

// All returns a copy of all stored gateway transactions.
func (f *FakeGatewayTxRepo) All() []domainBilling.GatewayTransaction {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domainBilling.GatewayTransaction, 0, len(f.byID))
	for _, t := range f.byID {
		out = append(out, t)
	}
	return out
}

// ─── PaymentReader ──────────────────────────────────────────────────────

type FakePaymentReader struct {
	Payments []domainBilling.Payment
}

func (f *FakePaymentReader) ListByCustomer(context.Context, string, int) ([]domainBilling.Payment, error) {
	return f.Payments, nil
}

// NewFakePaymentReader constructs an empty reader.
func NewFakePaymentReader() *FakePaymentReader { return &FakePaymentReader{} }

// OTPCount exposes number of stored OTPs.
func (f *FakePortalRepo) OTPCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.otps)
}

// AllNotifications returns a copy of the queue (termasuk yang sudah sent).
func (f *FakeNotificationRepo) AllNotifications() []domainNotification.WANotification {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domainNotification.WANotification, len(f.queued))
	copy(out, f.queued)
	return out
}

// MarkFailedWithAttempt implements port.NotificationRetryRepository.
func (f *FakeNotificationRepo) MarkFailedWithAttempt(_ context.Context, id, errMsg string, attempts int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := range f.queued {
		if f.queued[i].ID == id {
			f.queued[i].Status = domainNotification.StatusFailed
			f.queued[i].ErrorMessage = errMsg
			f.queued[i].Attempts = attempts
			return nil
		}
	}
	return ErrFakeNotFound
}

// PendingWithRetryLimit implements port.NotificationRetryRepository.
func (f *FakeNotificationRepo) PendingWithRetryLimit(_ context.Context, limit, maxAttempts int) ([]domainNotification.WANotification, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domainNotification.WANotification
	for _, n := range f.queued {
		retryable := n.Status == domainNotification.StatusQueued ||
			(n.Status == domainNotification.StatusFailed && n.Attempts < maxAttempts)
		if retryable {
			out = append(out, n)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}
