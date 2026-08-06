package devicepb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type KnowledgeItem struct {
	Id        string   `json:"id"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Category  string   `json:"category"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type ListKnowledgeRequest struct {
	Category    string `json:"category"`
	SearchQuery string `json:"search_query"`
}
type ListKnowledgeResponse struct {
	Items []*KnowledgeItem `json:"items"`
}

type GetKnowledgeRequest struct {
	Id string `json:"id"`
}
type GetKnowledgeResponse struct {
	Item *KnowledgeItem `json:"item"`
}

type CreateKnowledgeRequest struct {
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
}
type CreateKnowledgeResponse struct {
	Item *KnowledgeItem `json:"item"`
}

type UpdateKnowledgeRequest struct {
	Id       string   `json:"id"`
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
}
type UpdateKnowledgeResponse struct {
	Item *KnowledgeItem `json:"item"`
}

type DeleteKnowledgeRequest struct {
	Id string `json:"id"`
}
type DeleteKnowledgeResponse struct {
	Message string `json:"message"`
}

type LLMConfig struct {
	Id            string  `json:"id"`
	Provider      string  `json:"provider"`
	ModelName     string  `json:"model_name"`
	ApiKeyMasked  string  `json:"api_key_masked"`
	BaseUrl       string  `json:"base_url"`
	IsActive      bool    `json:"is_active"`
	Temperature   float64 `json:"temperature"`
	MaxTokens     int32   `json:"max_tokens"`
	SystemPrompt  string  `json:"system_prompt"`
}

type ListLLMConfigsRequest struct{}
type ListLLMConfigsResponse struct {
	Configs []*LLMConfig `json:"configs"`
}

type CreateLLMConfigRequest struct {
	Provider     string  `json:"provider"`
	ModelName    string  `json:"model_name"`
	ApiKey       string  `json:"api_key"`
	BaseUrl      string  `json:"base_url"`
	Temperature  float64 `json:"temperature"`
	MaxTokens    int32   `json:"max_tokens"`
	SystemPrompt string  `json:"system_prompt"`
}
type CreateLLMConfigResponse struct {
	Config *LLMConfig `json:"config"`
}

type UpdateLLMConfigRequest struct {
	Id           string  `json:"id"`
	Provider     string  `json:"provider"`
	ModelName    string  `json:"model_name"`
	ApiKey       string  `json:"api_key"`
	BaseUrl      string  `json:"base_url"`
	Temperature  float64 `json:"temperature"`
	MaxTokens    int32   `json:"max_tokens"`
	SystemPrompt string  `json:"system_prompt"`
}
type UpdateLLMConfigResponse struct {
	Config *LLMConfig `json:"config"`
}

type ActivateLLMConfigRequest struct {
	Id string `json:"id"`
}
type ActivateLLMConfigResponse struct {
	Message string `json:"message"`
}

type TestLLMConfigRequest struct {
	Id         string `json:"id"`
	TestPrompt string `json:"test_prompt"`
}
type TestLLMConfigResponse struct {
	Success      bool   `json:"success"`
	ResponseText string `json:"response_text"`
	ErrorMessage string `json:"error_message"`
}

type DeleteLLMConfigRequest struct {
	Id string `json:"id"`
}
type DeleteLLMConfigResponse struct {
	Message string `json:"message"`
}

type Technician struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	PhoneNumber   string `json:"phone_number"`
	Email         string `json:"email"`
	IsActive      bool   `json:"is_active"`
	ActiveTickets int32  `json:"active_tickets"`
}

type ListTechniciansRequest struct{}
type ListTechniciansResponse struct {
	Technicians []*Technician `json:"technicians"`
}

type CreateTechnicianRequest struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
}
type CreateTechnicianResponse struct {
	Technician *Technician `json:"technician"`
}

type UpdateTechnicianRequest struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
}
type UpdateTechnicianResponse struct {
	Technician *Technician `json:"technician"`
}

