package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/quixiq/polyglot/internal/adapter/auth"
	authConnect "github.com/quixiq/polyglot/internal/adapter/connect/auth"
	billingConnect "github.com/quixiq/polyglot/internal/adapter/connect/billing"
	botConnect "github.com/quixiq/polyglot/internal/adapter/connect/bot"
	customerConnect "github.com/quixiq/polyglot/internal/adapter/connect/customer"
	deviceConnect "github.com/quixiq/polyglot/internal/adapter/connect/device"
	hotspotConnect "github.com/quixiq/polyglot/internal/adapter/connect/hotspot"
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
	convUC "github.com/quixiq/polyglot/internal/usecase/conversation"
	customerUC "github.com/quixiq/polyglot/internal/usecase/customer"
	deviceUC "github.com/quixiq/polyglot/internal/usecase/device"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
	knowledgeUC "github.com/quixiq/polyglot/internal/usecase/knowledge"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
)

type App struct {
	cfg        config.Config
	httpServer *http.Server
	registry   *registry.Registry
	waManager  *whatsapp.SessionManager
	pgStore    *postgres.Store
	redisStore *redisAdapter.Store
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	log.Println("[Polyglot Engine] Initializing application container...")

	dbURL := strings.TrimSpace(cfg.DatabaseURL)
	pgStore, err := postgres.NewStore(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	redisURL := strings.TrimSpace(cfg.RedisURL)
	redisStore, err := redisAdapter.NewStore(redisURL)
	if err != nil {
		log.Printf("[Warning] Failed to connect to Redis (%v). Falling back to in-memory cache.", err)
	}

	jwtService := auth.NewJWTService(cfg.JWTSecret, cfg.JWTExpiryHours)

	casbinEnforcer, err := auth.NewCasbinEnforcer(ctx, pgStore.DB())
	if err != nil {
		log.Printf("[Warning] Failed to initialize Casbin enforcer: %v", err)
	}

	sseHub := wsAdapter.NewSSEHub()

	eventHandler := whatsapp.NewEventHandler(pgStore, sseHub)
	waManager, err := whatsapp.NewSessionManager(cfg.DatabaseURL, nil, eventHandler.MakeStatusCallback())
	if err != nil {
		log.Printf("[Warning] Failed to initialize WhatsApp SessionManager: %v", err)
	}

	convService := convUC.NewConversationService(pgStore)
	knowledgeRetriever := knowledgeUC.NewKeywordRetriever(pgStore)
	botEngine := botUC.NewEngine(cfg, redisStore, waManager, convService, knowledgeRetriever, pgStore, sseHub)

	if waManager != nil {
		sessions, err := pgStore.FindAllSessions()
		if err == nil && len(sessions) > 0 {
			_ = waManager.RestoreAllSessions(sessions)
		}
	}
	_ = botEngine

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

	devUC := deviceUC.NewManageDeviceUseCase(repo, vault, reg)
	hotUC := hotspotUC.NewHotspotUseCase("internal/templates")

	customerRepo := postgres.NewCustomerRepository(pgStore.DB())
	custUC := customerUC.NewManageCustomerUseCase(customerRepo)

	r := gin.Default()
	r.Use(middleware.CORS(cfg.CORSOrigins, cfg.AppEnv))

	connectDriverProvider := func(ctx context.Context, deviceID string) (port.DeviceDriver, error) {
		return reg.Get(ctx, deviceID)
	}

	mcpServer := mcp.New(reg, nil).
		WithMikhmonUseCase(hotUC).
		WithCustomerRepository(customerRepo)
	r.Any("/mcp", gin.WrapH(mcpServer.HTTPHandler()))

	openTermUC := networkUC.NewOpenTerminalUseCase(repo, vault)
	wsAdapter.RegisterEventRoutes(r, sseHub, openTermUC)

	connectGroup := r.Group("/")
	connectGroup.Use(middleware.AuthenticateJWT(jwtService))

	connectPath, connectHandler := deviceConnect.NewDeviceServiceHandler(devUC, connectDriverProvider)
	connectGroup.Any(connectPath+"*action", gin.WrapH(connectHandler))

	custConnectPath, custConnectHandler := customerConnect.NewCustomerServiceHandler(custUC)
	connectGroup.Any(custConnectPath+"*action", gin.WrapH(custConnectHandler))

	// Auth service remains public
	authConnectPath, authConnectHandler := authConnect.NewAuthServiceHandler(pgStore, jwtService)
	r.Any(authConnectPath+"*action", gin.WrapH(authConnectHandler))

	rbacConnectPath, rbacConnectHandler := authConnect.NewRBACServiceHandler(casbinEnforcer)
	connectGroup.Any(rbacConnectPath+"*action", gin.WrapH(rbacConnectHandler))

	billingConnectPath, billingConnectHandler := billingConnect.NewBillingServiceHandler()
	connectGroup.Any(billingConnectPath+"*action", gin.WrapH(billingConnectHandler))

	mikhmonConnectPath, mikhmonConnectHandler := hotspotConnect.NewHotspotServiceHandler(hotUC, connectDriverProvider)
	connectGroup.Any(mikhmonConnectPath+"*action", gin.WrapH(mikhmonConnectHandler))

	waConnectPath, waConnectHandler := botConnect.NewWhatsAppServiceHandler(pgStore, waManager)
	connectGroup.Any(waConnectPath+"*action", gin.WrapH(waConnectHandler))

	botConnectPath, botConnectHandler := botConnect.NewBotServiceHandler(convService)
	connectGroup.Any(botConnectPath+"*action", gin.WrapH(botConnectHandler))

	knwConnectPath, knwConnectHandler := botConnect.NewKnowledgeServiceHandler()
	connectGroup.Any(knwConnectPath+"*action", gin.WrapH(knwConnectHandler))

	probeConnectPath, probeConnectHandler := deviceConnect.NewProbeServiceHandler()
	connectGroup.Any(probeConnectPath+"*action", gin.WrapH(probeConnectHandler))

	httpAddr := envOr("PORT", ":"+cfg.Port)
	if httpAddr[0] != ':' {
		httpAddr = ":" + httpAddr
	}

	httpSrv := &http.Server{
		Addr:    httpAddr,
		Handler: r,
	}

	return &App{
		cfg:        cfg,
		httpServer: httpSrv,
		registry:   reg,
		waManager:  waManager,
		pgStore:    pgStore,
		redisStore: redisStore,
	}, nil
}

func (a *App) Run() error {
	log.Printf("polyglot: Engine starting on http://localhost%s (REST, ConnectRPC, WebSockets, WhatsApp Gateway & MCP)", a.httpServer.Addr)
	if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server failed: %w", err)
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	log.Println("polyglot: shutting down server...")

	if err := a.httpServer.Shutdown(ctx); err != nil {
		log.Printf("[Warning] Error shutting down HTTP server: %v", err)
	}

	if a.waManager != nil {
		// Logically disconnect active WhatsApp sessions if needed
	}

	if a.registry != nil {
		_ = a.registry.Close()
	}

	log.Println("polyglot: shutdown complete")
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadInitialDevices(pgStore *postgres.Store) (port.DeviceRepository, port.CredentialVault) {
	return postgres.NewDeviceRepository(pgStore.DB()), postgres.NewCredentialVault(pgStore.DB())
}
