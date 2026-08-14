package bot

import (
	"net/http"

	"connectrpc.com/connect"

	knowledgeuc "github.com/quixiq/polyglot/internal/usecase/knowledge"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
)

type KnowledgeConnectHandler struct {
	documents *knowledgeuc.DocumentManager
}

func NewKnowledgeConnectHandler(documents *knowledgeuc.DocumentManager) *KnowledgeConnectHandler {
	return &KnowledgeConnectHandler{documents: documents}
}

func NewKnowledgeServiceHandler(documents *knowledgeuc.DocumentManager) (string, http.Handler) {
	handler := NewKnowledgeConnectHandler(documents)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.KnowledgeService"
	mux.Handle("/"+serviceName+"/ListKnowledge", connect.NewUnaryHandler(
		"/"+serviceName+"/ListKnowledge",
		handler.ListKnowledge,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/GetKnowledge", connect.NewUnaryHandler(
		"/"+serviceName+"/GetKnowledge",
		handler.GetKnowledge,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/CreateKnowledge", connect.NewUnaryHandler(
		"/"+serviceName+"/CreateKnowledge",
		handler.CreateKnowledge,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/UpdateKnowledge", connect.NewUnaryHandler(
		"/"+serviceName+"/UpdateKnowledge",
		handler.UpdateKnowledge,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/DeleteKnowledge", connect.NewUnaryHandler(
		"/"+serviceName+"/DeleteKnowledge",
		handler.DeleteKnowledge,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/RetryEmbed", connect.NewUnaryHandler(
		"/"+serviceName+"/RetryEmbed",
		handler.RetryEmbed,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/ListLLMConfigs", connect.NewUnaryHandler(
		"/"+serviceName+"/ListLLMConfigs",
		handler.ListLLMConfigs,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/CreateLLMConfig", connect.NewUnaryHandler(
		"/"+serviceName+"/CreateLLMConfig",
		handler.CreateLLMConfig,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/UpdateLLMConfig", connect.NewUnaryHandler(
		"/"+serviceName+"/UpdateLLMConfig",
		handler.UpdateLLMConfig,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/ActivateLLMConfig", connect.NewUnaryHandler(
		"/"+serviceName+"/ActivateLLMConfig",
		handler.ActivateLLMConfig,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/TestLLMConfig", connect.NewUnaryHandler(
		"/"+serviceName+"/TestLLMConfig",
		handler.TestLLMConfig,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/DeleteLLMConfig", connect.NewUnaryHandler(
		"/"+serviceName+"/DeleteLLMConfig",
		handler.DeleteLLMConfig,
		codecOpt,
	))

	return "/" + serviceName + "/", mux
}
