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
		CustomerCode:  c.CustomerCode,
		Latitude:      c.Latitude,
		Longitude:     c.Longitude,
		Notes:         c.Notes,
	}
}

// fromProtoCustomer maps the wire representation into a domain entity.
func fromProtoCustomer(pb *devicepb.Customer) customer.Customer {
	if pb == nil {
		return customer.Customer{}
	}
	return customer.Customer{
		ID:           pb.Id,
		TenantID:     pb.TenantId,
		CustomerCode: pb.CustomerCode,
		Name:         pb.Name,
		Email:        pb.Email,
		Phone:        pb.Phone,
		Address:      pb.Address,
		Latitude:     pb.Latitude,
		Longitude:    pb.Longitude,
		Status:       pb.Status,
		Notes:        pb.Notes,
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
