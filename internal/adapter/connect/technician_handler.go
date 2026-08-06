package connectadapter

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
)

func (h *KnowledgeConnectHandler) ListTechnicians(ctx context.Context, req *connect.Request[devicepb.ListTechniciansRequest]) (*connect.Response[devicepb.ListTechniciansResponse], error) {
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

func (h *KnowledgeConnectHandler) CreateTechnician(ctx context.Context, req *connect.Request[devicepb.CreateTechnicianRequest]) (*connect.Response[devicepb.CreateTechnicianResponse], error) {
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

func (h *KnowledgeConnectHandler) UpdateTechnician(ctx context.Context, req *connect.Request[devicepb.UpdateTechnicianRequest]) (*connect.Response[devicepb.UpdateTechnicianResponse], error) {
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

func (h *KnowledgeConnectHandler) ToggleTechnicianActive(ctx context.Context, req *connect.Request[devicepb.ToggleTechnicianActiveRequest]) (*connect.Response[devicepb.ToggleTechnicianActiveResponse], error) {
	return connect.NewResponse(&devicepb.ToggleTechnicianActiveResponse{
		Message:  "technician active status updated",
		IsActive: req.Msg.IsActive,
	}), nil
}

func (h *KnowledgeConnectHandler) DeleteTechnician(ctx context.Context, req *connect.Request[devicepb.DeleteTechnicianRequest]) (*connect.Response[devicepb.DeleteTechnicianResponse], error) {
	return connect.NewResponse(&devicepb.DeleteTechnicianResponse{
		Message: "technician deleted successfully",
	}), nil
}
