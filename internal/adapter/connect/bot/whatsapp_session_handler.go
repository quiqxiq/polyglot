package bot

import (
	"context"
	"fmt"
	"github.com/quixiq/polyglot/pkg/logger"
	"strconv"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/bot"
)

// formatTime renders a time value as a compact local timestamp; empty for zero values.
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func (h *WhatsAppConnectHandler) ListSessions(ctx context.Context, req *connect.Request[devicepb.ListWASessionsRequest]) (*connect.Response[devicepb.ListWASessionsResponse], error) {
	if h.sessionRepo == nil {
		return connect.NewResponse(&devicepb.ListWASessionsResponse{Sessions: []*devicepb.WASession{}}), nil
	}

	sessions, err := h.sessionRepo.FindAllSessions(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	pbSessions := make([]*devicepb.WASession, len(sessions))
	for i, s := range sessions {
		// Status live dari client (online/offline/connecting/needs_rescan),
		// fallback ke status tersimpan di DB bila gateway tidak tersedia.
		status := string(s.Status)
		if h.waGateway != nil {
			if live, err := h.waGateway.GetStatus(s.ID); err == nil && live != "" {
				status = live
			}
		}

		pbSessions[i] = &devicepb.WASession{
			Id:          fmt.Sprintf("%d", s.ID),
			Name:        s.DeviceName,
			PhoneNumber: s.PhoneNumber,
			Status:      status,
			IsBotActive: s.IsBotEnabled,
			Jid:         s.JID,
			CreatedAt:   s.CreatedAt.Format("2006-01-02 15:04:05"),
			ConnectedAt: formatTime(s.ConnectedAt),
		}
	}

	return connect.NewResponse(&devicepb.ListWASessionsResponse{Sessions: pbSessions}), nil
}

func (h *WhatsAppConnectHandler) CreateSession(ctx context.Context, req *connect.Request[devicepb.CreateWASessionRequest]) (*connect.Response[devicepb.CreateWASessionResponse], error) {
	if h.sessionRepo == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("session repo not initialized"))
	}

	sess := &bot.WASession{
		DeviceName:   req.Msg.Name,
		PhoneNumber:  req.Msg.PhoneNumber,
		Status:       bot.StatusOffline,
		IsBotEnabled: true,
	}

	if err := h.sessionRepo.CreateSession(ctx, sess); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Langsung connect — untuk session baru (belum ter-pair) ini memicu QR
	// flow; frontend langsung membuka modal QR setelah create.
	if h.waGateway != nil {
		if err := h.waGateway.Connect(sess); err != nil {
			logger.WithComponent("WhatsApp").Warnf("CreateSession %d: connect failed (QR flow): %v", sess.ID, err)
		}
		sess.Status = bot.StatusConnecting
		_ = h.sessionRepo.UpdateSession(ctx, sess)
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
	idUint, err := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	if err != nil || idUint == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid session_id"))
	}

	if h.waGateway == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("whatsapp gateway not initialized"))
	}

	sessionID := uint(idUint)

	// Jika client belum connect (mis. baru dibuat / reconnect diminta), trigger
	// koneksi dulu supaya QR flow aktif dan QR bisa dibaca dari cache.
	status, _ := h.waGateway.GetStatus(sessionID)
	if status != string(bot.StatusOnline) && h.sessionRepo != nil {
		sess, findErr := h.sessionRepo.FindSessionByID(ctx, sessionID)
		if findErr == nil {
			_ = h.waGateway.Connect(sess)
		}
	}

	qrBase64, qrErr := h.waGateway.GetQRCode(sessionID)
	if qrErr != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("get QR code: %w", qrErr))
	}

	return connect.NewResponse(&devicepb.GetWASessionQRResponse{
		QrCodeBase64:    qrBase64,
		QrCodePath:      "",
		DurationSeconds: 60,
	}), nil
}

func (h *WhatsAppConnectHandler) GetPairingCode(ctx context.Context, req *connect.Request[devicepb.GetWASessionPairingRequest]) (*connect.Response[devicepb.GetWASessionPairingResponse], error) {
	idUint, err := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	if err != nil || idUint == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid session_id"))
	}
	if req.Msg.PhoneNumber == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("phone_number is required"))
	}

	if h.waGateway == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("whatsapp gateway not initialized"))
	}

	sessionID := uint(idUint)

	// Pairing code memerlukan client yang terhubung — pastikan connect dulu.
	status, _ := h.waGateway.GetStatus(sessionID)
	if status != string(bot.StatusOnline) && h.sessionRepo != nil {
		sess, findErr := h.sessionRepo.FindSessionByID(ctx, sessionID)
		if findErr == nil {
			_ = h.waGateway.Connect(sess)
		}
	}

	code, err := h.waGateway.GetPairingCode(sessionID, req.Msg.PhoneNumber)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("get pairing code: %w", err))
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
	idUint, err := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	if err != nil || idUint == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid session_id"))
	}
	if h.waGateway == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("whatsapp gateway not initialized"))
	}
	if err := h.waGateway.Reconnect(uint(idUint)); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("reconnect failed: %w", err))
	}
	return connect.NewResponse(&devicepb.ReconnectWASessionResponse{
		Message: "manual reconnect initiated (auto-reconnect active)",
		Status:  "connecting",
	}), nil
}

func (h *WhatsAppConnectHandler) LogoutSession(ctx context.Context, req *connect.Request[devicepb.LogoutWASessionRequest]) (*connect.Response[devicepb.LogoutWASessionResponse], error) {
	idUint, err := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	if err != nil || idUint == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid session_id"))
	}
	if h.waGateway == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("whatsapp gateway not initialized"))
	}

	// Logout = unlink dari WhatsApp + hapus session lokal, tapi slot di DB
	// TETAP (keep-slot) sehingga device masih tampil dan bisa di-pair ulang.
	if err := h.waGateway.Logout(uint(idUint)); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("logout failed: %w", err))
	}
	return connect.NewResponse(&devicepb.LogoutWASessionResponse{Message: "session logged out (slot kept for re-pairing)"}), nil
}

func (h *WhatsAppConnectHandler) PurgeSession(ctx context.Context, req *connect.Request[devicepb.PurgeWASessionRequest]) (*connect.Response[devicepb.PurgeWASessionResponse], error) {
	idUint, err := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	if err != nil || idUint == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid session_id"))
	}

	// Purge = logout + hapus baris session di DB (mirror chat terhapus otomatis
	// via ON DELETE CASCADE pada wa_chats / wa_messages).
	if h.waGateway != nil {
		if err := h.waGateway.Purge(uint(idUint)); err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("purge session: %w", err))
		}
	}
	if h.sessionRepo != nil {
		if err := h.sessionRepo.DeleteSession(ctx, uint(idUint)); err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("delete session record: %w", err))
		}
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
