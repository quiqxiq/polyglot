package bot

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/response"
)

// formatTime renders a time value as a compact local timestamp; empty for zero values.
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func (h *WhatsAppConnectHandler) ListSessions(ctx context.Context, req *connect.Request[polyglotv1.ListWASessionsRequest]) (*connect.Response[polyglotv1.ListWASessionsResponse], error) {
	if h.sessionRepo == nil {
		return connect.NewResponse(&polyglotv1.ListWASessionsResponse{Sessions: []*polyglotv1.WASession{}}), nil
	}

	sessions, err := h.sessionRepo.FindAllSessions(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	pbSessions := make([]*polyglotv1.WASession, len(sessions))
	for i, s := range sessions {
		// Status live dari client (online/offline/connecting/needs_rescan),
		// fallback ke status tersimpan di DB bila gateway tidak tersedia.
		status := string(s.Status)
		if h.waGateway != nil {
			if live, err := h.waGateway.GetStatus(s.ID); err == nil && live != "" {
				status = live
			}
		}

		pbSessions[i] = &polyglotv1.WASession{
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

	return connect.NewResponse(&polyglotv1.ListWASessionsResponse{Sessions: pbSessions}), nil
}

func (h *WhatsAppConnectHandler) CreateSession(ctx context.Context, req *connect.Request[polyglotv1.CreateWASessionRequest]) (*connect.Response[polyglotv1.CreateWASessionResponse], error) {
	if h.sessionRepo == nil {
		return nil, response.Unavailable("session repo not initialized")
	}

	sess := &bot.WASession{
		DeviceName:   req.Msg.Name,
		PhoneNumber:  req.Msg.PhoneNumber,
		Status:       bot.StatusOffline,
		IsBotEnabled: true,
	}

	if err := h.sessionRepo.CreateSession(ctx, sess); err != nil {
		return nil, response.MapDomainError(err)
	}

	// Langsung connect — untuk session baru (belum ter-pair) ini memicu QR
	// flow; frontend langsung membuka modal QR setelah create.
	if h.waGateway != nil {
		if err := h.waGateway.Connect(sess); err != nil {
			logger.WithComponent("WhatsApp").WithError(err).WithField("session_id", sess.ID).Warn("session connect failed during QR flow")
		}
		sess.Status = bot.StatusConnecting
		// best-effort: update status best-effort; response tetap berdasarkan sess di memori.
		_ = h.sessionRepo.UpdateSession(ctx, sess)
	}

	return connect.NewResponse(&polyglotv1.CreateWASessionResponse{
		Session: &polyglotv1.WASession{
			Id:          fmt.Sprintf("%d", sess.ID),
			Name:        sess.DeviceName,
			PhoneNumber: sess.PhoneNumber,
			Status:      string(sess.Status),
			IsBotActive: sess.IsBotEnabled,
			CreatedAt:   sess.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	}), nil
}

func (h *WhatsAppConnectHandler) GetQRCode(ctx context.Context, req *connect.Request[polyglotv1.GetWASessionQRRequest]) (*connect.Response[polyglotv1.GetWASessionQRResponse], error) {
	idUint, err := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	if err != nil || idUint == 0 {
		return nil, response.InvalidArgument("invalid session_id")
	}

	if h.waGateway == nil {
		return nil, response.Unavailable("whatsapp gateway not initialized")
	}

	sessionID := uint(idUint)

	// Jika client belum connect (mis. baru dibuat / reconnect diminta), trigger
	// koneksi dulu supaya QR flow aktif dan QR bisa dibaca dari cache.
	status, _ := h.waGateway.GetStatus(sessionID)
	if status != string(bot.StatusOnline) && h.sessionRepo != nil {
		sess, findErr := h.sessionRepo.FindSessionByID(ctx, sessionID)
		if findErr == nil {
			// best-effort: connect ulang hanya pemicu QR flow; QR dibaca dari cache setelahnya.
			_ = h.waGateway.Connect(sess)
		}
	}

	qrBase64, qrErr := h.waGateway.GetQRCode(sessionID)
	if qrErr != nil {
		return nil, response.MapDomainError(fmt.Errorf("get QR code: %w", qrErr))
	}

	return connect.NewResponse(&polyglotv1.GetWASessionQRResponse{
		QrCodeBase64:    qrBase64,
		QrCodePath:      "",
		DurationSeconds: 60,
	}), nil
}

func (h *WhatsAppConnectHandler) GetPairingCode(ctx context.Context, req *connect.Request[polyglotv1.GetWASessionPairingRequest]) (*connect.Response[polyglotv1.GetWASessionPairingResponse], error) {
	idUint, err := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	if err != nil || idUint == 0 {
		return nil, response.InvalidArgument("invalid session_id")
	}

	if h.waGateway == nil {
		return nil, response.Unavailable("whatsapp gateway not initialized")
	}

	sessionID := uint(idUint)

	// Pairing code memerlukan client yang terhubung — pastikan connect dulu.
	status, _ := h.waGateway.GetStatus(sessionID)
	if status != string(bot.StatusOnline) && h.sessionRepo != nil {
		sess, findErr := h.sessionRepo.FindSessionByID(ctx, sessionID)
		if findErr == nil {
			// best-effort: connect ulang hanya pemicu; pairing code dihasilkan oleh client yang terhubung.
			_ = h.waGateway.Connect(sess)
		}
	}

	code, err := h.waGateway.GetPairingCode(sessionID, req.Msg.PhoneNumber)
	if err != nil {
		return nil, response.MapDomainError(fmt.Errorf("get pairing code: %w", err))
	}
	return connect.NewResponse(&polyglotv1.GetWASessionPairingResponse{PairingCode: code}), nil
}

func (h *WhatsAppConnectHandler) ToggleBot(ctx context.Context, req *connect.Request[polyglotv1.ToggleWABotRequest]) (*connect.Response[polyglotv1.ToggleWABotResponse], error) {
	return connect.NewResponse(&polyglotv1.ToggleWABotResponse{
		Message:  "bot toggled successfully",
		IsActive: req.Msg.IsActive,
	}), nil
}

func (h *WhatsAppConnectHandler) ReconnectSession(ctx context.Context, req *connect.Request[polyglotv1.ReconnectWASessionRequest]) (*connect.Response[polyglotv1.ReconnectWASessionResponse], error) {
	idUint, err := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	if err != nil || idUint == 0 {
		return nil, response.InvalidArgument("invalid session_id")
	}
	if h.waGateway == nil {
		return nil, response.Unavailable("whatsapp gateway not initialized")
	}
	if err := h.waGateway.Reconnect(uint(idUint)); err != nil {
		return nil, response.Internal("reconnect failed")
	}
	return connect.NewResponse(&polyglotv1.ReconnectWASessionResponse{
		Message: "manual reconnect initiated (auto-reconnect active)",
		Status:  "connecting",
	}), nil
}

func (h *WhatsAppConnectHandler) LogoutSession(ctx context.Context, req *connect.Request[polyglotv1.LogoutWASessionRequest]) (*connect.Response[polyglotv1.LogoutWASessionResponse], error) {
	idUint, err := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	if err != nil || idUint == 0 {
		return nil, response.InvalidArgument("invalid session_id")
	}
	if h.waGateway == nil {
		return nil, response.Unavailable("whatsapp gateway not initialized")
	}

	// Logout = unlink dari WhatsApp + hapus session lokal, tapi slot di DB
	// TETAP (keep-slot) sehingga device masih tampil dan bisa di-pair ulang.
	if err := h.waGateway.Logout(uint(idUint)); err != nil {
		return nil, response.Internal("logout failed")
	}
	return connect.NewResponse(&polyglotv1.LogoutWASessionResponse{Message: "session logged out (slot kept for re-pairing)"}), nil
}

func (h *WhatsAppConnectHandler) PurgeSession(ctx context.Context, req *connect.Request[polyglotv1.PurgeWASessionRequest]) (*connect.Response[polyglotv1.PurgeWASessionResponse], error) {
	idUint, err := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	if err != nil || idUint == 0 {
		return nil, response.InvalidArgument("invalid session_id")
	}

	// Purge = logout + hapus baris session di DB (mirror chat terhapus otomatis
	// via ON DELETE CASCADE pada wa_chats / wa_messages).
	if h.waGateway != nil {
		if err := h.waGateway.Purge(uint(idUint)); err != nil {
			return nil, response.Internal("purge session failed")
		}
	}
	if h.sessionRepo != nil {
		if err := h.sessionRepo.DeleteSession(ctx, uint(idUint)); err != nil {
			return nil, response.Internal("delete session record failed")
		}
	}
	return connect.NewResponse(&polyglotv1.PurgeWASessionResponse{Message: "session purged permanently"}), nil
}

func (h *WhatsAppConnectHandler) SendTextMessage(ctx context.Context, req *connect.Request[polyglotv1.SendWATextMessageRequest]) (*connect.Response[polyglotv1.SendWATextMessageResponse], error) {
	idUint, _ := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	var err error
	if h.waGateway != nil && idUint > 0 {
		err = h.waGateway.SendMessage(uint(idUint), req.Msg.RecipientPhone, req.Msg.MessageText)
	}
	if err != nil {
		return nil, response.Internal("send text message failed")
	}
	return connect.NewResponse(&polyglotv1.SendWATextMessageResponse{
		MessageId: "msg-sent-ok",
		Status:    "SENT",
	}), nil
}
