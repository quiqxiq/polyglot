package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/pkg/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logLevel := envOr("LOG_LEVEL", "info")
	appEnv := envOr("APP_ENV", "production")
	logger.Init(logLevel, appEnv)

	serverURL := envOr("PROBE_SERVER_URL", "http://localhost:8080")
	probeID := envOr("PROBE_ID", "pop-remote-01")

	logger.WithComponent("ProbeAgent").WithFields(map[string]any{
		"probe_id":   probeID,
		"server_url": serverURL,
	}).Info("starting lightweight probe agent")

	client := newJSONConnectClient(serverURL)

	// Heartbeat goroutine
	go runHeartbeatLoop(ctx, client, probeID)

	// Telemetry streaming goroutine
	go runTelemetryLoop(ctx, client, probeID)

	<-ctx.Done()
	logger.WithComponent("ProbeAgent").Info("shutting down probe agent")
}

func runHeartbeatLoop(ctx context.Context, client *probeClient, probeID string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			uptimeSec := int64(time.Since(startTime).Seconds())
			req := connect.NewRequest(&devicepb.ProbeStatusRequest{
				ProbeId:       probeID,
				Version:       "v1.0.0-probe",
				UptimeSeconds: uptimeSec,
			})
			resp, err := client.ReportStatus(ctx, req)
			if err != nil {
				logger.WithComponent("ProbeAgent").WithError(err).Warn("failed to report probe status")
			} else {
				logger.WithComponent("ProbeAgent").WithFields(map[string]any{
					"acknowledged":     resp.Msg.Acknowledged,
					"server_time_unix": resp.Msg.ServerTimeUnix,
				}).Debug("probe heartbeat acknowledged")
			}
		}
	}
}

func runTelemetryLoop(ctx context.Context, client *probeClient, probeID string) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	targetIPs := []string{"1.1.1.1", "8.8.8.8", "192.168.1.1"}
	idx := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			target := targetIPs[idx%len(targetIPs)]
			idx++

			latency, alive := measureTargetLatency(target)
			logger.WithComponent("ProbeAgent").WithFields(map[string]any{
				"target":     target,
				"latency_ms": latency,
				"alive":      alive,
			}).Debug("probe telemetry polled target")
		}
	}
}

func measureTargetLatency(target string) (int64, bool) {
	start := time.Now()
	port := "53"
	if target == "192.168.1.1" {
		port = "80"
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(target, port), 2*time.Second)
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		return elapsed, false
	}
	_ = conn.Close()
	return elapsed, true
}

type probeClient struct {
	baseURL    string
	httpClient *http.Client
}

func newJSONConnectClient(baseURL string) *probeClient {
	return &probeClient{
		baseURL:    baseURL,
		httpClient: http.DefaultClient,
	}
}

func (c *probeClient) ReportStatus(ctx context.Context, req *connect.Request[devicepb.ProbeStatusRequest]) (*connect.Response[devicepb.ProbeStatusResponse], error) {
	serviceName := "polyglot.v1.ProbeService"
	url := fmt.Sprintf("%s/%s/ReportStatus", c.baseURL, serviceName)

	client := connect.NewClient[devicepb.ProbeStatusRequest, devicepb.ProbeStatusResponse](
		c.httpClient,
		url,
	)

	return client.CallUnary(ctx, req)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
