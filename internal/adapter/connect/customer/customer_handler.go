package customer

import (
	"context"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
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

func (h *CustomerConnectHandler) toEnrichedProto(ctx context.Context, c *domainCustomer.Customer) *polyglotv1.Customer {
	if c == nil {
		return nil
	}
	if h.customerUC != nil {
		detail := h.customerUC.Enrich(ctx, *c)
		return toProtoCustomerDetail(&detail)
	}
	return toProtoCustomer(c)
}

func (h *CustomerConnectHandler) ListCustomers(ctx context.Context, req *connect.Request[polyglotv1.ListCustomersRequest]) (*connect.Response[polyglotv1.ListCustomersResponse], error) {
	if h.customerUC == nil {
		return nil, response.Unavailable("customer usecase unavailable")
	}
	customers, err := h.customerUC.ListCustomers(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.ListCustomersResponse{
		Customers: toProtoCustomerDetailList(customers),
	}), nil
}

func (h *CustomerConnectHandler) GetCustomer(ctx context.Context, req *connect.Request[polyglotv1.GetCustomerRequest]) (*connect.Response[polyglotv1.GetCustomerResponse], error) {
	if h.customerUC == nil {
		return nil, response.Unavailable("customer usecase unavailable")
	}
	detail, err := h.customerUC.GetCustomer(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.GetCustomerResponse{
		Customer: toProtoCustomerDetail(&detail),
	}), nil
}

func (h *CustomerConnectHandler) CreateCustomer(ctx context.Context, req *connect.Request[polyglotv1.CreateCustomerRequest]) (*connect.Response[polyglotv1.CreateCustomerResponse], error) {
	if h.customerUC == nil {
		return nil, response.Unavailable("customer usecase unavailable")
	}
	c := fromProtoCustomer(req.Msg.Customer)
	created, err := h.customerUC.CreateCustomer(ctx, c)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.CreateCustomerResponse{
		Customer: h.toEnrichedProto(ctx, &created),
		Message:  "Customer created successfully",
	}), nil
}

func (h *CustomerConnectHandler) UpdateCustomer(ctx context.Context, req *connect.Request[polyglotv1.UpdateCustomerRequest]) (*connect.Response[polyglotv1.UpdateCustomerResponse], error) {
	if h.customerUC == nil {
		return nil, response.Unavailable("customer usecase unavailable")
	}
	c := fromProtoCustomer(req.Msg.Customer)
	updated, err := h.customerUC.UpdateCustomer(ctx, c)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.UpdateCustomerResponse{
		Customer: h.toEnrichedProto(ctx, &updated),
		Message:  "Customer updated successfully",
	}), nil
}

func (h *CustomerConnectHandler) DeleteCustomer(ctx context.Context, req *connect.Request[polyglotv1.DeleteCustomerRequest]) (*connect.Response[polyglotv1.DeleteCustomerResponse], error) {
	if h.customerUC == nil {
		return nil, response.Unavailable("customer usecase unavailable")
	}
	if err := h.customerUC.DeleteCustomer(ctx, req.Msg.Id); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.DeleteCustomerResponse{
		Message: "Customer deleted successfully",
	}), nil
}

func (h *CustomerConnectHandler) FindByPhone(ctx context.Context, req *connect.Request[polyglotv1.FindCustomerByPhoneRequest]) (*connect.Response[polyglotv1.FindCustomerResponse], error) {
	c, err := h.customerUC.FindByPhone(ctx, req.Msg.Phone)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.FindCustomerResponse{Customer: h.toEnrichedProto(ctx, &c)}), nil
}

func (h *CustomerConnectHandler) FindByCustomerCode(ctx context.Context, req *connect.Request[polyglotv1.FindCustomerByCodeRequest]) (*connect.Response[polyglotv1.FindCustomerResponse], error) {
	c, err := h.customerUC.FindByCustomerCode(ctx, req.Msg.CustomerCode)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.FindCustomerResponse{Customer: h.toEnrichedProto(ctx, &c)}), nil
}

func (h *CustomerConnectHandler) FindByPortalCode(ctx context.Context, req *connect.Request[polyglotv1.FindCustomerByPortalCodeRequest]) (*connect.Response[polyglotv1.FindCustomerResponse], error) {
	c, err := h.customerUC.FindByPortalCode(ctx, req.Msg.PortalAccessCode)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.FindCustomerResponse{Customer: h.toEnrichedProto(ctx, &c)}), nil
}

func (h *CustomerConnectHandler) ListSubscriptions(ctx context.Context, req *connect.Request[polyglotv1.ListSubscriptionsRequest]) (*connect.Response[polyglotv1.ListSubscriptionsResponse], error) {
	return connect.NewResponse(&polyglotv1.ListSubscriptionsResponse{}), nil
}

// GetCustomer360 mengembalikan data pelanggan terpadu (profil, langganan,
// tagihan, pembayaran, dan riwayat PSB) — planning v2 Fase 5.
func (h *CustomerConnectHandler) GetCustomer360(ctx context.Context, req *connect.Request[polyglotv1.GetCustomer360Request]) (*connect.Response[polyglotv1.GetCustomer360Response], error) {
	if h.customerUC == nil {
		return nil, response.Unavailable("customer usecase unavailable")
	}
	data, err := h.customerUC.GetCustomer360(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	resp := &polyglotv1.GetCustomer360Response{
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
		resp.Subscriptions = append(resp.Subscriptions, &polyglotv1.Customer360Subscription{
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
		resp.Invoices = append(resp.Invoices, &polyglotv1.Customer360Invoice{
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
		resp.Payments = append(resp.Payments, &polyglotv1.Customer360Payment{
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
		resp.Registrations = append(resp.Registrations, &polyglotv1.Customer360Registration{
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
