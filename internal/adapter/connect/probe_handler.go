package connectadapter

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
	"github.com/quixiq/polyglot/internal/adapter/connect/codec"
	"github.com/quixiq/polyglot/pkg/logger"
)

type ProbeConnectHandler struct{}

func NewProbeConnectHandler() *ProbeConnectHandler {
	return &ProbeConnectHandler{}
}

func (h *ProbeConnectHandler) ReportStatus(ctx context.Context, req *connect.Request[devicepb.ProbeStatusRequest]) (*connect.Response[devicepb.ProbeStatusResponse], error) {
	if req.Msg.ProbeId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("probe_id is required"))
	}

	logger.FromContext(ctx).WithFields(logger.Fields{
		"probe_id":   req.Msg.ProbeId,
		"version":    req.Msg.Version,
		"uptime_sec": req.Msg.UptimeSeconds,
	}).Info("Received heartbeat from probe")

	return connect.NewResponse(&devicepb.ProbeStatusResponse{
		Acknowledged:   true,
		ServerTimeUnix: time.Now().Unix(),
	}), nil
}

func (h *ProbeConnectHandler) StreamTelemetry(ctx context.Context, stream *connect.BidiStream[devicepb.PingTelemetry, devicepb.ProbeControlCommand]) error {
	for {
		msg, err := stream.Receive()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		logger.FromContext(ctx).WithFields(logger.Fields{
			"probe_id":   msg.ProbeId,
			"target_ip":  msg.TargetIp,
			"latency_ms": msg.LatencyMs,
			"alive":      msg.IsAlive,
		}).Debug("Telemetry received from probe")
	}
}

func NewProbeServiceHandler() (string, http.Handler) {
	handler := NewProbeConnectHandler()
	mux := http.NewServeMux()
	codecOpt := codec.Option()

	serviceName := "polyglot.v1.ProbeService"
	mux.Handle("/"+serviceName+"/ReportStatus", connect.NewUnaryHandler(
		"/"+serviceName+"/ReportStatus",
		handler.ReportStatus,
		codecOpt,
	))
	mux.Handle("/"+serviceName+"/StreamTelemetry", connect.NewBidiStreamHandler(
		"/"+serviceName+"/StreamTelemetry",
		handler.StreamTelemetry,
		codecOpt,
	))

	return "/" + serviceName + "/", mux
}
