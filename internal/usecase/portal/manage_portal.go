package portal

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/phone"
)

// Error sentinels untuk alur portal tinggal di
// internal/domain/customer/errors.go (ErrPortalBadCredentials,
// ErrCustomerNotFound) per DEVELOPMENT-GUIDELINES.md §6.

// UseCase menyediakan login OTP-WA dan akses data mandiri pelanggan.
type UseCase struct {
	portals port.PortalRepository
	customs port.CustomerRepository
	subs    port.SubscriptionRepository
	invs    port.InvoiceRepository
	pays    port.PaymentReader
	sender  port.NotificationSender
	reader  port.SettingReader
}

// NewUseCase wires dependencies.
func NewUseCase(
	portals port.PortalRepository,
	customs port.CustomerRepository,
	subs port.SubscriptionRepository,
	invs port.InvoiceRepository,
	pays port.PaymentReader,
	sender port.NotificationSender,
	reader port.SettingReader,
) *UseCase {
	return &UseCase{portals: portals, customs: customs, subs: subs, invs: invs,
		pays: pays, sender: sender, reader: reader}
}

// ─── OTP & Login ────────────────────────────────────────────────────────

// RequestOTP mengirim kode 6 digit via WhatsApp ke pelanggan yang cocok
// dengan kode portal ATAU nomor HP. Mengembalikan nomor tersamar.
func (u *UseCase) RequestOTP(ctx context.Context, identifier string) (string, error) {
	cust, err := u.resolve(ctx, identifier)
	if err != nil {
		return "", domainCustomer.ErrPortalBadCredentials // jangan bocorkan keberadaan akun
	}
	code := generateDigits(6)
	ttlMin := atoiDefault(u.reader.GetValue(ctx, "isp.otp_ttl_minutes", "5"), 5)
	now := time.Now()
	o := domainCustomer.PortalOTP{
		ID:        idgen.New("otp"),
		TenantID:  cust.TenantID,
		Phone:     cust.Phone,
		CodeHash:  hashCode(code),
		Purpose:   "PORTAL_LOGIN",
		ExpiresAt: now.Add(time.Duration(ttlMin) * time.Minute),
		CreatedAt: now,
	}
	if err := u.portals.SaveOTP(ctx, o); err != nil {
		return "", err
	}
	msg := fmt.Sprintf("Kode login portal Anda: %s (berlaku %d menit). Jangan bagikan kode ini.", code, ttlMin)
	if err := u.sender.Send(ctx, cust.Phone, msg); err != nil {
		return "", fmt.Errorf("send OTP via WhatsApp: %w", err)
	}
	return maskPhone(cust.Phone), nil
}

// Login memverifikasi OTP lalu membuat sesi portal baru.
func (u *UseCase) Login(ctx context.Context, identifier, otp string) (token string, cust domainCustomer.Customer, err error) {
	cust, err = u.resolve(ctx, identifier)
	if err != nil {
		return "", cust, domainCustomer.ErrPortalBadCredentials
	}
	maxAttempts := atoiDefault(u.reader.GetValue(ctx, "isp.otp_max_attempts", "5"), 5)
	ok, verr := u.portals.ConsumeOTP(ctx, cust.Phone, hashCode(otp), maxAttempts)
	if verr != nil {
		return "", cust, verr
	}
	if !ok {
		return "", cust, domainCustomer.ErrPortalBadCredentials
	}
	hours := atoiDefault(u.reader.GetValue(ctx, "isp.portal_session_hours", "12"), 12)
	now := time.Now()
	session := domainCustomer.PortalSession{
		ID:           idgen.New("pses"),
		TenantID:     cust.TenantID,
		CustomerID:   cust.ID,
		SessionToken: newToken(),
		ExpiresAt:    now.Add(time.Duration(hours) * time.Hour),
		CreatedAt:    now,
	}
	if err := u.portals.SaveSession(ctx, session); err != nil {
		return "", cust, err
	}
	return session.SessionToken, cust, nil
}

// Authenticate memvalidasi bearer token dan mengembalikan pelanggannya.
func (u *UseCase) Authenticate(ctx context.Context, bearer string) (domainCustomer.Customer, error) {
	if bearer == "" {
		return domainCustomer.Customer{}, domainCustomer.ErrPortalBadCredentials
	}
	session, err := u.portals.FindValidSession(ctx, bearer)
	if err != nil {
		return domainCustomer.Customer{}, domainCustomer.ErrPortalBadCredentials
	}
	return u.customs.FindByID(ctx, session.CustomerID)
}

