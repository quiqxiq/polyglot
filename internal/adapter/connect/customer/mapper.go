package customer

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	customerUC "github.com/quixiq/polyglot/internal/usecase/customer"
)

func toProtoCustomer(c *domainCustomer.Customer) *devicepb.Customer {
	if c == nil {
		return nil
	}
	var lat, lon float64
	hasCoord := false
	if c.Latitude != nil {
		lat = *c.Latitude
		hasCoord = true
	}
	if c.Longitude != nil {
		lon = *c.Longitude
	}
	var registeredAtUnix int64
	if !c.RegisteredAt.IsZero() {
		registeredAtUnix = c.RegisteredAt.Unix()
	}
	return &devicepb.Customer{
		Id:               c.ID,
		TenantId:         c.TenantID,
		CustomerCode:     c.CustomerCode,
		Name:             c.Name,
		Phone:            c.Phone,
		Email:            c.Email,
		Address:          c.Address,
		Latitude:         lat,
		Longitude:        lon,
		HasCoordinates:   hasCoord,
		PortalAccessCode: c.PortalAccessCode,
		Status:           c.Status,
		Notes:            c.Notes,
		RegisteredAtUnix: registeredAtUnix,
		CreatedAtUnix:    c.CreatedAt.Unix(),
	}
}

func toProtoCustomerDetail(cd *customerUC.Detail) *devicepb.Customer {
	if cd == nil {
		return nil
	}
	pb := toProtoCustomer(&cd.Customer)
	if pb == nil {
		return nil
	}
	pb.ActiveSubscriptionsCount = int32(cd.ActiveSubscriptionsCount)
	pb.UnpaidInvoicesCount = int32(cd.UnpaidInvoicesCount)
	return pb
}

func toProtoCustomerDetailList(details []customerUC.Detail) []*devicepb.Customer {
	out := make([]*devicepb.Customer, len(details))
	for i := range details {
		out[i] = toProtoCustomerDetail(&details[i])
	}
	return out
}

func toProtoCustomerList(customers []domainCustomer.Customer) []*devicepb.Customer {
	out := make([]*devicepb.Customer, len(customers))
	for i := range customers {
		out[i] = toProtoCustomer(&customers[i])
	}
	return out
}

func fromProtoCustomer(pb *devicepb.Customer) domainCustomer.Customer {
	if pb == nil {
		return domainCustomer.Customer{}
	}
	var lat, lon *float64
	if pb.HasCoordinates || pb.Latitude != 0 || pb.Longitude != 0 {
		vLat, vLon := pb.Latitude, pb.Longitude
		lat, lon = &vLat, &vLon
	}
	return domainCustomer.Customer{
		ID:               pb.Id,
		TenantID:         pb.TenantId,
		CustomerCode:     pb.CustomerCode,
		Name:             pb.Name,
		Phone:            pb.Phone,
		Email:            pb.Email,
		Address:          pb.Address,
		Latitude:         lat,
		Longitude:        lon,
		PortalAccessCode: pb.PortalAccessCode,
		Status:           pb.Status,
		Notes:            pb.Notes,
	}
}
