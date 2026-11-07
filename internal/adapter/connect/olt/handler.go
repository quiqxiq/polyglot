// Package olt memasang ConnectRPC OltService: inventori OLT/ONU HiOSO,
// sync, riwayat optik, dan aksi tulis async (job queue).
package olt

import (
	"context"
	"strconv"
	"time"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	domainOLT "github.com/quixiq/polyglot/internal/domain/olt"
	"github.com/quixiq/polyglot/internal/domain/ont"
	"github.com/quixiq/polyglot/internal/port"
	oltUC "github.com/quixiq/polyglot/internal/usecase/olt"
	"github.com/quixiq/polyglot/pkg/response"
)

// ConnectHandler implements OltService.
type ConnectHandler struct {
	svc      *oltUC.Service
	settings port.SettingReader
}

// NewConnectHandler wires the OLT handler.
func NewConnectHandler(svc *oltUC.Service, settings port.SettingReader) *ConnectHandler {
	return &ConnectHandler{svc: svc, settings: settings}
}

func toProtoOLT(o oltUC.OLT) *polyglotv1.OltDevice {
	return &polyglotv1.OltDevice{
		Id:          o.DeviceID,
		Name:        o.Name,
		Host:        o.Host,
		TelnetPort:  int32(o.TelnetPort),
		SshPort:     int32(o.SSHPort),
		SnmpPort:    int32(o.SNMPPort),
		Transport:   o.Transport,
		Variant:     o.Variant,
		PonCount:    int32(o.PONCount),
		Enabled:     o.Enabled,
		TotalOnus:   int32(o.Counts.Total),
		OnlineOnus:  int32(o.Counts.Online),
		OfflineOnus: int32(o.Counts.Offline),
		UnknownOnus: int32(o.Counts.Unknown),
		RemovedOnus: int32(o.Counts.Removed),
	}
}

func toProtoHealth(h oltUC.Health) *polyglotv1.OltHealth {
	return &polyglotv1.OltHealth{
		Reachable: h.Reachable,
		Variant:   h.Variant,
		Firmware:  h.Firmware,
		SysDescr:  h.SysDescr,
		Uptime:    h.Uptime,
		LatencyMs: h.LatencyMS,
	}
}

func (h *ConnectHandler) rxClass(ctx context.Context, rx *float64) string {
	warning := 27.0
	critical := 30.0
	if v, err := strconv.ParseFloat(h.settings.GetValue(ctx, "olt.rx_warning_dbm", "-27"), 64); err == nil {
		warning = v
	}
	if v, err := strconv.ParseFloat(h.settings.GetValue(ctx, "olt.rx_critical_dbm", "-30"), 64); err == nil {
		critical = v
	}
	return string(ont.ClassifyRx(rx, warning, critical))
}

func (h *ConnectHandler) toProtoONU(ctx context.Context, o ont.ONT) *polyglotv1.Onu {
	pb := &polyglotv1.Onu{
		Id:             o.ID,
		OltDeviceId:    o.OLTDeviceID,
		PonIndex:       int32(o.PONIndex),
		OnuIndex:       int32(o.ONUIndex),
		MacAddress:     o.MACAddress,
		SerialNumber:   o.SerialNumber,
		Name:           o.Name,
		Model:          o.Model,
		Firmware:       o.Firmware,
		State:          string(o.State),
		AdminState:     string(o.AdminState),
		MissedPolls:    int32(o.MissedPolls),
		OfflineStrikes: int32(o.OfflineStrikes),
		RxClass:        h.rxClass(ctx, o.RxPowerDBM),
	}
	if o.LLID != nil {
		pb.Llid = int32(*o.LLID)
	}
	if o.RxPowerDBM != nil {
		v := *o.RxPowerDBM
		pb.RxPowerDbm = &v
	}
	if o.TxPowerDBM != nil {
		v := *o.TxPowerDBM
		pb.TxPowerDbm = &v
	}
	if o.DistanceM != nil {
		v := int32(*o.DistanceM)
		pb.DistanceM = &v
	}
	if o.TemperatureC != nil {
		v := *o.TemperatureC
		pb.TemperatureC = &v
	}
	if o.LastSeenAt != nil {
		pb.LastSeenAt = o.LastSeenAt.Format(time.RFC3339)
	}
	if o.LastPolledAt != nil {
		pb.LastPolledAt = o.LastPolledAt.Format(time.RFC3339)
	}
	if o.CustomerID != nil {
		pb.CustomerId = *o.CustomerID
	}
	if o.SubscriptionID != nil {
		pb.SubscriptionId = *o.SubscriptionID
	}
	pb.PppoeUsername = o.PPPoEUsername
	if o.ACSDeviceID != nil {
		pb.AcsDeviceId = *o.ACSDeviceID
	}
	if o.ACSLastInform != nil {
		pb.AcsLastInformAt = o.ACSLastInform.Format(time.RFC3339)
	}
	return pb
}

