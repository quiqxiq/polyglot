package connectadapter

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
	"github.com/quixiq/polyglot/internal/adapter/connect/codec"
	"github.com/quixiq/polyglot/internal/adapter/connect/mapper"
	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/port"
)

type WhatsAppConnectHandler struct {
	repo      port.WASessionRepository
	waGateway port.WhatsAppGateway
}

func NewWhatsAppConnectHandler(repo port.WASessionRepository, waGateway port.WhatsAppGateway) *WhatsAppConnectHandler {
	return &WhatsAppConnectHandler{
		repo:      repo,
		waGateway: waGateway,
	}
}

func (h *WhatsAppConnectHandler) ListSessions(ctx context.Context, req *connect.Request[devicepb.ListWASessionsRequest]) (*connect.Response[devicepb.ListWASessionsResponse], error) {
	if h.repo == nil {
		return connect.NewResponse(&devicepb.ListWASessionsResponse{Sessions: []*devicepb.WASession{}}), nil
	}

	sessions, err := h.repo.FindAll()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&devicepb.ListWASessionsResponse{
		Sessions: mapper.WASessionListToProto(sessions),
	}), nil
}

func (h *WhatsAppConnectHandler) CreateSession(ctx context.Context, req *connect.Request[devicepb.CreateWASessionRequest]) (*connect.Response[devicepb.CreateWASessionResponse], error) {
	if req.Msg.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("session name is required"))
	}

	sess := &bot.WASession{
		DeviceName:   req.Msg.Name,
		PhoneNumber:  req.Msg.PhoneNumber,
		Status:       bot.StatusNeedsRescan,
		IsBotEnabled: true,
	}

	if h.repo != nil {
		if err := h.repo.Create(sess); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	if h.waGateway != nil {
		_ = h.waGateway.Connect(sess)
	}

	return connect.NewResponse(&devicepb.CreateWASessionResponse{
		Session: mapper.WASessionToProto(*sess),
	}), nil
}

func (h *WhatsAppConnectHandler) GetQRCode(ctx context.Context, req *connect.Request[devicepb.GetWASessionQRRequest]) (*connect.Response[devicepb.GetWASessionQRResponse], error) {
	idUint, _ := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	qrBase64 := ""
	if h.waGateway != nil && idUint > 0 {
		qr, err := h.waGateway.GetQRCode(uint(idUint))
		if err == nil {
			qrBase64 = qr
		}
	}

	return connect.NewResponse(&devicepb.GetWASessionQRResponse{
		QrCodeBase64:    qrBase64,
		DurationSeconds: 60,
	}), nil
}

func (h *WhatsAppConnectHandler) GetPairingCode(ctx context.Context, req *connect.Request[devicepb.GetWASessionPairingRequest]) (*connect.Response[devicepb.GetWASessionPairingResponse], error) {
	idUint, _ := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	code := ""
	if h.waGateway != nil && idUint > 0 {
		c, err := h.waGateway.GetPairingCode(uint(idUint), req.Msg.PhoneNumber)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		code = c
	}

	return connect.NewResponse(&devicepb.GetWASessionPairingResponse{
		PairingCode: code,
	}), nil
}

func (h *WhatsAppConnectHandler) ToggleBot(ctx context.Context, req *connect.Request[devicepb.ToggleWABotRequest]) (*connect.Response[devicepb.ToggleWABotResponse], error) {
	idUint, _ := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	if h.repo != nil && idUint > 0 {
		sess, err := h.repo.FindByID(uint(idUint))
		if err == nil && sess != nil {
			sess.IsBotEnabled = req.Msg.IsActive
			_ = h.repo.Update(sess)
		}
	}

	return connect.NewResponse(&devicepb.ToggleWABotResponse{
		Message:  "bot status toggled successfully",
		IsActive: req.Msg.IsActive,
	}), nil
}

