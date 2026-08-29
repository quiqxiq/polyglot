package device

import (
	"context"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	iauth "github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/pkg/response"
)

// GetDevicePingConfig handles fetching ping configuration and TimescaleDB availability.
func (h *DeviceConnectHandler) GetDevicePingConfig(
	ctx context.Context,
	req *connect.Request[devicepb.GetDevicePingConfigRequest],
) (*connect.Response[devicepb.GetDevicePingConfigResponse], error) {
	callerID, callerRoles, _ := iauth.IdentityFromContext(ctx)
	cfg, timescaleAvailable, err := h.metricsUC.GetPingConfig(ctx, req.Msg.DeviceId, callerID, callerRoles)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	res := &devicepb.GetDevicePingConfigResponse{
		Config: &devicepb.DevicePingConfig{
			Enabled:       cfg.Enabled,
			Target:        cfg.Target,
			RetentionDays: int32(cfg.RetentionDays),
		},
		TimescaledbAvailable: timescaleAvailable,
	}
	return connect.NewResponse(res), nil
}

// UpdateDevicePingConfig handles updating ping settings for a router.
func (h *DeviceConnectHandler) UpdateDevicePingConfig(
	ctx context.Context,
	req *connect.Request[devicepb.UpdateDevicePingConfigRequest],
) (*connect.Response[devicepb.UpdateDevicePingConfigResponse], error) {
	callerID, callerRoles, _ := iauth.IdentityFromContext(ctx)
	cfg := device.DevicePingConfig{
		Enabled:       req.Msg.Config.Enabled,
		Target:        req.Msg.Config.Target,
		RetentionDays: int(req.Msg.Config.RetentionDays),
	}

	savedCfg, err := h.metricsUC.UpdatePingConfig(ctx, req.Msg.DeviceId, cfg, callerID, callerRoles)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	res := &devicepb.UpdateDevicePingConfigResponse{
		Config: &devicepb.DevicePingConfig{
			Enabled:       savedCfg.Enabled,
			Target:        savedCfg.Target,
			RetentionDays: int32(savedCfg.RetentionDays),
		},
		Message: "Ping configuration updated successfully",
	}
	return connect.NewResponse(res), nil
}

// QueryDevicePingMetrics retrieves historical time-series ping telemetry within the filter window.
func (h *DeviceConnectHandler) QueryDevicePingMetrics(
	ctx context.Context,
	req *connect.Request[devicepb.QueryDevicePingMetricsRequest],
) (*connect.Response[devicepb.QueryDevicePingMetricsResponse], error) {
	var startTime, endTime time.Time
	if req.Msg.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, req.Msg.StartTime); err == nil {
			startTime = t
		}
	}
	if req.Msg.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, req.Msg.EndTime); err == nil {
			endTime = t
		}
	}

	filter := device.PingMetricsFilter{
		DeviceID:       req.Msg.DeviceId,
		StartTime:      startTime,
		EndTime:        endTime,
		BucketInterval: req.Msg.BucketInterval,
	}

	callerID, callerRoles, _ := iauth.IdentityFromContext(ctx)
	points, summary, timescaleAvailable, err := h.metricsUC.QueryPingMetrics(ctx, filter, callerID, callerRoles)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	pbPoints := make([]*devicepb.PingMetricPointData, len(points))
	for i, p := range points {
		var minRTT, avgRTT, maxRTT float32
		if p.MinRTTMS != nil {
			minRTT = *p.MinRTTMS
		}
		if p.AvgRTTMS != nil {
			avgRTT = *p.AvgRTTMS
		}
		if p.MaxRTTMS != nil {
			maxRTT = *p.MaxRTTMS
		}

		pbPoints[i] = &devicepb.PingMetricPointData{
			Timestamp:  p.RecordedAt.Format(time.RFC3339),
			Target:     p.Target,
			Seq:        int32(p.Seq),
			Size:       int32(p.Size),
			Ttl:        int32(p.TTL),
			RttMs:      p.RTTMS,
			Status:     p.Status,
			Sent:       int32(p.Sent),
			Received:   int32(p.Received),
			PacketLoss: int32(p.PacketLoss),
			MinRttMs:   minRTT,
			AvgRttMs:   avgRTT,
			MaxRttMs:   maxRTT,
		}
	}

	res := &devicepb.QueryDevicePingMetricsResponse{
		Points:               pbPoints,
		MinRtt:               summary.MinRTT,
		AvgRtt:               summary.AvgRTT,
		MaxRtt:               summary.MaxRTT,
		PacketLossPct:        summary.PacketLossPct,
		TotalSamples:         summary.TotalSamples,
		TimescaledbAvailable: timescaleAvailable,
	}
	return connect.NewResponse(res), nil
}
