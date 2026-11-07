package customer

import (
	"context"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
)

func (h *CustomerConnectHandler) ListTechnicians(ctx context.Context, req *connect.Request[polyglotv1.ListTechniciansRequest]) (*connect.Response[polyglotv1.ListTechniciansResponse], error) {
	return connect.NewResponse(&polyglotv1.ListTechniciansResponse{
		Technicians: []*polyglotv1.Technician{
			{
				Id:          "tech-1",
				Name:        "Ahmad Field Tech",
				PhoneNumber: "62899887766",
				Email:       "ahmad@isp.net",
				IsActive:    true,
			},
		},
	}), nil
}

func (h *CustomerConnectHandler) CreateTechnician(ctx context.Context, req *connect.Request[polyglotv1.CreateTechnicianRequest]) (*connect.Response[polyglotv1.CreateTechnicianResponse], error) {
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

func (h *CustomerConnectHandler) UpdateTechnician(ctx context.Context, req *connect.Request[polyglotv1.UpdateTechnicianRequest]) (*connect.Response[polyglotv1.UpdateTechnicianResponse], error) {
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

func (h *CustomerConnectHandler) ToggleTechnicianActive(ctx context.Context, req *connect.Request[polyglotv1.ToggleTechnicianActiveRequest]) (*connect.Response[polyglotv1.ToggleTechnicianActiveResponse], error) {
	return connect.NewResponse(&polyglotv1.ToggleTechnicianActiveResponse{
		Message:  "technician active status updated",
		IsActive: req.Msg.IsActive,
	}), nil
}

func (h *CustomerConnectHandler) DeleteTechnician(ctx context.Context, req *connect.Request[polyglotv1.DeleteTechnicianRequest]) (*connect.Response[polyglotv1.DeleteTechnicianResponse], error) {
	return connect.NewResponse(&polyglotv1.DeleteTechnicianResponse{
		Message: "technician deleted successfully",
	}), nil
}
