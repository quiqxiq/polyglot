// Package ont memasang ConnectRPC OntService: konfigurasi ACS, matching
// device TR-069, telemetri, aksi remote, dan penautan ONT ke pelanggan.
package ont

import (
	"context"
	"time"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	domainONT "github.com/quixiq/polyglot/internal/domain/ont"
	ontUC "github.com/quixiq/polyglot/internal/usecase/ont"
	"github.com/quixiq/polyglot/pkg/response"
)

// ConnectHandler implements OntService.
type ConnectHandler struct {
	svc *ontUC.Service
}

// NewConnectHandler wires the ONT handler.
func NewConnectHandler(svc *ontUC.Service) *ConnectHandler {
	return &ConnectHandler{svc: svc}
}

func (h *ConnectHandler) GetAcsConfig(ctx context.Context, _ *connect.Request[polyglotv1.GetAcsConfigRequest]) (*connect.Response[polyglotv1.GetAcsConfigResponse], error) {
	cfg := h.svc.Config(ctx)
	return connect.NewResponse(&polyglotv1.GetAcsConfigResponse{Config: &polyglotv1.AcsConfig{
		Enabled:                cfg.Enabled,
		ServerDeviceId:         cfg.ServerDeviceID,
		AcsUrl:                 cfg.ACSURL,
		PeriodicInformInterval: int32(cfg.PeriodicInform),
		PresetName:             cfg.PresetName,
	}}), nil
}

func (h *ConnectHandler) SyncAcsDevices(ctx context.Context, _ *connect.Request[polyglotv1.SyncAcsDevicesRequest]) (*connect.Response[polyglotv1.SyncAcsDevicesResponse], error) {
	job, err := h.svc.EnqueueSyncDevices(ctx, "")
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.SyncAcsDevicesResponse{JobId: job.ID}), nil
}

func (h *ConnectHandler) ListAcsDevices(ctx context.Context, req *connect.Request[polyglotv1.ListAcsDevicesRequest]) (*connect.Response[polyglotv1.ListAcsDevicesResponse], error) {
	limit := int(req.Msg.GetLimit())
	if limit <= 0 {
		limit = 200
	}
	list, err := h.svc.ListACSDevices(ctx, req.Msg.GetOnlyUnmatched(), req.Msg.GetSearch(), limit)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*polyglotv1.AcsDevice, 0, len(list))
	for _, d := range list {
		item := &polyglotv1.AcsDevice{
			Id: d.ID, AcsDeviceId: d.ACSDeviceID, Oui: d.OUI,
			ProductClass: d.ProductClass, SerialNumber: d.SerialNumber,
			SoftwareVersion: d.SoftwareVersion, Tags: d.Tags, MacCandidates: d.MACCandidates,
			MatchSource: d.MatchSource,
		}
		if d.LastInformAt != nil {
			item.LastInformAt = d.LastInformAt.Format(time.RFC3339)
		}
		if d.MatchedONTID != nil {
			item.MatchedOntId = *d.MatchedONTID
		}
		out = append(out, item)
	}
	return connect.NewResponse(&polyglotv1.ListAcsDevicesResponse{Devices: out}), nil
}

func (h *ConnectHandler) RefreshOntTelemetry(ctx context.Context, req *connect.Request[polyglotv1.RefreshOntTelemetryRequest]) (*connect.Response[polyglotv1.RefreshOntTelemetryResponse], error) {
	job, err := h.svc.EnqueueRefreshTelemetry(ctx, req.Msg.GetOntId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.RefreshOntTelemetryResponse{JobId: job.ID}), nil
}

func (h *ConnectHandler) RebootOnt(ctx context.Context, req *connect.Request[polyglotv1.RebootOntRequest]) (*connect.Response[polyglotv1.RebootOntResponse], error) {
	job, err := h.svc.EnqueueReboot(ctx, req.Msg.GetOntId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.RebootOntResponse{JobId: job.ID}), nil
}

func (h *ConnectHandler) SetOntWifi(ctx context.Context, req *connect.Request[polyglotv1.SetOntWifiRequest]) (*connect.Response[polyglotv1.SetOntWifiResponse], error) {
	job, err := h.svc.EnqueueSetWiFi(ctx, req.Msg.GetOntId(), req.Msg.GetSsid(), req.Msg.GetPassword(), req.Msg.GetBand())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.SetOntWifiResponse{JobId: job.ID}), nil
}

func (h *ConnectHandler) PushOntPppoe(ctx context.Context, req *connect.Request[polyglotv1.PushOntPppoeRequest]) (*connect.Response[polyglotv1.PushOntPppoeResponse], error) {
	job, err := h.svc.EnqueuePushPPPoE(ctx, req.Msg.GetOntId(), req.Msg.GetUsername(), req.Msg.GetPassword())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.PushOntPppoeResponse{JobId: job.ID}), nil
}

func (h *ConnectHandler) LinkOnt(ctx context.Context, req *connect.Request[polyglotv1.LinkOntRequest]) (*connect.Response[polyglotv1.LinkOntResponse], error) {
	err := h.svc.Link(ctx, req.Msg.GetOntId(), ontUC.LinkInput{
		CustomerID:     req.Msg.GetCustomerId(),
		SubscriptionID: req.Msg.GetSubscriptionId(),
		PPPoEUsername:  req.Msg.GetPppoeUsername(),
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.LinkOntResponse{Message: "ONT ditautkan"}), nil
}

func (h *ConnectHandler) UnlinkOnt(ctx context.Context, req *connect.Request[polyglotv1.UnlinkOntRequest]) (*connect.Response[polyglotv1.UnlinkOntResponse], error) {
	if err := h.svc.Unlink(ctx, req.Msg.GetOntId()); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.UnlinkOntResponse{Message: "tautan ONT dilepas"}), nil
}

func (h *ConnectHandler) PinAcsDevice(ctx context.Context, req *connect.Request[polyglotv1.PinAcsDeviceRequest]) (*connect.Response[polyglotv1.PinAcsDeviceResponse], error) {
	if err := h.svc.PinACSDevice(ctx, req.Msg.GetOntId(), req.Msg.GetAcsDeviceId()); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.PinAcsDeviceResponse{Message: "device ACS disematkan"}), nil
}

func (h *ConnectHandler) UnpinAcsDevice(ctx context.Context, req *connect.Request[polyglotv1.UnpinAcsDeviceRequest]) (*connect.Response[polyglotv1.UnpinAcsDeviceResponse], error) {
	if err := h.svc.UnpinACSDevice(ctx, req.Msg.GetOntId()); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.UnpinAcsDeviceResponse{Message: "pin device ACS dilepas"}), nil
}

func (h *ConnectHandler) EnsureAcsPreset(ctx context.Context, _ *connect.Request[polyglotv1.EnsureAcsPresetRequest]) (*connect.Response[polyglotv1.EnsureAcsPresetResponse], error) {
	if err := h.svc.BootstrapPreset(ctx); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.EnsureAcsPresetResponse{Message: "preset ACS disinkronkan"}), nil
}

var _ = domainONT.ErrACSNotConfigured
