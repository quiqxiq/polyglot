package registration

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainReg "github.com/quixiq/polyglot/internal/domain/registration"
	regUC "github.com/quixiq/polyglot/internal/usecase/registration"
	"github.com/quixiq/polyglot/pkg/response"
)

// RegistrationConnectHandler implements polyglot.v1.RegistrationService.
type RegistrationConnectHandler struct {
	useCase *regUC.RegistrationService
}

func NewRegistrationConnectHandler(uc *regUC.RegistrationService) *RegistrationConnectHandler {
	return &RegistrationConnectHandler{useCase: uc}
}

func (h *RegistrationConnectHandler) CreateRegistration(ctx context.Context, req *connect.Request[devicepb.CreateRegistrationRequest]) (*connect.Response[devicepb.CreateRegistrationResponse], error) {
	if req.Msg.Registration == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("registration is required"))
	}
	created, err := h.useCase.Create(ctx, fromProto(req.Msg.Registration))
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.CreateRegistrationResponse{
		Registration: toProto(&created),
		Message:      "registration created",
	}), nil
}

func (h *RegistrationConnectHandler) ListRegistrations(ctx context.Context, req *connect.Request[devicepb.ListRegistrationsRequest]) (*connect.Response[devicepb.ListRegistrationsResponse], error) {
	regs, err := h.useCase.List(ctx, req.Msg.Status, int(req.Msg.Limit))
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*devicepb.Registration, len(regs))
	for i := range regs {
		out[i] = toProto(&regs[i])
	}
	return connect.NewResponse(&devicepb.ListRegistrationsResponse{Registrations: out}), nil
}

func (h *RegistrationConnectHandler) GetRegistration(ctx context.Context, req *connect.Request[devicepb.GetRegistrationRequest]) (*connect.Response[devicepb.GetRegistrationResponse], error) {
	reg, err := h.useCase.Get(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetRegistrationResponse{Registration: toProto(&reg)}), nil
}

func (h *RegistrationConnectHandler) ReviewRegistration(ctx context.Context, req *connect.Request[devicepb.ReviewRegistrationRequest]) (*connect.Response[devicepb.ReviewRegistrationResponse], error) {
	var scheduled *time.Time
	if req.Msg.ScheduledInstallDateUnix > 0 {
		t := time.Unix(req.Msg.ScheduledInstallDateUnix, 0)
		scheduled = &t
	}
	reg, err := h.useCase.Review(ctx, req.Msg.Id, req.Msg.Approve, req.Msg.ReviewedBy, req.Msg.Notes, scheduled, req.Msg.AssignedTechnicianId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	msg := "registration approved"
	if !req.Msg.Approve {
		msg = "registration rejected"
	}
	return connect.NewResponse(&devicepb.ReviewRegistrationResponse{
		Registration: toProto(&reg), Message: msg,
	}), nil
}

func (h *RegistrationConnectHandler) MarkInstalled(ctx context.Context, req *connect.Request[devicepb.MarkInstalledRequest]) (*connect.Response[devicepb.MarkInstalledResponse], error) {
	res, err := h.useCase.Install(ctx, regUC.MarkInstalledInput{
		ID:              req.Msg.Id,
		DeviceID:        req.Msg.DeviceId,
		RemoteUsername:  req.Msg.RemoteUsername,
		Password:        req.Msg.Password,
		TechnicianNotes: req.Msg.TechnicianNotes,
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.MarkInstalledResponse{
		Registration:   toProto(&res.Registration),
		CustomerId:     res.CustomerID,
		SubscriptionId: res.SubscriptionID,
		Message:        "installation complete; account provisioned on router",
	}), nil
}

func (h *RegistrationConnectHandler) CancelRegistration(ctx context.Context, req *connect.Request[devicepb.CancelRegistrationRequest]) (*connect.Response[devicepb.CancelRegistrationResponse], error) {
	reg, err := h.useCase.Cancel(ctx, req.Msg.Id, req.Msg.Reason)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.CancelRegistrationResponse{
		Registration: toProto(&reg), Message: "registration cancelled",
	}), nil
}

// ─── mappers ─────────────────────────────────────────────────────────────

func toProto(r *domainReg.Registration) *devicepb.Registration {
	if r == nil {
		return nil
	}
	pb := &devicepb.Registration{
		Id:              r.ID,
		TenantId:        r.TenantID,
		RegistrationNo:  r.RegistrationNo,
		PlanId:          r.PlanID,
		FullName:        r.FullName,
		Phone:           r.Phone,
		Address:         r.Address,
		Latitude:        r.Latitude,
		Longitude:       r.Longitude,
		Notes:           r.Notes,
		Status:          r.Status,
		AdminNotes:      r.AdminNotes,
		DeviceId:        r.DeviceID,
		TechnicianNotes: r.TechnicianNotes,
		CustomerId:      r.CustomerID,
		SubscriptionId:  r.SubscriptionID,
		RejectedReason:  r.RejectedReason,
		CancelReason:    r.CancelReason,
		CreatedAtUnix:   r.CreatedAt.Unix(),
	}
	if r.ReviewedBy != nil {
		pb.ReviewedBy = *r.ReviewedBy
	}
	if r.ReviewedAt != nil {
		pb.ReviewedAtUnix = r.ReviewedAt.Unix()
	}
	if r.ScheduledInstallDate != nil {
		pb.ScheduledInstallDateUnix = r.ScheduledInstallDate.Unix()
	}
	if r.AssignedTechnicianID != nil {
		pb.AssignedTechnicianId = *r.AssignedTechnicianID
	}
	if r.InstalledAt != nil {
		pb.InstalledAtUnix = r.InstalledAt.Unix()
	}
	if r.RejectedAt != nil {
		pb.RejectedAtUnix = r.RejectedAt.Unix()
	}
	if r.CancelledAt != nil {
		pb.CancelledAtUnix = r.CancelledAt.Unix()
	}
	return pb
}

func fromProto(pb *devicepb.Registration) domainReg.Registration {
	if pb == nil {
		return domainReg.Registration{}
	}
	return domainReg.Registration{
		ID:             pb.Id,
		TenantID:       pb.TenantId,
		RegistrationNo: pb.RegistrationNo,
		PlanID:         pb.PlanId,
		FullName:       pb.FullName,
		Phone:          pb.Phone,
		Address:        pb.Address,
		Latitude:       pb.Latitude,
		Longitude:      pb.Longitude,
		Notes:          pb.Notes,
	}
}
