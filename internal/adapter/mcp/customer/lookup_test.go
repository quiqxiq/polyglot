package customer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

type memCustomerRepo struct {
	customers []domainCustomer.Customer
	subs      map[string][]subscription.Subscription
}

func (m *memCustomerRepo) Save(_ context.Context, c domainCustomer.Customer) error {
	m.customers = append(m.customers, c)
	return nil
}

func (m *memCustomerRepo) FindByID(_ context.Context, id string) (domainCustomer.Customer, error) {
	for _, c := range m.customers {
		if c.ID == id {
			return c, nil
		}
	}
	return domainCustomer.Customer{}, domainCustomer.ErrCustomerNotFound
}

func (m *memCustomerRepo) FindAll(_ context.Context) ([]domainCustomer.Customer, error) {
	return m.customers, nil
}

func (m *memCustomerRepo) FindPaged(_ context.Context, _ port.PageFilter) ([]domainCustomer.Customer, error) {
	return m.customers, nil
}

func (m *memCustomerRepo) Delete(_ context.Context, _ string) error {
	return nil
}

func (m *memCustomerRepo) FindSubscriptions(_ context.Context, customerID string) ([]subscription.Subscription, error) {
	return m.subs[customerID], nil
}

func (m *memCustomerRepo) FindByPortalAccessCode(_ context.Context, _ string) (domainCustomer.Customer, error) {
	return domainCustomer.Customer{}, domainCustomer.ErrCustomerNotFound
}

func (m *memCustomerRepo) FindByPhone(_ context.Context, _ string) (domainCustomer.Customer, error) {
	return domainCustomer.Customer{}, domainCustomer.ErrCustomerNotFound
}

func (m *memCustomerRepo) FindByCustomerCode(_ context.Context, _ string) (domainCustomer.Customer, error) {
	return domainCustomer.Customer{}, domainCustomer.ErrCustomerNotFound
}

func TestCustomerLookup_NoCriteria(t *testing.T) {
	h := NewHandler(nil)
	res, out, err := h.CustomerLookup(context.Background(), nil, customerLookupArgs{})
	require.NoError(t, err)
	assert.True(t, res.IsError)
	assert.Equal(t, "error", out.Status)
	assert.Contains(t, out.Summary, "at least one search criterion")
}

func TestCustomerLookup_NoRepoConfigured(t *testing.T) {
	h := NewHandler(nil)
	res, out, err := h.CustomerLookup(context.Background(), nil, customerLookupArgs{Phone: "081234"})
	require.NoError(t, err)
	assert.False(t, res.IsError)
	assert.Equal(t, "success", out.Status)
	assert.Contains(t, out.Summary, "not configured")
}

func TestCustomerLookup_Success(t *testing.T) {
	repo := &memCustomerRepo{
		customers: []domainCustomer.Customer{
			{ID: "cust-1", Name: "John Doe", Phone: "0812345678", Email: "john@example.com", Status: "active"},
			{ID: "cust-2", Name: "Jane Smith", Phone: "0898765432", Email: "jane@example.com", Status: "active"},
		},
		subs: map[string][]subscription.Subscription{
			"cust-1": {{ID: "sub-1", PlanID: "plan-50mbps"}},
		},
	}
	h := NewHandler(func() port.CustomerRepository { return repo })

	// By Phone
	res, out, err := h.CustomerLookup(context.Background(), nil, customerLookupArgs{Phone: "1234"})
	require.NoError(t, err)
	assert.False(t, res.IsError)
	assert.Len(t, out.Customers, 1)
	assert.Equal(t, "John Doe", out.Customers[0].Name)
	assert.Len(t, out.Customers[0].Subscriptions, 1)

	// By CustomerID
	res, out, err = h.CustomerLookup(context.Background(), nil, customerLookupArgs{CustomerID: "cust-2"})
	require.NoError(t, err)
	assert.False(t, res.IsError)
	assert.Len(t, out.Customers, 1)
	assert.Equal(t, "Jane Smith", out.Customers[0].Name)

	// Not Found
	res, out, err = h.CustomerLookup(context.Background(), nil, customerLookupArgs{Name: "NonExistent"})
	require.NoError(t, err)
	assert.False(t, res.IsError)
	assert.Empty(t, out.Customers)
	assert.Contains(t, out.Summary, "No customers matched")
}
