package llm

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	llmUC "github.com/quixiq/polyglot/internal/usecase/llm"
)

// NewLLMConfigServiceHandler mounts LLMConfigService Connect handlers.
func NewLLMConfigServiceHandler(
	useCase *llmUC.ManageConfigUseCase,
) (string, http.Handler) {
	handler := NewLLMConnectHandler(useCase)
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.LLMConfigService"

	mux.Handle("/"+serviceName+"/ListLLMConfigs", connect.NewUnaryHandler("/"+serviceName+"/ListLLMConfigs", handler.ListLLMConfigs, opts...))
	mux.Handle("/"+serviceName+"/CreateLLMConfig", connect.NewUnaryHandler("/"+serviceName+"/CreateLLMConfig", handler.CreateLLMConfig, opts...))
	mux.Handle("/"+serviceName+"/UpdateLLMConfig", connect.NewUnaryHandler("/"+serviceName+"/UpdateLLMConfig", handler.UpdateLLMConfig, opts...))
	mux.Handle("/"+serviceName+"/ActivateLLMConfig", connect.NewUnaryHandler("/"+serviceName+"/ActivateLLMConfig", handler.ActivateLLMConfig, opts...))
	mux.Handle("/"+serviceName+"/TestLLMConfig", connect.NewUnaryHandler("/"+serviceName+"/TestLLMConfig", handler.TestLLMConfig, opts...))
	mux.Handle("/"+serviceName+"/DeleteLLMConfig", connect.NewUnaryHandler("/"+serviceName+"/DeleteLLMConfig", handler.DeleteLLMConfig, opts...))

	return "/" + serviceName + "/", mux
}
