package customer

import (
	"context"
	"fmt"
	"time"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// Detail menggabungkan domain Customer dengan agregasi counter aktif
// (Zero Waterfall untuk tampilan tabel & CRM).
type Detail struct {
	Customer                 customer.Customer
	ActiveSubscriptionsCount int
	UnpaidInvoicesCount      int
}

// ManageCustomerUseCase orchestrates customer CRUD and subscription management.
type ManageCustomerUseCase struct {
	repo    port.CustomerRepository
	subRepo port.SubscriptionRepository
	invRepo port.InvoiceRepository
	router  port.RouterAccountManager
}

// NewManageCustomerUseCase constructs a new ManageCustomerUseCase.
func NewManageCustomerUseCase(
	repo port.CustomerRepository,
	subRepo port.SubscriptionRepository,
	invRepo port.InvoiceRepository,
	router port.RouterAccountManager,
) *ManageCustomerUseCase {
	return &ManageCustomerUseCase{
		repo:    repo,
		subRepo: subRepo,
		invRepo: invRepo,
		router:  router,
	}
}

// Enrich menghitung activeSubscriptionsCount dan unpaidInvoicesCount untuk 1 customer.
func (uc *ManageCustomerUseCase) Enrich(ctx context.Context, c customer.Customer) Detail {
	var activeSubs int
	if uc.subRepo != nil {
		if subs, err := uc.subRepo.FindByCustomerID(ctx, c.ID); err == nil {
			for _, sub := range subs {
				if sub.Status != subscription.StatusTerminated && sub.Status != subscription.StatusCancelled {
					activeSubs++
				}
			}
		}
	}
	var unpaidInvs int
	if uc.invRepo != nil {
		if invoices, err := uc.invRepo.FindByCustomerID(ctx, c.ID); err == nil {
			for _, inv := range invoices {
				if inv.Status != domainBilling.StatusPaid {
					unpaidInvs++
				}
			}
		}
	}
	return Detail{
		Customer:                 c,
		ActiveSubscriptionsCount: activeSubs,
		UnpaidInvoicesCount:      unpaidInvs,
	}
}

// ListCustomers mengembalikan daftar pelanggan beserta counter langganan dan tagihan aktif.
func (uc *ManageCustomerUseCase) ListCustomers(ctx context.Context) ([]Detail, error) {
	customers, err := uc.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list customers: %w", err)
	}

	subCountByCust := make(map[string]int)
	if uc.subRepo != nil {
		if subs, err := uc.subRepo.FindAll(ctx); err == nil {
			for _, sub := range subs {
				if sub.Status != subscription.StatusTerminated && sub.Status != subscription.StatusCancelled {
					subCountByCust[sub.CustomerID]++
				}
			}
		}
	}

	unpaidCountByCust := make(map[string]int)
	if uc.invRepo != nil {
		if invoices, err := uc.invRepo.FindAll(ctx); err == nil {
			for _, inv := range invoices {
				if inv.Status != domainBilling.StatusPaid {
					unpaidCountByCust[inv.CustomerID]++
				}
			}
		}
	}

	details := make([]Detail, len(customers))
	for i, c := range customers {
		details[i] = Detail{
			Customer:                 c,
			ActiveSubscriptionsCount: subCountByCust[c.ID],
			UnpaidInvoicesCount:      unpaidCountByCust[c.ID],
		}
	}
	return details, nil
}

// GetCustomer mengambil detail pelanggan beserta counternya.
func (uc *ManageCustomerUseCase) GetCustomer(ctx context.Context, id string) (Detail, error) {
	if id == "" {
		return Detail{}, customer.ErrInvalidInput
	}
	result, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return Detail{}, fmt.Errorf("find customer %s: %w", id, err)
	}
	return uc.Enrich(ctx, result), nil
}

