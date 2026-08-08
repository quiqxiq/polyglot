package bot

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
)

func (h *KnowledgeConnectHandler) ListKnowledge(ctx context.Context, req *connect.Request[devicepb.ListKnowledgeRequest]) (*connect.Response[devicepb.ListKnowledgeResponse], error) {
	return connect.NewResponse(&devicepb.ListKnowledgeResponse{
		Items: []*devicepb.KnowledgeItem{},
	}), nil
}

func (h *KnowledgeConnectHandler) GetKnowledge(ctx context.Context, req *connect.Request[devicepb.GetKnowledgeRequest]) (*connect.Response[devicepb.GetKnowledgeResponse], error) {
	return connect.NewResponse(&devicepb.GetKnowledgeResponse{
		Item: &devicepb.KnowledgeItem{
			Id:       req.Msg.Id,
			Title:    "Prosedur Reset Router Mikrotik",
			Content:  "Langkah-langkah reset RouterOS...",
			Category: "Mikrotik",
		},
	}), nil
}

func (h *KnowledgeConnectHandler) CreateKnowledge(ctx context.Context, req *connect.Request[devicepb.CreateKnowledgeRequest]) (*connect.Response[devicepb.CreateKnowledgeResponse], error) {
	return connect.NewResponse(&devicepb.CreateKnowledgeResponse{
		Item: &devicepb.KnowledgeItem{
			Id:       "knw-001",
			Title:    req.Msg.Title,
			Content:  req.Msg.Content,
			Category: req.Msg.Category,
			Tags:     req.Msg.Tags,
		},
	}), nil
}

func (h *KnowledgeConnectHandler) UpdateKnowledge(ctx context.Context, req *connect.Request[devicepb.UpdateKnowledgeRequest]) (*connect.Response[devicepb.UpdateKnowledgeResponse], error) {
	return connect.NewResponse(&devicepb.UpdateKnowledgeResponse{
		Item: &devicepb.KnowledgeItem{
			Id:       req.Msg.Id,
			Title:    req.Msg.Title,
			Content:  req.Msg.Content,
			Category: req.Msg.Category,
			Tags:     req.Msg.Tags,
		},
	}), nil
}

func (h *KnowledgeConnectHandler) DeleteKnowledge(ctx context.Context, req *connect.Request[devicepb.DeleteKnowledgeRequest]) (*connect.Response[devicepb.DeleteKnowledgeResponse], error) {
	return connect.NewResponse(&devicepb.DeleteKnowledgeResponse{
		Message: "knowledge item deleted successfully",
	}), nil
}
