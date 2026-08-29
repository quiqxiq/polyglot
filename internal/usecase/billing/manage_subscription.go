package billing

import (
	"context"
	"fmt"
	"strings"
	"time"

	domainAudit "github.com/quixiq/polyglot/internal/domain/audit"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/logger"
)

// CreateInput adalah payload pembuatan langganan baru. Kredensial boleh
// kosong — akan di-generate otomatis (pola convert.go: "pg"+digits /
// digits+"pg").
type CreateInput struct {
	CustomerID     string
	PlanID         string
	DeviceID       *string
	ServiceType    string
	RemoteUsername string
	RemotePassword string
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
	CustomPrice    *float64
	BillingCycle   *string
	BillingDay     *int
	DeviceID       *string
	Notes          *string
}

// ManageSubscriptionUseCase mengelola CRUD langganan (Create/Update/Delete)
// di atas proto billing v1. Create memvalidasi customer & plan, generate
// kredensial bila kosong, dan memprovisikan akun router bila device
// ditugaskan — gagal provisi BUKAN error (PENDING, worker retry).
// Delete di-guard oleh keberadaan tagihan.
type ManageSubscriptionUseCase struct {
	subs      port.SubscriptionRepository
	plans     port.ServicePlanRepository
	customers port.CustomerRepository
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
	manager port.RouterAccountManager,
	auditW port.AuditLogWriter,
	invoices port.InvoiceRepository,
) *ManageSubscriptionUseCase {
	return &ManageSubscriptionUseCase{
		subs: subs, plans: plans, customers: customers,
		manager: manager, audit: auditW, invoices: invoices,
		now: time.Now,
	}
}

// Create memvalidasi input, menyimpan langganan baru, dan (bila deviceID
// diberikan) memprovisikan akun PPP/hotspot ke router target.
func (u *ManageSubscriptionUseCase) Create(ctx context.Context, in CreateInput) (domainSubscription.Subscription, error) {
	if in.CustomerID == "" || in.PlanID == "" {
		return domainSubscription.Subscription{}, fmt.Errorf("%w: customer_id and plan_id are required", domainBilling.ErrInvalidInput)
	}
	if _, err := u.customers.FindByID(ctx, in.CustomerID); err != nil {
		return domainSubscription.Subscription{}, fmt.Errorf("%w: customer %s not found", domainBilling.ErrInvalidInput, in.CustomerID)
	}
	pl, err := u.plans.FindByID(ctx, in.PlanID)
	if err != nil {
		return domainSubscription.Subscription{}, fmt.Errorf("%w: plan %s not found", domainBilling.ErrInvalidInput, in.PlanID)
	}
	if !pl.IsActive {
		return domainSubscription.Subscription{}, fmt.Errorf("%w: plan %s is inactive", domainBilling.ErrInvalidInput, pl.ID)
	}
	serviceType := in.ServiceType
	if serviceType == "" {
		serviceType = pl.ServiceType
	}
	if !strings.EqualFold(serviceType, pl.ServiceType) {
		return domainSubscription.Subscription{}, fmt.Errorf("%w: service_type %s does not match plan (%s)",
			domainBilling.ErrInvalidInput, serviceType, pl.ServiceType)
	}

	username, password := in.RemoteUsername, in.RemotePassword
	if username == "" {
		username = "pg" + idgen.Digits(6)
	}
	if password == "" {
		password = idgen.Digits(6) + "pg"
	}

	now := u.now()
	sub := domainSubscription.Subscription{
		ID:              "SUB-" + idgen.Digits(8),
		TenantID:        pl.TenantID,
		CustomerID:      in.CustomerID,
		PlanID:          in.PlanID,
		DeviceID:        in.DeviceID,
		ServiceType:     strings.ToUpper(serviceType),
		RemoteUsername:  username,
		RemotePassword:  password,
		RouterProfile:   pl.Name,
		ProvisionStatus: domainSubscription.ProvisionNone,
		BillingCycle:    in.BillingCycle,
		BillingDay:      in.BillingDay,
		Status:          domainSubscription.StatusPending,
		StartDate:       now,
		CustomPrice:     in.CustomPrice,
		Notes:           in.Notes,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if sub.DeviceID != nil && *sub.DeviceID != "" {
		acct := subscriberAccountFromPlan(sub, pl)
		acct.Comment = "polyglot:" + sub.ID
		if perr := u.manager.Provision(ctx, *sub.DeviceID, sub.ServiceType, acct); perr != nil {
			// Gagal provisi bukan error: PENDING agar worker mencoba ulang.
			logger.WithComponent("ManageSubscriptionUC").WithError(perr).WithFields(map[string]any{
				"subscription_id": sub.ID,
				"device_id":       *sub.DeviceID,
			}).Warn("provisi awal gagal; worker akan mencoba ulang")
			sub.ProvisionStatus = domainSubscription.ProvisionPending
		} else {
			sub.ProvisionStatus = domainSubscription.ProvisionOK
			sub.Status = domainSubscription.StatusActive
		}
	}

	if err := u.subs.Save(ctx, sub); err != nil {
		return domainSubscription.Subscription{}, fmt.Errorf("save subscription: %w", err)
	}
	u.writeAudit(ctx, "CREATE_SUBSCRIPTION", "subscription", sub.ID)
	return sub, nil
}

// Update menerapkan hanya field pointer non-nil pada langganan existing.
func (u *ManageSubscriptionUseCase) Update(ctx context.Context, subID string, in UpdateInput) (domainSubscription.Subscription, error) {
	sub, err := u.subs.FindByID(ctx, subID)
	if err != nil {
		return domainSubscription.Subscription{}, domainBilling.ErrNotFound
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
	hasInvoice, err := u.invoices.HasForSubscription(ctx, subID)
	if err != nil {
		return fmt.Errorf("check subscription invoices %s: %w", subID, err)
	}
	if hasInvoice {
		return fmt.Errorf("%w: subscription still has invoices", domainBilling.ErrInvalidInput)
	}

	sub, err := u.subs.FindByID(ctx, subID)
	if err != nil {
		return domainBilling.ErrNotFound
	}
	if provisioned(sub) {
		if terr := u.manager.Terminate(ctx, derefDevice(sub.DeviceID), sub.ServiceType, sub.RemoteUsername); terr != nil {
			// Best-effort: baris DB tetap dihapus, akun yatim ditangani
			// worker sinkronisasi router.
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