// Logout menghapus sesi aktif.
func (u *UseCase) Logout(ctx context.Context, token string) error {
	session, err := u.portals.FindValidSession(ctx, token)
	if err != nil {
		return nil // sudah tidak valid — anggap sukses
	}
	return u.portals.DeleteSession(ctx, session.ID)
}

// ─── Data mandiri ───────────────────────────────────────────────────────

type CustomerOverview struct {
	Customer       domainCustomer.Customer `json:"customer"`
	Status         string                  `json:"status"`
	Subscription   *SubscriptionView       `json:"subscription,omitempty"`
	UnpaidInvoices []InvoiceView           `json:"unpaid_invoices"`
	PaymentURL     string                  `json:"payment_url,omitempty"`
}

// SubscriptionView represents the customer's subscription summary for portal view.
type SubscriptionView struct {
	ID          string  `json:"id"`
	PlanID      string  `json:"plan_id"`
	ServiceType string  `json:"service_type"`
	Status      string  `json:"status"`
	RateLimit   string  `json:"rate_limit,omitempty"`
	BillingDay  int     `json:"billing_day"`
	EndDate     *string `json:"end_date,omitempty"`
}

// InvoiceView represents an unpaid invoice summary for portal view.
type InvoiceView struct {
	ID                string  `json:"id"`
	InvoiceNumber     string  `json:"invoice_number"`
	Period            string  `json:"period"`
	Total             float64 `json:"total"`
	PaidAmount        float64 `json:"paid_amount"`
	Outstanding       float64 `json:"outstanding"`
	DueDate           string  `json:"due_date"`
	Status            string  `json:"status"`
	ManualPaymentCode string  `json:"manual_payment_code"`
}

// Overview merangkum profil + langganan + tagihan tertunggak + URL bayar.
func (u *UseCase) Overview(ctx context.Context, customerID string) (CustomerOverview, error) {
	cust, err := u.customs.FindByID(ctx, customerID)
	if err != nil {
		return CustomerOverview{}, domainCustomer.ErrCustomerNotFound
	}
	ov := CustomerOverview{Customer: cust, Status: cust.Status}

	if all, err := u.subs.FindByCustomerID(ctx, customerID); err == nil && len(all) > 0 {
		var live []domainSubscription.Subscription
		for _, s := range all {
			if s.Status == domainSubscription.StatusActive ||
				s.Status == domainSubscription.StatusIsolated ||
				s.Status == domainSubscription.StatusSuspended {
				live = append(live, s)
			}
		}
		if len(live) == 0 {
			return ov, nil
		}
		s := live[0]
		endDate := ""
		if s.EndDate != nil {
			endDate = s.EndDate.Format("2006-01-02")
		}
		ov.Subscription = &SubscriptionView{
			ID: s.ID, PlanID: s.PlanID, ServiceType: s.ServiceType,
			Status: s.Status, RateLimit: s.RateLimit, BillingDay: s.BillingDay,
			EndDate: strPtrOrNil(endDate),
		}
	}

	invoices, _ := u.invs.FindByCustomerID(ctx, customerID)
	for _, inv := range invoices {
		outstanding := inv.Total - inv.PaidAmount
		if outstanding <= 0.001 {
			continue
		}
		ov.UnpaidInvoices = append(ov.UnpaidInvoices, InvoiceView{
			ID: inv.ID, InvoiceNumber: inv.InvoiceNumber, Period: inv.Period,
			Total: inv.Total, PaidAmount: inv.PaidAmount, Outstanding: outstanding,
			DueDate: inv.DueDate.Format("2006-01-02"), Status: inv.Status,
			ManualPaymentCode: inv.ManualPaymentCode,
		})
	}
	if cust.Status == domainCustomer.StatusIsolated && len(ov.UnpaidInvoices) > 0 {
		ov.PaymentURL = u.reader.GetValue(ctx, "isp.payment_redirect_url", "")
	}
	return ov, nil
}

// Invoices seluruh riwayat faktur milik pelanggan.
func (u *UseCase) Invoices(ctx context.Context, customerID string) ([]domainBilling.Invoice, error) {
	return u.invs.FindByCustomerID(ctx, customerID)
}

// Payments riwayat pembayaran pelanggan.
func (u *UseCase) Payments(ctx context.Context, customerID string, limit int) ([]domainBilling.Payment, error) {
	return u.pays.ListByCustomer(ctx, customerID, limit)
}

