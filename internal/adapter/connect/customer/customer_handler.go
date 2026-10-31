package customer

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	customerUC "github.com/quixiq/polyglot/internal/usecase/customer"
	"github.com/quixiq/polyglot/pkg/response"
)

type CustomerConnectHandler struct {
	customerUC *customerUC.ManageCustomerUseCase
}

func NewCustomerConnectHandler(uc *customerUC.ManageCustomerUseCase) *CustomerConnectHandler {
	return &CustomerConnectHandler{customerUC: uc}
}

func (h *CustomerConnectHandler) toEnrichedProto(ctx context.Context, c *domainCustomer.Customer) *devicepb.Customer {
	if c == nil {
		return nil
	}
	if h.customerUC != nil {
		detail := h.customerUC.Enrich(ctx, *c)
		return toProtoCustomerDetail(&detail)
	}
	return toProtoCustomer(c)
}

func (h *CustomerConnectHandler) ListCustomers(ctx context.Context, req *connect.Request[devicepb.ListCustomersRequest]) (*connect.Response[devicepb.ListCustomersResponse], error) {
	if h.customerUC == nil {
		return nil, response.Unavailable("customer usecase unavailable")
	}
	customers, err := h.customerUC.ListCustomers(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListCustomersResponse{
		Customers: toProtoCustomerDetailList(customers),
	}), nil
}

func (h *CustomerConnectHandler) GetCustomer(ctx context.Context, req *connect.Request[devicepb.GetCustomerRequest]) (*connect.Response[devicepb.GetCustomerResponse], error) {
	if h.customerUC == nil {
		return nil, response.Unavailable("customer usecase unavailable")
	}
	detail, err := h.customerUC.GetCustomer(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetCustomerResponse{
		Customer: toProtoCustomerDetail(&detail),
	}), nil
}

func (h *CustomerConnectHandler) CreateCustomer(ctx context.Context, req *connect.Request[devicepb.CreateCustomerRequest]) (*connect.Response[devicepb.CreateCustomerResponse], error) {
	if h.customerUC == nil {
		return nil, response.Unavailable("customer usecase unavailable")
	}
	c := fromProtoCustomer(req.Msg.Customer)
	created, err := h.customerUC.CreateCustomer(ctx, c)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.CreateCustomerResponse{
		Customer: h.toEnrichedProto(ctx, &created),
		Message:  "Customer created successfully",
	}), nil
}

func (h *CustomerConnectHandler) UpdateCustomer(ctx context.Context, req *connect.Request[devicepb.UpdateCustomerRequest]) (*connect.Response[devicepb.UpdateCustomerResponse], error) {
	if h.customerUC == nil {
		return nil, response.Unavailable("customer usecase unavailable")
	}
	c := fromProtoCustomer(req.Msg.Customer)
	updated, err := h.customerUC.UpdateCustomer(ctx, c)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.UpdateCustomerResponse{
		Customer: h.toEnrichedProto(ctx, &updated),
		Message:  "Customer updated successfully",
	}), nil
}

func (h *CustomerConnectHandler) DeleteCustomer(ctx context.Context, req *connect.Request[devicepb.DeleteCustomerRequest]) (*connect.Response[devicepb.DeleteCustomerResponse], error) {
	if h.customerUC == nil {
		return nil, response.Unavailable("customer usecase unavailable")
	}
	if err := h.customerUC.DeleteCustomer(ctx, req.Msg.Id); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.DeleteCustomerResponse{
		Message: "Customer deleted successfully",
	}), nil
}

func (h *CustomerConnectHandler) FindByPhone(ctx context.Context, req *connect.Request[devicepb.FindCustomerByPhoneRequest]) (*connect.Response[devicepb.FindCustomerResponse], error) {
	c, err := h.customerUC.FindByPhone(ctx, req.Msg.Phone)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.FindCustomerResponse{Customer: h.toEnrichedProto(ctx, &c)}), nil
}

func (h *CustomerConnectHandler) FindByCustomerCode(ctx context.Context, req *connect.Request[devicepb.FindCustomerByCodeRequest]) (*connect.Response[devicepb.FindCustomerResponse], error) {
	c, err := h.customerUC.FindByCustomerCode(ctx, req.Msg.CustomerCode)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.FindCustomerResponse{Customer: h.toEnrichedProto(ctx, &c)}), nil
}

