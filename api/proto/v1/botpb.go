package devicepb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
	ListConversations(ctx context.Context, in *ListConversationsRequest, opts ...grpc.CallOption) (*ListConversationsResponse, error)
	GetConversation(ctx context.Context, in *GetConversationRequest, opts ...grpc.CallOption) (*GetConversationResponse, error)
	TakeOverConversation(ctx context.Context, in *TakeOverConversationRequest, opts ...grpc.CallOption) (*TakeOverConversationResponse, error)
	ResetConversationBot(ctx context.Context, in *ResetConversationBotRequest, opts ...grpc.CallOption) (*ResetConversationBotResponse, error)
	CloseConversation(ctx context.Context, in *CloseConversationRequest, opts ...grpc.CallOption) (*CloseConversationResponse, error)
}

type BotServiceServer interface {
	ListConversations(context.Context, *ListConversationsRequest) (*ListConversationsResponse, error)
	GetConversation(context.Context, *GetConversationRequest) (*GetConversationResponse, error)
	TakeOverConversation(context.Context, *TakeOverConversationRequest) (*TakeOverConversationResponse, error)
	ResetConversationBot(context.Context, *ResetConversationBotRequest) (*ResetConversationBotResponse, error)
	CloseConversation(context.Context, *CloseConversationRequest) (*CloseConversationResponse, error)
}

type UnimplementedBotServiceServer struct{}

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
