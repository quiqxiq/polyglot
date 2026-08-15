package mapper

import (
	devicepb "github.com/quixiq/polyglot/api/proto/v1"
	"github.com/quixiq/polyglot/internal/domain/customer"
)

// CustomerToProto maps a domain Customer entity to its Protobuf message representation.
func CustomerToProto(c customer.Customer) *devicepb.Customer {
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

// CustomerListToProto maps a slice of domain Customer entities to Protobuf messages.
func CustomerListToProto(customers []customer.Customer) []*devicepb.Customer {
	res := make([]*devicepb.Customer, len(customers))
	for i, c := range customers {
		res[i] = CustomerToProto(c)
	}
	return res
}

// ProtoToCustomer maps a Protobuf Customer message to a domain Customer entity.
func ProtoToCustomer(pb *devicepb.Customer) customer.Customer {
	if pb == nil {
		return customer.Customer{}
	}
	return customer.Customer{
		ID:       pb.Id,
		TenantID: pb.TenantId,
		Name:     pb.Name,
		Email:    pb.Email,
		Phone:    pb.Phone,
		Address:  pb.Address,
		Status:   pb.Status,
	}
}
