package billing

import (
	"context"
	"fmt"
	"strings"
	"time"

	domainAudit "github.com/quixiq/polyglot/internal/domain/audit"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainTenant "github.com/quixiq/polyglot/internal/domain/tenant"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/logger"
)

// Default approval routing untuk bukti transfer manual.
const (
	defaultManualCashAccountID    = "ca-1001-kas-kantor"
	defaultManualIncomeCategoryID = "cc-tagihan"
)

// PaymentSubmissionUseCase mengelola bukti transfer manual (planning v2
// Fase 3): pelanggan mengunggah bukti, finance menyetujui/menolak. Approve
// memakai PaymentProcessor atomik sehingga buku kas, kwitansi, notifikasi,
// dan un-isolir router berjalan seperti pembayaran kasir biasa.
type PaymentSubmissionUseCase struct {
	subs      port.PaymentSubmissionRepository
	invoices  port.InvoiceRepository
	processor port.PaymentProcessor
	settings  port.SettingReader
	audit     port.AuditLogWriter

	now func() time.Time
}

// NewPaymentSubmissionUseCase wires dependencies.
func NewPaymentSubmissionUseCase(
	subs port.PaymentSubmissionRepository,
	invoices port.InvoiceRepository,
	processor port.PaymentProcessor,
	settings port.SettingReader,
) *PaymentSubmissionUseCase {
	return &PaymentSubmissionUseCase{
		subs: subs, invoices: invoices, processor: processor,
		settings: settings, now: time.Now,
	}
}

// WithAudit menautkan audit log.
func (u *PaymentSubmissionUseCase) WithAudit(w port.AuditLogWriter) *PaymentSubmissionUseCase {
	u.audit = w
	return u
}

// WithClock mengganti sumber waktu (untuk test deterministik).
func (u *PaymentSubmissionUseCase) WithClock(now func() time.Time) *PaymentSubmissionUseCase {
	u.now = now
	return u
}

// SubmitManualPaymentInput adalah input upload bukti transfer oleh pelanggan.
type SubmitManualPaymentInput struct {
	ManualPaymentCode string
	Amount            float64
	SenderName        string
	SenderBank        string
	ProofImageURL     string
	Notes             string
	BankAccountID     *string
}

// ResolveInvoice mencari invoice lewat kode bayar manual untuk ditampilkan
// pada halaman pembayaran publik (data minimal, tanpa PII berlebih).
func (u *PaymentSubmissionUseCase) ResolveInvoice(ctx context.Context, code string) (domainBilling.Invoice, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return domainBilling.Invoice{}, fmt.Errorf("%w: manual_payment_code is required", domainBilling.ErrInvalidInput)
	}
	inv, err := u.invoices.FindByPaymentCode(ctx, code)
	if err != nil {
		return domainBilling.Invoice{}, domainBilling.ErrNotFound
	}
	return inv, nil
}

