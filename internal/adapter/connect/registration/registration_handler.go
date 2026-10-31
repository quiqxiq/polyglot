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
}

func NewRegistrationConnectHandler(
	managerUC *uc.ManageRegistrationUseCase,
	convertUC *uc.ConvertUseCase,
) *RegistrationConnectHandler {
	return &RegistrationConnectHandler{managerUC: managerUC, convertUC: convertUC}
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
		FullName:       req.Msg.FullName,
		Phone:          req.Msg.Phone,
		Email:          req.Msg.Email,
		Address:        req.Msg.Address,
		Latitude:       lat,
		Longitude:      lon,
		Notes:          req.Msg.Notes,
		PlanID:         req.Msg.PlanId,
		NIK:            req.Msg.Nik,
		IDCardPhotoURL: req.Msg.IdCardPhotoUrl,
		HousePhotoURL:  req.Msg.HousePhotoUrl,
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
	if h.managerUC == nil {
		return nil, response.Unavailable("registration usecase unavailable")
	}
	list, err := h.managerUC.List(ctx, port.RegistrationFilter{
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
	if h.managerUC == nil {
		return nil, response.Unavailable("registration usecase unavailable")
	}
	reg, err := h.managerUC.FindByID(ctx, req.Msg.Id)
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
	installed, err := h.managerUC.MarkInstalledWithEvidence(ctx, req.Msg.Id, uc.MarkInstalledInput{
		DeviceID:             req.Msg.DeviceId,
		TechnicianNotes:      req.Msg.TechnicianNotes,
		InstallationPhotoURL: req.Msg.InstallationPhotoUrl,
		ONTMACAddress:        req.Msg.OntMacAddress,
		ONTSerialNumber:      req.Msg.OntSerialNumber,
		OLTDeviceID:          req.Msg.GetOltDeviceId(),
		PONIndex:             intPtrOrNil(req.Msg.GetPonIndex()),
		ONUIndex:             intPtrOrNil(req.Msg.GetOnuIndex()),
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.MarkInstalledResponse{
		Registration: toProtoRegistration(&installed),
	}), nil
}

// RecordSurvey mencatat hasil survey lokasi (PSB v2): PENDING/APPROVED → SURVEY.
func (h *RegistrationConnectHandler) RecordSurvey(ctx context.Context, req *connect.Request[devicepb.RecordSurveyRequest]) (*connect.Response[devicepb.RecordSurveyResponse], error) {
	if h.managerUC == nil {
		return nil, response.Unavailable("registration usecase unavailable")
	}
	in := uc.SurveyInput{
		SurveyStatus:              req.Msg.SurveyStatus,
		SurveyNotes:               req.Msg.SurveyNotes,
		EstimatedCableLengthMeter: int(req.Msg.EstimatedCableLengthMeter),
	}
	if req.Msg.SurveyorId != "" {
		if id, err := strconv.ParseUint(req.Msg.SurveyorId, 10, 64); err == nil {
			v := uint(id)
			in.SurveyorID = &v
		}
	}
	if req.Msg.SurveyDateUnix > 0 {
		d := time.Unix(req.Msg.SurveyDateUnix, 0).UTC()
		in.SurveyDate = &d
	}
	if req.Msg.SurveyTimeHhmm != "" {
		if t, err := time.Parse("15:04", req.Msg.SurveyTimeHhmm); err == nil {
			in.SurveyTime = &t
		}
	}
	surveyed, err := h.managerUC.RecordSurvey(ctx, req.Msg.Id, in)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.RecordSurveyResponse{
		Registration: toProtoRegistration(&surveyed),
	}), nil
}

// StartInstallation memulai pekerjaan lapangan (APPROVED → ON_PROGRESS).
func (h *RegistrationConnectHandler) StartInstallation(ctx context.Context, req *connect.Request[devicepb.StartInstallationRequest]) (*connect.Response[devicepb.StartInstallationResponse], error) {
	if h.managerUC == nil {
		return nil, response.Unavailable("registration usecase unavailable")
	}
	started, err := h.managerUC.StartInstallation(ctx, req.Msg.Id, "1")
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.StartInstallationResponse{
		Registration: toProtoRegistration(&started),
	}), nil
}

// TrackRegistration — PUBLIC: lacak status pendaftaran via token tanpa login.
func (h *RegistrationConnectHandler) TrackRegistration(ctx context.Context, req *connect.Request[devicepb.TrackRegistrationRequest]) (*connect.Response[devicepb.TrackRegistrationResponse], error) {
	if h.managerUC == nil {
		return nil, response.Unavailable("registration usecase unavailable")
	}
	reg, err := h.managerUC.TrackByPublicToken(ctx, req.Msg.PublicToken)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	events, err := h.managerUC.ListEvents(ctx, reg.ID)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.TrackRegistrationResponse{
		Registration: toProtoRegistration(&reg),
		Events:       toProtoRegistrationEvents(events),
	}), nil
}

// ListPublicPlans — PUBLIC: pilihan paket aktif untuk formulir pendaftaran.
func (h *RegistrationConnectHandler) ListPublicPlans(ctx context.Context, req *connect.Request[devicepb.ListPublicPlansRequest]) (*connect.Response[devicepb.ListPublicPlansResponse], error) {
	if h.managerUC == nil {
		return nil, response.Unavailable("registration usecase unavailable")
	}
	plans, err := h.managerUC.PublicPlans(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*devicepb.PublicPlan, len(plans))
	for i := range plans {
		p := plans[i]
		out[i] = &devicepb.PublicPlan{
			Id:                    p.ID,
			Name:                  p.Name,
			ServiceType:           p.ServiceType,
			Price:                 p.Price,
			BandwidthDownloadKbps: int32(p.BandwidthDownloadKbps),
			BandwidthUploadKbps:   int32(p.BandwidthUploadKbps),
			InstallationFee:       p.InstallationFee,
		}
	}
	return connect.NewResponse(&devicepb.ListPublicPlansResponse{Plans: out}), nil
}

// ListRegistrationEvents mengembalikan timeline PSB satu pendaftaran.
func (h *RegistrationConnectHandler) ListRegistrationEvents(ctx context.Context, req *connect.Request[devicepb.ListRegistrationEventsRequest]) (*connect.Response[devicepb.ListRegistrationEventsResponse], error) {
	if h.managerUC == nil {
		return nil, response.Unavailable("registration usecase unavailable")
	}
	events, err := h.managerUC.ListEvents(ctx, req.Msg.Id)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListRegistrationEventsResponse{
		Events: toProtoRegistrationEvents(events),
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

// intPtrOrNil mengubah nilai 0 (tidak diisi) menjadi nil untuk field opsional.
func intPtrOrNil(v int32) *int {
	if v <= 0 {
		return nil
	}
	n := int(v)
	return &n
}
