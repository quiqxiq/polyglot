package device

import (
	"context"
	"fmt"
	"io"
	"github.com/quixiq/polyglot/pkg/logger"
	"net/http"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
)

type ProbeConnectHandler struct{}

func NewProbeConnectHandler() *ProbeConnectHandler {
	return &ProbeConnectHandler{}
}

func (h *ProbeConnectHandler) ReportStatus(ctx context.Context, req *connect.Request[devicepb.ProbeStatusRequest]) (*connect.Response[devicepb.ProbeStatusResponse], error) {
	if req.Msg.ProbeId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("probe_id is required"))
	}

	logger.WithComponent("ProbeServer").Infof("Received heartbeat from Probe %s (Version: %s, Uptime: %ds)", req.Msg.ProbeId, req.Msg.Version, req.Msg.UptimeSeconds)

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

		logger.WithComponent("ProbeServer").Infof("Telemetry from probe %s: Target=%s Latency=%dms Alive=%v",
			msg.ProbeId, msg.TargetIp, msg.LatencyMs, msg.IsAlive)
	}
}

func NewProbeServiceHandler() (string, http.Handler) {
	handler := NewProbeConnectHandler()
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

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
