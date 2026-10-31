package registration

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/audit"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	planUC "github.com/quixiq/polyglot/internal/usecase/plan"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/logger"
)

func orTenant(t string) string {
	if t == "" {
		return "tenant-default"
	}
	return t
}

// planServiceType memetakan tipe paket ke service_type langganan.
func planServiceType(pl domainPlan.ServicePlan) string {
	switch pl.ServiceType {
	case domainPlan.TypeHotspot:
		return "HOTSPOT"
	case domainPlan.TypeDedicated:
		return "DEDICATED"
	default:
		return "PPPOE"
	}
}

func (u *ConvertUseCase) buildCustomer(ctx context.Context, reg domainRegistration.Registration, now time.Time) (domainCustomer.Customer, error) {
	cust := domainCustomer.Customer{
		ID:           idgen.New("cust"),
		TenantID:     orTenant(reg.TenantID),
		Name:         reg.FullName,
		Phone:        reg.Phone,
		Email:        reg.Email,
		Address:      reg.Address,
		Latitude:     reg.Latitude,
		Longitude:    reg.Longitude,
		Status:       domainCustomer.StatusActive,
		Notes:        reg.Notes,
		RegisteredAt: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	for attempt := 0; attempt < 8; attempt++ {
		cust.CustomerCode = u.genCode()
		if _, err := u.deps.Customers.FindByCustomerCode(ctx, cust.CustomerCode); err != nil {
			break // kode belum terpakai
		}
	}
	for attempt := 0; attempt < 8; attempt++ {
		cust.PortalAccessCode = u.genPortal()
		if _, err := u.deps.Customers.FindByPortalAccessCode(ctx, cust.PortalAccessCode); err != nil {
			break
		}
	}
	return cust, nil
}

func (u *ConvertUseCase) buildSubscription(
	ctx context.Context,
	reg domainRegistration.Registration,
	pl domainPlan.ServicePlan,
	cust domainCustomer.Customer,
	deviceID string,
	now time.Time,
) (domainSubscription.Subscription, error) {
	pattern, prefix, pwMode := "{initials}{digits4}", "", "digits6"
	if u.deps.Settings != nil {
		pattern = u.deps.Settings.GetValue(ctx, "isp.pppoe_username_pattern", pattern)
		prefix = u.deps.Settings.GetValue(ctx, "isp.pppoe_username_prefix", prefix)
		pwMode = u.deps.Settings.GetValue(ctx, "isp.pppoe_password_mode", pwMode)
	}
	username := idgen.GenerateUsername(reg.FullName, pattern, prefix, cust.CustomerCode)
	password := idgen.Password(pwMode, cust.Phone)
	if u.genSecret != nil {
		password = u.genSecret()
	}
	endDate := now.AddDate(0, 0, 30)
	// Anchor billing global (planning v2 D1): hari anchor dari setting
	// isp.billing_cycle_start_day (default 1) sebagai acuan siklus tagihan.
	anchor := 1
	if u.deps.Settings != nil {
		anchor = atoiClamped(u.deps.Settings.GetValue(ctx, "isp.billing_cycle_start_day", "1"), 1, 28)
	}
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	nextDue := nextAnchorDate(start, anchor)
	provStatus := domainSubscription.ProvisionNone
	var devRef *string
	if deviceID != "" {
		devRef = &deviceID
		provStatus = domainSubscription.ProvisionPending // worker retry fallback
	}
	sub := domainSubscription.Subscription{
		ID:                 idgen.New("sub"),
		TenantID:           orTenant(reg.TenantID),
		CustomerID:         cust.ID,
		PlanID:             pl.ID,
		DeviceID:           devRef,
		ServiceType:        planServiceType(pl),
		RemoteUsername:     username,
		RemotePassword:     password,
		BillingCycle:       domainSubscription.CycleMonthly,
		BillingDay:         anchor,
		BillingAnchorDay:   anchor,
		NextDueDate:        &nextDue,
		AutoIsolate:        true,
		IsolationGraceDays: 3,
		Status:             domainSubscription.StatusActive,
		StartDate:          start,
		EndDate:            &endDate,
		ProvisionStatus:    provStatus,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	return sub, nil
}

// provisionSubscription membangun spec kanonik dari planUC (termasuk burst,
// parent queue, timeout, pool, dan config bertipe) lalu menjalankannya ke
// router. Dipisah dari buildSubscription agar provisi berjalan setelah
// transaksi konversi commit.
func provisionSubscription(ctx context.Context, mgr port.RouterAccountManager, deviceID string, sub domainSubscription.Subscription, pl domainPlan.ServicePlan) error {
	switch {
	case strings.EqualFold(sub.ServiceType, "HOTSPOT"):
		spec := planUC.BuildHotspotProvisionSpec(sub, pl)
		if err := mgr.ProvisionHotspot(ctx, deviceID, spec); err != nil {
			return fmt.Errorf("provision hotspot %s: %w", sub.RemoteUsername, err)
		}
	case strings.EqualFold(sub.ServiceType, "DEDICATED"):
		spec := planUC.BuildDedicatedProvisionSpec(sub, pl)
		if err := mgr.ProvisionDedicated(ctx, deviceID, spec); err != nil {
			return fmt.Errorf("provision dedicated %s: %w", sub.RemoteUsername, err)
		}
	default:
		spec := planUC.BuildPPPoEProvisionSpec(sub, pl)
		if err := mgr.ProvisionPPPoE(ctx, deviceID, spec); err != nil {
			return fmt.Errorf("provision pppoe %s: %w", sub.RemoteUsername, err)
		}
	}
	return nil
}

// buildInvoice menyusun faktur pasang baru (planning v2 D3): biaya prorata
// paket (hari aktif s/d H-1 anchor berikutnya) + biaya pasang − diskon,
// lalu PPN dari subtotal setelah diskon. Due date = tgl terbit + isp.first_invoice_due_days.
func (u *ConvertUseCase) buildInvoice(ctx context.Context, reg domainRegistration.Registration, pl domainPlan.ServicePlan, subscriptionID string, now time.Time) (domainBilling.Invoice, []domainBilling.InvoiceItem) {
	anchor := 1
	firstDueDays := 7
	if u.deps.Settings != nil {
		anchor = atoiClamped(u.deps.Settings.GetValue(ctx, "isp.billing_cycle_start_day", "1"), 1, 28)
		firstDueDays = atoiClamped(u.deps.Settings.GetValue(ctx, "isp.first_invoice_due_days", "7"), 0, 60)
	}
	installDate := now
	if reg.InstalledAt != nil {
		installDate = *reg.InstalledAt
	}
	mode := domainBilling.ProrateMode(pl.ProrateMode)
	if mode == "" {
		mode = domainBilling.ProrateModeCalendarDay
	}
	discountType, discountValue := domainBilling.DiscountTypeNone, 0.0
	if reg.DiscountAmount > 0 {
		discountType, discountValue = domainBilling.DiscountTypeAmount, reg.DiscountAmount
	}
	inv, items := domainBilling.BuildFirstInvoice(domainBilling.FirstInvoiceInput{
		InvoiceID:         idgen.New("inv"),
		InvoiceNumber:     fmt.Sprintf("INV-%s-%s", now.Format("200601"), idgen.Digits(6)),
		ManualPaymentCode: "PAY-" + idgen.Digits(6),
		ItemIDFee:         idgen.New("itm"),
		ItemIDInstall:     idgen.New("itm"),
		TenantID:          orTenant(reg.TenantID),
		SubscriptionID:    subscriptionID,
		PlanName:          pl.Name,
		PlanPrice:         pl.Price,
		InstallationFee:   pl.InstallationFee,
		TaxPercent:        pl.TaxPercent,
		ProrateMode:       mode,
		DiscountType:      discountType,
		DiscountValue:     discountValue,
		InstallDate:       installDate,
		AnchorDay:         anchor,
		FirstDueDays:      firstDueDays,
		Now:               now,
	})
	return inv, items
}

// nextAnchorDate returns the first date strictly after start whose day equals
// anchorDay (dijepit ke hari terakhir bulan pendek, mis. anchor 28 di Februari).
func nextAnchorDate(start time.Time, anchorDay int) time.Time {
	anchorDay = atoiClamped(fmt.Sprint(anchorDay), 1, 28)
	d := time.Date(start.Year(), start.Month(), anchorDay, 0, 0, 0, 0, time.UTC)
	if !d.After(start) {
		d = time.Date(start.Year(), start.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	}
	last := time.Date(d.Year(), d.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if anchorDay > last {
		d = time.Date(d.Year(), d.Month(), last, 0, 0, 0, 0, time.UTC)
	}
	return d
}

// atoiClamped parses s as an integer clamped to [min, max]; fallback min on error.
func atoiClamped(s string, min, max int) int {
	var n int
	if _, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &n); err != nil {
		return min
	}
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

func writeAudit(ctx context.Context, w port.AuditLogWriter, actorID, action, entityType, entityID string) {
	if w == nil {
		return
	}
	err := w.Write(ctx, audit.AuditLog{
		TenantID:   "tenant-default",
		ActorType:  audit.ActorUser,
		ActorID:    actorID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
	})
	if err != nil {
		logger.WithComponent("RegistrationUC").WithError(err).Warn("audit log write failed")
	}
}
