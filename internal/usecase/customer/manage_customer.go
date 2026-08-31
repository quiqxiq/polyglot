package customer

import (
	"context"
	"fmt"
	"time"

	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// ManageCustomerUseCase orchestrates customer CRUD and subscription management.
type ManageCustomerUseCase struct {
	repo port.CustomerRepository
}

// NewManageCustomerUseCase constructs a new ManageCustomerUseCase.
func NewManageCustomerUseCase(repo port.CustomerRepository) *ManageCustomerUseCase {
	return &ManageCustomerUseCase{repo: repo}
}

func (uc *ManageCustomerUseCase) ListCustomers(ctx context.Context) ([]customer.Customer, error) {
	customers, err := uc.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list customers: %w", err)
	}
	return customers, nil
}

func (uc *ManageCustomerUseCase) GetCustomer(ctx context.Context, id string) (customer.Customer, error) {
	if id == "" {
		return customer.Customer{}, customer.ErrInvalidInput
	}
	result, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return customer.Customer{}, fmt.Errorf("find customer %s: %w", id, err)
	}
	return result, nil
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

func (uc *ManageCustomerUseCase) DeleteCustomer(ctx context.Context, id string) error {
	if id == "" {
		return customer.ErrInvalidInput
	}
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
