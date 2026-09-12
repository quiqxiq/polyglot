package mocktest

import (
	"context"
	"sort"
	"sync"

	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

// FakeCustomerRepo is an in-memory customer repository for tests.
type FakeCustomerRepo struct {
	mu   sync.Mutex
	byID map[string]domainCustomer.Customer
}

// NewFakeCustomerRepo constructs an empty fake customer repository.
func NewFakeCustomerRepo() *FakeCustomerRepo {
	return &FakeCustomerRepo{byID: map[string]domainCustomer.Customer{}}
}

// Save stores a customer in memory.
func (f *FakeCustomerRepo) Save(_ context.Context, c domainCustomer.Customer) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[c.ID] = c
	return nil
}

// FindByID returns a customer by ID.
func (f *FakeCustomerRepo) FindByID(_ context.Context, id string) (domainCustomer.Customer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	c, ok := f.byID[id]
	if !ok {
		return domainCustomer.Customer{}, ErrFakeNotFound
	}
	return c, nil
}

// FindAll returns all in-memory customers.
func (f *FakeCustomerRepo) FindAll(_ context.Context) ([]domainCustomer.Customer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domainCustomer.Customer, 0, len(f.byID))
	for _, c := range f.byID {
		out = append(out, c)
	}
	return out, nil
}

// Delete removes a customer by ID.
func (f *FakeCustomerRepo) Delete(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.byID, id)
	return nil
}

// FindPaged implements tenant + limit/offset filtering (F6-8).
func (f *FakeCustomerRepo) FindPaged(_ context.Context, filter port.PageFilter) ([]domainCustomer.Customer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domainCustomer.Customer, 0, len(f.byID))
	for _, c := range f.byID {
		if filter.TenantID != "" && c.TenantID != filter.TenantID {
			continue
		}
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	if filter.Offset > 0 {
		if filter.Offset >= len(out) {
			return []domainCustomer.Customer{}, nil
		}
		out = out[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(out) {
		out = out[:filter.Limit]
	}
	return out, nil
}

// FindSubscriptions returns subscriptions for a customer.
func (f *FakeCustomerRepo) FindSubscriptions(_ context.Context, customerID string) ([]domainSubscription.Subscription, error) {
	return nil, nil
}

// FindByPortalAccessCode finds a customer by portal code.
func (f *FakeCustomerRepo) FindByPortalAccessCode(_ context.Context, code string) (domainCustomer.Customer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.byID {
		if c.PortalAccessCode == code {
			return c, nil
		}
	}
	return domainCustomer.Customer{}, ErrFakeNotFound
}

// FindByPhone finds a customer by phone number.
func (f *FakeCustomerRepo) FindByPhone(_ context.Context, phone string) (domainCustomer.Customer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.byID {
		if c.Phone == phone {
			return c, nil
		}
	}
	return domainCustomer.Customer{}, ErrFakeNotFound
}

// FindByCustomerCode finds a customer by customer code.
func (f *FakeCustomerRepo) FindByCustomerCode(_ context.Context, code string) (domainCustomer.Customer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.byID {
		if c.CustomerCode == code {
			return c, nil
		}
	}
	return domainCustomer.Customer{}, ErrFakeNotFound
}
