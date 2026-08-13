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
	llmadapter "github.com/quixiq/polyglot/internal/adapter/llm"
	"github.com/quixiq/polyglot/internal/adapter/mcp"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
	redisAdapter "github.com/quixiq/polyglot/internal/adapter/redis"
	wsAdapter "github.com/quixiq/polyglot/internal/adapter/ws"
	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/device"
	domainllm "github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/driver/genieacs"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/driver/whatsapp"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/registry"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
	botUC "github.com/quixiq/polyglot/internal/usecase/bot"
	chatUC "github.com/quixiq/polyglot/internal/usecase/chat"
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
	sseHub     *wsAdapter.SSEHub
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
	chatService := chatUC.NewChatService(pgStore)
	waManager, err := whatsapp.NewSessionManager(cfg.DatabaseURL, pgStore, nil, eventHandler.MakeStatusCallback())
	if err != nil {
		log.Printf("[Warning] Failed to initialize WhatsApp SessionManager: %v", err)
	}

	convService := convUC.NewConversationService(pgStore)
	// Broadcast SSE `conversation_status` tiap kali status percakapan berubah
	// (take-over/return bot/close/escalation) — dikonsumsi useWARealtimeStream.
	convService.SetPublisher(sseHub)
	knowledgeRetriever := knowledgeUC.NewKeywordRetriever(pgStore)
	// Factory LLM di-inject agar usecase/bot tetap bersih dari adapter layer.
	llmFactory := func(c *domainllm.LLMConfig) (port.LLMProvider, error) {
		return llmadapter.NewProvider(c, cfg.EncryptionKey)
	}
	botEngine := botUC.NewEngine(cfg, redisStore, waManager, convService, knowledgeRetriever, pgStore, pgStore, sseHub, llmFactory)

	if waManager != nil {
		// Hubungkan pesan masuk WhatsApp ke engine bot. Dilakukan setelah engine
		// dibangun (bukan saat NewSessionManager) karena ada circular dependency.
		waManager.SetMessageCallback(eventHandler.MakeMessageCallback(botEngine.HandleIncomingMessage))
		// Broadcast SSE `chat_update` setiap kali mirror chat berubah, supaya
		// Inbox frontend ter-update instan (lihat useWARealtimeStream).
		waManager.SetChatUpdateCallback(eventHandler.MakeChatUpdateCallback())
		sessions, err := pgStore.FindAllSessions()
		if err == nil && len(sessions) > 0 {
			_ = waManager.RestoreAllSessions(sessions)
		}
	}

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

	invRepo := postgres.NewInvoiceRepository(pgStore.DB())
	subRepo := postgres.NewSubscriptionRepository(pgStore.DB())
	invUC := billingUC.NewInvoiceUsecase(invRepo)
	subUC := billingUC.NewSubscriptionUsecase(subRepo)

	billingConnectPath, billingConnectHandler := billingConnect.NewBillingServiceHandler(invUC, subUC)
	connectGroup.Any(billingConnectPath+"*action", gin.WrapH(billingConnectHandler))

	mikhmonConnectPath, mikhmonConnectHandler := hotspotConnect.NewHotspotServiceHandler(hotUC, connectDriverProvider)
	connectGroup.Any(mikhmonConnectPath+"*action", gin.WrapH(mikhmonConnectHandler))

	waConnectPath, waConnectHandler := botConnect.NewWhatsAppServiceHandler(pgStore, waManager, chatService)
	connectGroup.Any(waConnectPath+"*action", gin.WrapH(waConnectHandler))

	botConnectPath, botConnectHandler := botConnect.NewBotServiceHandler(convService, botEngine)
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
		sseHub:     sseHub,
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

	// Putuskan semua koneksi SSE (EventSource) LEBIH DULU. Tanpa ini
	// http.Server.Shutdown menunggu koneksi streaming yang tidak pernah
	// selesai sampai context deadline habis (5 detik) lalu gagal dengan
	// "context deadline exceeded". Setelah Close, stream tiap client berakhir
	// (channel tertutup) sehingga koneksi jadi idle dan shutdown selesai cepat.
	if a.sseHub != nil {
		a.sseHub.Close()
	}

	if err := a.httpServer.Shutdown(ctx); err != nil {
		log.Printf("[Warning] Error shutting down HTTP server: %v", err)
		// Jaring pengaman: paksa tutup koneksi yang tersisa (mis. request LLM
		// panjang yang masih jalan) supaya proses benar-benar bisa keluar.
		_ = a.httpServer.Close()
	}

	if a.waManager != nil {
		// Putuskan koneksi WhatsApp dengan rapi (pairing & session lokal
		// dipertahankan — restore berikutnya tidak perlu scan ulang).
		a.waManager.DisconnectAll()
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