func (h *ConnectHandler) ListOlts(ctx context.Context, _ *connect.Request[polyglotv1.ListOltsRequest]) (*connect.Response[polyglotv1.ListOltsResponse], error) {
	list, err := h.svc.List(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*polyglotv1.OltDevice, 0, len(list))
	for _, o := range list {
		out = append(out, toProtoOLT(o))
	}
	return connect.NewResponse(&polyglotv1.ListOltsResponse{Olts: out}), nil
}

func (h *ConnectHandler) GetOlt(ctx context.Context, req *connect.Request[polyglotv1.GetOltRequest]) (*connect.Response[polyglotv1.GetOltResponse], error) {
	o, err := h.svc.Get(ctx, req.Msg.GetId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	health, errProbe := h.svc.TestConnection(ctx, req.Msg.GetId())
	if errProbe != nil {
		health = oltUC.Health{Reachable: false}
	}
	return connect.NewResponse(&polyglotv1.GetOltResponse{Olt: toProtoOLT(o), Health: toProtoHealth(health)}), nil
}

func (h *ConnectHandler) CreateOlt(ctx context.Context, req *connect.Request[polyglotv1.CreateOltRequest]) (*connect.Response[polyglotv1.CreateOltResponse], error) {
	in := oltUC.Input{
		Name: req.Msg.GetName(), Host: req.Msg.GetHost(),
		TelnetPort: int(req.Msg.GetTelnetPort()), SSHPort: int(req.Msg.GetSshPort()), SNMPPort: int(req.Msg.GetSnmpPort()),
		Transport: req.Msg.GetTransport(), Variant: req.Msg.GetVariant(), PONCount: int(req.Msg.GetPonCount()),
		Username: req.Msg.GetUsername(), Password: req.Msg.GetPassword(),
		CommunityRO: req.Msg.GetCommunityRo(), CommunityRW: req.Msg.GetCommunityRw(),
		Enabled: req.Msg.GetEnabled(),
	}
	o, err := h.svc.Create(ctx, in)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.CreateOltResponse{Olt: toProtoOLT(o)}), nil
}

func (h *ConnectHandler) UpdateOlt(ctx context.Context, req *connect.Request[polyglotv1.UpdateOltRequest]) (*connect.Response[polyglotv1.UpdateOltResponse], error) {
	in := oltUC.Input{
		Name: req.Msg.GetName(), Host: req.Msg.GetHost(),
		TelnetPort: int(req.Msg.GetTelnetPort()), SSHPort: int(req.Msg.GetSshPort()), SNMPPort: int(req.Msg.GetSnmpPort()),
		Transport: req.Msg.GetTransport(), Variant: req.Msg.GetVariant(), PONCount: int(req.Msg.GetPonCount()),
		Username: req.Msg.GetUsername(), Password: req.Msg.GetPassword(),
		CommunityRO: req.Msg.GetCommunityRo(), CommunityRW: req.Msg.GetCommunityRw(),
		Enabled: req.Msg.GetEnabled(),
	}
	o, err := h.svc.Update(ctx, req.Msg.GetId(), in)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.UpdateOltResponse{Olt: toProtoOLT(o)}), nil
}