// Submit menyimpan bukti transfer manual sebagai PENDING setelah memvalidasi
// kode bayar dan sisa tagihan.
func (u *PaymentSubmissionUseCase) Submit(ctx context.Context, in SubmitManualPaymentInput) (domainBilling.PaymentSubmission, error) {
	code := strings.TrimSpace(in.ManualPaymentCode)
	if code == "" {
		return domainBilling.PaymentSubmission{}, fmt.Errorf("%w: manual_payment_code is required", domainBilling.ErrInvalidInput)
	}
	if strings.TrimSpace(in.ProofImageURL) == "" {
		return domainBilling.PaymentSubmission{}, fmt.Errorf("%w: proof_image_url is required", domainBilling.ErrInvalidInput)
	}
	if strings.TrimSpace(in.SenderName) == "" {
		return domainBilling.PaymentSubmission{}, fmt.Errorf("%w: sender_name is required", domainBilling.ErrInvalidInput)
	}
	inv, err := u.invoices.FindByPaymentCode(ctx, code)
	if err != nil {
		return domainBilling.PaymentSubmission{}, domainBilling.ErrNotFound
	}
	switch inv.Status {
	case domainBilling.StatusUnpaid, domainBilling.StatusPartial, domainBilling.StatusOverdue:
	default:
		return domainBilling.PaymentSubmission{}, fmt.Errorf("%w: invoice status %s cannot be paid", domainBilling.ErrInvalidTransition, inv.Status)
	}
	outstanding := inv.Total - inv.PaidAmount
	if outstanding <= 0 {
		return domainBilling.PaymentSubmission{}, domainBilling.ErrInvoiceAlreadyPaid
	}
	if in.Amount <= 0 {
		return domainBilling.PaymentSubmission{}, fmt.Errorf("%w: amount must be greater than zero", domainBilling.ErrInvalidInput)
	}
	if in.Amount > outstanding+1e-9 {
		return domainBilling.PaymentSubmission{}, fmt.Errorf("%w: outstanding %.2f, got %.2f", domainBilling.ErrOverpayment, outstanding, in.Amount)
	}
	now := u.now().UTC()
	sub := domainBilling.PaymentSubmission{
		ID:            idgen.New("psub"),
		TenantID:      orTenantID(inv.TenantID),
		InvoiceID:     inv.ID,
		Amount:        in.Amount,
		BankAccountID: in.BankAccountID,
		SenderBank:    strings.TrimSpace(in.SenderBank),
		SenderName:    strings.TrimSpace(in.SenderName),
		ProofImageURL: strings.TrimSpace(in.ProofImageURL),
		Notes:         in.Notes,
		Status:        domainBilling.PaymentSubmissionPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := u.subs.Save(ctx, sub); err != nil {
		return domainBilling.PaymentSubmission{}, fmt.Errorf("save payment submission: %w", err)
	}
	return sub, nil
}

// Approve menyetujui bukti transfer: pembayaran diproses atomik oleh
// PaymentProcessor (kas, kwitansi, notifikasi, un-isolir) lalu submisi
// ditandai APPROVED.
func (u *PaymentSubmissionUseCase) Approve(ctx context.Context, id string, reviewerID *uint) (domainBilling.PaymentSubmission, error) {
	sub, err := u.subs.FindByID(ctx, id)
	if err != nil {
		return domainBilling.PaymentSubmission{}, domainBilling.ErrNotFound
	}
	if sub.Status != domainBilling.PaymentSubmissionPending {
		return domainBilling.PaymentSubmission{}, fmt.Errorf("%w: got %s", domainBilling.ErrPaymentSubmissionNotPending, sub.Status)
	}
	cashAccountID := u.setting(ctx, "isp.manual_payment_cash_account_id", defaultManualCashAccountID)
	incomeCategoryID := u.setting(ctx, "isp.manual_payment_income_category_id", defaultManualIncomeCategoryID)
	methodID := u.setting(ctx, "isp.manual_payment_method_id", "pm-bank")
	pay, err := u.processor.ProcessCashPayment(ctx, port.CashPaymentCommand{
		TenantID:         sub.TenantID,
		InvoiceID:        sub.InvoiceID,
		PaymentMethodID:  &methodID,
		Amount:           sub.Amount,
		CashAccountID:    cashAccountID,
		IncomeCategoryID: incomeCategoryID,
		ReceivedBy:       reviewerID,
		ScanMethod:       domainBilling.ScanManual,
		Reference:        "SUBMISSION:" + sub.ID,
		Notes:            fmt.Sprintf("Bukti transfer manual disetujui (%s)", sub.SenderName),
	})
	if err != nil {
		return domainBilling.PaymentSubmission{}, fmt.Errorf("process approved submission: %w", err)
	}
	now := u.now().UTC()
	sub.Status = domainBilling.PaymentSubmissionApproved
	sub.ReviewedBy = reviewerID
	sub.ReviewedAt = &now
	sub.PaymentID = &pay.ID
	sub.UpdatedAt = now
	if err := u.subs.Save(ctx, sub); err != nil {
		return domainBilling.PaymentSubmission{}, fmt.Errorf("update submission approved: %w", err)
	}
	u.writeAudit(ctx, actorStr(reviewerID), "APPROVE_PAYMENT_SUBMISSION", sub.ID)
	return sub, nil
}

// Reject menolak bukti transfer tanpa menyentuh buku kas.
func (u *PaymentSubmissionUseCase) Reject(ctx context.Context, id, reason string, reviewerID *uint) (domainBilling.PaymentSubmission, error) {
	if strings.TrimSpace(reason) == "" {
		return domainBilling.PaymentSubmission{}, fmt.Errorf("%w: rejection reason is required", domainBilling.ErrInvalidInput)
	}
	sub, err := u.subs.FindByID(ctx, id)
	if err != nil {
		return domainBilling.PaymentSubmission{}, domainBilling.ErrNotFound
	}
	if sub.Status != domainBilling.PaymentSubmissionPending {
		return domainBilling.PaymentSubmission{}, fmt.Errorf("%w: got %s", domainBilling.ErrPaymentSubmissionNotPending, sub.Status)
	}
	now := u.now().UTC()
	sub.Status = domainBilling.PaymentSubmissionRejected
	sub.RejectionReason = reason
	sub.ReviewedBy = reviewerID
	sub.ReviewedAt = &now
	sub.UpdatedAt = now
	if err := u.subs.Save(ctx, sub); err != nil {
		return domainBilling.PaymentSubmission{}, fmt.Errorf("update submission rejected: %w", err)
	}
	u.writeAudit(ctx, actorStr(reviewerID), "REJECT_PAYMENT_SUBMISSION", sub.ID)
	return sub, nil
}

// List mengembalikan submisi terbaru (opsional per status).
func (u *PaymentSubmissionUseCase) List(ctx context.Context, status string) ([]domainBilling.PaymentSubmission, error) {
	return u.subs.List(ctx, port.PaymentSubmissionFilter{Status: strings.ToUpper(strings.TrimSpace(status)), Limit: 200})
}

func (u *PaymentSubmissionUseCase) setting(ctx context.Context, key, fallback string) string {
	if u.settings == nil {
		return fallback
	}
	v := strings.TrimSpace(u.settings.GetValue(ctx, key, fallback))
	if v == "" {
		return fallback
	}
	return v
}

func (u *PaymentSubmissionUseCase) writeAudit(ctx context.Context, actorID, action, entityID string) {
	if u.audit == nil {
		return
	}
	err := u.audit.Write(ctx, domainAudit.AuditLog{
		TenantID:   domainTenant.DefaultID,
		ActorType:  domainAudit.ActorUser,
		ActorID:    actorID,
		Action:     action,
		EntityType: "payment_submission",
		EntityID:   entityID,
	})
	if err != nil {
		logger.WithComponent("PaymentSubmissionUC").WithError(err).Warn("audit log write failed")
	}
}

func actorStr(id *uint) string {
	if id == nil {
		return ""
	}
	return fmt.Sprint(*id)
}
