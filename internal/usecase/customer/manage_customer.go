package customer

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
)

// customerRepository narrows the persistence surface for customer CRUD.
type customerRepository interface {
	Save(ctx context.Context, c domainCustomer.Customer) error
	FindByID(ctx context.Context, id string) (domainCustomer.Customer, error)
	FindAll(ctx context.Context) ([]domainCustomer.Customer, error)
	SoftDelete(ctx context.Context, id string, at time.Time) error
	FindByPhone(ctx context.Context, phone string) (domainCustomer.Customer, error)
	NextCustomerCode(ctx context.Context) (string, error)
	FindSubscriptions(ctx context.Context, customerID string) ([]domainSub.Subscription, error)
}

// ManageCustomerUseCase orchestrates full customer CRUD. Network accounts
// are NOT managed here — they live behind subscription mapping.
type ManageCustomerUseCase struct {
	repo customerRepository
}

func NewManageCustomerUseCase(repo customerRepository) *ManageCustomerUseCase {
	return &ManageCustomerUseCase{repo: repo}
}

func (uc *ManageCustomerUseCase) ListCustomers(ctx context.Context) ([]domainCustomer.Customer, error) {
	return uc.repo.FindAll(ctx)
}

func (uc *ManageCustomerUseCase) GetCustomer(ctx context.Context, id string) (domainCustomer.Customer, error) {
	if id == "" {
		return domainCustomer.Customer{}, domainCustomer.ErrNotFound
	}
	return uc.repo.FindByID(ctx, id)
}

// CreateCustomer validates, generates ID/code when absent, and persists.
func (uc *ManageCustomerUseCase) CreateCustomer(ctx context.Context, c domainCustomer.Customer) (domainCustomer.Customer, error) {
	if err := c.Validate(); err != nil {
		return domainCustomer.Customer{}, err
	}
	if existing, err := uc.repo.FindByPhone(ctx, c.Phone); err == nil && existing.ID != "" {
		return domainCustomer.Customer{}, fmt.Errorf("%w: %s", domainCustomer.ErrAlreadyExists, "phone already registered")
	}
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	if c.TenantID == "" {
		c.TenantID = "tenant-default"
	}
	if c.CustomerCode == "" {
		code, err := uc.repo.NextCustomerCode(ctx)
		if err != nil {
			return domainCustomer.Customer{}, fmt.Errorf("generate customer code: %w", err)
		}
		c.CustomerCode = code
	}
	if c.Status == "" {
		c.Status = "ACTIVE"
	}
	now := time.Now()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	if err := uc.repo.Save(ctx, c); err != nil {
		return domainCustomer.Customer{}, fmt.Errorf("save customer: %w", err)
	}
	return c, nil
}

func (uc *ManageCustomerUseCase) UpdateCustomer(ctx context.Context, c domainCustomer.Customer) (domainCustomer.Customer, error) {
	if c.ID == "" {
		return domainCustomer.Customer{}, domainCustomer.ErrNotFound
	}
	current, err := uc.repo.FindByID(ctx, c.ID)
	if err != nil {
		return domainCustomer.Customer{}, err
	}
	if err := c.Validate(); err != nil {
		return domainCustomer.Customer{}, err
	}
	// Immutable fields stay under system control.
	c.CreatedAt = current.CreatedAt
	c.TenantID = current.TenantID
	if c.Status == "" {
		c.Status = current.Status
	}
	if err := uc.repo.Save(ctx, c); err != nil {
		return domainCustomer.Customer{}, fmt.Errorf("update customer: %w", err)
	}
	return c, nil
}

// DeleteCustomer soft-deletes so invoices/subscriptions keep referential
// integrity; hard delete remains a maintenance operation on the repo.
func (uc *ManageCustomerUseCase) DeleteCustomer(ctx context.Context, id string) error {
	if id == "" {
		return domainCustomer.ErrNotFound
	}
	subs, err := uc.repo.FindSubscriptions(ctx, id)
	if err != nil {
		return fmt.Errorf("check subscriptions: %w", err)
	}
	for _, s := range subs {
		if s.Status != domainSub.StatusTerminated {
			return fmt.Errorf("customer still has %d active subscription(s); terminate them first", len(subs))
		}
	}
	if err := uc.repo.SoftDelete(ctx, id, time.Now()); err != nil {
		return fmt.Errorf("delete customer: %w", err)
	}
	return nil
}

func (uc *ManageCustomerUseCase) ListSubscriptions(ctx context.Context, customerID string) ([]domainSub.Subscription, error) {
	return uc.repo.FindSubscriptions(ctx, customerID)
}
