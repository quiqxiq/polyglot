package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	grpcAdapter "github.com/quixiq/polyglot/internal/adapter/grpc"
	mcpAdapter "github.com/quixiq/polyglot/internal/adapter/http"
	"github.com/quixiq/polyglot/internal/adapter/http/middleware"
	"github.com/quixiq/polyglot/internal/adapter/mcp"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
	redisAdapter "github.com/quixiq/polyglot/internal/adapter/redis"
	wsAdapter "github.com/quixiq/polyglot/internal/adapter/ws"
	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/driver/genieacs"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/driver/whatsapp"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/registry"
	botUC "github.com/quixiq/polyglot/internal/usecase/bot"
	"github.com/quixiq/polyglot/internal/usecase/business"
	knowledgeUC "github.com/quixiq/polyglot/internal/usecase/knowledge"
	"github.com/quixiq/polyglot/internal/usecase/network"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	httpAddr := envOr("PORT", ":"+cfg.Port)
	if httpAddr[0] != ':' {
		httpAddr = ":" + httpAddr
	}

	// 1. Initialize Database & Storage Adapters
	log.Println("[Polyglot Engine] Initializing database and cache stores...")
	pgStore, err := postgres.NewStore(cfg.DatabaseURL)
	if err != nil {
		log.Printf("[Warning] Failed to connect to PostgreSQL (%v). Features requiring database will fail until DB is online.", err)
	}

	redisStore, err := redisAdapter.NewStore(cfg.RedisURL)
	if err != nil {
		log.Printf("[Warning] Failed to connect to Redis (%v). Falling back to in-memory cache.", err)
	}

	jwtService := auth.NewJWTService(cfg.JWTSecret, cfg.JWTExpiryHours)

	var casbinEnforcer *auth.CasbinEnforcer
	if pgStore != nil {
		casbinEnforcer, err = auth.NewCasbinEnforcer(ctx, pgStore.DB())
		if err != nil {
			log.Printf("[Warning] Failed to initialize Casbin enforcer: %v", err)
		}
	}

	sseHub := wsAdapter.NewSSEHub()

	// 2. WhatsApp Gateway & AI Bot Engine Setup
	var waManager *whatsapp.SessionManager
	if pgStore != nil {
		eventHandler := whatsapp.NewEventHandler(pgStore, sseHub)
		var err error
		waManager, err = whatsapp.NewSessionManager(cfg.DatabaseURL, nil, eventHandler.OnStatusChanged)
		if err != nil {
			log.Printf("[Warning] Failed to initialize WhatsApp SessionManager: %v", err)
		}

		convService := business.NewConversationService(pgStore)
		knowledgeRetriever := knowledgeUC.NewKeywordRetriever(pgStore)
		botEngine := botUC.NewEngine(cfg, pgStore, redisStore, waManager, convService, knowledgeRetriever, sseHub)

		if waManager != nil {
			sessions, err := pgStore.FindAllSessions()
			if err == nil && len(sessions) > 0 {
				_ = waManager.RestoreAllSessions(sessions)
			}
			_ = botEngine
		}
	}

	// 3. Network & Infrastructure Drivers Setup
	repo, vault := loadInitialDevices(pgStore)
	factories := map[string]registry.DriverFactory{
		"mikrotik": func(ctx context.Context, target device.Target) (port.DeviceDriver, error) {
			return mikrotik.NewDriver(ctx, target)
		},
		"genieacs": func(ctx context.Context, target device.Target) (port.DeviceDriver, error) {
			return genieacs.NewDriver(ctx, target)
		},
	}

	reg := registry.New(repo, vault, factories)

	// Usecases
	deviceUC := business.NewManageDeviceUseCase(repo, vault, reg)
	mikhmonUC := network.NewMikhmonUseCase("internal/templates")

	// Driver Providers
	httpDriverProvider := func(c *gin.Context, deviceID string) (port.DeviceDriver, error) {
		return reg.Get(c.Request.Context(), deviceID)
	}

	streamDriverProvider := func(ctx context.Context, deviceID string) (port.StreamingDeviceDriver, error) {
		driver, err := reg.Get(ctx, deviceID)
		if err != nil {
			return nil, err
		}
		sd, ok := driver.(port.StreamingDeviceDriver)
		if !ok {
			return nil, fmt.Errorf("driver for device %q does not support streaming", deviceID)
		}
		return sd, nil
	}

	// HTTP Handlers
	deviceHandler := mcpAdapter.NewDeviceHandler(deviceUC, httpDriverProvider)
	mikhmonHandler := mcpAdapter.NewMikhmonHandler(mikhmonUC, httpDriverProvider)
	mikhmonStreamHandler := wsAdapter.NewMikhmonStreamHandler(streamDriverProvider)
	deviceStreamHandler := wsAdapter.NewDeviceStreamHandler(deviceUC, streamDriverProvider)

	// Gin Router Setup
	r := gin.Default()
	r.Use(middleware.CORS(cfg.CORSOrigins))

	// Auth & RBAC Handlers
	if pgStore != nil && casbinEnforcer != nil {
		authHandler := mcpAdapter.NewAuthHandler(pgStore, jwtService)
		rbacHandler := mcpAdapter.NewRBACHandler(casbinEnforcer)
		sessionHandler := mcpAdapter.NewSessionHandler(pgStore, waManager)
		convService := business.NewConversationService(pgStore)
		convHandler := mcpAdapter.NewConversationHandler(convService, waManager, sseHub)
		knowledgeHandler := mcpAdapter.NewKnowledgeHandler(pgStore)
		llmHandler := mcpAdapter.NewLLMConfigHandler(pgStore, cfg)
		technicianHandler := mcpAdapter.NewTechnicianHandler(pgStore)

		mcpAdapter.RegisterAuthRoutes(r, authHandler, jwtService)
		mcpAdapter.RegisterRBACRoutes(r, rbacHandler, jwtService, casbinEnforcer)
		mcpAdapter.RegisterBotRoutes(r, sessionHandler, convHandler, knowledgeHandler, llmHandler, technicianHandler, sseHub, jwtService, casbinEnforcer)
	}

	// Network REST Routes
	mcpAdapter.RegisterDeviceRoutes(r, deviceHandler)
	mcpAdapter.RegisterMikhmonRoutes(r, mikhmonHandler)

	// Realtime WebSockets & SSE Streaming Routes
	wsAdapter.RegisterStreamingRoutes(r, mikhmonStreamHandler)
	wsAdapter.RegisterDeviceStreamingRoutes(r, deviceStreamHandler)

	// MCP Protocol Handler
	mcpServer := mcp.New(reg, nil)
	r.Any("/mcp", gin.WrapH(mcpServer.HTTPHandler()))

	// gRPC Server Adapter
	grpcAddr := envOr("GRPC_ADDR", ":50051")
	grpcSrv, err := grpcAdapter.NewServer(grpcAddr, deviceUC, deviceStreamHandler)
	if err != nil {
		log.Printf("[gRPC Adapter Warning] Failed to initialize gRPC server: %v", err)
	} else {
		grpcSrv.Start()
		defer grpcSrv.Stop()
	}

	httpSrv := &http.Server{
		Addr:    httpAddr,
		Handler: r,
	}

	log.Printf("polyglot: Engine starting on http://localhost%s (REST, WebSockets, WhatsApp Gateway & MCP) and gRPC on %s", httpAddr, grpcAddr)
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server http: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("polyglot: shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = httpSrv.Shutdown(shutdownCtx)
	_ = reg.Close()
	log.Printf("polyglot: shutdown complete")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadInitialDevices(pgStore *postgres.Store) (port.DeviceRepository, port.CredentialVault) {
	if pgStore != nil {
		return postgres.NewDeviceRepository(pgStore.DB()), postgres.NewCredentialVault(pgStore.DB())
	}
	return &memRepo{devices: make(map[string]device.Device)}, &memVault{creds: make(map[string]device.Credentials)}
}

type memRepo struct {
	mu      sync.RWMutex
	devices map[string]device.Device
}

func (r *memRepo) Save(_ context.Context, d device.Device) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.devices[d.ID] = d
	return nil
}

func (r *memRepo) FindByID(_ context.Context, id string) (device.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.devices[id]
	if !ok {
		return device.Device{}, device.ErrNotFound
	}
	return d, nil
}

func (r *memRepo) FindAll(_ context.Context) ([]device.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]device.Device, 0, len(r.devices))
	for _, d := range r.devices {
		list = append(list, d)
	}
	return list, nil
}

func (r *memRepo) Update(ctx context.Context, d device.Device) error {
	return r.Save(ctx, d)
}

func (r *memRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.devices, id)
	return nil
}

type memVault struct {
	mu    sync.RWMutex
	creds map[string]device.Credentials
}

func (v *memVault) Get(_ context.Context, deviceID string) (device.Credentials, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	c, ok := v.creds[deviceID]
	if !ok {
		return device.Credentials{}, device.ErrNotFound
	}
	return c, nil
}

func (v *memVault) Save(_ context.Context, deviceID string, c device.Credentials) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.creds[deviceID] = c
	return nil
}
