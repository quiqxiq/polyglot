package registration

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/quixiq/polyglot/internal/domain/audit"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
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
	default:
		return "PPPOE"
	}
}

func (u *ConvertUseCase) createCustomer(ctx context.Context, reg domainRegistration.Registration, now time.Time) (domainCustomer.Customer, error) {
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
	if err := u.deps.Customers.Save(ctx, cust); err != nil {
		return domainCustomer.Customer{}, fmt.Errorf("create customer: %w", err)
	}
	writeAudit(ctx, u.deps.Audit, "", "CREATE_CUSTOMER", "customer", cust.ID)
	return cust, nil
}

func (u *ConvertUseCase) createSubscription(ctx context.Context, reg domainRegistration.Registration, pl domainPlan.ServicePlan, customerID, deviceID string, now time.Time) (domainSubscription.Subscription, error) {
	username := idgen.Slug(reg.FullName)
	if username == "" {
		username = "USER"
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
		CustomerID:         customerID,
		PlanID:             pl.ID,
		DeviceID:           devRef,
		ServiceType:        planServiceType(pl),
		RemoteUsername:     username,
		RemotePassword:     u.genSecret(),
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
	if err := u.deps.Subs.Save(ctx, sub); err != nil {
		return domainSubscription.Subscription{}, fmt.Errorf("create subscription: %w", err)
	}

	// Provisioning langsung bila manager tersedia & device ditugaskan.
	if deviceID != "" && u.deps.Manager != nil {
		err := u.deps.Manager.Provision(ctx, deviceID, sub.ServiceType, port.SubscriberAccount{
			Username:  sub.RemoteUsername,
			Password:  sub.RemotePassword,
			Profile:   pl.Name,
			RateLimit: planRate(pl),
			Comment:   "polyglot:" + sub.ID,
		})
		if err != nil {
			logger.WithComponent("ConvertUC").WithError(err).Warn("provisioning gagal; worker akan mencoba ulang")
		} else {
			sub.ProvisionStatus = domainSubscription.ProvisionOK
			sub.RouterProfile = pl.Name
			if serr := u.deps.Subs.Save(ctx, sub); serr != nil {
				return domainSubscription.Subscription{}, serr
			}
		}
	}
	writeAudit(ctx, u.deps.Audit, "", "CREATE_SUBSCRIPTION", "subscription", sub.ID)
	return sub, nil
}

// buildInvoice menyusun faktur bulan pertama: fee langganan (+ biaya pasang
// bila ada), pajak dari tax_percent paket.
func buildInvoice(reg domainRegistration.Registration, pl domainPlan.ServicePlan, subscriptionID string, now time.Time) (domainBilling.Invoice, []domainBilling.InvoiceItem) {
	base := pl.Price
	tax := base * pl.TaxPercent / 100
	total := base + tax

	period := now.Format("2006-01")
	due := endOfMonth(now)

	invID := idgen.New("inv")
	inv := domainBilling.Invoice{
		ID:                invID,
		TenantID:          orTenant(reg.TenantID),
		InvoiceNumber:     fmt.Sprintf("INV-%s-%04d", now.Format("200601"), now.UnixNano()%10000),
		CustomerID:        "", // diisi pemanggil setelah customer dibuat
		SubscriptionID:    &subscriptionID,
		Period:            period,
		Subtotal:          base,
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
	if pl.InstallationFee > 0 {
		items = append(items, domainBilling.InvoiceItem{
			ID:          idgen.New("itm"),
			InvoiceID:   invID,
			Description: "Biaya pemasangan awal",
			Quantity:    1,
			UnitPrice:   pl.InstallationFee,
			Amount:      pl.InstallationFee,
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

// planRate memformat rate-limit MikroTik dari bandwidth paket ("5M/5M").
func planRate(pl domainPlan.ServicePlan) string {
	side := func(kbps int) string {
		if kbps <= 0 {
			return ""
		}
		if kbps >= 1000 {
			return strconv.Itoa((kbps+500)/1000) + "M"
		}
		return strconv.Itoa(kbps) + "k"
	}
	dl, ul := side(pl.BandwidthDownloadKbps), side(pl.BandwidthUploadKbps)
	if dl == "" && ul == "" {
		return ""
	}
	return dl + "/" + ul
}
