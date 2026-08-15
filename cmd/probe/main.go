package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
	"github.com/quixiq/polyglot/pkg/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverURL := envOr("PROBE_SERVER_URL", "http://localhost:8080")
	probeID := envOr("PROBE_ID", "pop-remote-01")

	logger.WithFields(logger.Fields{
		"probe_id":   probeID,
		"server_url": serverURL,
	}).Info("[POP Remote Probe Agent] Starting lightweight probe agent...")

	client := newJSONConnectClient(serverURL)

	// Heartbeat goroutine
	go runHeartbeatLoop(ctx, client, probeID)

	// Telemetry streaming goroutine
	go runTelemetryLoop(ctx, client, probeID)

	<-ctx.Done()
	logger.Info("[POP Remote Probe Agent] Shutting down probe agent...")
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
				logger.WithError(err).Warn("[Probe Heartbeat Warning] Failed to report status")
			} else {
				logger.WithFields(logger.Fields{
					"ack":         resp.Msg.Acknowledged,
					"server_time": resp.Msg.ServerTimeUnix,
				}).Info("[Probe Heartbeat OK] Server ack received")
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

			// Simulate ICMP Ping measurement
			latency := int64(5 + (idx * 3) % 25)
			logger.WithFields(logger.Fields{
				"target":     target,
				"latency_ms": latency,
				"alive":      true,
			}).Info("[Probe Telemetry] Polled target")
		}
	}
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