type ToggleTechnicianActiveRequest struct {
	Id       string `json:"id"`
	IsActive bool   `json:"is_active"`
}
type ToggleTechnicianActiveResponse struct {
	Message  string `json:"message"`
	IsActive bool   `json:"is_active"`
}

type DeleteTechnicianRequest struct {
	Id string `json:"id"`
}
type DeleteTechnicianResponse struct {
	Message string `json:"message"`
}

type KnowledgeServiceClient interface {
	ListKnowledge(ctx context.Context, in *ListKnowledgeRequest, opts ...grpc.CallOption) (*ListKnowledgeResponse, error)
	GetKnowledge(ctx context.Context, in *GetKnowledgeRequest, opts ...grpc.CallOption) (*GetKnowledgeResponse, error)
	CreateKnowledge(ctx context.Context, in *CreateKnowledgeRequest, opts ...grpc.CallOption) (*CreateKnowledgeResponse, error)
	UpdateKnowledge(ctx context.Context, in *UpdateKnowledgeRequest, opts ...grpc.CallOption) (*UpdateKnowledgeResponse, error)
	DeleteKnowledge(ctx context.Context, in *DeleteKnowledgeRequest, opts ...grpc.CallOption) (*DeleteKnowledgeResponse, error)

	ListLLMConfigs(ctx context.Context, in *ListLLMConfigsRequest, opts ...grpc.CallOption) (*ListLLMConfigsResponse, error)
	CreateLLMConfig(ctx context.Context, in *CreateLLMConfigRequest, opts ...grpc.CallOption) (*CreateLLMConfigResponse, error)
	UpdateLLMConfig(ctx context.Context, in *UpdateLLMConfigRequest, opts ...grpc.CallOption) (*UpdateLLMConfigResponse, error)
	ActivateLLMConfig(ctx context.Context, in *ActivateLLMConfigRequest, opts ...grpc.CallOption) (*ActivateLLMConfigResponse, error)
	TestLLMConfig(ctx context.Context, in *TestLLMConfigRequest, opts ...grpc.CallOption) (*TestLLMConfigResponse, error)
	DeleteLLMConfig(ctx context.Context, in *DeleteLLMConfigRequest, opts ...grpc.CallOption) (*DeleteLLMConfigResponse, error)

	ListTechnicians(ctx context.Context, in *ListTechniciansRequest, opts ...grpc.CallOption) (*ListTechniciansResponse, error)
	CreateTechnician(ctx context.Context, in *CreateTechnicianRequest, opts ...grpc.CallOption) (*CreateTechnicianResponse, error)
	UpdateTechnician(ctx context.Context, in *UpdateTechnicianRequest, opts ...grpc.CallOption) (*UpdateTechnicianResponse, error)
	ToggleTechnicianActive(ctx context.Context, in *ToggleTechnicianActiveRequest, opts ...grpc.CallOption) (*ToggleTechnicianActiveResponse, error)
	DeleteTechnician(ctx context.Context, in *DeleteTechnicianRequest, opts ...grpc.CallOption) (*DeleteTechnicianResponse, error)
}

type KnowledgeServiceServer interface {
	ListKnowledge(context.Context, *ListKnowledgeRequest) (*ListKnowledgeResponse, error)
	GetKnowledge(context.Context, *GetKnowledgeRequest) (*GetKnowledgeResponse, error)
	CreateKnowledge(context.Context, *CreateKnowledgeRequest) (*CreateKnowledgeResponse, error)
	UpdateKnowledge(context.Context, *UpdateKnowledgeRequest) (*UpdateKnowledgeResponse, error)
	DeleteKnowledge(context.Context, *DeleteKnowledgeRequest) (*DeleteKnowledgeResponse, error)

	ListLLMConfigs(context.Context, *ListLLMConfigsRequest) (*ListLLMConfigsResponse, error)
	CreateLLMConfig(context.Context, *CreateLLMConfigRequest) (*CreateLLMConfigResponse, error)
	UpdateLLMConfig(context.Context, *UpdateLLMConfigRequest) (*UpdateLLMConfigResponse, error)
	ActivateLLMConfig(context.Context, *ActivateLLMConfigRequest) (*ActivateLLMConfigResponse, error)
	TestLLMConfig(context.Context, *TestLLMConfigRequest) (*TestLLMConfigResponse, error)
	DeleteLLMConfig(context.Context, *DeleteLLMConfigRequest) (*DeleteLLMConfigResponse, error)

	ListTechnicians(context.Context, *ListTechniciansRequest) (*ListTechniciansResponse, error)
	CreateTechnician(context.Context, *CreateTechnicianRequest) (*CreateTechnicianResponse, error)
	UpdateTechnician(context.Context, *UpdateTechnicianRequest) (*UpdateTechnicianResponse, error)
	ToggleTechnicianActive(context.Context, *ToggleTechnicianActiveRequest) (*ToggleTechnicianActiveResponse, error)
	DeleteTechnician(context.Context, *DeleteTechnicianRequest) (*DeleteTechnicianResponse, error)
}

