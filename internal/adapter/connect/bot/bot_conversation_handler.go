package bot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/bot"
)

func (h *BotConnectHandler) ListConversations(ctx context.Context, req *connect.Request[devicepb.ListConversationsRequest]) (*connect.Response[devicepb.ListConversationsResponse], error) {
	if h.convService == nil {
		return connect.NewResponse(&devicepb.ListConversationsResponse{Conversations: []*devicepb.Conversation{}}), nil
	}

	var convs []bot.Conversation
	var err error
	if req.Msg.SessionId != "" {
		var sessionID uint
		_, _ = fmt.Sscanf(req.Msg.SessionId, "%d", &sessionID)
		convs, err = h.convService.ListConversationsBySession(ctx, sessionID)
	} else {
		convs, err = h.convService.ListConversations(ctx, "")
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbConvs := make([]*devicepb.Conversation, len(convs))
	for i, c := range convs {
		pbConvs[i] = &devicepb.Conversation{
			Id:          fmt.Sprintf("%d", c.ID),
			SessionId:   fmt.Sprintf("%d", c.SessionID),
			ClientPhone: c.CustomerWANumber,
			Status:      string(c.Status),
			UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
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

	c, err := h.convService.GetConversationWithMessages(ctx, convID)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	pbMsgs := make([]*devicepb.ConversationMessage, len(c.Messages))
	for i, m := range c.Messages {
		pbMsgs[i] = &devicepb.ConversationMessage{
			Id:     fmt.Sprintf("%d", m.ID),
			Sender: string(m.SenderType),
			Text:   m.Content,
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

func (h *BotConnectHandler) GetConversationContext(ctx context.Context, req *connect.Request[devicepb.GetConversationContextRequest]) (*connect.Response[devicepb.GetConversationContextResponse], error) {
	var convID uint
	_, _ = fmt.Sscanf(req.Msg.Id, "%d", &convID)
	if convID == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("conversation id is required"))
	}
	if h.contextProvider == nil {
		return nil, connect.NewError(connect.CodeUnimplemented, errors.New("conversation context provider is not available"))
	}

	info, err := h.contextProvider.GetConversationContext(ctx, convID)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	pbMsgs := make([]*devicepb.ConversationContextMessage, len(info.RecentMessages))
	for i, m := range info.RecentMessages {
		pbMsg := &devicepb.ConversationContextMessage{
			Id:        fmt.Sprintf("%d", m.ID),
			Sender:    string(m.SenderType),
			Text:      m.Content,
			TokenIn:   int64(m.TokenIn),
			TokenOut:  int64(m.TokenOut),
			Timestamp: m.CreatedAt.Format(time.RFC3339),
		}
		if m.LLMConfigID != nil {
			pbMsg.LlmConfigId = fmt.Sprintf("%d", *m.LLMConfigID)
		}
		pbMsgs[i] = pbMsg
	}

	return connect.NewResponse(&devicepb.GetConversationContextResponse{
		ConversationId: fmt.Sprintf("%d", info.ConversationID),
		Status:         string(info.Status),
		ClientPhone:    info.ClientPhone,
		Summary:        info.Summary,
		RecentMessages: pbMsgs,
		TotalTokenIn:   info.TotalTokenIn,
		TotalTokenOut:  info.TotalTokenOut,
		TotalLlmCalls:  info.TotalLLMCalls,
		UpdatedAt:      info.UpdatedAt.Format(time.RFC3339),
	}), nil
}

func (h *BotConnectHandler) TakeOverConversation(ctx context.Context, req *connect.Request[devicepb.TakeOverConversationRequest]) (*connect.Response[devicepb.TakeOverConversationResponse], error) {
	if h.convService != nil {
		var convID uint
		_, _ = fmt.Sscanf(req.Msg.Id, "%d", &convID)
		_ = h.convService.TakeOver(ctx, convID, 1)
	}
	return connect.NewResponse(&devicepb.TakeOverConversationResponse{
		Message: "conversation taken over by technician",
	}), nil
}

func (h *BotConnectHandler) ResetConversationBot(ctx context.Context, req *connect.Request[devicepb.ResetConversationBotRequest]) (*connect.Response[devicepb.ResetConversationBotResponse], error) {
	if h.convService != nil {
		var convID uint
		_, _ = fmt.Sscanf(req.Msg.Id, "%d", &convID)
		_ = h.convService.ResetBot(ctx, convID)
	}
	return connect.NewResponse(&devicepb.ResetConversationBotResponse{
		Message: "conversation bot control reset to automatic",
	}), nil
}

func (h *BotConnectHandler) CloseConversation(ctx context.Context, req *connect.Request[devicepb.CloseConversationRequest]) (*connect.Response[devicepb.CloseConversationResponse], error) {
	if h.convService != nil {
		var convID uint
		_, _ = fmt.Sscanf(req.Msg.Id, "%d", &convID)
		_ = h.convService.CloseConversation(ctx, convID)
	}
	return connect.NewResponse(&devicepb.CloseConversationResponse{
		Message: "conversation closed",
	}), nil
}