func (h *ConnectHandler) DeleteOlt(ctx context.Context, req *connect.Request[polyglotv1.DeleteOltRequest]) (*connect.Response[polyglotv1.DeleteOltResponse], error) {
	if err := h.svc.Delete(ctx, req.Msg.GetId()); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.DeleteOltResponse{Message: "OLT dihapus"}), nil
}

func (h *ConnectHandler) TestOltConnection(ctx context.Context, req *connect.Request[polyglotv1.TestOltConnectionRequest]) (*connect.Response[polyglotv1.TestOltConnectionResponse], error) {
	health, err := h.svc.TestConnection(ctx, req.Msg.GetId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.TestOltConnectionResponse{Health: toProtoHealth(health)}), nil
}

func (h *ConnectHandler) ListOnus(ctx context.Context, req *connect.Request[polyglotv1.ListOnusRequest]) (*connect.Response[polyglotv1.ListOnusResponse], error) {
	filter := port.ONTFilter{
		OLTDeviceID: req.Msg.GetOltDeviceId(),
		PONIndex:    int(req.Msg.GetPonIndex()),
		State:       req.Msg.GetState(),
		Search:      req.Msg.GetSearch(),
		Limit:       int(req.Msg.GetLimit()),
		Offset:      int(req.Msg.GetOffset()),
	}
	list, err := h.svc.ListONTs(ctx, filter)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*polyglotv1.Onu, 0, len(list))
	for _, o := range list {
		out = append(out, h.toProtoONU(ctx, o))
	}
	return connect.NewResponse(&polyglotv1.ListOnusResponse{Onus: out, Total: int32(len(out))}), nil
}

func (h *ConnectHandler) GetOnu(ctx context.Context, req *connect.Request[polyglotv1.GetOnuRequest]) (*connect.Response[polyglotv1.GetOnuResponse], error) {
	o, err := h.svc.GetONT(ctx, req.Msg.GetId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.GetOnuResponse{Onu: h.toProtoONU(ctx, o)}), nil
}

func (h *ConnectHandler) ListOnuEvents(ctx context.Context, req *connect.Request[polyglotv1.ListOnuEventsRequest]) (*connect.Response[polyglotv1.ListOnuEventsResponse], error) {
	events, err := h.svc.ListONTEvents(ctx, req.Msg.GetOnuId(), int(req.Msg.GetLimit()))
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*polyglotv1.OnuEvent, 0, len(events))
	for _, e := range events {
		out = append(out, &polyglotv1.OnuEvent{
			Id: e.ID, OnuId: e.ONTID, EventType: e.EventType,
			OldState: e.OldState, NewState: e.NewState, Message: e.Message,
			CreatedAt: e.CreatedAt.Format(time.RFC3339),
		})
	}
	return connect.NewResponse(&polyglotv1.ListOnuEventsResponse{Events: out}), nil
}

