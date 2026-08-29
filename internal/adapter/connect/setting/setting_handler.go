package setting

import (
	"context"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/setting"
	settingUC "github.com/quixiq/polyglot/internal/usecase/setting"
	"github.com/quixiq/polyglot/pkg/response"
)

type SettingConnectHandler struct {
	settingUC *settingUC.ManageSettingUseCase
}

func NewSettingConnectHandler(settingUC *settingUC.ManageSettingUseCase) *SettingConnectHandler {
	return &SettingConnectHandler{
		settingUC: settingUC,
	}
}

func (h *SettingConnectHandler) GetAllSettings(ctx context.Context, req *connect.Request[devicepb.GetAllSettingsRequest]) (*connect.Response[devicepb.GetAllSettingsResponse], error) {
	settings, err := h.settingUC.GetAllSettings(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.GetAllSettingsResponse{
		Settings: DomainSettingsToPb(settings),
	}), nil
}

func (h *SettingConnectHandler) GetSettingsByCategory(ctx context.Context, req *connect.Request[devicepb.GetSettingsByCategoryRequest]) (*connect.Response[devicepb.GetSettingsByCategoryResponse], error) {
	if req.Msg.Category == "" {
		return nil, response.InvalidArgument("category is required")
	}

	settings, err := h.settingUC.GetSettingsByCategory(ctx, req.Msg.Category)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.GetSettingsByCategoryResponse{
		Settings: DomainSettingsToPb(settings),
	}), nil
}

func (h *SettingConnectHandler) UpdateSetting(ctx context.Context, req *connect.Request[devicepb.UpdateSettingRequest]) (*connect.Response[devicepb.UpdateSettingResponse], error) {
	if req.Msg.Key == "" {
		return nil, response.InvalidArgument("key is required")
	}

	updated, err := h.settingUC.UpdateSetting(ctx, req.Msg.Key, req.Msg.Value)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.UpdateSettingResponse{
		Setting: DomainSettingToPb(updated),
	}), nil
}

func (h *SettingConnectHandler) BatchUpdateSettings(ctx context.Context, req *connect.Request[devicepb.BatchUpdateSettingsRequest]) (*connect.Response[devicepb.BatchUpdateSettingsResponse], error) {
	domainSettings := make([]setting.Setting, 0, len(req.Msg.Settings))
	for _, pb := range req.Msg.Settings {
		if d := PbSettingToDomain(pb); d != nil {
			domainSettings = append(domainSettings, *d)
		}
	}

	if err := h.settingUC.BatchUpdateSettings(ctx, domainSettings); err != nil {
		return nil, response.MapDomainError(err)
	}

	allSettings, err := h.settingUC.GetAllSettings(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.BatchUpdateSettingsResponse{
		Settings: DomainSettingsToPb(allSettings),
	}), nil
}

func (h *SettingConnectHandler) GetBotSettings(ctx context.Context, req *connect.Request[devicepb.GetBotSettingsRequest]) (*connect.Response[devicepb.GetBotSettingsResponse], error) {
	s, err := h.settingUC.GetBotSettings(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.GetBotSettingsResponse{
		Settings: DomainBotSettingsToPb(s),
	}), nil
}

func (h *SettingConnectHandler) UpdateBotSettings(ctx context.Context, req *connect.Request[devicepb.UpdateBotSettingsRequest]) (*connect.Response[devicepb.UpdateBotSettingsResponse], error) {
	if req.Msg.Settings == nil {
		return nil, response.InvalidArgument("settings is required")
	}

	domainSettings := PbBotSettingsToDomain(req.Msg.Settings)
	if err := h.settingUC.UpdateBotSettings(ctx, domainSettings); err != nil {
		return nil, response.MapDomainError(err)
	}

	latest, err := h.settingUC.GetBotSettings(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.UpdateBotSettingsResponse{
		Settings: DomainBotSettingsToPb(latest),
	}), nil
}
