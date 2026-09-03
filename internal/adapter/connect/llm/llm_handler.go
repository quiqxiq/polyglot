package llm

import (
	"context"
	"strconv"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	llmUC "github.com/quixiq/polyglot/internal/usecase/llm"
	"github.com/quixiq/polyglot/pkg/response"
)

// LLMConnectHandler implements the LLMConfigService ConnectRPC service.
//
//nolint:revive // Explicit transport role is part of the project naming convention.
type LLMConnectHandler struct {
	useCase *llmUC.ManageConfigUseCase
}

// NewLLMConnectHandler constructs an LLMConnectHandler.
func NewLLMConnectHandler(useCase *llmUC.ManageConfigUseCase) *LLMConnectHandler {
	return &LLMConnectHandler{
		useCase: useCase,
	}
}

// ListLLMConfigs returns all configured LLM models and credentials.
func (h *LLMConnectHandler) ListLLMConfigs(ctx context.Context, req *connect.Request[devicepb.ListLLMConfigsRequest]) (*connect.Response[devicepb.ListLLMConfigsResponse], error) {
	if h.useCase == nil {
		return connect.NewResponse(&devicepb.ListLLMConfigsResponse{}), nil
	}
	configs, err := h.useCase.List(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListLLMConfigsResponse{
		Configs: ToProtoLLMConfigList(configs),
	}), nil
}

// CreateLLMConfig adds a new LLM provider/model configuration.
func (h *LLMConnectHandler) CreateLLMConfig(ctx context.Context, req *connect.Request[devicepb.CreateLLMConfigRequest]) (*connect.Response[devicepb.CreateLLMConfigResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("llm usecase not initialized")
	}

	cfg, err := h.useCase.Create(ctx, llmUC.CreateConfigInput{
		Provider:       req.Msg.Provider,
		ModelName:      req.Msg.ModelName,
		APIKey:         req.Msg.ApiKey,
		BaseURL:        req.Msg.BaseUrl,
		Temperature:    req.Msg.Temperature,
		MaxTokens:      int(req.Msg.MaxTokens),
		SystemPrompt:   req.Msg.SystemPrompt,
		SkillsMode:     req.Msg.SkillsMode,
		EnableSkills:   req.Msg.EnableSkills,
		SkillsPrompt:   req.Msg.SkillsPrompt,
		SelectedSkills: req.Msg.SelectedSkills,
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.CreateLLMConfigResponse{
		Config: ToProtoLLMConfig(cfg),
	}), nil
}

// UpdateLLMConfig updates an existing LLM model configuration.
func (h *LLMConnectHandler) UpdateLLMConfig(ctx context.Context, req *connect.Request[devicepb.UpdateLLMConfigRequest]) (*connect.Response[devicepb.UpdateLLMConfigResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("llm usecase not initialized")
	}
	idNum, _ := strconv.ParseUint(req.Msg.Id, 10, 32)
	cfg, err := h.useCase.Update(ctx, llmUC.UpdateConfigInput{
		ID:             uint(idNum),
		Provider:       req.Msg.Provider,
		ModelName:      req.Msg.ModelName,
		APIKey:         req.Msg.ApiKey,
		BaseURL:        req.Msg.BaseUrl,
		Temperature:    req.Msg.Temperature,
		MaxTokens:      int(req.Msg.MaxTokens),
		SystemPrompt:   req.Msg.SystemPrompt,
		SkillsMode:     req.Msg.SkillsMode,
		EnableSkills:   req.Msg.EnableSkills,
		SkillsPrompt:   req.Msg.SkillsPrompt,
		SelectedSkills: req.Msg.SelectedSkills,
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.UpdateLLMConfigResponse{
		Config: ToProtoLLMConfig(cfg),
	}), nil
}

// ActivateLLMConfig activates an LLM configuration as the default active model.
func (h *LLMConnectHandler) ActivateLLMConfig(ctx context.Context, req *connect.Request[devicepb.ActivateLLMConfigRequest]) (*connect.Response[devicepb.ActivateLLMConfigResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("llm usecase not initialized")
	}
	idNum, _ := strconv.ParseUint(req.Msg.Id, 10, 32)
	if err := h.useCase.SetActive(ctx, uint(idNum)); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ActivateLLMConfigResponse{
		Message: "llm config activated successfully",
	}), nil
}

// TestLLMConfig tests network and authentication connectivity for an LLM configuration.
func (h *LLMConnectHandler) TestLLMConfig(ctx context.Context, req *connect.Request[devicepb.TestLLMConfigRequest]) (*connect.Response[devicepb.TestLLMConfigResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("llm usecase not initialized")
	}
	idNum, _ := strconv.ParseUint(req.Msg.Id, 10, 32)
	msg, err := h.useCase.TestConnection(ctx, uint(idNum))
	if err != nil {
		return connect.NewResponse(&devicepb.TestLLMConfigResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}), nil
	}

	return connect.NewResponse(&devicepb.TestLLMConfigResponse{
		Success:      true,
		ResponseText: msg,
	}), nil
}

// DeleteLLMConfig deletes an LLM configuration by ID.
func (h *LLMConnectHandler) DeleteLLMConfig(ctx context.Context, req *connect.Request[devicepb.DeleteLLMConfigRequest]) (*connect.Response[devicepb.DeleteLLMConfigResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("llm usecase not initialized")
	}
	idNum, _ := strconv.ParseUint(req.Msg.Id, 10, 32)
	if err := h.useCase.Delete(ctx, uint(idNum)); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.DeleteLLMConfigResponse{
		Message: "llm config deleted successfully",
	}), nil
}
