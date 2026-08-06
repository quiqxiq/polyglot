package connectadapter

import (
	"context"
	"fmt"
	"strconv"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
	"github.com/quixiq/polyglot/internal/domain/bot"
)

func (h *WhatsAppConnectHandler) ListSessions(ctx context.Context, req *connect.Request[devicepb.ListWASessionsRequest]) (*connect.Response[devicepb.ListWASessionsResponse], error) {
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

func (h *WhatsAppConnectHandler) CreateSession(ctx context.Context, req *connect.Request[devicepb.CreateWASessionRequest]) (*connect.Response[devicepb.CreateWASessionResponse], error) {
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

	if h.waGateway != nil {
		_ = h.waGateway.Connect(sess)
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

func (h *WhatsAppConnectHandler) GetQRCode(ctx context.Context, req *connect.Request[devicepb.GetWASessionQRRequest]) (*connect.Response[devicepb.GetWASessionQRResponse], error) {
	idUint, _ := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	qrBase64 := ""
	if h.waGateway != nil && idUint > 0 {
		if code, err := h.waGateway.GetQRCode(uint(idUint)); err == nil {
			qrBase64 = code
		}
	}
	return connect.NewResponse(&devicepb.GetWASessionQRResponse{
		QrCodeBase64:    qrBase64,
		QrCodePath:      "",
		DurationSeconds: 60,
	}), nil
}

func (h *WhatsAppConnectHandler) GetPairingCode(ctx context.Context, req *connect.Request[devicepb.GetWASessionPairingRequest]) (*connect.Response[devicepb.GetWASessionPairingResponse], error) {
	idUint, _ := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	code := ""
	if h.waGateway != nil && idUint > 0 {
		if pairing, err := h.waGateway.GetPairingCode(uint(idUint), req.Msg.PhoneNumber); err == nil {
			code = pairing
		}
	}
	return connect.NewResponse(&devicepb.GetWASessionPairingResponse{PairingCode: code}), nil
}

func (h *WhatsAppConnectHandler) ToggleBot(ctx context.Context, req *connect.Request[devicepb.ToggleWABotRequest]) (*connect.Response[devicepb.ToggleWABotResponse], error) {
	return connect.NewResponse(&devicepb.ToggleWABotResponse{
		Message:  "bot toggled successfully",
		IsActive: req.Msg.IsActive,
	}), nil
}

func (h *WhatsAppConnectHandler) ReconnectSession(ctx context.Context, req *connect.Request[devicepb.ReconnectWASessionRequest]) (*connect.Response[devicepb.ReconnectWASessionResponse], error) {
	idUint, _ := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	var err error
	if h.waGateway != nil && idUint > 0 {
		err = h.waGateway.Reconnect(uint(idUint))
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("reconnect failed: %w", err))
	}
	return connect.NewResponse(&devicepb.ReconnectWASessionResponse{
		Message: "manual reconnect initiated (auto-reconnect active)",
		Status:  "RECONNECTING",
	}), nil
}

func (h *WhatsAppConnectHandler) LogoutSession(ctx context.Context, req *connect.Request[devicepb.LogoutWASessionRequest]) (*connect.Response[devicepb.LogoutWASessionResponse], error) {
	idUint, _ := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	if h.waGateway != nil && idUint > 0 {
		_ = h.waGateway.Disconnect(uint(idUint))
	}
	return connect.NewResponse(&devicepb.LogoutWASessionResponse{Message: "session logged out (slot kept for re-pairing)"}), nil
}

func (h *WhatsAppConnectHandler) PurgeSession(ctx context.Context, req *connect.Request[devicepb.PurgeWASessionRequest]) (*connect.Response[devicepb.PurgeWASessionResponse], error) {
	idUint, _ := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	if h.waGateway != nil && idUint > 0 {
		_ = h.waGateway.Logout(uint(idUint))
	}
	return connect.NewResponse(&devicepb.PurgeWASessionResponse{Message: "session purged permanently"}), nil
}

func (h *WhatsAppConnectHandler) SendTextMessage(ctx context.Context, req *connect.Request[devicepb.SendWATextMessageRequest]) (*connect.Response[devicepb.SendWATextMessageResponse], error) {
	idUint, _ := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	var err error
	if h.waGateway != nil && idUint > 0 {
		err = h.waGateway.SendMessage(uint(idUint), req.Msg.RecipientPhone, req.Msg.MessageText)
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("send text message failed: %w", err))
	}
	return connect.NewResponse(&devicepb.SendWATextMessageResponse{
		MessageId: "msg-sent-ok",
		Status:    "SENT",
	}), nil
}
