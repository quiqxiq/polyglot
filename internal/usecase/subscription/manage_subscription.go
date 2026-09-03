package subscription

import (
	"context"
	"fmt"
	"strings"
	"time"

	domainAudit "github.com/quixiq/polyglot/internal/domain/audit"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainDevice "github.com/quixiq/polyglot/internal/domain/device"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	planUC "github.com/quixiq/polyglot/internal/usecase/plan"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/logger"
)

// Detail menggabungkan Subscription dengan metadata Plan, Customer, dan Device
// untuk denormalisasi respon tampilan (Zero Waterfall di Frontend).
type Detail struct {
	Subscription domainSub.Subscription
	Plan         *domainPlan.ServicePlan
	Customer     *domainCustomer.Customer
	Device       *domainDevice.Device
}

// CreateInput adalah payload pembuatan langganan baru. Kredensial boleh
// kosong — akan di-generate otomatis berbasis inisial/nama pelanggan.
type CreateInput struct {
	CustomerID     string
	PlanID         string
	DeviceID       *string
	ServiceType    string
	RemoteUsername string
	RemotePassword string
	LocalAddress   string
	RemoteAddress  string
	RateLimit      string
	CustomPrice    *float64
	BillingCycle   string
	BillingDay     int
	Notes          string
}

// UpdateInput memakai pointer semantics: hanya field non-nil yang diubah.
// RemotePassword pointer ke "" berarti TIDAK diubah (skip).
type UpdateInput struct {
	RemoteUsername *string
	RemotePassword *string // kosong = tidak diubah
	LocalAddress   *string
	RemoteAddress  *string
	RateLimit      *string
	CustomPrice    *float64
	BillingCycle   *string
	BillingDay     *int
	DeviceID       *string
	Notes          *string
}

func isHotspot(s string) bool   { return strings.EqualFold(s, "HOTSPOT") }
func isDedicated(s string) bool { return strings.EqualFold(s, "DEDICATED") }

func provisioned(sub domainSub.Subscription) bool {
	return sub.DeviceID != nil && *sub.DeviceID != "" &&
		sub.ProvisionStatus == domainSub.ProvisionOK
}

func derefDevice(d *string) string {
	if d == nil {
		return ""
	}
	return *d
}

// ManageSubscriptionUseCase mengelola CRUD langganan (Create/Update/Delete/List/Get/Enrich)
// di atas domain subscription.
type ManageSubscriptionUseCase struct {
	subs      port.SubscriptionRepository
	plans     port.ServicePlanRepository
	customers port.CustomerRepository
	devices   port.DeviceRepository
	manager   port.RouterAccountManager
	audit     port.AuditLogWriter
	invoices  port.InvoiceRepository // guard delete

	now func() time.Time
}

// NewManageSubscriptionUseCase wires dependencies.
func NewManageSubscriptionUseCase(
	subs port.SubscriptionRepository,
	plans port.ServicePlanRepository,
	customers port.CustomerRepository,
	devices port.DeviceRepository,
	manager port.RouterAccountManager,
	auditW port.AuditLogWriter,
	invoices port.InvoiceRepository,
) *ManageSubscriptionUseCase {
	return &ManageSubscriptionUseCase{
		subs:      subs,
		plans:     plans,
		customers: customers,
		devices:   devices,
		manager:   manager,
		audit:     auditW,
		invoices:  invoices,
		now:       time.Now,
	}
}

// Enrich memperkaya model domain Subscription dengan metadata Plan, Customer, dan Device.
func (u *ManageSubscriptionUseCase) Enrich(ctx context.Context, sub domainSub.Subscription) Detail {
	detail := Detail{Subscription: sub}
	if u.plans != nil && sub.PlanID != "" {
		if pl, err := u.plans.FindByID(ctx, sub.PlanID); err == nil {
			detail.Plan = &pl
		}
	}
	if u.customers != nil && sub.CustomerID != "" {
		if cust, err := u.customers.FindByID(ctx, sub.CustomerID); err == nil {
			detail.Customer = &cust
		}
	}
	if u.devices != nil && sub.DeviceID != nil && *sub.DeviceID != "" {
		if dev, err := u.devices.FindByID(ctx, *sub.DeviceID); err == nil {
			detail.Device = &dev
		}
	}
	return detail
}