// PublicBillView merangkum data tagihan publik untuk halaman isolir / lookup cepat.
type PublicBillView struct {
	CustomerName      string  `json:"customer_name"`
	CustomerCode      string  `json:"customer_code"`
	InvoiceID         string  `json:"invoice_id"`
	InvoiceNumber     string  `json:"invoice_number"`
	Period            string  `json:"period"`
	Total             float64 `json:"total"`
	PaidAmount        float64 `json:"paid_amount"`
	Outstanding       float64 `json:"outstanding"`
	DueDate           string  `json:"due_date"`
	Status            string  `json:"status"`
	ManualPaymentCode string  `json:"manual_payment_code"`
}

// LookupBill mencari tagihan terbuka berdasarkan kode bayar, no faktur, no HP, atau kode portal.
func (u *UseCase) LookupBill(ctx context.Context, identifier string) (PublicBillView, error) {
	ident := strings.TrimSpace(identifier)
	if ident == "" {
		return PublicBillView{}, domainCustomer.ErrInvalidInput
	}

	// 1. Coba cari faktur langsung via kode bayar manual atau nomor faktur
	inv, err := u.invs.FindByPaymentCode(ctx, ident)
	if err != nil {
		inv, err = u.invs.FindByID(ctx, ident)
	}
	if err == nil {
		cust, _ := u.customs.FindByID(ctx, inv.CustomerID)
		outstanding := inv.Total - inv.PaidAmount
		return PublicBillView{
			CustomerName:      maskName(cust.Name),
			CustomerCode:      cust.CustomerCode,
			InvoiceID:         inv.ID,
			InvoiceNumber:     inv.InvoiceNumber,
			Period:            inv.Period,
			Total:             inv.Total,
			PaidAmount:        inv.PaidAmount,
			Outstanding:       outstanding,
			DueDate:           inv.DueDate.Format("2006-01-02"),
			Status:            inv.Status,
			ManualPaymentCode: inv.ManualPaymentCode,
		}, nil
	}

	// 2. Coba cari pelanggan via no HP, kode portal, atau kode pelanggan
	cust, err := u.resolve(ctx, ident)
	if err != nil {
		cust, err = u.customs.FindByCustomerCode(ctx, ident)
	}
	if err != nil {
		return PublicBillView{}, domainCustomer.ErrCustomerNotFound
	}

	invoices, err := u.invs.FindByCustomerID(ctx, cust.ID)
	if err != nil || len(invoices) == 0 {
		return PublicBillView{}, domainCustomer.ErrCustomerNotFound
	}

	// Cari tagihan tertua yang belum lunas
	for _, inv := range invoices {
		outstanding := inv.Total - inv.PaidAmount
		if outstanding > 0.01 {
			return PublicBillView{
				CustomerName:      maskName(cust.Name),
				CustomerCode:      cust.CustomerCode,
				InvoiceID:         inv.ID,
				InvoiceNumber:     inv.InvoiceNumber,
				Period:            inv.Period,
				Total:             inv.Total,
				PaidAmount:        inv.PaidAmount,
				Outstanding:       outstanding,
				DueDate:           inv.DueDate.Format("2006-01-02"),
				Status:            inv.Status,
				ManualPaymentCode: inv.ManualPaymentCode,
			}, nil
		}
	}

	return PublicBillView{}, domainCustomer.ErrCustomerNotFound
}

// ─── helpers ────────────────────────────────────────────────────────────

func (u *UseCase) resolve(ctx context.Context, identifier string) (domainCustomer.Customer, error) {
	id := strings.TrimSpace(identifier)
	candidates := []string{id}
	if n := phone.Normalize(id); n != "" && n != id {
		candidates = append(candidates, n)
	}
	// Coba tiap kandidat sebagai nomor HP, lalu sebagai kode portal.
	for _, cand := range candidates {
		if c, err := u.customs.FindByPhone(ctx, cand); err == nil {
			return c, nil
		}
	}
	for _, cand := range candidates {
		if c, err := u.customs.FindByPortalAccessCode(ctx, cand); err == nil {
			return c, nil
		}
	}
	return domainCustomer.Customer{}, domainCustomer.ErrCustomerNotFound
}

func generateDigits(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = '0' + b[i]%10
	}
	return string(b)
}

func hashCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func newToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func maskPhone(p string) string {
	if len(p) < 5 {
		return "***"
	}
	return p[:3] + "****" + p[len(p)-3:]
}

func atoiDefault(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func maskName(name string) string {
	parts := strings.Fields(name)
	for i, p := range parts {
		r := []rune(p)
		if len(r) > 2 {
			for j := 1; j < len(r)-1; j++ {
				r[j] = '*'
			}
			parts[i] = string(r)
		} else if len(r) == 2 {
			parts[i] = string(r[0]) + "*"
		}
	}
	return strings.Join(parts, " ")
}
