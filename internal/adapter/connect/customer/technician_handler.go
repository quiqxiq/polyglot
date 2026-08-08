package customer

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
)

func (h *CustomerConnectHandler) ListTechnicians(ctx context.Context, req *connect.Request[devicepb.ListTechniciansRequest]) (*connect.Response[devicepb.ListTechniciansResponse], error) {
	return connect.NewResponse(&devicepb.ListTechniciansResponse{
		Technicians: []*devicepb.Technician{
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

func (h *CustomerConnectHandler) CreateTechnician(ctx context.Context, req *connect.Request[devicepb.CreateTechnicianRequest]) (*connect.Response[devicepb.CreateTechnicianResponse], error) {
	return connect.NewResponse(&devicepb.CreateTechnicianResponse{
		Technician: &devicepb.Technician{
			Id:          "tech-new",
			Name:        req.Msg.Name,
			PhoneNumber: req.Msg.PhoneNumber,
			Email:       req.Msg.Email,
			IsActive:    true,
		},
	}), nil
}

func (h *CustomerConnectHandler) UpdateTechnician(ctx context.Context, req *connect.Request[devicepb.UpdateTechnicianRequest]) (*connect.Response[devicepb.UpdateTechnicianResponse], error) {
	return connect.NewResponse(&devicepb.UpdateTechnicianResponse{
		Technician: &devicepb.Technician{
			Id:          req.Msg.Id,
			Name:        req.Msg.Name,
			PhoneNumber: req.Msg.PhoneNumber,
			Email:       req.Msg.Email,
			IsActive:    true,
		},
	}), nil
}

func (h *CustomerConnectHandler) ToggleTechnicianActive(ctx context.Context, req *connect.Request[devicepb.ToggleTechnicianActiveRequest]) (*connect.Response[devicepb.ToggleTechnicianActiveResponse], error) {
	return connect.NewResponse(&devicepb.ToggleTechnicianActiveResponse{
		Message:  "technician active status updated",
		IsActive: req.Msg.IsActive,
	}), nil
}

func (h *CustomerConnectHandler) DeleteTechnician(ctx context.Context, req *connect.Request[devicepb.DeleteTechnicianRequest]) (*connect.Response[devicepb.DeleteTechnicianResponse], error) {
	return connect.NewResponse(&devicepb.DeleteTechnicianResponse{
		Message: "technician deleted successfully",
	}), nil
}
