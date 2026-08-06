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
	QrCodeBase64 string `json:"qr_code_base64"`
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

type LogoutWASessionRequest struct {
	SessionId string `json:"session_id"`
}
type LogoutWASessionResponse struct {
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

type ConversationMessage struct {
	Id        string `json:"id"`
	Sender    string `json:"sender"`
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
}

type Conversation struct {
	Id                 string                 `json:"id"`
	SessionId          string                 `json:"session_id"`
	ClientPhone        string                 `json:"client_phone"`
	ClientName         string                 `json:"client_name"`
	Status             string                 `json:"status"`
	AssignedTechnician string                 `json:"assigned_technician"`
	Messages           []*ConversationMessage `json:"messages"`
	UpdatedAt          string                 `json:"updated_at"`
}

type ListConversationsRequest struct {
	SessionId string `json:"session_id"`
}
type ListConversationsResponse struct {
	Conversations []*Conversation `json:"conversations"`
}

type GetConversationRequest struct {
	Id string `json:"id"`
}
type GetConversationResponse struct {
	Conversation *Conversation `json:"conversation"`
}

type TakeOverConversationRequest struct {
	Id           string `json:"id"`
	TechnicianId string `json:"technician_id"`
}
type TakeOverConversationResponse struct {
	Message string `json:"message"`
}

type ResetConversationBotRequest struct {
	Id string `json:"id"`
}
type ResetConversationBotResponse struct {
	Message string `json:"message"`
}

type CloseConversationRequest struct {
	Id string `json:"id"`
}
type CloseConversationResponse struct {
	Message string `json:"message"`
}

type BotServiceClient interface {
	ListSessions(ctx context.Context, in *ListWASessionsRequest, opts ...grpc.CallOption) (*ListWASessionsResponse, error)
	CreateSession(ctx context.Context, in *CreateWASessionRequest, opts ...grpc.CallOption) (*CreateWASessionResponse, error)
	GetQRCode(ctx context.Context, in *GetWASessionQRRequest, opts ...grpc.CallOption) (*GetWASessionQRResponse, error)
	GetPairingCode(ctx context.Context, in *GetWASessionPairingRequest, opts ...grpc.CallOption) (*GetWASessionPairingResponse, error)
	ToggleBot(ctx context.Context, in *ToggleWABotRequest, opts ...grpc.CallOption) (*ToggleWABotResponse, error)
	LogoutSession(ctx context.Context, in *LogoutWASessionRequest, opts ...grpc.CallOption) (*LogoutWASessionResponse, error)
	SendTextMessage(ctx context.Context, in *SendWATextMessageRequest, opts ...grpc.CallOption) (*SendWATextMessageResponse, error)

	ListConversations(ctx context.Context, in *ListConversationsRequest, opts ...grpc.CallOption) (*ListConversationsResponse, error)
	GetConversation(ctx context.Context, in *GetConversationRequest, opts ...grpc.CallOption) (*GetConversationResponse, error)
	TakeOverConversation(ctx context.Context, in *TakeOverConversationRequest, opts ...grpc.CallOption) (*TakeOverConversationResponse, error)
	ResetConversationBot(ctx context.Context, in *ResetConversationBotRequest, opts ...grpc.CallOption) (*ResetConversationBotResponse, error)
	CloseConversation(ctx context.Context, in *CloseConversationRequest, opts ...grpc.CallOption) (*CloseConversationResponse, error)
}

type BotServiceServer interface {
	ListSessions(context.Context, *ListWASessionsRequest) (*ListWASessionsResponse, error)
	CreateSession(context.Context, *CreateWASessionRequest) (*CreateWASessionResponse, error)
	GetQRCode(context.Context, *GetWASessionQRRequest) (*GetWASessionQRResponse, error)
	GetPairingCode(context.Context, *GetWASessionPairingRequest) (*GetWASessionPairingResponse, error)
	ToggleBot(context.Context, *ToggleWABotRequest) (*ToggleWABotResponse, error)
	LogoutSession(context.Context, *LogoutWASessionRequest) (*LogoutWASessionResponse, error)
	SendTextMessage(context.Context, *SendWATextMessageRequest) (*SendWATextMessageResponse, error)

	ListConversations(context.Context, *ListConversationsRequest) (*ListConversationsResponse, error)
	GetConversation(context.Context, *GetConversationRequest) (*GetConversationResponse, error)
	TakeOverConversation(context.Context, *TakeOverConversationRequest) (*TakeOverConversationResponse, error)
	ResetConversationBot(context.Context, *ResetConversationBotRequest) (*ResetConversationBotResponse, error)
	CloseConversation(context.Context, *CloseConversationRequest) (*CloseConversationResponse, error)
}

type UnimplementedBotServiceServer struct{}

func (UnimplementedBotServiceServer) ListSessions(context.Context, *ListWASessionsRequest) (*ListWASessionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListSessions not implemented")
}
func (UnimplementedBotServiceServer) CreateSession(context.Context, *CreateWASessionRequest) (*CreateWASessionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSession not implemented")
}
func (UnimplementedBotServiceServer) GetQRCode(context.Context, *GetWASessionQRRequest) (*GetWASessionQRResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetQRCode not implemented")
}
func (UnimplementedBotServiceServer) GetPairingCode(context.Context, *GetWASessionPairingRequest) (*GetWASessionPairingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPairingCode not implemented")
}
func (UnimplementedBotServiceServer) ToggleBot(context.Context, *ToggleWABotRequest) (*ToggleWABotResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ToggleBot not implemented")
}
func (UnimplementedBotServiceServer) LogoutSession(context.Context, *LogoutWASessionRequest) (*LogoutWASessionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LogoutSession not implemented")
}
func (UnimplementedBotServiceServer) SendTextMessage(context.Context, *SendWATextMessageRequest) (*SendWATextMessageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendTextMessage not implemented")
}
func (UnimplementedBotServiceServer) ListConversations(context.Context, *ListConversationsRequest) (*ListConversationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListConversations not implemented")
}
func (UnimplementedBotServiceServer) GetConversation(context.Context, *GetConversationRequest) (*GetConversationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetConversation not implemented")
}
func (UnimplementedBotServiceServer) TakeOverConversation(context.Context, *TakeOverConversationRequest) (*TakeOverConversationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TakeOverConversation not implemented")
}
func (UnimplementedBotServiceServer) ResetConversationBot(context.Context, *ResetConversationBotRequest) (*ResetConversationBotResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResetConversationBot not implemented")
}
func (UnimplementedBotServiceServer) CloseConversation(context.Context, *CloseConversationRequest) (*CloseConversationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CloseConversation not implemented")
}
