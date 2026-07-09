package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/quixiq/polyglot/internal/adapter/mcp"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/driver/genieacs"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/registry"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	transport := envOr("MCP_TRANSPORT", "http")
	httpAddr := envOr("MCP_HTTP_ADDR", ":8080")

	repo, vault := loadDemoDevices()
	factories := map[string]registry.DriverFactory{
		"mikrotik": func(ctx context.Context, target device.Target) (port.DeviceDriver, error) {
			return mikrotik.NewDriver(ctx, target)
		},
		"genieacs": func(ctx context.Context, target device.Target) (port.DeviceDriver, error) {
			return genieacs.NewDriver(ctx, target)
		},
	}

	reg := registry.New(repo, vault, factories)
	srv := mcp.New(reg, nil)

	switch transport {
	case "stdio":
		log.Printf("polyglot: MCP server starting on stdio transport")
		if err := srv.RunStdio(ctx); err != nil {
			log.Fatalf("mcp stdio: %v", err)
		}
	case "http":
		handler := srv.HTTPHandler()
		httpSrv := &http.Server{Addr: httpAddr, Handler: handler}
		log.Printf("polyglot: MCP server starting on http://%s/mcp", httpAddr)
		go func() {
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("mcp http: %v", err)
			}
		}()
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdownCtx)
	default:
		log.Fatalf("unknown MCP_TRANSPORT %q (use \"stdio\" or \"http\")", transport)
	}

	_ = reg.Close()
	log.Printf("polyglot: shutdown complete")
}

// envOr reads an env var, returning fallback if empty.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// loadDemoDevices builds an in-memory device repository and credential
// vault from env vars. This is a temporary composition-root shim — when
// the Postgres-backed adapters (internal/adapter/postgres + vault) are
// ready, this function is replaced with real database connections.
//
// Demo device env vars (all optional — server starts with zero devices
// if unset, and MCP tools return "device not found" until devices are
// configured via the REST API or database):
//
//	POLYGLOT_DEMO_DEVICE_ID   = device ID (e.g. "mtk-1")
//	POLYGLOT_DEMO_DRIVER_TYPE = "mikrotik" or "genieacs"
//	POLYGLOT_DEMO_HOST        = host or IP
//	POLYGLOT_DEMO_PORT        = port (0 = driver default)
//	POLYGLOT_DEMO_USERNAME    = SSH/API username
//	POLYGLOT_DEMO_PASSWORD    = SSH/API password
//	POLYGLOT_DEMO_EXTRA_*     = extra params (e.g. POLYGLOT_DEMO_EXTRA_use_tls=true)
func loadDemoDevices() (port.DeviceRepository, port.CredentialVault) {
	devices := make(map[string]device.Device)
	creds := make(map[string]device.Credentials)

	if id := os.Getenv("POLYGLOT_DEMO_DEVICE_ID"); id != "" {
		extra := loadExtraFromEnv("POLYGLOT_DEMO_EXTRA_")
		devices[id] = device.Device{
			ID:         id,
			Name:       id,
			Vendor:     envOr("POLYGLOT_DEMO_DRIVER_TYPE", "mikrotik"),
			DriverType: envOr("POLYGLOT_DEMO_DRIVER_TYPE", "mikrotik"),
			Host:       envOr("POLYGLOT_DEMO_HOST", "127.0.0.1"),
			Port:       envInt("POLYGLOT_DEMO_PORT", 0),
			TimeoutMS:  30000,
			Enabled:    true,
			Extra:      extra,
		}
		creds[id] = device.Credentials{
			Username: os.Getenv("POLYGLOT_DEMO_USERNAME"),
			Password: os.Getenv("POLYGLOT_DEMO_PASSWORD"),
		}
		log.Printf("polyglot: loaded demo device %q (%s @ %s)", id, devices[id].DriverType, devices[id].Host)
	}

	return &memRepo{devices: devices}, &memVault{creds: creds}
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return fallback
}

func loadExtraFromEnv(prefix string) map[string]string {
	extra := make(map[string]string)
	for _, env := range os.Environ() {
		if len(env) > len(prefix) && env[:len(prefix)] == prefix {
			key := env[len(prefix):]
			if eq := indexByte(key, '='); eq >= 0 {
				extra[key[:eq]] = key[eq+1:]
			}
		}
	}
	return extra
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

// memRepo is a temporary in-memory port.DeviceRepository for the composition
// root. Replace with internal/adapter/postgres when that adapter is wired.
type memRepo struct {
	devices map[string]device.Device
}

func (r *memRepo) FindByID(_ context.Context, id string) (device.Device, error) {
	d, ok := r.devices[id]
	if !ok {
		return device.Device{}, device.ErrNotFound
	}
	return d, nil
}

// memVault is a temporary in-memory port.CredentialVault for the composition
// root. Replace with internal/adapter/vault when that adapter is wired.
type memVault struct {
	creds map[string]device.Credentials
}

func (v *memVault) Get(_ context.Context, deviceID string) (device.Credentials, error) {
	c, ok := v.creds[deviceID]
	if !ok {
		return device.Credentials{}, device.ErrNotFound
	}
	return c, nil
}
