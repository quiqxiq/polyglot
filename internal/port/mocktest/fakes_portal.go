package mocktest

import (
	"context"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
	"github.com/quixiq/polyglot/internal/port"
	"sync"
	"time"
)

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
			return false, domainCustomer.ErrOTPLocked
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
			return false, domainCustomer.ErrOTPLocked
		}
		return false, nil
	}
	return false, domainCustomer.ErrOTPNotFound
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
