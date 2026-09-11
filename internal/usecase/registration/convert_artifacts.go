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
	billingDay := now.Day()
	if billingDay > 28 {
		billingDay = 28
	}
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
		BillingDay:         billingDay,
		AutoIsolate:        true,
		IsolationGraceDays: 3,
		Status:             domainSubscription.StatusActive,
		StartDate:          time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
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

// buildInvoice menyusun faktur bulan pertama: fee langganan + biaya pasang
// (bila ada), lalu pajak dari tax_percent paket dihitung dari subtotal.
func buildInvoice(reg domainRegistration.Registration, pl domainPlan.ServicePlan, subscriptionID string, now time.Time) (domainBilling.Invoice, []domainBilling.InvoiceItem) {
	base := pl.Price
	fee := pl.InstallationFee
	subtotal := base + fee
	tax := subtotal * pl.TaxPercent / 100
	total := subtotal + tax

	period := now.Format("2006-01")
	due := endOfMonth(now)

	invID := idgen.New("inv")
	inv := domainBilling.Invoice{
		ID:                invID,
		TenantID:          orTenant(reg.TenantID),
		InvoiceNumber:     fmt.Sprintf("INV-%s-%s", now.Format("200601"), idgen.Digits(6)),
		CustomerID:        "", // diisi pemanggil setelah customer dibuat
		SubscriptionID:    &subscriptionID,
		Period:            period,
		Subtotal:          subtotal,
		TaxAmount:         tax,
		Total:             total,
		DueDate:           due,
		Status:            domainBilling.StatusUnpaid,
		QRPayload:         "polyglot://invoice/" + invID,
		ManualPaymentCode: "PAY-" + idgen.Digits(6),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	items := []domainBilling.InvoiceItem{{
		ID:          idgen.New("itm"),
		InvoiceID:   invID,
		Description: fmt.Sprintf("Paket %s (%s)", pl.Name, period),
		Quantity:    1,
		UnitPrice:   base,
		Amount:      base,
		ItemType:    domainBilling.ItemTypeSubscriptionFee,
		CreatedAt:   now,
	}}
	if fee > 0 {
		items = append(items, domainBilling.InvoiceItem{
			ID:          idgen.New("itm"),
			InvoiceID:   invID,
			Description: "Biaya pemasangan awal",
			Quantity:    1,
			UnitPrice:   fee,
			Amount:      fee,
			ItemType:    domainBilling.ItemTypeInstallationFee,
			CreatedAt:   now,
		})
	}
	return inv, items
}

func endOfMonth(t time.Time) time.Time {
	firstNext := time.Date(t.Year(), t.Month()+1, 1, 23, 59, 59, 0, t.Location())
	return firstNext.AddDate(0, 0, -1)
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