// ListSubscriptions mengembalikan daftar langganan yang diperkaya metadata plan/customer/device.
func (u *ManageSubscriptionUseCase) ListSubscriptions(ctx context.Context, customerID string) ([]Detail, error) {
	if u.subs == nil {
		return nil, domainSub.ErrNotFound
	}
	var subs []domainSub.Subscription
	var err error
	if customerID != "" {
		subs, err = u.subs.FindByCustomerID(ctx, customerID)
	} else {
		subs, err = u.subs.FindAll(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("find subscriptions: %w", err)
	}

	out := make([]Detail, len(subs))
	for i, sub := range subs {
		out[i] = u.Enrich(ctx, sub)
	}
	return out, nil
}

// GetSubscription mengembalikan detail langganan berdasarkan ID yang diperkaya metadata.
func (u *ManageSubscriptionUseCase) GetSubscription(ctx context.Context, id string) (Detail, error) {
	if u.subs == nil {
		return Detail{}, domainSub.ErrNotFound
	}
	sub, err := u.subs.FindByID(ctx, id)
	if err != nil {
		return Detail{}, fmt.Errorf("find subscription %s: %w", id, err)
	}
	return u.Enrich(ctx, sub), nil
}

// Create memvalidasi input, menyimpan langganan baru, dan (bila deviceID
// diberikan) memprovisikan akun PPP/hotspot ke router target.
func (u *ManageSubscriptionUseCase) Create(ctx context.Context, in CreateInput) (domainSub.Subscription, error) {
	if in.CustomerID == "" || in.PlanID == "" {
		return domainSub.Subscription{}, fmt.Errorf("%w: customer_id and plan_id are required", domainSub.ErrInvalidInput)
	}
	cust, err := u.customers.FindByID(ctx, in.CustomerID)
	if err != nil {
		return domainSub.Subscription{}, fmt.Errorf("%w: customer %s not found", domainSub.ErrInvalidInput, in.CustomerID)
	}
	pl, err := u.plans.FindByID(ctx, in.PlanID)
	if err != nil {
		return domainSub.Subscription{}, fmt.Errorf("%w: plan %s not found", domainSub.ErrInvalidInput, in.PlanID)
	}
	if !pl.IsActive {
		return domainSub.Subscription{}, fmt.Errorf("%w: plan %s is inactive", domainSub.ErrInvalidInput, pl.ID)
	}
	serviceType := in.ServiceType
	if serviceType == "" {
		serviceType = pl.ServiceType
	}
	if !strings.EqualFold(serviceType, pl.ServiceType) {
		return domainSub.Subscription{}, fmt.Errorf("%w: service_type %s does not match plan (%s)",
			domainSub.ErrInvalidInput, serviceType, pl.ServiceType)
	}

	username, password := in.RemoteUsername, in.RemotePassword
	if username == "" {
		username = idgen.GenerateUsername(cust.Name, "{initials}{digits4}", "", cust.CustomerCode)
	}
	if password == "" {
		password = idgen.Digits(6) + "pg"
	}

	now := u.now()
	sub := domainSub.Subscription{
		ID:              "SUB-" + idgen.Digits(8),
		TenantID:        pl.TenantID,
		CustomerID:      in.CustomerID,
		PlanID:          in.PlanID,
		DeviceID:        in.DeviceID,
		ServiceType:     strings.ToUpper(serviceType),
		RemoteUsername:  username,
		RemotePassword:  password,
		LocalAddress:    in.LocalAddress,
		RemoteAddress:   in.RemoteAddress,
		RateLimit:       in.RateLimit,
		RouterProfile:   pl.Name,
		ProvisionStatus: domainSub.ProvisionNone,
		BillingCycle:    in.BillingCycle,
		BillingDay:      in.BillingDay,
		Status:          domainSub.StatusPending,
		StartDate:       now,
		CustomPrice:     in.CustomPrice,
		Notes:           in.Notes,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if sub.DeviceID != nil && *sub.DeviceID != "" && u.manager != nil {
		var perr error
		if isHotspot(sub.ServiceType) {
			hotSpec := planUC.BuildHotspotProvisionSpec(sub, pl)
			perr = u.manager.ProvisionHotspot(ctx, *sub.DeviceID, hotSpec)
		} else if isDedicated(sub.ServiceType) {
			dedSpec := planUC.BuildDedicatedProvisionSpec(sub, pl)
			perr = u.manager.ProvisionDedicated(ctx, *sub.DeviceID, dedSpec)
		} else {
			pppSpec := planUC.BuildPPPoEProvisionSpec(sub, pl)
			perr = u.manager.ProvisionPPPoE(ctx, *sub.DeviceID, pppSpec)
		}
		if perr != nil {
			logger.WithComponent("ManageSubscriptionUC").WithError(perr).WithFields(map[string]any{
				"subscription_id": sub.ID,
				"device_id":       *sub.DeviceID,
			}).Warn("provisi awal gagal; worker akan mencoba ulang")
			sub.ProvisionStatus = domainSub.ProvisionPending
		} else {
			sub.ProvisionStatus = domainSub.ProvisionOK
			sub.Status = domainSub.StatusActive
		}
	}

	if err := u.subs.Save(ctx, sub); err != nil {
		return domainSub.Subscription{}, fmt.Errorf("save subscription: %w", err)
	}
	u.writeAudit(ctx, "CREATE_SUBSCRIPTION", "subscription", sub.ID)
	return sub, nil
}

// Update menerapkan hanya field pointer non-nil pada langganan existing.
func (u *ManageSubscriptionUseCase) Update(ctx context.Context, subID string, in UpdateInput) (domainSub.Subscription, error) {
	sub, err := u.subs.FindByID(ctx, subID)
	if err != nil {
		return domainSub.Subscription{}, domainSub.ErrNotFound
	}

	if in.RemoteUsername != nil && *in.RemoteUsername != "" {
		sub.RemoteUsername = *in.RemoteUsername
	}
	if in.RemotePassword != nil && *in.RemotePassword != "" { // kosong = skip
		sub.RemotePassword = *in.RemotePassword
	}
	if in.CustomPrice != nil {
		sub.CustomPrice = in.CustomPrice
	}
	if in.BillingCycle != nil && *in.BillingCycle != "" {
		sub.BillingCycle = *in.BillingCycle
	}
	if in.BillingDay != nil {
		sub.BillingDay = *in.BillingDay
	}
	if in.DeviceID != nil {
		sub.DeviceID = in.DeviceID
	}
	if in.LocalAddress != nil {
		sub.LocalAddress = *in.LocalAddress
	}
	if in.RemoteAddress != nil {
		sub.RemoteAddress = *in.RemoteAddress
	}
	if in.RateLimit != nil {
		sub.RateLimit = *in.RateLimit
	}
	if in.Notes != nil {
		sub.Notes = *in.Notes
	}
	sub.UpdatedAt = u.now()

	if err := u.subs.Save(ctx, sub); err != nil {
		return sub, fmt.Errorf("save subscription: %w", err)
	}
	u.writeAudit(ctx, "UPDATE_SUBSCRIPTION", "subscription", sub.ID)
	return sub, nil
}

// Delete menghapus langganan: guard tagihan dulu, terminate akun router
// best-effort (gagal hanya warn), lalu hard delete baris DB.
func (u *ManageSubscriptionUseCase) Delete(ctx context.Context, subID string) error {
	if u.invoices != nil {
		hasInvoice, err := u.invoices.HasForSubscription(ctx, subID)
		if err != nil {
			return fmt.Errorf("check subscription invoices %s: %w", subID, err)
		}
		if hasInvoice {
			return fmt.Errorf("%w: subscription still has invoices", domainSub.ErrInvalidInput)
		}
	}

	sub, err := u.subs.FindByID(ctx, subID)
	if err != nil {
		return domainSub.ErrNotFound
	}
	if provisioned(sub) && u.manager != nil {
		if terr := u.manager.Terminate(ctx, derefDevice(sub.DeviceID), sub.ServiceType, sub.RemoteUsername); terr != nil {
			logger.WithComponent("ManageSubscriptionUC").WithError(terr).WithFields(map[string]any{
				"subscription_id": subID,
				"username":        sub.RemoteUsername,
			}).Warn("terminate akun router gagal saat delete; lanjut hapus DB")
		}
	}

	if err := u.subs.Delete(ctx, subID); err != nil {
		return fmt.Errorf("delete subscription %s: %w", subID, err)
	}
	u.writeAudit(ctx, "DELETE_SUBSCRIPTION", "subscription", subID)
	return nil
}

func (u *ManageSubscriptionUseCase) writeAudit(ctx context.Context, action, entityType, entityID string) {
	if u.audit == nil {
		return
	}
	err := u.audit.Write(ctx, domainAudit.AuditLog{
		TenantID:   "tenant-default",
		ActorType:  domainAudit.ActorUser,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
	})
	if err != nil {
		logger.WithComponent("ManageSubscriptionUC").WithError(err).Warn("audit log write failed")
	}
}
