package setting

import (
	"context"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
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

func (h *SettingConnectHandler) GetAllSettings(ctx context.Context, req *connect.Request[polyglotv1.GetAllSettingsRequest]) (*connect.Response[polyglotv1.GetAllSettingsResponse], error) {
	settings, err := h.settingUC.GetAllSettings(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.GetAllSettingsResponse{
		Settings: DomainSettingsToPb(settings),
	}), nil
}

func (h *SettingConnectHandler) GetSettingsByCategory(ctx context.Context, req *connect.Request[polyglotv1.GetSettingsByCategoryRequest]) (*connect.Response[polyglotv1.GetSettingsByCategoryResponse], error) {
	settings, err := h.settingUC.GetSettingsByCategory(ctx, req.Msg.Category)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.GetSettingsByCategoryResponse{
		Settings: DomainSettingsToPb(settings),
	}), nil
}

func (h *SettingConnectHandler) UpdateSetting(ctx context.Context, req *connect.Request[polyglotv1.UpdateSettingRequest]) (*connect.Response[polyglotv1.UpdateSettingResponse], error) {
	updated, err := h.settingUC.UpdateSetting(ctx, req.Msg.Key, req.Msg.Value)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.UpdateSettingResponse{
		Setting: DomainSettingToPb(updated),
	}), nil
}

func (h *SettingConnectHandler) BatchUpdateSettings(ctx context.Context, req *connect.Request[polyglotv1.BatchUpdateSettingsRequest]) (*connect.Response[polyglotv1.BatchUpdateSettingsResponse], error) {
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

	return connect.NewResponse(&polyglotv1.BatchUpdateSettingsResponse{
		Settings: DomainSettingsToPb(allSettings),
	}), nil
}

func (h *SettingConnectHandler) GetBotSettings(ctx context.Context, req *connect.Request[polyglotv1.GetBotSettingsRequest]) (*connect.Response[polyglotv1.GetBotSettingsResponse], error) {
	s, err := h.settingUC.GetBotSettings(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.GetBotSettingsResponse{
		Settings: DomainBotSettingsToPb(s),
	}), nil
}

func (h *SettingConnectHandler) UpdateBotSettings(ctx context.Context, req *connect.Request[polyglotv1.UpdateBotSettingsRequest]) (*connect.Response[polyglotv1.UpdateBotSettingsResponse], error) {
	domainSettings := PbBotSettingsToDomain(req.Msg.Settings)
	if err := h.settingUC.UpdateBotSettings(ctx, domainSettings); err != nil {
		return nil, response.MapDomainError(err)
	}

	latest, err := h.settingUC.GetBotSettings(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.UpdateBotSettingsResponse{
		Settings: DomainBotSettingsToPb(latest),
	}), nil
}
