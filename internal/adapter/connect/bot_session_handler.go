package connectadapter

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
	"github.com/quixiq/polyglot/internal/domain/bot"
)

func (h *BotConnectHandler) ListSessions(ctx context.Context, req *connect.Request[devicepb.ListWASessionsRequest]) (*connect.Response[devicepb.ListWASessionsResponse], error) {
	if h.pgStore == nil {
		return connect.NewResponse(&devicepb.ListWASessionsResponse{Sessions: []*devicepb.WASession{}}), nil
	}

	sessions, err := h.pgStore.FindAllSessions()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbSessions := make([]*devicepb.WASession, len(sessions))
	for i, s := range sessions {
		pbSessions[i] = &devicepb.WASession{
			Id:          fmt.Sprintf("%d", s.ID),
			Name:        s.DeviceName,
			PhoneNumber: s.PhoneNumber,
			Status:      string(s.Status),
			IsBotActive: s.IsBotEnabled,
			WebhookUrl:  s.WebhookURL,
			CreatedAt:   s.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return connect.NewResponse(&devicepb.ListWASessionsResponse{Sessions: pbSessions}), nil
}

func (h *BotConnectHandler) CreateSession(ctx context.Context, req *connect.Request[devicepb.CreateWASessionRequest]) (*connect.Response[devicepb.CreateWASessionResponse], error) {
	if h.pgStore == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("postgres store not initialized"))
	}

	sess := &bot.WASession{
		DeviceName:   req.Msg.Name,
		PhoneNumber:  req.Msg.PhoneNumber,
		Status:       bot.StatusOffline,
		IsBotEnabled: true,
	}

	if err := h.pgStore.CreateSession(sess); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&devicepb.CreateWASessionResponse{
		Session: &devicepb.WASession{
			Id:          fmt.Sprintf("%d", sess.ID),
			Name:        sess.DeviceName,
			PhoneNumber: sess.PhoneNumber,
			Status:      string(sess.Status),
			IsBotActive: sess.IsBotEnabled,
			CreatedAt:   sess.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}), nil
}

func (h *BotConnectHandler) GetQRCode(ctx context.Context, req *connect.Request[devicepb.GetWASessionQRRequest]) (*connect.Response[devicepb.GetWASessionQRResponse], error) {
	return connect.NewResponse(&devicepb.GetWASessionQRResponse{QrCodeBase64: ""}), nil
}

func (h *BotConnectHandler) GetPairingCode(ctx context.Context, req *connect.Request[devicepb.GetWASessionPairingRequest]) (*connect.Response[devicepb.GetWASessionPairingResponse], error) {
	return connect.NewResponse(&devicepb.GetWASessionPairingResponse{PairingCode: ""}), nil
}

func (h *BotConnectHandler) ToggleBot(ctx context.Context, req *connect.Request[devicepb.ToggleWABotRequest]) (*connect.Response[devicepb.ToggleWABotResponse], error) {
	return connect.NewResponse(&devicepb.ToggleWABotResponse{
		Message:  "bot toggled successfully",
		IsActive: req.Msg.IsActive,
	}), nil
}

func (h *BotConnectHandler) LogoutSession(ctx context.Context, req *connect.Request[devicepb.LogoutWASessionRequest]) (*connect.Response[devicepb.LogoutWASessionResponse], error) {
	return connect.NewResponse(&devicepb.LogoutWASessionResponse{Message: "logged out"}), nil
}

func (h *BotConnectHandler) SendTextMessage(ctx context.Context, req *connect.Request[devicepb.SendWATextMessageRequest]) (*connect.Response[devicepb.SendWATextMessageResponse], error) {
	return connect.NewResponse(&devicepb.SendWATextMessageResponse{
		MessageId: "msg-001",
		Status:    "SENT",
	}), nil
}