func (h *CustomerConnectHandler) FindByPortalCode(ctx context.Context, req *connect.Request[devicepb.FindCustomerByPortalCodeRequest]) (*connect.Response[devicepb.FindCustomerResponse], error) {
	c, err := h.customerUC.FindByPortalCode(ctx, req.Msg.PortalAccessCode)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.FindCustomerResponse{Customer: h.toEnrichedProto(ctx, &c)}), nil
}

func (h *CustomerConnectHandler) ListSubscriptions(ctx context.Context, req *connect.Request[devicepb.ListSubscriptionsRequest]) (*connect.Response[devicepb.ListSubscriptionsResponse], error) {
	return connect.NewResponse(&devicepb.ListSubscriptionsResponse{}), nil
}

// GetCustomer360 mengembalikan data pelanggan terpadu (profil, langganan,
// tagihan, pembayaran, dan riwayat PSB) — planning v2 Fase 5.
func (h *CustomerConnectHandler) GetCustomer360(ctx context.Context, req *connect.Request[devicepb.GetCustomer360Request]) (*connect.Response[devicepb.GetCustomer360Response], error) {
	if h.customerUC == nil {
		return nil, response.Unavailable("customer usecase unavailable")
	}
	data, err := h.customerUC.GetCustomer360(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	resp := &devicepb.GetCustomer360Response{
		Customer: toProtoCustomer(&data.Customer),
	}
	for i := range data.Subscriptions {
		sub := data.Subscriptions[i]
		deviceID := ""
		if sub.DeviceID != nil {
			deviceID = *sub.DeviceID
		}
		nextDue := ""
		if sub.NextDueDate != nil {
			nextDue = sub.NextDueDate.Format("2006-01-02")
		}
		resp.Subscriptions = append(resp.Subscriptions, &devicepb.Customer360Subscription{
			Id:                 sub.ID,
			PlanId:             sub.PlanID,
			ServiceType:        sub.ServiceType,
			Status:             sub.Status,
			ProvisionStatus:    sub.ProvisionStatus,
			LastProvisionError: sub.LastProvisionError,
			ProvisionAttempts:  int32(sub.ProvisionAttempts),
			RouterProfile:      sub.RouterProfile,
			DeviceId:           deviceID,
			StartDateUnix:      sub.StartDate.Unix(),
			NextDueDate:        nextDue,
		})
	}
	for i := range data.Invoices {
		inv := data.Invoices[i]
		resp.Invoices = append(resp.Invoices, &devicepb.Customer360Invoice{
			Id:            inv.ID,
			InvoiceNumber: inv.InvoiceNumber,
			Period:        inv.Period,
			Total:         inv.Total,
			PaidAmount:    inv.PaidAmount,
			Status:        inv.Status,
			DueDateUnix:   inv.DueDate.Unix(),
		})
	}
	for i := range data.Payments {
		pay := data.Payments[i]
		resp.Payments = append(resp.Payments, &devicepb.Customer360Payment{
			Id:              pay.ID,
			PaymentNo:       pay.PaymentNo,
			InvoiceId:       pay.InvoiceID,
			Amount:          pay.Amount,
			PaymentDateUnix: pay.PaymentDate.Unix(),
			ScanMethod:      pay.ScanMethod,
			Voided:          pay.VoidedAt != nil,
			VoidReason:      pay.VoidReason,
		})
	}
	for i := range data.Registrations {
		reg := data.Registrations[i]
		var installedAt int64
		if reg.InstalledAt != nil {
			installedAt = reg.InstalledAt.Unix()
		}
		resp.Registrations = append(resp.Registrations, &devicepb.Customer360Registration{
			Id:              reg.ID,
			RegistrationNo:  reg.RegistrationNo,
			Status:          reg.Status,
			CreatedAtUnix:   reg.CreatedAt.Unix(),
			InstalledAtUnix: installedAt,
			SurveyStatus:    reg.SurveyStatus,
			PublicToken:     reg.PublicToken,
		})
	}
	return connect.NewResponse(resp), nil
}