type UnimplementedKnowledgeServiceServer struct{}

func (UnimplementedKnowledgeServiceServer) ListKnowledge(context.Context, *ListKnowledgeRequest) (*ListKnowledgeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListKnowledge not implemented")
}
func (UnimplementedKnowledgeServiceServer) GetKnowledge(context.Context, *GetKnowledgeRequest) (*GetKnowledgeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetKnowledge not implemented")
}
func (UnimplementedKnowledgeServiceServer) CreateKnowledge(context.Context, *CreateKnowledgeRequest) (*CreateKnowledgeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateKnowledge not implemented")
}
func (UnimplementedKnowledgeServiceServer) UpdateKnowledge(context.Context, *UpdateKnowledgeRequest) (*UpdateKnowledgeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateKnowledge not implemented")
}
func (UnimplementedKnowledgeServiceServer) DeleteKnowledge(context.Context, *DeleteKnowledgeRequest) (*DeleteKnowledgeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteKnowledge not implemented")
}
func (UnimplementedKnowledgeServiceServer) ListLLMConfigs(context.Context, *ListLLMConfigsRequest) (*ListLLMConfigsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListLLMConfigs not implemented")
}
func (UnimplementedKnowledgeServiceServer) CreateLLMConfig(context.Context, *CreateLLMConfigRequest) (*CreateLLMConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateLLMConfig not implemented")
}
func (UnimplementedKnowledgeServiceServer) UpdateLLMConfig(context.Context, *UpdateLLMConfigRequest) (*UpdateLLMConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateLLMConfig not implemented")
}
func (UnimplementedKnowledgeServiceServer) ActivateLLMConfig(context.Context, *ActivateLLMConfigRequest) (*ActivateLLMConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ActivateLLMConfig not implemented")
}
func (UnimplementedKnowledgeServiceServer) TestLLMConfig(context.Context, *TestLLMConfigRequest) (*TestLLMConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TestLLMConfig not implemented")
}
func (UnimplementedKnowledgeServiceServer) DeleteLLMConfig(context.Context, *DeleteLLMConfigRequest) (*DeleteLLMConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteLLMConfig not implemented")
}
func (UnimplementedKnowledgeServiceServer) ListTechnicians(context.Context, *ListTechniciansRequest) (*ListTechniciansResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListTechnicians not implemented")
}
func (UnimplementedKnowledgeServiceServer) CreateTechnician(context.Context, *CreateTechnicianRequest) (*CreateTechnicianResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateTechnician not implemented")
}
func (UnimplementedKnowledgeServiceServer) UpdateTechnician(context.Context, *UpdateTechnicianRequest) (*UpdateTechnicianResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateTechnician not implemented")
}
func (UnimplementedKnowledgeServiceServer) ToggleTechnicianActive(context.Context, *ToggleTechnicianActiveRequest) (*ToggleTechnicianActiveResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ToggleTechnicianActive not implemented")
}
func (UnimplementedKnowledgeServiceServer) DeleteTechnician(context.Context, *DeleteTechnicianRequest) (*DeleteTechnicianResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteTechnician not implemented")
}
