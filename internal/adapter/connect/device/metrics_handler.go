package device

import (
	"context"
	"time"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	iauth "github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/pkg/response"
)

// GetDevicePingConfig handles fetching ping configuration and TimescaleDB availability.
func (h *DeviceConnectHandler) GetDevicePingConfig(
	ctx context.Context,
	req *connect.Request[polyglotv1.GetDevicePingConfigRequest],
) (*connect.Response[polyglotv1.GetDevicePingConfigResponse], error) {
	callerID, callerRoles, _ := iauth.IdentityFromContext(ctx)
	cfg, timescaleAvailable, err := h.metricsUC.GetPingConfig(ctx, req.Msg.DeviceId, callerID, callerRoles)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	res := &polyglotv1.GetDevicePingConfigResponse{
		Config: &polyglotv1.DevicePingConfig{
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
	req *connect.Request[polyglotv1.UpdateDevicePingConfigRequest],
) (*connect.Response[polyglotv1.UpdateDevicePingConfigResponse], error) {
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

	res := &polyglotv1.UpdateDevicePingConfigResponse{
		Config: &polyglotv1.DevicePingConfig{
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
	req *connect.Request[polyglotv1.QueryDevicePingMetricsRequest],
) (*connect.Response[polyglotv1.QueryDevicePingMetricsResponse], error) {
	var startTime, endTime time.Time
	var err error
	if startTime, err = time.Parse(time.RFC3339, req.Msg.StartTime); err != nil {
		return nil, response.MapDomainError(device.ErrInvalidMetricsRange)
	}
	if endTime, err = time.Parse(time.RFC3339, req.Msg.EndTime); err != nil {
		return nil, response.MapDomainError(device.ErrInvalidMetricsRange)
	}
	if !startTime.Before(endTime) {
		return nil, response.MapDomainError(device.ErrInvalidMetricsRange)
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

	pbPoints := make([]*polyglotv1.PingMetricPointData, len(points))
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

		pbPoints[i] = &polyglotv1.PingMetricPointData{
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

	res := &polyglotv1.QueryDevicePingMetricsResponse{
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
