package connectadapter

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
)

func (h *KnowledgeConnectHandler) ListLLMConfigs(ctx context.Context, req *connect.Request[devicepb.ListLLMConfigsRequest]) (*connect.Response[devicepb.ListLLMConfigsResponse], error) {
	return connect.NewResponse(&devicepb.ListLLMConfigsResponse{
		Configs: []*devicepb.LLMConfig{
			{
				Id:           "llm-default",
				Provider:     "openai",
				ModelName:    "gpt-4o-mini",
				IsActive:     true,
				Temperature:  0.7,
				MaxTokens:    1024,
				SystemPrompt: "You are GNET Bot assistant.",
			},
		},
	}), nil
}

func (h *KnowledgeConnectHandler) CreateLLMConfig(ctx context.Context, req *connect.Request[devicepb.CreateLLMConfigRequest]) (*connect.Response[devicepb.CreateLLMConfigResponse], error) {
	return connect.NewResponse(&devicepb.CreateLLMConfigResponse{
		Config: &devicepb.LLMConfig{
			Id:           "llm-new",
			Provider:     req.Msg.Provider,
			ModelName:    req.Msg.ModelName,
			IsActive:     false,
			Temperature:  req.Msg.Temperature,
			MaxTokens:    req.Msg.MaxTokens,
			SystemPrompt: req.Msg.SystemPrompt,
		},
	}), nil
}

func (h *KnowledgeConnectHandler) UpdateLLMConfig(ctx context.Context, req *connect.Request[devicepb.UpdateLLMConfigRequest]) (*connect.Response[devicepb.UpdateLLMConfigResponse], error) {
	return connect.NewResponse(&devicepb.UpdateLLMConfigResponse{
		Config: &devicepb.LLMConfig{
			Id:           req.Msg.Id,
			Provider:     req.Msg.Provider,
			ModelName:    req.Msg.ModelName,
			IsActive:     true,
			Temperature:  req.Msg.Temperature,
			MaxTokens:    req.Msg.MaxTokens,
			SystemPrompt: req.Msg.SystemPrompt,
		},
	}), nil
}

func (h *KnowledgeConnectHandler) ActivateLLMConfig(ctx context.Context, req *connect.Request[devicepb.ActivateLLMConfigRequest]) (*connect.Response[devicepb.ActivateLLMConfigResponse], error) {
	return connect.NewResponse(&devicepb.ActivateLLMConfigResponse{
		Message: "llm config activated successfully",
	}), nil
}

func (h *KnowledgeConnectHandler) TestLLMConfig(ctx context.Context, req *connect.Request[devicepb.TestLLMConfigRequest]) (*connect.Response[devicepb.TestLLMConfigResponse], error) {
	return connect.NewResponse(&devicepb.TestLLMConfigResponse{
		Success:      true,
		ResponseText: "Hello! LLM Connection test successful.",
	}), nil
}

func (h *KnowledgeConnectHandler) DeleteLLMConfig(ctx context.Context, req *connect.Request[devicepb.DeleteLLMConfigRequest]) (*connect.Response[devicepb.DeleteLLMConfigResponse], error) {
	return connect.NewResponse(&devicepb.DeleteLLMConfigResponse{
		Message: "llm config deleted successfully",
	}), nil
}
