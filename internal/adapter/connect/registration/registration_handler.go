package registration

import (
	"context"
	"strconv"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
	"github.com/quixiq/polyglot/internal/port"
	uc "github.com/quixiq/polyglot/internal/usecase/registration"
	"github.com/quixiq/polyglot/pkg/response"
)

type RegistrationConnectHandler struct {
	managerUC *uc.ManageRegistrationUseCase
	convertUC *uc.ConvertUseCase
	repo      port.RegistrationRepository
}

func NewRegistrationConnectHandler(
	managerUC *uc.ManageRegistrationUseCase,
	convertUC *uc.ConvertUseCase,
	repo port.RegistrationRepository,
) *RegistrationConnectHandler {
	return &RegistrationConnectHandler{managerUC: managerUC, convertUC: convertUC, repo: repo}
}

// SubmitRegistration — PUBLIC (calon pelanggan).
func (h *RegistrationConnectHandler) SubmitRegistration(ctx context.Context, req *connect.Request[devicepb.SubmitRegistrationRequest]) (*connect.Response[devicepb.SubmitRegistrationResponse], error) {
	if h.managerUC == nil {
		return nil, response.Unavailable("registration usecase unavailable")
	}
	var lat, lon *float64
	if req.Msg.HasCoordinates {
		vLat, vLon := req.Msg.Latitude, req.Msg.Longitude
		lat, lon = &vLat, &vLon
	}
	reg := domainRegistration.Registration{
		FullName:  req.Msg.FullName,
		Phone:     req.Msg.Phone,
		Email:     req.Msg.Email,
		Address:   req.Msg.Address,
		Latitude:  lat,
		Longitude: lon,
		Notes:     req.Msg.Notes,
		PlanID:    req.Msg.PlanId,
	}
	submitted, err := h.managerUC.Submit(ctx, reg)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.SubmitRegistrationResponse{
		Registration: toProtoRegistration(&submitted),
	}), nil
}

func (h *RegistrationConnectHandler) ListRegistrations(ctx context.Context, req *connect.Request[devicepb.ListRegistrationsRequest]) (*connect.Response[devicepb.ListRegistrationsResponse], error) {
	if h.repo == nil {
		return nil, response.Unavailable("registration repository unavailable")
	}
	list, err := h.repo.List(ctx, port.RegistrationFilter{
		Status: req.Msg.Status,
		Phone:  req.Msg.Phone,
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListRegistrationsResponse{
		Registrations: toProtoRegistrationList(list),
	}), nil
}

func (h *RegistrationConnectHandler) GetRegistration(ctx context.Context, req *connect.Request[devicepb.GetRegistrationRequest]) (*connect.Response[devicepb.GetRegistrationResponse], error) {
	if h.repo == nil {
		return nil, response.Unavailable("registration repository unavailable")
	}
	reg, err := h.repo.FindByID(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetRegistrationResponse{
		Registration: toProtoRegistration(&reg),
	}), nil
}

func (h *RegistrationConnectHandler) ApproveRegistration(ctx context.Context, req *connect.Request[devicepb.ApproveRegistrationRequest]) (*connect.Response[devicepb.ApproveRegistrationResponse], error) {
	if h.managerUC == nil {
		return nil, response.Unavailable("registration usecase unavailable")
	}
	approved, err := h.managerUC.Approve(ctx, req.Msg.Id, 1, req.Msg.AdminNotes)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ApproveRegistrationResponse{
		Registration: toProtoRegistration(&approved),
	}), nil
}

func (h *RegistrationConnectHandler) ScheduleInstall(ctx context.Context, req *connect.Request[devicepb.ScheduleInstallRequest]) (*connect.Response[devicepb.ScheduleInstallResponse], error) {
	if h.managerUC == nil {
		return nil, response.Unavailable("registration usecase unavailable")
	}
	date := time.Unix(req.Msg.InstallDateUnix, 0).UTC()
	var timeOfDay *time.Time
	if req.Msg.InstallTimeHhmm != "" {
		if t, err := time.Parse("15:04", req.Msg.InstallTimeHhmm); err == nil {
			timeOfDay = &t
		}
	}
	var techID *uint
	if req.Msg.TechnicianId != "" {
		if u, err := strconv.ParseUint(req.Msg.TechnicianId, 10, 32); err == nil {
			val := uint(u)
			techID = &val
		}
	}
	scheduled, err := h.managerUC.ScheduleInstall(ctx, req.Msg.Id, date, timeOfDay, techID)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ScheduleInstallResponse{
		Registration: toProtoRegistration(&scheduled),
	}), nil
}

func (h *RegistrationConnectHandler) MarkInstalled(ctx context.Context, req *connect.Request[devicepb.MarkInstalledRequest]) (*connect.Response[devicepb.MarkInstalledResponse], error) {
	if h.managerUC == nil {
		return nil, response.Unavailable("registration usecase unavailable")
	}
	installed, err := h.managerUC.MarkInstalled(ctx, req.Msg.Id, nil, req.Msg.TechnicianNotes)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.MarkInstalledResponse{
		Registration: toProtoRegistration(&installed),
	}), nil
}

func (h *RegistrationConnectHandler) RejectRegistration(ctx context.Context, req *connect.Request[devicepb.RejectRegistrationRequest]) (*connect.Response[devicepb.RejectRegistrationResponse], error) {
	if h.managerUC == nil {
		return nil, response.Unavailable("registration usecase unavailable")
	}
	rejected, err := h.managerUC.Reject(ctx, req.Msg.Id, req.Msg.Reason, 1)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.RejectRegistrationResponse{
		Registration: toProtoRegistration(&rejected),
	}), nil
}

func (h *RegistrationConnectHandler) CancelRegistration(ctx context.Context, req *connect.Request[devicepb.CancelRegistrationRequest]) (*connect.Response[devicepb.CancelRegistrationResponse], error) {
	if h.managerUC == nil {
		return nil, response.Unavailable("registration usecase unavailable")
	}
	cancelled, err := h.managerUC.Cancel(ctx, req.Msg.Id, req.Msg.Reason)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.CancelRegistrationResponse{
		Registration: toProtoRegistration(&cancelled),
	}), nil
}

func (h *RegistrationConnectHandler) ConvertRegistration(ctx context.Context, req *connect.Request[devicepb.ConvertRegistrationRequest]) (*connect.Response[devicepb.ConvertRegistrationResponse], error) {
	if h.convertUC == nil {
		return nil, response.Unavailable("convert usecase unavailable")
	}
	converted, err := h.convertUC.ConvertWithDevice(ctx, req.Msg.Id, req.Msg.DeviceId, "1")
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ConvertRegistrationResponse{
		Registration:   toProtoRegistration(&converted),
		CustomerId:     converted.CustomerID,
		SubscriptionId: converted.SubscriptionID,
		InvoiceId:      converted.InvoiceID,
	}), nil
}
