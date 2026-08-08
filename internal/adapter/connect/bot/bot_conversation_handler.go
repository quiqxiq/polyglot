package bot

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
)

func (h *BotConnectHandler) ListConversations(ctx context.Context, req *connect.Request[devicepb.ListConversationsRequest]) (*connect.Response[devicepb.ListConversationsResponse], error) {
	if h.convService == nil {
		return connect.NewResponse(&devicepb.ListConversationsResponse{Conversations: []*devicepb.Conversation{}}), nil
	}

	convs, err := h.convService.ListConversations("")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbConvs := make([]*devicepb.Conversation, len(convs))
	for i, c := range convs {
		pbConvs[i] = &devicepb.Conversation{
			Id:          fmt.Sprintf("%d", c.ID),
			ClientPhone: c.CustomerWANumber,
			Status:      string(c.Status),
		}
	}
	return connect.NewResponse(&devicepb.ListConversationsResponse{Conversations: pbConvs}), nil
}

func (h *BotConnectHandler) GetConversation(ctx context.Context, req *connect.Request[devicepb.GetConversationRequest]) (*connect.Response[devicepb.GetConversationResponse], error) {
	if h.convService == nil {
		return connect.NewResponse(&devicepb.GetConversationResponse{}), nil
	}

	var convID uint
	_, _ = fmt.Sscanf(req.Msg.Id, "%d", &convID)

	c, err := h.convService.GetConversationWithMessages(convID)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	pbMsgs := make([]*devicepb.ConversationMessage, len(c.Messages))
	for i, m := range c.Messages {
		pbMsgs[i] = &devicepb.ConversationMessage{
			Id:      fmt.Sprintf("%d", m.ID),
			Sender:  string(m.SenderType),
			Text:    m.Content,
		}
	}

	return connect.NewResponse(&devicepb.GetConversationResponse{
		Conversation: &devicepb.Conversation{
			Id:          fmt.Sprintf("%d", c.ID),
			ClientPhone: c.CustomerWANumber,
			Status:      string(c.Status),
			Messages:    pbMsgs,
		},
	}), nil
}

func (h *BotConnectHandler) TakeOverConversation(ctx context.Context, req *connect.Request[devicepb.TakeOverConversationRequest]) (*connect.Response[devicepb.TakeOverConversationResponse], error) {
	if h.convService != nil {
		var convID uint
		_, _ = fmt.Sscanf(req.Msg.Id, "%d", &convID)
		_ = h.convService.TakeOver(convID, 1)
	}
	return connect.NewResponse(&devicepb.TakeOverConversationResponse{
		Message: "conversation taken over by technician",
	}), nil
}

func (h *BotConnectHandler) ResetConversationBot(ctx context.Context, req *connect.Request[devicepb.ResetConversationBotRequest]) (*connect.Response[devicepb.ResetConversationBotResponse], error) {
	if h.convService != nil {
		var convID uint
		_, _ = fmt.Sscanf(req.Msg.Id, "%d", &convID)
		_ = h.convService.ResetBot(convID)
	}
	return connect.NewResponse(&devicepb.ResetConversationBotResponse{
		Message: "conversation bot control reset to automatic",
	}), nil
}

func (h *BotConnectHandler) CloseConversation(ctx context.Context, req *connect.Request[devicepb.CloseConversationRequest]) (*connect.Response[devicepb.CloseConversationResponse], error) {
	if h.convService != nil {
		var convID uint
		_, _ = fmt.Sscanf(req.Msg.Id, "%d", &convID)
		_ = h.convService.CloseConversation(convID)
	}
	return connect.NewResponse(&devicepb.CloseConversationResponse{
		Message: "conversation closed",
	}), nil
}
