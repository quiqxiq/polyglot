package device

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/quixiq/polyglot/pkg/logger"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
)

// ProbeConnectHandler implements the probe ConnectRPC service.
type ProbeConnectHandler struct{}

// NewProbeConnectHandler constructs a probe ConnectRPC handler.
func NewProbeConnectHandler() *ProbeConnectHandler {
	return &ProbeConnectHandler{}
}

// ReportStatus acknowledges a probe heartbeat.
func (h *ProbeConnectHandler) ReportStatus(ctx context.Context, req *connect.Request[devicepb.ProbeStatusRequest]) (*connect.Response[devicepb.ProbeStatusResponse], error) {
	logger.WithComponent("ProbeServer").WithFields(map[string]any{
		"probe_id":       req.Msg.ProbeId,
		"version":        req.Msg.Version,
		"uptime_seconds": req.Msg.UptimeSeconds,
	}).Info("probe heartbeat received")

	return connect.NewResponse(&devicepb.ProbeStatusResponse{
		Acknowledged:   true,
		ServerTimeUnix: time.Now().Unix(),
	}), nil
}

// StreamTelemetry receives telemetry from a remote probe.
func (h *ProbeConnectHandler) StreamTelemetry(ctx context.Context, stream *connect.BidiStream[devicepb.PingTelemetry, devicepb.ProbeControlCommand]) error {
	for {
		msg, err := stream.Receive()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		logger.WithComponent("ProbeServer").WithFields(map[string]any{
			"probe_id":   msg.ProbeId,
			"target_ip":  msg.TargetIp,
			"latency_ms": msg.LatencyMs,
			"is_alive":   msg.IsAlive,
		}).Info("probe telemetry received")
	}
}

// NewProbeServiceHandler creates the probe ConnectRPC service handler.
func NewProbeServiceHandler() (string, http.Handler) {
	handler := NewProbeConnectHandler()
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.ProbeService"
	mux.Handle("/"+serviceName+"/ReportStatus", connect.NewUnaryHandler(
		"/"+serviceName+"/ReportStatus",
		handler.ReportStatus,
		opts...,
	))
	mux.Handle("/"+serviceName+"/StreamTelemetry", connect.NewBidiStreamHandler(
		"/"+serviceName+"/StreamTelemetry",
		handler.StreamTelemetry,
		opts...,
	))

	return "/" + serviceName + "/", mux
}
