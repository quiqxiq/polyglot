package bot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	chatUC "github.com/quixiq/polyglot/internal/usecase/chat"
)

// parseSessionID memvalidasi session_id dari request; kosong diizinkan (0).
func parseSessionID(raw string) (uint, error) {
	if raw == "" {
		return 0, nil
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid session_id %q", raw)
	}
	return uint(id), nil
}

func (h *WhatsAppConnectHandler) ListChats(ctx context.Context, req *connect.Request[devicepb.ListChatsRequest]) (*connect.Response[devicepb.ListChatsResponse], error) {
	if h.chatService == nil {
		return connect.NewResponse(&devicepb.ListChatsResponse{Chats: []*devicepb.WAChat{}}), nil
	}

	sessionID, err := parseSessionID(req.Msg.SessionId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	chats, err := h.chatService.ListChats(sessionID, int(req.Msg.Limit), int(req.Msg.Offset), req.Msg.Search)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbChats := make([]*devicepb.WAChat, len(chats))
	for i, c := range chats {
		lastTime := ""
		if !c.LastMessageTime.IsZero() {
			lastTime = c.LastMessageTime.Format(time.RFC3339)
		}
		pbChats[i] = &devicepb.WAChat{
			Id:                 fmt.Sprintf("%d", c.ID),
			SessionId:          fmt.Sprintf("%d", c.SessionID),
			ChatJid:            c.ChatJID,
			DisplayName:        c.DisplayName,
			IsGroup:            c.IsGroup,
			LastMessagePreview: c.LastMessagePreview,
			LastMessageTime:    lastTime,
			UnreadCount:        int32(c.UnreadCount),
			BotEnabled:         c.BotEnabled,
		}
	}

	// Total = ukuran halaman saat ini (belum ada COUNT query untuk total sebenarnya).
	return connect.NewResponse(&devicepb.ListChatsResponse{Chats: pbChats, Total: int32(len(pbChats))}), nil
}

func (h *WhatsAppConnectHandler) GetChatMessages(ctx context.Context, req *connect.Request[devicepb.GetChatMessagesRequest]) (*connect.Response[devicepb.GetChatMessagesResponse], error) {
	if h.chatService == nil {
		return connect.NewResponse(&devicepb.GetChatMessagesResponse{Messages: []*devicepb.WAChatMessage{}}), nil
	}

	sessionID, err := parseSessionID(req.Msg.SessionId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	msgs, err := h.chatService.GetChatMessages(sessionID, req.Msg.ChatJid, int(req.Msg.Limit), int(req.Msg.Offset))
	if err != nil {
		if errors.Is(err, chatUC.ErrEmptyChatJID) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbMsgs := make([]*devicepb.WAChatMessage, len(msgs))
	for i, m := range msgs {
		ts := ""
		if !m.Timestamp.IsZero() {
			ts = m.Timestamp.Format(time.RFC3339)
		}
		pbMsgs[i] = &devicepb.WAChatMessage{
			Id:         fmt.Sprintf("%d", m.ID),
			ChatJid:    m.ChatJID,
			SenderJid:  m.SenderJID,
			SenderName: m.SenderName,
			Content:    m.Content,
			MediaType:  m.MediaType,
			IsFromMe:   m.IsFromMe,
			IsRead:     m.IsRead,
			Status:     m.Status,
			Timestamp:  ts,
		}
	}

	return connect.NewResponse(&devicepb.GetChatMessagesResponse{Messages: pbMsgs, Total: int32(len(pbMsgs))}), nil
}

func (h *WhatsAppConnectHandler) ToggleChatBot(ctx context.Context, req *connect.Request[devicepb.ToggleChatBotRequest]) (*connect.Response[devicepb.ToggleChatBotResponse], error) {
	if h.chatService == nil {
		return connect.NewResponse(&devicepb.ToggleChatBotResponse{Message: "chat service not available", IsActive: req.Msg.IsActive}), nil
	}

	sessionID, err := parseSessionID(req.Msg.SessionId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if err := h.chatService.ToggleChatBot(sessionID, req.Msg.ChatJid, req.Msg.IsActive); err != nil {
		if errors.Is(err, chatUC.ErrEmptyChatJID) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&devicepb.ToggleChatBotResponse{
		Message:  "chat bot toggled",
		IsActive: req.Msg.IsActive,
	}), nil
}

func (h *WhatsAppConnectHandler) MarkChatRead(ctx context.Context, req *connect.Request[devicepb.MarkChatReadRequest]) (*connect.Response[devicepb.MarkChatReadResponse], error) {
	if h.chatService == nil {
		return connect.NewResponse(&devicepb.MarkChatReadResponse{Message: "chat marked as read"}), nil
	}

	sessionID, err := parseSessionID(req.Msg.SessionId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if err := h.chatService.MarkRead(sessionID, req.Msg.ChatJid); err != nil {
		if errors.Is(err, chatUC.ErrEmptyChatJID) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&devicepb.MarkChatReadResponse{Message: "chat marked as read"}), nil
}
