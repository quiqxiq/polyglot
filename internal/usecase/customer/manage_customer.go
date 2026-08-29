package customer

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
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
	return uc.repo.FindAll(ctx)
}

func (uc *ManageCustomerUseCase) GetCustomer(ctx context.Context, id string) (customer.Customer, error) {
	if id == "" {
		return customer.Customer{}, customer.ErrInvalidInput
	}
	return uc.repo.FindByID(ctx, id)
}

func (uc *ManageCustomerUseCase) CreateCustomer(ctx context.Context, c customer.Customer) error {
	if c.ID == "" || c.Name == "" {
		return customer.ErrInvalidInput
	}
	return uc.repo.Save(ctx, c)
}

func (uc *ManageCustomerUseCase) UpdateCustomer(ctx context.Context, c customer.Customer) error {
	if c.ID == "" {
		return customer.ErrInvalidInput
	}
	return uc.repo.Save(ctx, c)
}

func (uc *ManageCustomerUseCase) DeleteCustomer(ctx context.Context, id string) error {
	if id == "" {
		return customer.ErrInvalidInput
	}
	return uc.repo.Delete(ctx, id)
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
