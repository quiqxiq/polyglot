package connectadapter

import (
	"net/http"

	"connectrpc.com/connect"
)

type KnowledgeConnectHandler struct{}

func NewKnowledgeConnectHandler() *KnowledgeConnectHandler {
	return &KnowledgeConnectHandler{}
}

func NewKnowledgeServiceHandler() (string, http.Handler) {
	handler := NewKnowledgeConnectHandler()
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(connectJSONCodec{})

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
	mux.Handle("/"+serviceName+"/ListTechnicians", connect.NewUnaryHandler(
		"/"+serviceName+"/ListTechnicians",
		handler.ListTechnicians,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/CreateTechnician", connect.NewUnaryHandler(
		"/"+serviceName+"/CreateTechnician",
		handler.CreateTechnician,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/UpdateTechnician", connect.NewUnaryHandler(
		"/"+serviceName+"/UpdateTechnician",
		handler.UpdateTechnician,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/ToggleTechnicianActive", connect.NewUnaryHandler(
		"/"+serviceName+"/ToggleTechnicianActive",
		handler.ToggleTechnicianActive,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/DeleteTechnician", connect.NewUnaryHandler(
		"/"+serviceName+"/DeleteTechnician",
		handler.DeleteTechnician,
		codecOpt,
	))

	return "/" + serviceName + "/", mux
}
