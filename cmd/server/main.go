package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	connectAdapter "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/adapter/http/middleware"
	"github.com/quixiq/polyglot/internal/adapter/mcp"
	"github.com/quixiq/polyglot/internal/adapter/memory"
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
	authUC "github.com/quixiq/polyglot/internal/usecase/auth"
	botUC "github.com/quixiq/polyglot/internal/usecase/bot"
	"github.com/quixiq/polyglot/internal/usecase/business"
	knowledgeUC "github.com/quixiq/polyglot/internal/usecase/knowledge"
	"github.com/quixiq/polyglot/internal/usecase/network"
	"github.com/quixiq/polyglot/pkg/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	// Initialize centralized Logrus logger
	logger.Init(cfg.AppEnv, "debug", os.Stdout)
	logger.WithFields(logger.Fields{
		"env":  cfg.AppEnv,
		"port": cfg.Port,
	}).Info("Initializing Polyglot Engine backend...")

	httpAddr := envOr("PORT", ":"+cfg.Port)
	if httpAddr[0] != ':' {
		httpAddr = ":" + httpAddr
	}

	// 1. Initialize Database & Storage Adapters
	dbURL := strings.TrimSpace(cfg.DatabaseURL)
	pgStore, err := postgres.NewStore(dbURL)
	if err != nil {
		logger.WithError(err).Warn("PostgreSQL unavailable, running in fallback mode")
	}

	redisURL := strings.TrimSpace(cfg.RedisURL)
	redisStore, err := redisAdapter.NewStore(redisURL)
	if err != nil {
		logger.WithError(err).Warn("Redis unavailable, using in-memory cache")
	}

	jwtService := auth.NewJWTService(cfg.JWTSecret, cfg.JWTExpiryHours)

	var casbinEnforcer *auth.CasbinEnforcer
	if pgStore != nil {
		casbinEnforcer, err = auth.NewCasbinEnforcer(ctx, pgStore.DB())
		if err != nil {
			logger.WithError(err).Warn("Failed to initialize Casbin RBAC enforcer")
		}
	}

	sseHub := wsAdapter.NewSSEHub()

	// 2. WhatsApp Gateway & AI Bot Engine Setup
	var waManager *whatsapp.SessionManager
	if pgStore != nil {
		eventHandler := whatsapp.NewEventHandler(pgStore, sseHub)
		waManager, err = whatsapp.NewSessionManager(cfg.DatabaseURL, nil, eventHandler.MakeStatusCallback())
		if err != nil {
			logger.WithError(err).Warn("Failed to initialize WhatsApp SessionManager")
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

	// Repositories & Use Cases
	var userRepo port.UserRepository
	if pgStore != nil {
		userRepo = postgres.NewUserRepository(pgStore.DB())
	}
	authUseCase := authUC.NewAuthUseCase(userRepo, jwtService)
	deviceUC := business.NewManageDeviceUseCase(repo, vault, reg)
	mikhmonUC := network.NewMikhmonUseCase("internal/templates")

	var customerRepo port.CustomerRepository
	if pgStore != nil {
		customerRepo = postgres.NewCustomerRepository(pgStore.DB())
	} else {
		customerRepo = memory.NewCustomerRepository()
	}
	customerUC := business.NewManageCustomerUseCase(customerRepo)

	connectDriverProvider := func(ctx context.Context, deviceID string) (port.DeviceDriver, error) {
		return reg.Get(ctx, deviceID)
	}

	// 4. Native net/http.ServeMux Root Router (Go 1.22+)
	mux := http.NewServeMux()

	// MCP Protocol Handler
	mcpServer := mcp.New(reg, nil).
		WithMikhmonUseCase(mikhmonUC).
		WithCustomerRepository(customerRepo)
	mux.Handle("/mcp", mcpServer.HTTPHandler())

	// Realtime Event Streaming Route (SSE & WS Terminal)
	openTermUC := network.NewOpenTerminalUseCase(repo, vault)
	wsAdapter.RegisterEventRoutes(mux, sseHub, openTermUC)

	var waSessionRepo port.WASessionRepository
	if pgStore != nil {
		waSessionRepo = postgres.NewWASessionRepository(pgStore.DB())
	}

	// ConnectRPC Protocol Handlers
	mount := func(prefix string, h http.Handler) {
		mux.Handle(prefix, h)
	}

	mount(connectAdapter.NewDeviceServiceHandler(deviceUC, connectDriverProvider))
	mount(connectAdapter.NewCustomerServiceHandler(customerUC))
	mount(connectAdapter.NewAuthServiceHandler(authUseCase))
	mount(connectAdapter.NewRBACServiceHandler(casbinEnforcer))
	mount(connectAdapter.NewBillingServiceHandler())
	mount(connectAdapter.NewMikhmonServiceHandler(mikhmonUC, connectDriverProvider))
	mount(connectAdapter.NewWhatsAppServiceHandler(waSessionRepo, waManager))
	mount(connectAdapter.NewBotServiceHandler())
	mount(connectAdapter.NewKnowledgeServiceHandler())
	mount(connectAdapter.NewProbeServiceHandler())

	// 5. Global Middleware Chain (Recovery -> Logger -> CORS)
	var handler http.Handler = mux
	handler = middleware.CORS(cfg.CORSOrigins)(handler)
	handler = middleware.RequestLogger()(handler)
	handler = middleware.Recoverer()(handler)

	httpSrv := &http.Server{
		Addr:         httpAddr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.WithFields(logger.Fields{
		"address":  fmt.Sprintf("http://localhost%s", httpAddr),
		"protocol": "Standard net/http (REST, ConnectRPC, SSE, WebSocket, MCP)",
	}).Info("Polyglot NetOps Engine started successfully")

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.WithError(err).Error("HTTP server failed unexpectedly")
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down Polyglot server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = httpSrv.Shutdown(shutdownCtx)
	_ = reg.Close()
	logger.Info("Polyglot server shutdown complete")
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
	return memory.NewDeviceRepository(), memory.NewCredentialVault()
}
