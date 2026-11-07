package bot

import (
	"context"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
)

func (h *BotConnectHandler) ListTechnicians(ctx context.Context, req *connect.Request[polyglotv1.ListTechniciansRequest]) (*connect.Response[polyglotv1.ListTechniciansResponse], error) {
	return connect.NewResponse(&polyglotv1.ListTechniciansResponse{
		Technicians: []*polyglotv1.Technician{},
	}), nil
}

func (h *BotConnectHandler) CreateTechnician(ctx context.Context, req *connect.Request[polyglotv1.CreateTechnicianRequest]) (*connect.Response[polyglotv1.CreateTechnicianResponse], error) {
	return connect.NewResponse(&polyglotv1.CreateTechnicianResponse{
		Technician: &polyglotv1.Technician{
			Id:          "tech-new",
			Name:        req.Msg.Name,
			PhoneNumber: req.Msg.PhoneNumber,
			Email:       req.Msg.Email,
			IsActive:    true,
		},
	}), nil
}

func (h *BotConnectHandler) UpdateTechnician(ctx context.Context, req *connect.Request[polyglotv1.UpdateTechnicianRequest]) (*connect.Response[polyglotv1.UpdateTechnicianResponse], error) {
	return connect.NewResponse(&polyglotv1.UpdateTechnicianResponse{
		Technician: &polyglotv1.Technician{
			Id:          req.Msg.Id,
			Name:        req.Msg.Name,
			PhoneNumber: req.Msg.PhoneNumber,
			Email:       req.Msg.Email,
			IsActive:    true,
		},
	}), nil
}

func (h *BotConnectHandler) ToggleTechnicianActive(ctx context.Context, req *connect.Request[polyglotv1.ToggleTechnicianActiveRequest]) (*connect.Response[polyglotv1.ToggleTechnicianActiveResponse], error) {
	return connect.NewResponse(&polyglotv1.ToggleTechnicianActiveResponse{
		Message:  "technician status updated",
		IsActive: req.Msg.IsActive,
	}), nil
}

func (h *BotConnectHandler) DeleteTechnician(ctx context.Context, req *connect.Request[polyglotv1.DeleteTechnicianRequest]) (*connect.Response[polyglotv1.DeleteTechnicianResponse], error) {
	return connect.NewResponse(&polyglotv1.DeleteTechnicianResponse{
		Message: "technician deleted successfully",
	}), nil
}
