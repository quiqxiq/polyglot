package bot

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/pkg/response"
)

// parseID mengonversi ID berformat string dari wire ke uint. Parse gagal
// dianggap 0 (best-effort): ID aplikasi ini selalu numerik, dan 0 akan
// ditolak/diabaikan oleh use case di bawahnya.
func parseID(s string) uint {
	var id uint
	// parse error sengaja diabaikan: string kosong/non-numeric diperlakukan sebagai 0 (tidak valid),
	// validasi selanjutnya di use case yang akan menolak nilai 0.
	_, _ = fmt.Sscanf(s, "%d", &id)
	return id
}

func (h *BotConnectHandler) ListConversations(ctx context.Context, req *connect.Request[polyglotv1.ListConversationsRequest]) (*connect.Response[polyglotv1.ListConversationsResponse], error) {
	if h.convService == nil {
		return connect.NewResponse(&polyglotv1.ListConversationsResponse{Conversations: []*polyglotv1.Conversation{}}), nil
	}

	var convs []bot.Conversation
	var err error
	if req.Msg.SessionId != "" {
		sessionID := parseID(req.Msg.SessionId)
		convs, err = h.convService.ListConversationsBySession(ctx, sessionID)
	} else {
		convs, err = h.convService.ListConversations(ctx, "")
	}
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	pbConvs := make([]*polyglotv1.Conversation, len(convs))
	for i, c := range convs {
		pbConvs[i] = &polyglotv1.Conversation{
			Id:          fmt.Sprintf("%d", c.ID),
			SessionId:   fmt.Sprintf("%d", c.SessionID),
			ClientPhone: c.CustomerWANumber,
			Status:      string(c.Status),
			UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
		}
	}
	return connect.NewResponse(&polyglotv1.ListConversationsResponse{Conversations: pbConvs}), nil
}

func (h *BotConnectHandler) GetConversation(ctx context.Context, req *connect.Request[polyglotv1.GetConversationRequest]) (*connect.Response[polyglotv1.GetConversationResponse], error) {
	if h.convService == nil {
		return connect.NewResponse(&polyglotv1.GetConversationResponse{}), nil
	}

	convID := parseID(req.Msg.Id)

	c, err := h.convService.GetConversationWithMessages(ctx, convID)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	pbMsgs := make([]*polyglotv1.ConversationMessage, len(c.Messages))
	for i, m := range c.Messages {
		pbMsgs[i] = &polyglotv1.ConversationMessage{
			Id:     fmt.Sprintf("%d", m.ID),
			Sender: string(m.SenderType),
			Text:   m.Content,
		}
	}

	return connect.NewResponse(&polyglotv1.GetConversationResponse{
		Conversation: &polyglotv1.Conversation{
			Id:          fmt.Sprintf("%d", c.ID),
			ClientPhone: c.CustomerWANumber,
			Status:      string(c.Status),
			Messages:    pbMsgs,
		},
	}), nil
}

func (h *BotConnectHandler) GetConversationContext(ctx context.Context, req *connect.Request[polyglotv1.GetConversationContextRequest]) (*connect.Response[polyglotv1.GetConversationContextResponse], error) {
	convID := parseID(req.Msg.Id)
	if convID == 0 {
		return nil, response.InvalidArgument("conversation id is required")
	}
	if h.contextProvider == nil {
		return nil, response.Unimplemented("conversation context provider is not available")
	}

	info, err := h.contextProvider.GetConversationContext(ctx, convID)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	pbMsgs := make([]*polyglotv1.ConversationContextMessage, len(info.RecentMessages))
	for i, m := range info.RecentMessages {
		pbMsg := &polyglotv1.ConversationContextMessage{
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

	return connect.NewResponse(&polyglotv1.GetConversationContextResponse{
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

func (h *BotConnectHandler) TakeOverConversation(ctx context.Context, req *connect.Request[polyglotv1.TakeOverConversationRequest]) (*connect.Response[polyglotv1.TakeOverConversationResponse], error) {
	if h.convService != nil {
		convID := parseID(req.Msg.Id)
		// best-effort: takeover tetap dilaporkan sukses meski eksekusi gagal.
		_ = h.convService.TakeOver(ctx, convID, 1)
	}
	return connect.NewResponse(&polyglotv1.TakeOverConversationResponse{
		Message: "conversation taken over by technician",
	}), nil
}

func (h *BotConnectHandler) ResetConversationBot(ctx context.Context, req *connect.Request[polyglotv1.ResetConversationBotRequest]) (*connect.Response[polyglotv1.ResetConversationBotResponse], error) {
	if h.convService != nil {
		convID := parseID(req.Msg.Id)
		// best-effort: reset bot tetap dilaporkan sukses meski eksekusi gagal.
		_ = h.convService.ResetBot(ctx, convID)
	}
	return connect.NewResponse(&polyglotv1.ResetConversationBotResponse{
		Message: "conversation bot control reset to automatic",
	}), nil
}

func (h *BotConnectHandler) CloseConversation(ctx context.Context, req *connect.Request[polyglotv1.CloseConversationRequest]) (*connect.Response[polyglotv1.CloseConversationResponse], error) {
	if h.convService != nil {
		convID := parseID(req.Msg.Id)
		// best-effort: penutupan tetap dilaporkan sukses meski eksekusi gagal.
		_ = h.convService.CloseConversation(ctx, convID)
	}
	return connect.NewResponse(&polyglotv1.CloseConversationResponse{
		Message: "conversation closed",
	}), nil
}

func (h *BotConnectHandler) ResetRateLimit(ctx context.Context, req *connect.Request[polyglotv1.ResetRateLimitRequest]) (*connect.Response[polyglotv1.ResetRateLimitResponse], error) {
	if h.contextProvider != nil {
		if err := h.contextProvider.ResetRateLimit(ctx, req.Msg.PhoneNumber); err != nil {
			return nil, response.MapDomainError(err)
		}
	}

	return connect.NewResponse(&polyglotv1.ResetRateLimitResponse{
		Success: true,
		Message: fmt.Sprintf("rate limit and daily quota reset for %s", req.Msg.PhoneNumber),
	}), nil
}

func (h *BotConnectHandler) GetRateLimitStatus(ctx context.Context, req *connect.Request[polyglotv1.GetRateLimitStatusRequest]) (*connect.Response[polyglotv1.GetRateLimitStatusResponse], error) {
	if h.contextProvider == nil {
		return connect.NewResponse(&polyglotv1.GetRateLimitStatusResponse{
			PhoneNumber: req.Msg.PhoneNumber,
		}), nil
	}

	info, err := h.contextProvider.GetRateLimitStatus(ctx, req.Msg.PhoneNumber)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.GetRateLimitStatusResponse{
		PhoneNumber:     info.PhoneNumber,
		IsMuted:         info.IsMuted,
		DailyChatCount:  int32(info.DailyChatCount),
		DailyQuotaLimit: int32(info.DailyQuotaLimit),
		IsWhitelisted:   info.IsWhitelisted,
	}), nil
}
