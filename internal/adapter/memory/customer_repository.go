package memory

import (
	"context"
	"fmt"
	"sync"

	customerDomain "github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/port"
)

// MemCustomerRepository is an in-memory implementation of port.CustomerRepository.
type MemCustomerRepository struct {
	mu        sync.RWMutex
	customers map[string]customerDomain.Customer
}

var _ port.CustomerRepository = (*MemCustomerRepository)(nil)

// NewCustomerRepository constructs a new MemCustomerRepository.
func NewCustomerRepository() *MemCustomerRepository {
	return &MemCustomerRepository{
		customers: make(map[string]customerDomain.Customer),
	}
}

func (r *MemCustomerRepository) Save(_ context.Context, c customerDomain.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.customers[c.ID] = c
	return nil
}

func (r *MemCustomerRepository) FindByID(_ context.Context, id string) (customerDomain.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.customers[id]
	if !ok {
		return customerDomain.Customer{}, fmt.Errorf("customer not found")
	}
	return c, nil
}

func (r *MemCustomerRepository) FindAll(_ context.Context) ([]customerDomain.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]customerDomain.Customer, 0, len(r.customers))
	for _, c := range r.customers {
		list = append(list, c)
	}
	return list, nil
}

func (r *MemCustomerRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.customers, id)
	return nil
}

func (r *MemCustomerRepository) FindSubscriptions(_ context.Context, _ string) ([]customerDomain.Subscription, error) {
	return []customerDomain.Subscription{}, nil
}
