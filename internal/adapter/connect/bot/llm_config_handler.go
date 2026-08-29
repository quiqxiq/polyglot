package bot

import (
	"context"
	"fmt"
	"strconv"

	"connectrpc.com/connect"
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/adapter/llm/genkit"
	"github.com/quixiq/polyglot/internal/config"
	domainllm "github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/pkg/llmcost"
	"github.com/quixiq/polyglot/pkg/response"
)

func (h *BotConnectHandler) ListLLMConfigs(ctx context.Context, req *connect.Request[devicepb.ListLLMConfigsRequest]) (*connect.Response[devicepb.ListLLMConfigsResponse], error) {
	if h.llmRepo == nil {
		return connect.NewResponse(&devicepb.ListLLMConfigsResponse{}), nil
	}
	configs, err := h.llmRepo.FindAll(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListLLMConfigsResponse{
		Configs: toProtoLLMConfigList(configs),
	}), nil
}

func (h *BotConnectHandler) CreateLLMConfig(ctx context.Context, req *connect.Request[devicepb.CreateLLMConfigRequest]) (*connect.Response[devicepb.CreateLLMConfigResponse], error) {
	if h.llmRepo == nil {
		return nil, response.Unavailable("llm repository not initialized")
	}

	var encryptedKey string
	if req.Msg.ApiKey != "" {
		enc, err := config.Encrypt(req.Msg.ApiKey, h.encryptionKey)
		if err != nil {
			return nil, response.Internal("failed to encrypt api key")
		}
		encryptedKey = enc
	}

	inRate, outRate := llmcost.GetDefaultPricing(req.Msg.Provider, req.Msg.ModelName)

	isActive := false
	if activeCfg, _ := h.llmRepo.FindActive(ctx); activeCfg == nil {
		isActive = true
	}

	cfg := &domainllm.Config{
		Provider:        req.Msg.Provider,
		Model:           req.Msg.ModelName,
		APIKeyEncrypted: encryptedKey,
		BaseURL:         req.Msg.BaseUrl,
		Temperature:     req.Msg.Temperature,
		MaxOutputTokens: int(req.Msg.MaxTokens),
		SystemPrompt:    req.Msg.SystemPrompt,
		SkillsMode:      req.Msg.SkillsMode,
		EnableSkills:    req.Msg.EnableSkills,
		SkillsPrompt:    req.Msg.SkillsPrompt,
		SelectedSkills:  req.Msg.SelectedSkills,
		CostPer1MInput:  inRate,
		CostPer1MOutput: outRate,
		IsActive:        isActive,
	}

	if err := h.llmRepo.Create(ctx, cfg); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.CreateLLMConfigResponse{
		Config: toProtoLLMConfig(cfg),
	}), nil
}

func (h *BotConnectHandler) UpdateLLMConfig(ctx context.Context, req *connect.Request[devicepb.UpdateLLMConfigRequest]) (*connect.Response[devicepb.UpdateLLMConfigResponse], error) {
	if h.llmRepo == nil {
		return nil, response.Unavailable("llm repository not initialized")
	}
	idNum, _ := strconv.ParseUint(req.Msg.Id, 10, 32)
	cfg, err := h.llmRepo.FindByID(ctx, uint(idNum))
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	cfg.Provider = req.Msg.Provider
	cfg.Model = req.Msg.ModelName
	cfg.BaseURL = req.Msg.BaseUrl
	cfg.Temperature = req.Msg.Temperature
	cfg.SystemPrompt = req.Msg.SystemPrompt
	cfg.SkillsMode = req.Msg.SkillsMode
	cfg.EnableSkills = req.Msg.EnableSkills
	cfg.SkillsPrompt = req.Msg.SkillsPrompt
	cfg.SelectedSkills = req.Msg.SelectedSkills

	inRate, outRate := llmcost.GetDefaultPricing(req.Msg.Provider, req.Msg.ModelName)
	cfg.CostPer1MInput = inRate
	cfg.CostPer1MOutput = outRate

	if req.Msg.MaxTokens > 0 {
		cfg.MaxOutputTokens = int(req.Msg.MaxTokens)
	}
	if req.Msg.ApiKey != "" {
		enc, err := config.Encrypt(req.Msg.ApiKey, h.encryptionKey)
		if err != nil {
			return nil, response.Internal("failed to encrypt api key")
		}
		cfg.APIKeyEncrypted = enc
	}

	if err := h.llmRepo.Update(ctx, cfg); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.UpdateLLMConfigResponse{
		Config: toProtoLLMConfig(cfg),
	}), nil
}

func (h *BotConnectHandler) ActivateLLMConfig(ctx context.Context, req *connect.Request[devicepb.ActivateLLMConfigRequest]) (*connect.Response[devicepb.ActivateLLMConfigResponse], error) {
	if h.llmRepo == nil {
		return nil, response.Unavailable("llm repository not initialized")
	}
	idNum, _ := strconv.ParseUint(req.Msg.Id, 10, 32)
	if err := h.llmRepo.SetActive(ctx, uint(idNum)); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ActivateLLMConfigResponse{
		Message: "llm config activated successfully",
	}), nil
}

func (h *BotConnectHandler) TestLLMConfig(ctx context.Context, req *connect.Request[devicepb.TestLLMConfigRequest]) (*connect.Response[devicepb.TestLLMConfigResponse], error) {
	if h.llmRepo == nil {
		return nil, response.Unavailable("llm repository not initialized")
	}
	idNum, _ := strconv.ParseUint(req.Msg.Id, 10, 32)
	cfg, err := h.llmRepo.FindByID(ctx, uint(idNum))
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	var apiKey string
	if cfg.APIKeyEncrypted != "" {
		dec, err := config.Decrypt(cfg.APIKeyEncrypted, h.encryptionKey)
		if err != nil {
			return connect.NewResponse(&devicepb.TestLLMConfigResponse{
				Success:      false,
				ErrorMessage: fmt.Sprintf("Gagal decrypt API key: %v", err),
			}), nil
		}
		apiKey = dec
	}

	prov, err := genkit.NewProvider(ctx, cfg, apiKey)
	if err != nil {
		return connect.NewResponse(&devicepb.TestLLMConfigResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}), nil
	}

	if err := prov.TestConnection(ctx); err != nil {
		return connect.NewResponse(&devicepb.TestLLMConfigResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}), nil
	}

	return connect.NewResponse(&devicepb.TestLLMConfigResponse{
		Success:      true,
		ResponseText: fmt.Sprintf("Koneksi ke model AI Genkit (%s - %s) berhasil diuji!", cfg.Provider, cfg.Model),
	}), nil
}

func (h *BotConnectHandler) DeleteLLMConfig(ctx context.Context, req *connect.Request[devicepb.DeleteLLMConfigRequest]) (*connect.Response[devicepb.DeleteLLMConfigResponse], error) {
	if h.llmRepo == nil {
		return nil, response.Unavailable("llm repository not initialized")
	}
	idNum, _ := strconv.ParseUint(req.Msg.Id, 10, 32)
	if err := h.llmRepo.Delete(ctx, uint(idNum)); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.DeleteLLMConfigResponse{
		Message: "llm config deleted successfully",
	}), nil
}
