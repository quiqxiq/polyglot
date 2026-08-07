package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

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
	dbURL := strings.TrimSpace(cfg.DatabaseURL)
	pgStore, err := postgres.NewStore(dbURL)
	if err != nil {
		log.Printf("[Warning] Failed to connect to PostgreSQL (%v). Features requiring database will fail until DB is online.", err)
	}

	redisURL := strings.TrimSpace(cfg.RedisURL)
	redisStore, err := redisAdapter.NewStore(redisURL)
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
		waManager, err = whatsapp.NewSessionManager(cfg.DatabaseURL, nil, eventHandler.MakeStatusCallback())
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

	var customerRepo port.CustomerRepository
	if pgStore != nil {
		customerRepo = postgres.NewCustomerRepository(pgStore.DB())
	} else {
		customerRepo = memory.NewCustomerRepository()
	}
	customerUC := business.NewManageCustomerUseCase(customerRepo)

	// Streaming Driver Provider

	// Gin Router Setup
	r := gin.Default()
	r.Use(middleware.CORS(cfg.CORSOrigins))

	connectDriverProvider := func(ctx context.Context, deviceID string) (port.DeviceDriver, error) {
		return reg.Get(ctx, deviceID)
	}

	// MCP Protocol Handler
	mcpServer := mcp.New(reg, nil).
		WithMikhmonUseCase(mikhmonUC).
		WithCustomerRepository(customerRepo)
	r.Any("/mcp", gin.WrapH(mcpServer.HTTPHandler()))

	openTermUC := network.NewOpenTerminalUseCase(repo, vault)

	// Realtime Event Streaming Route (SSE & WS Terminal)
	wsAdapter.RegisterEventRoutes(r, sseHub, openTermUC)

	// ConnectRPC Protocol Handler (Buf / Connect RPC served over standard HTTP on :8080)
	connectPath, connectHandler := connectAdapter.NewDeviceServiceHandler(deviceUC, connectDriverProvider)
	r.Any(connectPath+"*action", gin.WrapH(connectHandler))

	custConnectPath, custConnectHandler := connectAdapter.NewCustomerServiceHandler(customerUC)
	r.Any(custConnectPath+"*action", gin.WrapH(custConnectHandler))

	authConnectPath, authConnectHandler := connectAdapter.NewAuthServiceHandler(pgStore, jwtService)
	r.Any(authConnectPath+"*action", gin.WrapH(authConnectHandler))

	rbacConnectPath, rbacConnectHandler := connectAdapter.NewRBACServiceHandler(casbinEnforcer)
	r.Any(rbacConnectPath+"*action", gin.WrapH(rbacConnectHandler))

	billingConnectPath, billingConnectHandler := connectAdapter.NewBillingServiceHandler()
	r.Any(billingConnectPath+"*action", gin.WrapH(billingConnectHandler))

	mikhmonConnectPath, mikhmonConnectHandler := connectAdapter.NewMikhmonServiceHandler(mikhmonUC, connectDriverProvider)
	r.Any(mikhmonConnectPath+"*action", gin.WrapH(mikhmonConnectHandler))

	waConnectPath, waConnectHandler := connectAdapter.NewWhatsAppServiceHandler(pgStore, waManager)
	r.Any(waConnectPath+"*action", gin.WrapH(waConnectHandler))

	botConnectPath, botConnectHandler := connectAdapter.NewBotServiceHandler()
	r.Any(botConnectPath+"*action", gin.WrapH(botConnectHandler))

	knwConnectPath, knwConnectHandler := connectAdapter.NewKnowledgeServiceHandler()
	r.Any(knwConnectPath+"*action", gin.WrapH(knwConnectHandler))

	probeConnectPath, probeConnectHandler := connectAdapter.NewProbeServiceHandler()
	r.Any(probeConnectPath+"*action", gin.WrapH(probeConnectHandler))

	httpSrv := &http.Server{
		Addr:    httpAddr,
		Handler: r,
	}

	log.Printf("polyglot: Engine starting on http://localhost%s (REST, ConnectRPC, WebSockets, WhatsApp Gateway & MCP)", httpAddr)
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
	return memory.NewDeviceRepository(), memory.NewCredentialVault()
}
