package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quixiq/polyglot/internal/adapter/auth"
	authConnect "github.com/quixiq/polyglot/internal/adapter/connect/auth"
	billingConnect "github.com/quixiq/polyglot/internal/adapter/connect/billing"
	botConnect "github.com/quixiq/polyglot/internal/adapter/connect/bot"
	customerConnect "github.com/quixiq/polyglot/internal/adapter/connect/customer"
	deviceConnect "github.com/quixiq/polyglot/internal/adapter/connect/device"
	hotspotConnect "github.com/quixiq/polyglot/internal/adapter/connect/hotspot"
	"github.com/quixiq/polyglot/internal/adapter/http/middleware"
	knowledgeadapter "github.com/quixiq/polyglot/internal/adapter/knowledge"
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

	// Refresh token opaque di Redis: rotasi tiap refresh, revoke di logout.
	// Redis wajib ada untuk fitur ini (refresh token tidak bisa di-fallback
	// ke memory store lintas restart). Kalau Redis mati, pakai access token
	// yang masih valid sampai kedaluwarsa — login baru tetap jalan.
	var refreshSvc *auth.RefreshTokenService
	if redisStore != nil {
		refreshSvc = auth.NewRefreshTokenService(redisStore, time.Duration(cfg.RefreshTokenTTLHours)*time.Hour)
	}

	casbinEnforcer, err := auth.NewCasbinEnforcer(ctx, pgStore.DB())
	if err != nil {
		log.Printf("[Warning] Failed to initialize Casbin enforcer: %v", err)
	} else {
		// Seed policy format baru (resource:action) + sync role assignment
		// dari tabel users. Idempotent — aman dipanggil setiap startup.
		auth.SeedSystemPolicies(casbinEnforcer)
		if users, err := pgStore.FindAllUsers(); err == nil {
			refs := make([]*auth.UserRef, 0, len(users))
			for _, u := range users {
				refs = append(refs, &auth.UserRef{ID: fmt.Sprintf("%d", u.ID), Role: u.Role})
			}
			auth.EnsureUserRoleAssignments(casbinEnforcer, refs)
		} else {
			log.Printf("[Warning] Failed to load users for role assignment sync: %v", err)
		}
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
	// Retriever pengetahuan = HYBRID: keyword retriever (tabel knowledge
	// Postgres, untuk dokumen lokal yang tidak di-embed) + vector retriever
	// AnythingLLM (untuk dokumen dengan embed_to_llm = true). Hybrid dipakai
	// karena dengan embed per-dokumen, either/or akan membuat dokumen lokal
	// tidak pernah ter-retrieve saat AnythingLLM aktif.
	keywordRetriever := knowledgeUC.NewKeywordRetriever(pgStore)
	retrievers := []port.KnowledgeRetriever{keywordRetriever}
	// Manager tulis ke AnythingLLM (raw-text/remove-documents) untuk fitur
	// admin knowledge. Nil kalau API key tidak di-set — dokumen tetap bisa
	// dikelola sebagai knowledge lokal (embed_to_llm = false).
	var knowledgeDocManager port.KnowledgeDocumentManager
	if cfg.AnythingLLMAPIKey != "" {
		anyRetriever, err := knowledgeadapter.NewRetriever(
			cfg.AnythingLLMBaseURL,
			cfg.AnythingLLMAPIKey,
			cfg.AnythingLLMWorkspace,
			cfg.AnythingLLMTopN,
		)
		if err != nil {
			log.Printf("[Warning] AnythingLLM retriever disabled (%v).", err)
		} else {
			retrievers = append(retrievers, anyRetriever)
			log.Printf("[Knowledge] AnythingLLM retriever aktif untuk workspace %q (%s)", cfg.AnythingLLMWorkspace, cfg.AnythingLLMBaseURL)
		}
		if anyManager, err := knowledgeadapter.NewManager(
			cfg.AnythingLLMBaseURL,
			cfg.AnythingLLMAPIKey,
			cfg.AnythingLLMWorkspace,
		); err != nil {
			log.Printf("[Warning] AnythingLLM document manager disabled (%v). Embed dari admin tidak tersedia.", err)
		} else {
			knowledgeDocManager = anyManager
			log.Printf("[Knowledge] AnythingLLM document manager aktif — embed per-dokumen dari admin tersedia")
		}
	}
	knowledgeRetriever := knowledgeUC.NewHybridRetriever(retrievers...)
	knowledgeRepo := postgres.NewKnowledgeRepository(pgStore)
	knowledgeDocUC := knowledgeUC.NewDocumentManager(knowledgeRepo, knowledgeDocManager)
	// Chat AnythingLLM = otak utama bot (RAG + LLM + history dalam satu
	// panggilan). Nil kalau API key tidak di-set → engine otomatis memakai
	// LLM lokal proyek (llm_configs) sebagai satu-satunya path.
	var knowledgeChat port.KnowledgeChat
	if cfg.AnythingLLMAPIKey != "" {
		if anyChat, err := knowledgeadapter.NewChatClient(
			cfg.AnythingLLMBaseURL,
			cfg.AnythingLLMAPIKey,
			cfg.AnythingLLMWorkspace,
		); err != nil {
			log.Printf("[Warning] AnythingLLM chat client disabled (%v) — bot pakai LLM lokal.", err)
		} else {
			knowledgeChat = anyChat
			log.Printf("[Bot] AnythingLLM chat aktif sebagai primary LLM (workspace %q); fallback ke LLM lokal bila down", cfg.AnythingLLMWorkspace)
		}
	}
	// Factory LLM di-inject agar usecase/bot tetap bersih dari adapter layer.
	llmFactory := func(c *domainllm.LLMConfig) (port.LLMProvider, error) {
		return llmadapter.NewProvider(c, cfg.EncryptionKey)
	}
	botEngine := botUC.NewEngine(cfg, redisStore, waManager, convService, knowledgeRetriever, knowledgeChat, pgStore, pgStore, sseHub, llmFactory)

	if waManager != nil {
		// Hubungkan pesan masuk WhatsApp ke engine bot. Dilakukan setelah engine
		// dibangun (bukan saat NewSessionManager) karena ada circular dependency.
		waManager.SetMessageCallback(eventHandler.MakeMessageCallback(botEngine.HandleIncomingMessage))
		// Broadcast SSE `chat_update` setiap kali mirror chat berubah, supaya
		// Inbox frontend ter-update instan (lihat useWARealtimeStream).
		waManager.SetChatUpdateCallback(eventHandler.MakeChatUpdateCallback())
		// Broadcast SSE `chat_presence` (typing/recording indicator) ke frontend.
		waManager.SetChatPresenceCallback(eventHandler.MakeChatPresenceCallback())
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
	connectGroup.Use(middleware.AuthorizeProcedure(casbinEnforcer))

	connectPath, connectHandler := deviceConnect.NewDeviceServiceHandler(devUC, connectDriverProvider)
	connectGroup.Any(connectPath+"*action", gin.WrapH(connectHandler))

	custConnectPath, custConnectHandler := customerConnect.NewCustomerServiceHandler(custUC)
	connectGroup.Any(custConnectPath+"*action", gin.WrapH(custConnectHandler))

	// Auth service remains public
	authConnectPath, authConnectHandler := authConnect.NewAuthServiceHandler(
		pgStore,
		jwtService,
		refreshSvc,
		redisStore,
		casbinEnforcer,
		cfg.AppEnv == "production",
	)
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

	knwConnectPath, knwConnectHandler := botConnect.NewKnowledgeServiceHandler(knowledgeDocUC)
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
