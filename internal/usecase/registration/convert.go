package registration

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/logger"
)

// ConvertDeps adalah port persistensi untuk konversi pendaftaran menjadi
// pelanggan aktif.
type ConvertDeps struct {
	Repo      port.RegistrationRepository
	Plans     port.ServicePlanRepository
	Customers port.CustomerRepository
	Subs      port.SubscriptionRepository
	Audit     port.AuditLogWriter

	// Writer mempersistensikan customer + subscription + invoice + registrasi
	// secara atomik (satu transaksi DB). Wajib diisi.
	Writer port.ConversionWriter

	// Settings opsional: pola username/password otomatis (migrasi 000023).
	Settings port.SettingReader

	// Manager opsional: bila diisi dan deviceID diberikan, akun pelanggan
	// diprovisikan ke router segera setelah artefak DB ter-commit.
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
		// genSecret nil = ikuti setting isp.pppoe_password_mode / default digits6.
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

// ConvertWithDevice executes INSTALLED → ACTIVE secara atomik: customer,
// subscription, invoice pertama, dan tautan registrasi disimpan dalam satu
// transaksi lewat port.ConversionWriter. Bila deviceID diberikan, akun
// diprovisikan ke router best-effort SETELAH commit — kegagalan router tidak
// menggagalkan konversi (provision_status PENDING untuk retry worker).
func (u *ConvertUseCase) ConvertWithDevice(ctx context.Context, regID, deviceID, actorID string) (domainRegistration.Registration, error) {
	reg, err := u.deps.Repo.FindByID(ctx, regID)
	if err != nil {
		return domainRegistration.Registration{}, domainRegistration.ErrNotFound
	}
	if reg.Status != domainRegistration.StatusInstalled {
		return domainRegistration.Registration{}, fmt.Errorf("%w: want %s, got %s",
			domainRegistration.ErrInvalidTransition, domainRegistration.StatusInstalled, reg.Status)
	}
	if reg.CustomerID != "" {
		return domainRegistration.Registration{}, fmt.Errorf("%w: already converted (%s)", domainRegistration.ErrInvalidTransition, reg.CustomerID)
	}
	if u.deps.Writer == nil {
		return domainRegistration.Registration{}, errors.New("convert: conversion writer not configured")
	}
	// Fallback router: pilihan teknisi saat pemasangan (F2-7) bila pemanggil
	// tidak menentukan device.
	if deviceID == "" {
		deviceID = reg.TargetDeviceID
	}

	pl, err := u.deps.Plans.FindByID(ctx, reg.PlanID)
	if err != nil {
		return domainRegistration.Registration{}, fmt.Errorf("plan %s: %w", reg.PlanID, err)
	}
	now := u.now()

	cust, err := u.buildCustomer(ctx, reg, now)
	if err != nil {
		return domainRegistration.Registration{}, err
	}
	sub, err := u.buildSubscription(ctx, reg, pl, cust, deviceID, now)
	if err != nil {
		return domainRegistration.Registration{}, err
	}
	inv, items := buildInvoice(reg, pl, sub.ID, now)
	inv.CustomerID = cust.ID

	reg.CustomerID = cust.ID
	reg.SubscriptionID = sub.ID
	reg.InvoiceID = inv.ID
	reg.Status = domainRegistration.StatusActive

	if err := u.deps.Writer.SaveConversion(ctx, port.ConversionArtifacts{
		Registration: reg,
		Customer:     cust,
		Subscription: sub,
		Invoice:      inv,
		Items:        items,
	}); err != nil {
		return domainRegistration.Registration{}, fmt.Errorf("save conversion: %w", err)
	}

	writeAudit(ctx, u.deps.Audit, "", "CREATE_CUSTOMER", "customer", cust.ID)
	writeAudit(ctx, u.deps.Audit, "", "CREATE_SUBSCRIPTION", "subscription", sub.ID)
	writeAudit(ctx, u.deps.Audit, actorID, "CONVERT_REGISTRATION", "registration", reg.ID)

	u.provisionAfterConvert(ctx, &sub, pl, deviceID)
	return reg, nil
}

// provisionAfterConvert menjalankan provisi router pasca-commit (best-effort):
// kegagalan hanya dicatat, status provisi tetap PENDING agar lifecycle worker
// mencoba ulang.
func (u *ConvertUseCase) provisionAfterConvert(ctx context.Context, sub *domainSubscription.Subscription, pl domainPlan.ServicePlan, deviceID string) {
	if deviceID == "" || u.deps.Manager == nil {
		return
	}
	if err := provisionSubscription(ctx, u.deps.Manager, deviceID, *sub, pl); err != nil {
		logger.WithComponent("ConvertUC").WithError(err).Warn("provisioning gagal; worker akan mencoba ulang")
		return
	}
	sub.ProvisionStatus = domainSubscription.ProvisionOK
	sub.RouterProfile = pl.Name
	if err := u.deps.Subs.Save(ctx, *sub); err != nil {
		logger.WithComponent("ConvertUC").WithError(err).Warn("gagal simpan provision_status; worker akan mencoba ulang")
	}
}
