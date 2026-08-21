package setting

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	settingUC "github.com/quixiq/polyglot/internal/usecase/setting"
)

// NewSettingServiceHandler creates the Connect http.Handler for SettingService.
func NewSettingServiceHandler(uc *settingUC.ManageSettingUseCase) (string, http.Handler) {
	handler := NewSettingConnectHandler(uc)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.SettingService"
	mux.Handle("/"+serviceName+"/GetAllSettings", connect.NewUnaryHandler("/"+serviceName+"/GetAllSettings", handler.GetAllSettings, codecOpt))
	mux.Handle("/"+serviceName+"/GetSettingsByCategory", connect.NewUnaryHandler("/"+serviceName+"/GetSettingsByCategory", handler.GetSettingsByCategory, codecOpt))
	mux.Handle("/"+serviceName+"/UpdateSetting", connect.NewUnaryHandler("/"+serviceName+"/UpdateSetting", handler.UpdateSetting, codecOpt))
	mux.Handle("/"+serviceName+"/BatchUpdateSettings", connect.NewUnaryHandler("/"+serviceName+"/BatchUpdateSettings", handler.BatchUpdateSettings, codecOpt))
	mux.Handle("/"+serviceName+"/GetBotSettings", connect.NewUnaryHandler("/"+serviceName+"/GetBotSettings", handler.GetBotSettings, codecOpt))
	mux.Handle("/"+serviceName+"/UpdateBotSettings", connect.NewUnaryHandler("/"+serviceName+"/UpdateBotSettings", handler.UpdateBotSettings, codecOpt))

	return "/" + serviceName + "/", mux
}
