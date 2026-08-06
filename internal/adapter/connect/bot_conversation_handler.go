package connectadapter

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
)

func (h *BotConnectHandler) ListConversations(ctx context.Context, req *connect.Request[devicepb.ListConversationsRequest]) (*connect.Response[devicepb.ListConversationsResponse], error) {
	return connect.NewResponse(&devicepb.ListConversationsResponse{
		Conversations: []*devicepb.Conversation{},
	}), nil
}

func (h *BotConnectHandler) GetConversation(ctx context.Context, req *connect.Request[devicepb.GetConversationRequest]) (*connect.Response[devicepb.GetConversationResponse], error) {
	return connect.NewResponse(&devicepb.GetConversationResponse{
		Conversation: &devicepb.Conversation{
			Id:          req.Msg.Id,
			ClientPhone: "628123456789",
			Status:      "bot",
		},
	}), nil
}

func (h *BotConnectHandler) TakeOverConversation(ctx context.Context, req *connect.Request[devicepb.TakeOverConversationRequest]) (*connect.Response[devicepb.TakeOverConversationResponse], error) {
	return connect.NewResponse(&devicepb.TakeOverConversationResponse{
		Message: "conversation taken over by technician",
	}), nil
}

func (h *BotConnectHandler) ResetConversationBot(ctx context.Context, req *connect.Request[devicepb.ResetConversationBotRequest]) (*connect.Response[devicepb.ResetConversationBotResponse], error) {
	return connect.NewResponse(&devicepb.ResetConversationBotResponse{
		Message: "conversation bot control reset to automatic",
	}), nil
}

func (h *BotConnectHandler) CloseConversation(ctx context.Context, req *connect.Request[devicepb.CloseConversationRequest]) (*connect.Response[devicepb.CloseConversationResponse], error) {
	return connect.NewResponse(&devicepb.CloseConversationResponse{
		Message: "conversation closed",
	}), nil
}