func (h *WhatsAppConnectHandler) ReconnectSession(ctx context.Context, req *connect.Request[devicepb.ReconnectWASessionRequest]) (*connect.Response[devicepb.ReconnectWASessionResponse], error) {
	idUint, _ := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	if h.waGateway != nil && idUint > 0 {
		_ = h.waGateway.Reconnect(uint(idUint))
	}
	return connect.NewResponse(&devicepb.ReconnectWASessionResponse{
		Message: "reconnect initiated",
		Status:  "reconnecting",
	}), nil
}

func (h *WhatsAppConnectHandler) LogoutSession(ctx context.Context, req *connect.Request[devicepb.LogoutWASessionRequest]) (*connect.Response[devicepb.LogoutWASessionResponse], error) {
	idUint, _ := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	if h.waGateway != nil && idUint > 0 {
		_ = h.waGateway.Logout(uint(idUint))
	}
	return connect.NewResponse(&devicepb.LogoutWASessionResponse{
		Message: "session logged out successfully",
	}), nil
}

func (h *WhatsAppConnectHandler) PurgeSession(ctx context.Context, req *connect.Request[devicepb.PurgeWASessionRequest]) (*connect.Response[devicepb.PurgeWASessionResponse], error) {
	idUint, _ := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	if h.repo != nil && idUint > 0 {
		_ = h.repo.Delete(uint(idUint))
	}
	return connect.NewResponse(&devicepb.PurgeWASessionResponse{
		Message: "session purged successfully",
	}), nil
}

func (h *WhatsAppConnectHandler) SendTextMessage(ctx context.Context, req *connect.Request[devicepb.SendWATextMessageRequest]) (*connect.Response[devicepb.SendWATextMessageResponse], error) {
	idUint, _ := strconv.ParseUint(req.Msg.SessionId, 10, 64)
	if req.Msg.RecipientPhone == "" || req.Msg.MessageText == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("recipient and message are required"))
	}

	if h.waGateway != nil && idUint > 0 {
		err := h.waGateway.SendMessage(uint(idUint), req.Msg.RecipientPhone, req.Msg.MessageText)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	return connect.NewResponse(&devicepb.SendWATextMessageResponse{
		MessageId: fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Status:    "sent",
	}), nil
}

func NewWhatsAppServiceHandler(repo port.WASessionRepository, waGateway port.WhatsAppGateway) (string, http.Handler) {
	handler := NewWhatsAppConnectHandler(repo, waGateway)
	mux := http.NewServeMux()
	codecOpt := codec.Option()

	serviceName := "polyglot.v1.WhatsAppService"
	mux.Handle("/"+serviceName+"/ListSessions", connect.NewUnaryHandler("/"+serviceName+"/ListSessions", handler.ListSessions, codecOpt))
	mux.Handle("/"+serviceName+"/CreateSession", connect.NewUnaryHandler("/"+serviceName+"/CreateSession", handler.CreateSession, codecOpt))
	mux.Handle("/"+serviceName+"/GetQRCode", connect.NewUnaryHandler("/"+serviceName+"/GetQRCode", handler.GetQRCode, codecOpt))
	mux.Handle("/"+serviceName+"/GetPairingCode", connect.NewUnaryHandler("/"+serviceName+"/GetPairingCode", handler.GetPairingCode, codecOpt))
	mux.Handle("/"+serviceName+"/ToggleBot", connect.NewUnaryHandler("/"+serviceName+"/ToggleBot", handler.ToggleBot, codecOpt))
	mux.Handle("/"+serviceName+"/ReconnectSession", connect.NewUnaryHandler("/"+serviceName+"/ReconnectSession", handler.ReconnectSession, codecOpt))
	mux.Handle("/"+serviceName+"/LogoutSession", connect.NewUnaryHandler("/"+serviceName+"/LogoutSession", handler.LogoutSession, codecOpt))
	mux.Handle("/"+serviceName+"/PurgeSession", connect.NewUnaryHandler("/"+serviceName+"/PurgeSession", handler.PurgeSession, codecOpt))
	mux.Handle("/"+serviceName+"/SendTextMessage", connect.NewUnaryHandler("/"+serviceName+"/SendTextMessage", handler.SendTextMessage, codecOpt))

	return "/" + serviceName + "/", mux
}