func (h *ConnectHandler) GetOnuOpticalHistory(ctx context.Context, req *connect.Request[polyglotv1.GetOnuOpticalHistoryRequest]) (*connect.Response[polyglotv1.GetOnuOpticalHistoryResponse], error) {
	hours := int(req.Msg.GetHours())
	if hours <= 0 {
		hours = 24
	}
	from := time.Now().Add(-time.Duration(hours) * time.Hour)
	samples, err := h.svc.ListONTSamples(ctx, req.Msg.GetOnuId(), from, 500)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	out := make([]*polyglotv1.OnuMetricSample, 0, len(samples))
	for _, s := range samples {
		item := &polyglotv1.OnuMetricSample{SampledAt: s.SampledAt.Format(time.RFC3339), Online: s.Online}
		if s.RxPowerDBM != nil {
			v := *s.RxPowerDBM
			item.RxPowerDbm = &v
		}
		if s.TxPowerDBM != nil {
			v := *s.TxPowerDBM
			item.TxPowerDbm = &v
		}
		if s.TemperatureC != nil {
			v := *s.TemperatureC
			item.TemperatureC = &v
		}
		if s.DistanceM != nil {
			v := int32(*s.DistanceM)
			item.DistanceM = &v
		}
		out = append(out, item)
	}
	return connect.NewResponse(&polyglotv1.GetOnuOpticalHistoryResponse{Samples: out}), nil
}

func (h *ConnectHandler) SyncOlt(ctx context.Context, req *connect.Request[polyglotv1.SyncOltRequest]) (*connect.Response[polyglotv1.SyncOltResponse], error) {
	job, err := h.svc.EnqueueSync(ctx, req.Msg.GetOltDeviceId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.SyncOltResponse{JobId: job.ID}), nil
}

func (h *ConnectHandler) SaveOltConfig(ctx context.Context, req *connect.Request[polyglotv1.SaveOltConfigRequest]) (*connect.Response[polyglotv1.SaveOltConfigResponse], error) {
	job, err := h.svc.EnqueueSaveConfig(ctx, req.Msg.GetOltDeviceId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.SaveOltConfigResponse{JobId: job.ID}), nil
}

func (h *ConnectHandler) onuPosition(ctx context.Context, onuID string) (ont.ONT, error) {
	o, err := h.svc.GetONT(ctx, onuID)
	if err != nil {
		return ont.ONT{}, err
	}
	if o.OLTDeviceID == "" {
		return ont.ONT{}, domainOLT.ErrNotFound
	}
	return o, nil
}

func (h *ConnectHandler) RenameOnu(ctx context.Context, req *connect.Request[polyglotv1.RenameOnuRequest]) (*connect.Response[polyglotv1.RenameOnuResponse], error) {
	o, err := h.onuPosition(ctx, req.Msg.GetOnuId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	job, err := h.svc.EnqueueRename(ctx, o.OLTDeviceID, o.PONIndex, o.ONUIndex, req.Msg.GetName())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.RenameOnuResponse{JobId: job.ID}), nil
}

func (h *ConnectHandler) RebootOnu(ctx context.Context, req *connect.Request[polyglotv1.RebootOnuRequest]) (*connect.Response[polyglotv1.RebootOnuResponse], error) {
	o, err := h.onuPosition(ctx, req.Msg.GetOnuId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	job, err := h.svc.EnqueueReboot(ctx, o.OLTDeviceID, o.PONIndex, o.ONUIndex)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.RebootOnuResponse{JobId: job.ID}), nil
}

func (h *ConnectHandler) SetOnuAdminState(ctx context.Context, req *connect.Request[polyglotv1.SetOnuAdminStateRequest]) (*connect.Response[polyglotv1.SetOnuAdminStateResponse], error) {
	o, err := h.onuPosition(ctx, req.Msg.GetOnuId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	job, err := h.svc.EnqueueSetAdmin(ctx, o.OLTDeviceID, o.PONIndex, o.ONUIndex, req.Msg.GetEnabled())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.SetOnuAdminStateResponse{JobId: job.ID}), nil
}

func (h *ConnectHandler) DeleteOnu(ctx context.Context, req *connect.Request[polyglotv1.DeleteOnuRequest]) (*connect.Response[polyglotv1.DeleteOnuResponse], error) {
	o, err := h.onuPosition(ctx, req.Msg.GetOnuId())
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	job, err := h.svc.EnqueueDelete(ctx, o.OLTDeviceID, o.PONIndex, o.ONUIndex)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.DeleteOnuResponse{JobId: job.ID}), nil
}
