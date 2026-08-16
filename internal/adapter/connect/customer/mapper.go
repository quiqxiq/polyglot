package customer

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/customer"
)

// toProtoCustomer maps a domain Customer entity to its Protobuf wire representation.
func toProtoCustomer(c *customer.Customer) *devicepb.Customer {
	if c == nil {
		return nil
	}
	return &devicepb.Customer{
		Id:            c.ID,
		TenantId:      c.TenantID,
		Name:          c.Name,
		Email:         c.Email,
		Phone:         c.Phone,
		Address:       c.Address,
		Status:        c.Status,
		CreatedAtUnix: c.CreatedAt.Unix(),
	}
}

// toProtoCustomerList maps a slice of domain Customer entities to a slice of Protobuf Customers.
func toProtoCustomerList(customers []customer.Customer) []*devicepb.Customer {
	pbCustomers := make([]*devicepb.Customer, len(customers))
	for i := range customers {
		pbCustomers[i] = toProtoCustomer(&customers[i])
	}
	return pbCustomers
}
