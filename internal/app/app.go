package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	authConnect "github.com/quixiq/polyglot/internal/adapter/connect/auth"
	billingConnect "github.com/quixiq/polyglot/internal/adapter/connect/billing"
	botConnect "github.com/quixiq/polyglot/internal/adapter/connect/bot"
	customerConnect "github.com/quixiq/polyglot/internal/adapter/connect/customer"
	deviceConnect "github.com/quixiq/polyglot/internal/adapter/connect/device"
	hotspotConnect "github.com/quixiq/polyglot/internal/adapter/connect/hotspot"
	pppConnect "github.com/quixiq/polyglot/internal/adapter/connect/ppp"
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
	"github.com/quixiq/polyglot/internal/driver/genericssh"
	"github.com/quixiq/polyglot/internal/driver/genieacs"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	mikhmon "github.com/quixiq/polyglot/internal/driver/mikrotik/hotspot"
	"github.com/quixiq/polyglot/internal/driver/whatsapp"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/registry"
	authUC "github.com/quixiq/polyglot/internal/usecase/auth"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
	botUC "github.com/quixiq/polyglot/internal/usecase/bot"
	chatUC "github.com/quixiq/polyglot/internal/usecase/chat"
	convUC "github.com/quixiq/polyglot/internal/usecase/conversation"
	customerUC "github.com/quixiq/polyglot/internal/usecase/customer"
	deviceUC "github.com/quixiq/polyglot/internal/usecase/device"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
	knowledgeUC "github.com/quixiq/polyglot/internal/usecase/knowledge"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
	pppUC "github.com/quixiq/polyglot/internal/usecase/ppp"
	userUC "github.com/quixiq/polyglot/internal/usecase/user"
	"github.com/quixiq/polyglot/pkg/logger"
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
	logger.Init(cfg.LogLevel, cfg.AppEnv)
	logger.WithComponent("Polyglot").Info("initializing application container...")

	dbURL := strings.TrimSpace(cfg.DatabaseURL)
	pgStore, err := postgres.NewStore(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	redisURL := strings.TrimSpace(cfg.RedisURL)
	redisStore, err := redisAdapter.NewStore(redisURL)
	if err != nil {
		logger.WithComponent("Polyglot").Warnf("failed to connect to Redis (%v). Falling back to in-memory cache.", err)
	}

	jwtService := auth.NewJWTService(cfg.JWTSecret, cfg.JWTExpiryHours)

	var refreshSvc *auth.RefreshTokenService
	if redisStore != nil {
		refreshSvc = auth.NewRefreshTokenService(redisStore, time.Duration(cfg.RefreshTokenTTLHours)*time.Hour)
	}

	casbinEnforcer, err := auth.NewCasbinEnforcer(ctx, pgStore.DB())
	if err != nil {
		logger.WithComponent("Polyglot").Warnf("failed to initialize Casbin enforcer: %v", err)
	} else {
		auth.SeedSystemPolicies(casbinEnforcer)
		if users, err := pgStore.FindAllUsers(ctx); err == nil {
			refs := make([]*auth.UserRef, 0, len(users))
			for _, u := range users {
				refs = append(refs, &auth.UserRef{ID: fmt.Sprintf("%d", u.ID), Role: u.Role})
			}
			auth.EnsureUserRoleAssignments(casbinEnforcer, refs)
		} else {
			logger.WithComponent("Polyglot").Warnf("failed to load users for role assignment sync: %v", err)
		}
	}

	sseHub := wsAdapter.NewSSEHub()

	eventHandler := whatsapp.NewEventHandler(pgStore, sseHub)
	chatService := chatUC.NewChatService(pgStore)
	waManager, err := whatsapp.NewSessionManager(cfg.DatabaseURL, pgStore, nil, eventHandler.MakeStatusCallback())
	if err != nil {
		logger.WithComponent("Polyglot").Warnf("failed to initialize WhatsApp SessionManager: %v", err)
	}

	convService := convUC.NewConversationService(pgStore)
	convService.SetPublisher(sseHub)

	keywordRetriever := knowledgeUC.NewKeywordRetriever(pgStore)
	retrievers := []port.KnowledgeRetriever{keywordRetriever}

	var knowledgeDocManager port.KnowledgeDocumentManager
	if cfg.AnythingLLMAPIKey != "" {
		anyRetriever, err := knowledgeadapter.NewRetriever(
			cfg.AnythingLLMBaseURL,
			cfg.AnythingLLMAPIKey,
			cfg.AnythingLLMWorkspace,
			cfg.AnythingLLMTopN,
		)
		if err != nil {
			logger.WithComponent("Knowledge").Warnf("AnythingLLM retriever disabled (%v)", err)
		} else {
			retrievers = append(retrievers, anyRetriever)
			logger.WithComponent("Knowledge").Infof("AnythingLLM retriever active for workspace %q", cfg.AnythingLLMWorkspace)
		}
		if anyManager, err := knowledgeadapter.NewManager(
			cfg.AnythingLLMBaseURL,
			cfg.AnythingLLMAPIKey,
			cfg.AnythingLLMWorkspace,
		); err != nil {
			logger.WithComponent("Knowledge").Warnf("AnythingLLM document manager disabled (%v)", err)
		} else {
			knowledgeDocManager = anyManager
			logger.WithComponent("Knowledge").Info("AnythingLLM document manager active")
		}
	}
	knowledgeRetriever := knowledgeUC.NewHybridRetriever(retrievers...)
	knowledgeRepo := postgres.NewKnowledgeRepository(pgStore)
	knowledgeDocUC := knowledgeUC.NewDocumentManager(knowledgeRepo, knowledgeDocManager)

	var knowledgeChat port.KnowledgeChat
	if cfg.AnythingLLMAPIKey != "" {
		if anyChat, err := knowledgeadapter.NewChatClient(
			cfg.AnythingLLMBaseURL,
			cfg.AnythingLLMAPIKey,
			cfg.AnythingLLMWorkspace,
		); err != nil {
			logger.WithComponent("Bot").Warnf("AnythingLLM chat client disabled (%v)", err)
		} else {
			knowledgeChat = anyChat
			logger.WithComponent("Bot").Infof("AnythingLLM chat active for workspace %q", cfg.AnythingLLMWorkspace)
		}
	}

	llmFactory := func(c *domainllm.Config) (port.LLMProvider, error) {
		return llmadapter.NewProvider(c, cfg.EncryptionKey)
	}
	botEngine := botUC.NewEngine(
		botUC.Config{
			SystemPrompt:       cfg.SystemPrompt,
			AllowedTopics:      cfg.AllowedTopics,
			LLMMaxOutputTokens: cfg.LLMMaxOutputTokens,
		},
		redisStore,
		waManager,
		convService,
		knowledgeRetriever,
		knowledgeChat,
		pgStore,
		pgStore,
		sseHub,
		llmFactory,
	)

	if waManager != nil {
		waManager.SetMessageCallback(eventHandler.MakeMessageCallback(botEngine.HandleIncomingMessage))
		waManager.SetChatUpdateCallback(eventHandler.MakeChatUpdateCallback())
		waManager.SetChatPresenceCallback(eventHandler.MakeChatPresenceCallback())
		sessions, err := pgStore.FindAllSessions(context.Background())
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

	// Fase 4 port seam: usecases depend only on port interfaces; vendor-native
	// command knowledge stays behind the gateways in the driver layer. The
	// executor is the policy-gated network.ExecuteCommand so destructive
	// commands keep requiring approval.
	exec := networkUC.ExecuteCommand
	hotGateway := mikhmon.NewGateway(exec)
	sessionGateway := mikrotik.NewGateway(exec) // implements SessionGateway + DeviceDiagnostics

	devUC := deviceUC.NewManageDeviceUseCase(repo, vault, reg, sessionGateway)
	hotUC := hotspotUC.New("internal/template", hotGateway)
	activeSessionsUC := networkUC.NewActiveSessionsUseCase(sessionGateway)
	openTermUC := networkUC.NewOpenTerminalUseCase(repo, vault, genericssh.DialSSHPty)

	customerRepo := postgres.NewCustomerRepository(pgStore.DB())
	custUC := customerUC.NewManageCustomerUseCase(customerRepo)

	userRepo := postgres.NewUserRepository(pgStore.DB())
	authUseCase := authUC.NewAuthUseCase(userRepo, jwtService, refreshSvc, redisStore, casbinEnforcer)
	refreshUseCase := authUC.NewRefreshTokenUseCase(userRepo, jwtService, refreshSvc, casbinEnforcer)
	manageUserUseCase := userUC.NewManageUserUseCase(userRepo, casbinEnforcer)

	connectDriverProvider := func(ctx context.Context, deviceID string) (port.DeviceDriver, error) {
		return reg.Get(ctx, deviceID)
	}

	rootMux := http.NewServeMux()

	// 1. Register Public Routes
	authPath, authHandler := authConnect.NewAuthServiceHandler(
		authUseCase,
		refreshUseCase,
		manageUserUseCase,
		cfg.AppEnv == "production",
	)
	rootMux.Handle(authPath, authHandler)

	mcpServer := mcp.New(reg, nil).
		WithMikhmonUseCase(hotUC).
		WithCustomerRepository(customerRepo)
	rootMux.Handle("/mcp", mcpServer.HTTPHandler())

	wsAdapter.RegisterEventRoutes(rootMux, sseHub, openTermUC)

	// 2. Register Protected ConnectRPC Services onto a Protected Sub-Mux
	protectedMux := http.NewServeMux()

	registerProtected := func(servicePath string, handler http.Handler) {
		protectedMux.Handle(servicePath, handler)
		// Mount pattern on rootMux forwarding through auth+rbac middlewares
	}

	devPath, devHandler := deviceConnect.NewDeviceServiceHandler(devUC, openTermUC, connectDriverProvider)
	registerProtected(devPath, devHandler)

	custPath, custHandler := customerConnect.NewCustomerServiceHandler(custUC)
	registerProtected(custPath, custHandler)

	rbacPath, rbacHandler := authConnect.NewRBACServiceHandler(casbinEnforcer)
	registerProtected(rbacPath, rbacHandler)

	userPath, userHandler := authConnect.NewUserServiceHandler(manageUserUseCase)
	registerProtected(userPath, userHandler)

	invRepo := postgres.NewInvoiceRepository(pgStore.DB())
	subRepo := postgres.NewSubscriptionRepository(pgStore.DB())
	invUC := billingUC.NewInvoiceUseCase(invRepo)
	subUC := billingUC.NewSubscriptionUseCase(subRepo)

	billingPath, billingHandler := billingConnect.NewBillingServiceHandler(invUC, subUC)
	registerProtected(billingPath, billingHandler)

	mikhmonPath, mikhmonHandler := hotspotConnect.NewHotspotServiceHandler(hotUC, activeSessionsUC, connectDriverProvider)
	registerProtected(mikhmonPath, mikhmonHandler)

	pppUseCase := pppUC.New(sessionGateway)
	pppPath, pppHandler := pppConnect.NewPPPServiceHandler(pppUseCase, connectDriverProvider)
	registerProtected(pppPath, pppHandler)

	waPath, waHandler := botConnect.NewWhatsAppServiceHandler(pgStore, waManager, chatService)
	registerProtected(waPath, waHandler)

	botPath, botHandler := botConnect.NewBotServiceHandler(convService, botEngine)
	registerProtected(botPath, botHandler)

	knwPath, knwHandler := botConnect.NewKnowledgeServiceHandler(knowledgeDocUC)
	registerProtected(knwPath, knwHandler)

	probePath, probeHandler := deviceConnect.NewProbeServiceHandler()
	registerProtected(probePath, probeHandler)

	// Wrap protected services with JWT and RBAC enforcement
	protectedHandler := middleware.Chain(
		protectedMux,
		middleware.AuthenticateJWT(jwtService),
		middleware.AuthorizeProcedure(casbinEnforcer),
	)

	// Mount protected handler for all protected service paths
	rootMux.Handle(devPath, protectedHandler)
	rootMux.Handle(custPath, protectedHandler)
	rootMux.Handle(rbacPath, protectedHandler)
	rootMux.Handle(userPath, protectedHandler)
	rootMux.Handle(billingPath, protectedHandler)
	rootMux.Handle(mikhmonPath, protectedHandler)
	rootMux.Handle(waPath, protectedHandler)
	rootMux.Handle(botPath, protectedHandler)
	rootMux.Handle(knwPath, protectedHandler)
	rootMux.Handle(probePath, protectedHandler)

	// 3. Wrap root handler with Global Middlewares (CORS, Logger, Recovery)
	finalHandler := middleware.Chain(
		rootMux,
		middleware.Recovery(),
		middleware.RequestLogger(),
		middleware.CORS(cfg.CORSOrigins, cfg.AppEnv),
	)

	httpAddr := envOr("PORT", ":"+cfg.Port)
	if httpAddr[0] != ':' {
		httpAddr = ":" + httpAddr
	}

	httpSrv := &http.Server{
		Addr:    httpAddr,
		Handler: finalHandler,
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
	logger.WithComponent("Polyglot").Infof("engine starting on http://localhost%s (net/http ServeMux, ConnectRPC, SSE, WebSockets, WhatsApp Gateway & MCP)", a.httpServer.Addr)
	if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server failed: %w", err)
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	logger.WithComponent("Polyglot").Info("shutting down server...")

	if a.sseHub != nil {
		a.sseHub.Close()
	}

	if err := a.httpServer.Shutdown(ctx); err != nil {
		logger.WithComponent("Polyglot").Warnf("error shutting down HTTP server: %v", err)
		_ = a.httpServer.Close()
	}

	if a.waManager != nil {
		a.waManager.DisconnectAll()
	}

	if a.registry != nil {
		_ = a.registry.Close()
	}

	logger.WithComponent("Polyglot").Info("shutdown complete")
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
