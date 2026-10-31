// Package ont memasang ConnectRPC OntService: konfigurasi ACS, matching
// device TR-069, telemetri, aksi remote, dan penautan ONT ke pelanggan.
package ont

import (
	"context"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
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

func (h *ConnectHandler) GetAcsConfig(ctx context.Context, _ *connect.Request[devicepb.GetAcsConfigRequest]) (*connect.Response[devicepb.GetAcsConfigResponse], error) {
	cfg := h.svc.Config(ctx)
	return connect.NewResponse(&devicepb.GetAcsConfigResponse{Config: &devicepb.AcsConfig{
		Enabled:                cfg.Enabled,
		ServerDeviceId:         cfg.ServerDeviceID,
		AcsUrl:                 cfg.ACSURL,
		PeriodicInformInterval: int32(cfg.PeriodicInform),
		PresetName:             cfg.PresetName,
	}}), nil
}

func (h *ConnectHandler) SyncAcsDevices(ctx context.Context, _ *connect.Request[devicepb.SyncAcsDevicesRequest]) (*connect.Response[devicepb.SyncAcsDevicesResponse], error) {
	job, err := h.svc.EnqueueSyncDevices(ctx, "")
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.SyncAcsDevicesResponse{JobId: job.ID}), nil
}

func (h *ConnectHandler) ListAcsDevices(ctx context.Context, req *connect.Request[devicepb.ListAcsDevicesRequest]) (*connect.Response[devicepb.ListAcsDevicesResponse], error) {
	limit := int(req.Msg.GetLimit())
	if limit <= 0 {
		limit = 200
	}
	list, err := h.svc.ListACSDevices(ctx, req.Msg.GetOnlyUnmatched(), req.Msg.GetSearch(), limit)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*devicepb.AcsDevice, 0, len(list))
	for _, d := range list {
		item := &devicepb.AcsDevice{
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
	return connect.NewResponse(&devicepb.ListAcsDevicesResponse{Devices: out}), nil
}

func (h *ConnectHandler) RefreshOntTelemetry(ctx context.Context, req *connect.Request[devicepb.RefreshOntTelemetryRequest]) (*connect.Response[devicepb.RefreshOntTelemetryResponse], error) {
	job, err := h.svc.EnqueueRefreshTelemetry(ctx, req.Msg.GetOntId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.RefreshOntTelemetryResponse{JobId: job.ID}), nil
}

func (h *ConnectHandler) RebootOnt(ctx context.Context, req *connect.Request[devicepb.RebootOntRequest]) (*connect.Response[devicepb.RebootOntResponse], error) {
	job, err := h.svc.EnqueueReboot(ctx, req.Msg.GetOntId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.RebootOntResponse{JobId: job.ID}), nil
}

func (h *ConnectHandler) SetOntWifi(ctx context.Context, req *connect.Request[devicepb.SetOntWifiRequest]) (*connect.Response[devicepb.SetOntWifiResponse], error) {
	job, err := h.svc.EnqueueSetWiFi(ctx, req.Msg.GetOntId(), req.Msg.GetSsid(), req.Msg.GetPassword(), req.Msg.GetBand())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.SetOntWifiResponse{JobId: job.ID}), nil
}

func (h *ConnectHandler) PushOntPppoe(ctx context.Context, req *connect.Request[devicepb.PushOntPppoeRequest]) (*connect.Response[devicepb.PushOntPppoeResponse], error) {
	job, err := h.svc.EnqueuePushPPPoE(ctx, req.Msg.GetOntId(), req.Msg.GetUsername(), req.Msg.GetPassword())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.PushOntPppoeResponse{JobId: job.ID}), nil
}

func (h *ConnectHandler) LinkOnt(ctx context.Context, req *connect.Request[devicepb.LinkOntRequest]) (*connect.Response[devicepb.LinkOntResponse], error) {
	err := h.svc.Link(ctx, req.Msg.GetOntId(), ontUC.LinkInput{
		CustomerID:     req.Msg.GetCustomerId(),
		SubscriptionID: req.Msg.GetSubscriptionId(),
		PPPoEUsername:  req.Msg.GetPppoeUsername(),
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.LinkOntResponse{Message: "ONT ditautkan"}), nil
}

func (h *ConnectHandler) UnlinkOnt(ctx context.Context, req *connect.Request[devicepb.UnlinkOntRequest]) (*connect.Response[devicepb.UnlinkOntResponse], error) {
	if err := h.svc.Unlink(ctx, req.Msg.GetOntId()); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.UnlinkOntResponse{Message: "tautan ONT dilepas"}), nil
}

func (h *ConnectHandler) PinAcsDevice(ctx context.Context, req *connect.Request[devicepb.PinAcsDeviceRequest]) (*connect.Response[devicepb.PinAcsDeviceResponse], error) {
	if err := h.svc.PinACSDevice(ctx, req.Msg.GetOntId(), req.Msg.GetAcsDeviceId()); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.PinAcsDeviceResponse{Message: "device ACS disematkan"}), nil
}

func (h *ConnectHandler) UnpinAcsDevice(ctx context.Context, req *connect.Request[devicepb.UnpinAcsDeviceRequest]) (*connect.Response[devicepb.UnpinAcsDeviceResponse], error) {
	if err := h.svc.UnpinACSDevice(ctx, req.Msg.GetOntId()); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.UnpinAcsDeviceResponse{Message: "pin device ACS dilepas"}), nil
}

func (h *ConnectHandler) EnsureAcsPreset(ctx context.Context, _ *connect.Request[devicepb.EnsureAcsPresetRequest]) (*connect.Response[devicepb.EnsureAcsPresetResponse], error) {
	if err := h.svc.BootstrapPreset(ctx); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.EnsureAcsPresetResponse{Message: "preset ACS disinkronkan"}), nil
}

var _ = domainONT.ErrACSNotConfigured
