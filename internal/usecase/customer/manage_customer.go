package customer

import (
	"context"
	"fmt"

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
		return customer.Customer{}, fmt.Errorf("customer id required")
	}
	return uc.repo.FindByID(ctx, id)
}

func (uc *ManageCustomerUseCase) CreateCustomer(ctx context.Context, c customer.Customer) error {
	if c.ID == "" || c.Name == "" {
		return fmt.Errorf("customer id and name are required")
	}
	return uc.repo.Save(ctx, c)
}

func (uc *ManageCustomerUseCase) UpdateCustomer(ctx context.Context, c customer.Customer) error {
	if c.ID == "" {
		return fmt.Errorf("customer id required")
	}
	return uc.repo.Save(ctx, c)
}

func (uc *ManageCustomerUseCase) DeleteCustomer(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("customer id required")
	}
	return uc.repo.Delete(ctx, id)
}

func (uc *ManageCustomerUseCase) ListSubscriptions(ctx context.Context, customerID string) ([]subscription.Subscription, error) {
	return uc.repo.FindSubscriptions(ctx, customerID)
}

