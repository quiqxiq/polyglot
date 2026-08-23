package registration

import (
	"context"
	"fmt"
	"time"

	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// ConvertDeps adalah port persistensi untuk konversi pendaftaran menjadi
// pelanggan aktif.
type ConvertDeps struct {
	Repo      port.RegistrationRepository
	Plans     port.ServicePlanRepository
	Customers port.CustomerRepository
	Subs      port.SubscriptionRepository
	Invoices  port.InvoiceRepository
	Audit     port.AuditLogWriter

	// Manager opsional: bila diisi dan deviceID diberikan, akun pelanggan
	// diprovisikan ke router segera setelah artefak DB jadi.
	Manager port.RouterAccountManager
}

// ConvertUseCase turns an INSTALLED registration into a live customer with
// subscription + first invoice, then links artifact IDs back to the
// registration row (DATABASE-SCHEMA-ISP.md §2.3 → §2.5–2.6).
type ConvertUseCase struct {
	deps ConvertDeps

	now       func() time.Time
	genCode   func() string // customer_code kandidat
	genPortal func() string // portal_access_code kandidat
	genSecret func() string // password akun jaringan
}

// NewConvertUseCase wires dependencies; generators default ke idgen.
func NewConvertUseCase(deps ConvertDeps) *ConvertUseCase {
	return &ConvertUseCase{
		deps:      deps,
		now:       time.Now,
		genCode:   func() string { return "CUST-" + idgen.Digits(5) },
		genPortal: func() string { return idgen.Digits(8) },
		genSecret: func() string { return idgen.Digits(6) + "pg" },
	}
}

// WithGenerators overrides generators (untuk test deterministik).
func (u *ConvertUseCase) WithGenerators(code, portal, secret func() string) {
	u.genCode, u.genPortal, u.genSecret = code, portal, secret
}

// Convert executes INSTALLED → ACTIVE tanpa penugasan device (akun belum
// diprovisikan; worker akan mencoba begitu device ditugaskan).
func (u *ConvertUseCase) Convert(ctx context.Context, regID, actorID string) (domainRegistration.Registration, error) {
	return u.ConvertWithDevice(ctx, regID, "", actorID)
}

// ConvertWithDevice executes INSTALLED → ACTIVE dan, bila deviceID diberikan,
// langsung memprovisikan akun ke router. Gagal router TIDAK menggagalkan
// konversi: status provisi PENDING dan worker lifecycle mencoba ulang.
func (u *ConvertUseCase) ConvertWithDevice(ctx context.Context, regID, deviceID, actorID string) (domainRegistration.Registration, error) {
	reg, err := u.deps.Repo.FindByID(ctx, regID)
	if err != nil {
		return domainRegistration.Registration{}, ErrNotFound
	}
	if reg.Status != domainRegistration.StatusInstalled {
		return domainRegistration.Registration{}, fmt.Errorf("%w: want %s, got %s",
			ErrInvalidTransition, domainRegistration.StatusInstalled, reg.Status)
	}
	if reg.CustomerID != "" {
		return domainRegistration.Registration{}, fmt.Errorf("%w: already converted (%s)", ErrInvalidTransition, reg.CustomerID)
	}

	pl, err := u.deps.Plans.FindByID(ctx, reg.PlanID)
	if err != nil {
		return domainRegistration.Registration{}, fmt.Errorf("plan %s: %w", reg.PlanID, err)
	}
	now := u.now()

	cust, err := u.createCustomer(ctx, reg, now)
	if err != nil {
		return domainRegistration.Registration{}, err
	}
	sub, err := u.createSubscription(ctx, reg, pl, cust.ID, deviceID, now)
	if err != nil {
		return domainRegistration.Registration{}, err
	}
	inv, items := buildInvoice(reg, pl, sub.ID, now)
	inv.CustomerID = cust.ID
	if err := u.deps.Invoices.SaveWithItems(ctx, inv, items); err != nil {
		return domainRegistration.Registration{}, fmt.Errorf("save invoice: %w", err)
	}

	reg.CustomerID = cust.ID
	reg.SubscriptionID = sub.ID
	reg.InvoiceID = inv.ID
	reg.Status = domainRegistration.StatusActive
	if err := u.deps.Repo.Save(ctx, reg); err != nil {
		return domainRegistration.Registration{}, err
	}
	writeAudit(ctx, u.deps.Audit, actorID, "CONVERT_REGISTRATION", "registration", reg.ID)
	return reg, nil
}
