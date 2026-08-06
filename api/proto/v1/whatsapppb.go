package devicepb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WASession struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Status      string `json:"status"`
	IsBotActive bool   `json:"is_bot_active"`
	WebhookUrl  string `json:"webhook_url"`
	CreatedAt   string `json:"created_at"`
}

type ListWASessionsRequest struct{}
type ListWASessionsResponse struct {
	Sessions []*WASession `json:"sessions"`
}

type CreateWASessionRequest struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
}
type CreateWASessionResponse struct {
	Session *WASession `json:"session"`
}

type GetWASessionQRRequest struct {
	SessionId string `json:"session_id"`
}
type GetWASessionQRResponse struct {
	QrCodeBase64    string `json:"qr_code_base64"`
	QrCodePath      string `json:"qr_code_path"`
	DurationSeconds int32  `json:"duration_seconds"`
}

type GetWASessionPairingRequest struct {
	SessionId   string `json:"session_id"`
	PhoneNumber string `json:"phone_number"`
}
type GetWASessionPairingResponse struct {
	PairingCode string `json:"pairing_code"`
}

type ToggleWABotRequest struct {
	SessionId string `json:"session_id"`
	IsActive  bool   `json:"is_active"`
}
type ToggleWABotResponse struct {
	Message  string `json:"message"`
	IsActive bool   `json:"is_active"`
}

type ReconnectWASessionRequest struct {
	SessionId string `json:"session_id"`
}
type ReconnectWASessionResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

type LogoutWASessionRequest struct {
	SessionId string `json:"session_id"`
}
type LogoutWASessionResponse struct {
	Message string `json:"message"`
}

type PurgeWASessionRequest struct {
	SessionId string `json:"session_id"`
}
type PurgeWASessionResponse struct {
	Message string `json:"message"`
}

type SendWATextMessageRequest struct {
	SessionId      string `json:"session_id"`
	RecipientPhone string `json:"recipient_phone"`
	MessageText    string `json:"message_text"`
}
type SendWATextMessageResponse struct {
	MessageId string `json:"message_id"`
	Status    string `json:"status"`
}

type WhatsAppServiceClient interface {
	ListSessions(ctx context.Context, in *ListWASessionsRequest, opts ...grpc.CallOption) (*ListWASessionsResponse, error)
	CreateSession(ctx context.Context, in *CreateWASessionRequest, opts ...grpc.CallOption) (*CreateWASessionResponse, error)
	GetQRCode(ctx context.Context, in *GetWASessionQRRequest, opts ...grpc.CallOption) (*GetWASessionQRResponse, error)
	GetPairingCode(ctx context.Context, in *GetWASessionPairingRequest, opts ...grpc.CallOption) (*GetWASessionPairingResponse, error)
	ToggleBot(ctx context.Context, in *ToggleWABotRequest, opts ...grpc.CallOption) (*ToggleWABotResponse, error)
	ReconnectSession(ctx context.Context, in *ReconnectWASessionRequest, opts ...grpc.CallOption) (*ReconnectWASessionResponse, error)
	LogoutSession(ctx context.Context, in *LogoutWASessionRequest, opts ...grpc.CallOption) (*LogoutWASessionResponse, error)
	PurgeSession(ctx context.Context, in *PurgeWASessionRequest, opts ...grpc.CallOption) (*PurgeWASessionResponse, error)
	SendTextMessage(ctx context.Context, in *SendWATextMessageRequest, opts ...grpc.CallOption) (*SendWATextMessageResponse, error)
}

type WhatsAppServiceServer interface {
	ListSessions(context.Context, *ListWASessionsRequest) (*ListWASessionsResponse, error)
	CreateSession(context.Context, *CreateWASessionRequest) (*CreateWASessionResponse, error)
	GetQRCode(context.Context, *GetWASessionQRRequest) (*GetWASessionQRResponse, error)
	GetPairingCode(context.Context, *GetWASessionPairingRequest) (*GetWASessionPairingResponse, error)
	ToggleBot(context.Context, *ToggleWABotRequest) (*ToggleWABotResponse, error)
	ReconnectSession(context.Context, *ReconnectWASessionRequest) (*ReconnectWASessionResponse, error)
	LogoutSession(context.Context, *LogoutWASessionRequest) (*LogoutWASessionResponse, error)
	PurgeSession(context.Context, *PurgeWASessionRequest) (*PurgeWASessionResponse, error)
	SendTextMessage(context.Context, *SendWATextMessageRequest) (*SendWATextMessageResponse, error)
}

type UnimplementedWhatsAppServiceServer struct{}

func (UnimplementedWhatsAppServiceServer) ListSessions(context.Context, *ListWASessionsRequest) (*ListWASessionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListSessions not implemented")
}
func (UnimplementedWhatsAppServiceServer) CreateSession(context.Context, *CreateWASessionRequest) (*CreateWASessionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSession not implemented")
}
func (UnimplementedWhatsAppServiceServer) GetQRCode(context.Context, *GetWASessionQRRequest) (*GetWASessionQRResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetQRCode not implemented")
}
func (UnimplementedWhatsAppServiceServer) GetPairingCode(context.Context, *GetWASessionPairingRequest) (*GetWASessionPairingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPairingCode not implemented")
}
func (UnimplementedWhatsAppServiceServer) ToggleBot(context.Context, *ToggleWABotRequest) (*ToggleWABotResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ToggleBot not implemented")
}
func (UnimplementedWhatsAppServiceServer) ReconnectSession(context.Context, *ReconnectWASessionRequest) (*ReconnectWASessionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReconnectSession not implemented")
}
func (UnimplementedWhatsAppServiceServer) LogoutSession(context.Context, *LogoutWASessionRequest) (*LogoutWASessionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LogoutSession not implemented")
}
func (UnimplementedWhatsAppServiceServer) PurgeSession(context.Context, *PurgeWASessionRequest) (*PurgeWASessionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PurgeSession not implemented")
}
func (UnimplementedWhatsAppServiceServer) SendTextMessage(context.Context, *SendWATextMessageRequest) (*SendWATextMessageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendTextMessage not implemented")
}