// CreateCustomer creates a new customer entity with generated codes and defaults.
func (uc *ManageCustomerUseCase) CreateCustomer(ctx context.Context, c customer.Customer) (customer.Customer, error) {
	if c.Name == "" {
		return customer.Customer{}, customer.ErrInvalidInput
	}
	if c.ID == "" {
		c.ID = idgen.New("cust")
	}
	if c.TenantID == "" {
		c.TenantID = "tenant-default"
	}
	if c.Status == "" {
		c.Status = customer.StatusActive
	}
	if c.CustomerCode == "" {
		for attempt := 0; attempt < 8; attempt++ {
			code := "CUST-" + idgen.Digits(5)
			if _, err := uc.repo.FindByCustomerCode(ctx, code); err != nil {
				c.CustomerCode = code
				break
			}
		}
	}
	if c.PortalAccessCode == "" {
		for attempt := 0; attempt < 8; attempt++ {
			portal := idgen.Digits(8)
			if _, err := uc.repo.FindByPortalAccessCode(ctx, portal); err != nil {
				c.PortalAccessCode = portal
				break
			}
		}
	}
	now := time.Now().UTC()
	if c.RegisteredAt.IsZero() {
		c.RegisteredAt = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	}
	c.CreatedAt = now
	c.UpdatedAt = now

	if err := uc.repo.Save(ctx, c); err != nil {
		return customer.Customer{}, fmt.Errorf("save customer: %w", err)
	}
	return c, nil
}

// UpdateCustomer updates an existing customer entity.
func (uc *ManageCustomerUseCase) UpdateCustomer(ctx context.Context, c customer.Customer) (customer.Customer, error) {
	if c.ID == "" || c.Name == "" {
		return customer.Customer{}, customer.ErrInvalidInput
	}
	existing, err := uc.repo.FindByID(ctx, c.ID)
	if err != nil {
		return customer.Customer{}, fmt.Errorf("find customer %s: %w", c.ID, err)
	}
	if c.TenantID == "" {
		c.TenantID = existing.TenantID
	}
	if c.CustomerCode == "" {
		c.CustomerCode = existing.CustomerCode
	}
	if c.PortalAccessCode == "" {
		c.PortalAccessCode = existing.PortalAccessCode
	}
	if c.Status == "" {
		c.Status = existing.Status
	}
	if c.RegisteredAt.IsZero() {
		c.RegisteredAt = existing.RegisteredAt
	}
	c.CreatedAt = existing.CreatedAt
	c.UpdatedAt = time.Now().UTC()

	if err := uc.repo.Save(ctx, c); err != nil {
		return customer.Customer{}, fmt.Errorf("update customer: %w", err)
	}
	return c, nil
}

// DeleteCustomer removes a customer and cascade-deletes their subscriptions (including router accounts) and invoices.
func (uc *ManageCustomerUseCase) DeleteCustomer(ctx context.Context, id string) error {
	if id == "" {
		return customer.ErrInvalidInput
	}

	// 1. Terminate router accounts & delete associated subscriptions
	if uc.subRepo != nil {
		subs, err := uc.subRepo.FindByCustomerID(ctx, id)
		if err == nil {
			for _, sub := range subs {
				if uc.router != nil && sub.DeviceID != nil && *sub.DeviceID != "" && sub.RemoteUsername != "" {
					_ = uc.router.Terminate(ctx, *sub.DeviceID, sub.ServiceType, sub.RemoteUsername)
				}
				_ = uc.subRepo.Delete(ctx, sub.ID)
			}
		}
	}

	// 2. Delete all invoices belonging to the customer
	if uc.invRepo != nil {
		if err := uc.invRepo.DeleteByCustomerID(ctx, id); err != nil {
			return fmt.Errorf("delete customer invoices: %w", err)
		}
	}

	// 3. Delete the customer record
	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete customer %s: %w", id, err)
	}
	return nil
}

func (uc *ManageCustomerUseCase) FindSubscriptions(ctx context.Context, customerID string) ([]subscription.Subscription, error) {
	return uc.repo.FindSubscriptions(ctx, customerID)
}

func (uc *ManageCustomerUseCase) FindByPhone(ctx context.Context, phone string) (customer.Customer, error) {
	if phone == "" {
		return customer.Customer{}, customer.ErrInvalidInput
	}
	return uc.repo.FindByPhone(ctx, phone)
}

func (uc *ManageCustomerUseCase) FindByCustomerCode(ctx context.Context, code string) (customer.Customer, error) {
	if code == "" {
		return customer.Customer{}, customer.ErrInvalidInput
	}
	return uc.repo.FindByCustomerCode(ctx, code)
}

func (uc *ManageCustomerUseCase) FindByPortalCode(ctx context.Context, portalCode string) (customer.Customer, error) {
	if portalCode == "" {
		return customer.Customer{}, customer.ErrInvalidInput
	}
	return uc.repo.FindByPortalAccessCode(ctx, portalCode)
}
