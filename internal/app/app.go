package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/adapter/auth"
	authConnect "github.com/quixiq/polyglot/internal/adapter/connect/auth"
	billingConnect "github.com/quixiq/polyglot/internal/adapter/connect/billing"
	botConnect "github.com/quixiq/polyglot/internal/adapter/connect/bot"
	cashbookConnect "github.com/quixiq/polyglot/internal/adapter/connect/cashbook"
	customerConnect "github.com/quixiq/polyglot/internal/adapter/connect/customer"
	deviceConnect "github.com/quixiq/polyglot/internal/adapter/connect/device"
	hotspotConnect "github.com/quixiq/polyglot/internal/adapter/connect/hotspot"
	ispadminConnect "github.com/quixiq/polyglot/internal/adapter/connect/ispadmin"
	notificationConnect "github.com/quixiq/polyglot/internal/adapter/connect/notification"
	portalConnect "github.com/quixiq/polyglot/internal/adapter/connect/portal"
	pppConnect "github.com/quixiq/polyglot/internal/adapter/connect/ppp"
	registrationConnect "github.com/quixiq/polyglot/internal/adapter/connect/registration"
	reportConnect "github.com/quixiq/polyglot/internal/adapter/connect/report"
	settingConnect "github.com/quixiq/polyglot/internal/adapter/connect/setting"
	adminapi "github.com/quixiq/polyglot/internal/adapter/http/adminapi"
	gatewayHTTP "github.com/quixiq/polyglot/internal/adapter/http/gateway"
	"github.com/quixiq/polyglot/internal/adapter/http/middleware"
	portalHTTP "github.com/quixiq/polyglot/internal/adapter/http/portal"
	reportsHTTP "github.com/quixiq/polyglot/internal/adapter/http/reports"
	webhookHTTP "github.com/quixiq/polyglot/internal/adapter/http/webhook"
	llmadapter "github.com/quixiq/polyglot/internal/adapter/llm"

	"github.com/quixiq/polyglot/internal/adapter/mcp"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/provisioner"
	redisAdapter "github.com/quixiq/polyglot/internal/adapter/redis"
	storageAdapter "github.com/quixiq/polyglot/internal/adapter/storage"
	tripay "github.com/quixiq/polyglot/internal/adapter/tripay"
	whatsappadapter "github.com/quixiq/polyglot/internal/adapter/whatsapp"
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
	"github.com/quixiq/polyglot/internal/usecase/importer"
	metricsUC "github.com/quixiq/polyglot/internal/usecase/metrics"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
	notificationUC "github.com/quixiq/polyglot/internal/usecase/notification"
	portalUC "github.com/quixiq/polyglot/internal/usecase/portal"
	pppUC "github.com/quixiq/polyglot/internal/usecase/ppp"
	registrationUC "github.com/quixiq/polyglot/internal/usecase/registration"
	settingUC "github.com/quixiq/polyglot/internal/usecase/setting"
	skillUC "github.com/quixiq/polyglot/internal/usecase/skill"
	userUC "github.com/quixiq/polyglot/internal/usecase/user"
	"github.com/quixiq/polyglot/pkg/logger"
)

type App struct {
	cfg           config.Config
	httpServer    *http.Server
	registry      *registry.Registry
	waManager     *whatsapp.SessionManager
	pgStore       *postgres.Store
	redisStore    *redisAdapter.Store
	sseHub        *wsAdapter.SSEHub
	pingStreamMgr *metricsUC.PingStreamWorker
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
		logger.WithComponent("Polyglot").WithError(err).Warn("failed to connect to Redis; falling back to in-memory cache")
	}

	jwtService := auth.NewJWTService(cfg.JWTSecret, cfg.JWTExpiryHours)

	var refreshSvc *auth.RefreshTokenService
	if redisStore != nil {
		refreshSvc = auth.NewRefreshTokenService(redisStore, time.Duration(cfg.RefreshTokenTTLHours)*time.Hour)
	}

	casbinEnforcer, err := auth.NewCasbinEnforcer(ctx, pgStore.DB())
	if err != nil {
		logger.WithComponent("Polyglot").WithError(err).Warn("failed to initialize Casbin enforcer")
	} else {
		auth.SeedSystemPolicies(casbinEnforcer)
		if users, err := pgStore.FindAllUsers(ctx); err == nil {
			refs := make([]*auth.UserRef, 0, len(users))
			for _, u := range users {
				refs = append(refs, &auth.UserRef{ID: fmt.Sprintf("%d", u.ID), Role: u.Role})
			}
			auth.EnsureUserRoleAssignments(casbinEnforcer, refs)
		} else {
			logger.WithComponent("Polyglot").WithError(err).Warn("failed to load users for role assignment sync")
		}
	}

	sseHub := wsAdapter.NewSSEHub()

	eventHandler := whatsapp.NewEventHandler(pgStore, sseHub)
	chatService := chatUC.NewChatUseCase(pgStore)
	waManager, err := whatsapp.NewSessionManager(cfg.DatabaseURL, pgStore, nil, eventHandler.MakeStatusCallback())
	if err != nil {
		logger.WithComponent("Polyglot").WithError(err).Warn("failed to initialize WhatsApp session manager")
	}

	convService := convUC.NewConversationUseCase(pgStore)
	convService.SetPublisher(sseHub)

	fsSkillStore, err := storageAdapter.NewFSSkillStore("data")
	if err != nil {
		logger.WithComponent("App").WithError(err).Warn("Failed to initialize FSSkillStore")
	}
	gitSyncer := storageAdapter.NewGitSyncer(2 * time.Minute)
	skillUseCase := skillUC.NewManageSkillUseCase(pgStore, fsSkillStore, gitSyncer)

	llmFactory := func(c *domainllm.Config) (port.LLMProvider, error) {
		return llmadapter.NewProvider(c, cfg.EncryptionKey)
	}
	userRepo := postgres.NewUserRepository(pgStore.DB())
	settingRepo := postgres.NewSettingRepository(pgStore.DB())
	botEngine := botUC.NewEngine(
		redisStore,
		waManager,
		convService,
		skillUseCase.Provider(),
		skillUseCase,
		pgStore,
		pgStore,
		userRepo,
		settingRepo,
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

	repo, vault := loadInitialDevices(pgStore, cfg.EncryptionKey)
	factories := map[string]registry.DriverFactory{
		"mikrotik": func(ctx context.Context, target device.Target) (port.DeviceDriver, error) {
			return mikrotik.NewDriver(ctx, target)
		},
		"genieacs": func(ctx context.Context, target device.Target) (port.DeviceDriver, error) {
			return genieacs.NewDriver(ctx, target)
		},
	}
	reg := registry.New(repo, vault, factories)

	exec := networkUC.ExecuteCommand
	hotGateway := mikhmon.NewGateway(exec)
	sessionGateway := mikrotik.NewGateway(exec)
	queueGateway := mikrotik.NewQueueGateway(exec)

	// ─── 1. Repositori & Manajer Infrastruktur ──────────────────────────────
	customerRepo := postgres.NewCustomerRepository(pgStore.DB())
	invRepo := postgres.NewInvoiceRepository(pgStore.DB())
	subRepo := postgres.NewSubscriptionRepository(pgStore.DB(), vault)
	planRepo := postgres.NewServicePlanRepository(pgStore.DB())
	notifRepo := postgres.NewNotificationRepository(pgStore.DB())
	cashRepo := postgres.NewCashbookRepository(pgStore.DB())
	reportingRepo := postgres.NewReportingRepository(pgStore.DB())
	auditLogRepo := postgres.NewAuditLogRepository(pgStore.DB())
	regRepo := postgres.NewRegistrationRepository(pgStore.DB())
	portalRepo := postgres.NewPortalRepository(pgStore.DB())
	gwtxRepo := postgres.NewGatewayTransactionRepository(pgStore.DB())
	paymentReader := postgres.NewPaymentReader(pgStore.DB())

	accountMgr := provisioner.New(reg, sessionGateway, hotGateway, sessionGateway, queueGateway)
	paymentProc := postgres.NewPaymentProcessor(pgStore.DB())
	paymentProc.OnPaid = provisioner.BuildOnPaidRestore(accountMgr, subRepo, planRepo, settingRepo)
	waSender := whatsappadapter.NewSenderAdapter(waManager, pgStore.FindAllSessions)
	tripayAdapter := tripay.NewAdapter(settingRepo)

	// ─── 2. Use Cases ───────────────────────────────────────────────────────
	authUseCase := authUC.NewAuthUseCase(userRepo, jwtService, refreshSvc, redisStore, casbinEnforcer)
	refreshUseCase := authUC.NewRefreshTokenUseCase(userRepo, jwtService, refreshSvc, casbinEnforcer)
	manageUserUseCase := userUC.NewManageUserUseCase(userRepo, casbinEnforcer)
	manageSettingUseCase := settingUC.NewManageSettingUseCase(settingRepo)
	custUC := customerUC.NewManageCustomerUseCase(customerRepo, subRepo, invRepo)
	devUC := deviceUC.NewManageDeviceUseCase(repo, vault, reg, sessionGateway)
	openTermUC := networkUC.NewOpenTerminalUseCase(repo, vault, genericssh.DialSSHPty)
	hotUC := hotspotUC.New("internal/template", hotGateway)
	activeSessionsUC := networkUC.NewActiveSessionsUseCase(sessionGateway)
	pppUseCase := pppUC.New(sessionGateway)

	invUC := billingUC.NewInvoiceUseCase(invRepo)
	subUC := billingUC.NewSubscriptionUseCase(subRepo, planRepo, customerRepo, repo)
	planUC := billingUC.NewPlanUseCase(planRepo, subRepo, accountMgr)
	checkoutUC := billingUC.NewCheckoutUseCase(invRepo, customerRepo, paymentProc)
	lifecycleUC := billingUC.NewSubscriptionLifecycleUseCase(subRepo, planRepo, accountMgr, auditLogRepo)
	runBillingUC := billingUC.NewRunBillingUseCase(subRepo, planRepo, invRepo).WithSettings(settingRepo)
	chargeUC := billingUC.NewGatewayChargeUseCase(invRepo, customerRepo, gwtxRepo, tripayAdapter, paymentProc, settingRepo)

	regManagerUC := registrationUC.NewManageRegistrationUseCase(regRepo, notifRepo, auditLogRepo)
	regConvertUC := registrationUC.NewConvertUseCase(registrationUC.ConvertDeps{
		Repo: regRepo, Plans: planRepo, Customers: customerRepo,
		Subs: subRepo, Invoices: invRepo, Audit: auditLogRepo, Manager: accountMgr,
	})

	portalUCase := portalUC.NewUseCase(portalRepo, customerRepo, subRepo, invRepo, paymentReader, waSender, settingRepo)

	upsertImport := importer.NewUpsertUseCase(planRepo, customerRepo, subRepo, auditLogRepo, "")
	upsertImport.SetDeviceResolver(func(deviceName string) (string, bool) {
		devices, err := repo.FindAll(context.Background())
		if err != nil {
			return "", false
		}
		for _, d := range devices {
			if strings.EqualFold(d.Name, deviceName) {
				return d.ID, true
			}
		}
		return "", false
	})
	routerSource := importer.NewRouterSource(sessionGateway)
	reconciler := importer.NewReconciler(subRepo, sessionGateway)
	exportUC := importer.NewExportUseCase(subRepo, customerRepo, planRepo)

	deviceAuthorizer := auth.NewDeviceAuthorizer(userRepo)
	metricsRepo := postgres.NewMetricsRepository(pgStore.DB())
	metricsUseCase := deviceUC.NewManageMetricsUseCase(repo, metricsRepo, deviceAuthorizer)
	manageIsolationUseCase := deviceUC.NewManageIsolationUseCase(accountMgr, repo)

	pingStreamMgr := metricsUC.NewPingStreamWorker(repo, metricsRepo, func(c context.Context, id string) (port.DeviceDriver, error) {
		return reg.Get(c, id)
	})
	pingStreamMgr.Start(ctx)

	connectDriverProvider := func(ctx context.Context, deviceID string) (port.DeviceDriver, error) {
		callerID, callerRoles, hasIdentity := auth.IdentityFromContext(ctx)
		if hasIdentity && deviceAuthorizer != nil {
			canAccess, err := deviceAuthorizer.CanAccessDevice(ctx, callerID, callerRoles, deviceID)
			if err != nil {
				return nil, err
			}
			if !canAccess {
				return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("access to device %s denied", deviceID))
			}
		}
		return reg.Get(ctx, deviceID)
	}

	driverResolverConnect := func(ctx context.Context, deviceID string) (port.DeviceDriver, bool) {
		drv, err := reg.Get(ctx, deviceID)
		if err != nil {
			return nil, false
		}
		return drv, true
	}

	// ─── 3. Routing (net/http ServeMux) ─────────────────────────────────────
	rootMux := http.NewServeMux()

	webhookHTTPHandler := webhookHTTP.NewHandler()
	webhookHTTPHandler.RegisterPublic(rootMux)

	authPath, authHandler := authConnect.NewAuthServiceHandler(
		authUseCase, refreshUseCase, manageUserUseCase, cfg.AppEnv == "production",
	)
	rootMux.Handle(authPath, authHandler)

	pubRegPath, pubRegHandler := registrationConnect.NewPublicSubmitHandler(regManagerUC)
	rootMux.Handle(pubRegPath, pubRegHandler)

	portalConnectPath, portalConnectHandler := portalConnect.NewPortalServiceHandler(portalUCase)
	rootMux.Handle(portalConnectPath, portalConnectHandler)

	portalHTTPHandler := portalHTTP.NewHandler(portalUCase)
	portalHTTPHandler.RegisterPublic(rootMux)
	portalHTTPHandler.RegisterAuthenticated(rootMux)

	gatewayHTTPHandler := gatewayHTTP.NewHandler(chargeUC)
	gatewayHTTPHandler.RegisterPublic(rootMux)

	mcpServer := mcp.New(reg, nil).
		WithMikhmonUseCase(hotUC).
		WithCustomerRepository(customerRepo)
	rootMux.Handle("/mcp", mcpServer.HTTPHandler())

	wsAdapter.RegisterEventRoutes(rootMux, sseHub, openTermUC)

	protectedMux := http.NewServeMux()

	var protectedPaths []string
	registerProtected := func(servicePath string, handler http.Handler) {
		protectedMux.Handle(servicePath, handler)
		protectedPaths = append(protectedPaths, servicePath)
	}

	devPath, devHandler := deviceConnect.NewDeviceServiceHandler(devUC, openTermUC, connectDriverProvider, metricsUseCase, manageIsolationUseCase)
	registerProtected(devPath, devHandler)

	custPath, custHandler := customerConnect.NewCustomerServiceHandler(custUC)
	registerProtected(custPath, custHandler)

	rbacPath, rbacHandler := authConnect.NewRBACServiceHandler(casbinEnforcer)
	registerProtected(rbacPath, rbacHandler)

	userPath, userHandler := authConnect.NewUserServiceHandler(manageUserUseCase)
	registerProtected(userPath, userHandler)

	settingPath, settingHandler := settingConnect.NewSettingServiceHandler(manageSettingUseCase)
	registerProtected(settingPath, settingHandler)

	billingPath, billingHandler := billingConnect.NewBillingServiceHandler(
		invUC, checkoutUC, subUC, lifecycleUC, planUC, runBillingUC,
		billingUC.NewManageSubscriptionUseCase(subRepo, planRepo, customerRepo, accountMgr, auditLogRepo, invRepo))
	registerProtected(billingPath, billingHandler)

	regPath, regHandler := registrationConnect.NewRegistrationServiceHandler(regManagerUC, regConvertUC, regRepo)
	registerProtected(regPath, regHandler)

	cashbookPath, cashbookHandler := cashbookConnect.NewCashbookServiceHandler(cashRepo)
	registerProtected(cashbookPath, cashbookHandler)

	notifPath, notifHandler := notificationConnect.NewNotificationServiceHandler(notifRepo, waSender)
	registerProtected(notifPath, notifHandler)

	reportPath, reportHandler := reportConnect.NewReportServiceHandler(reportingRepo, reportingRepo)
	registerProtected(reportPath, reportHandler)

	adminPath, adminHandler := ispadminConnect.NewIspAdminServiceHandler(
		upsertImport, routerSource, reconciler, exportUC, driverResolverConnect)
	registerProtected(adminPath, adminHandler)

	mikhmonPath, mikhmonHandler := hotspotConnect.NewHotspotServiceHandler(hotUC, activeSessionsUC, connectDriverProvider)
	registerProtected(mikhmonPath, mikhmonHandler)

	pppPath, pppHandler := pppConnect.NewPPPServiceHandler(pppUseCase, connectDriverProvider)
	registerProtected(pppPath, pppHandler)

	waPath, waHandler := botConnect.NewWhatsAppServiceHandler(pgStore, waManager, chatService)
	registerProtected(waPath, waHandler)

	botPath, botHandler := botConnect.NewBotServiceHandler(convService, botEngine, skillUseCase, pgStore, cfg.EncryptionKey)
	registerProtected(botPath, botHandler)

	probePath, probeHandler := deviceConnect.NewProbeServiceHandler()
	registerProtected(probePath, probeHandler)

	gatewayHTTPHandler.RegisterProtected(protectedMux)
	reportsHTTP.NewHandler(reportingRepo).Register(protectedMux)
	adminapi.NewHandler(upsertImport, routerSource, reconciler, reportingRepo, exportUC, driverResolverConnect).
		RegisterProtected(protectedMux)

	protectedHandler := middleware.Chain(
		protectedMux,
		middleware.AuthenticateJWT(jwtService),
		middleware.AuthorizeProcedure(casbinEnforcer),
	)
	for _, path := range protectedPaths {
		rootMux.Handle(path, protectedHandler)
	}
	for _, plainPath := range []string{
		"/api/cashier/charge",
		"/api/reports/daily", "/api/reports/monthly", "/api/reports/yearly",
		"/api/admin/import", "/api/admin/import-router", "/api/admin/reconcile",
		"/api/admin/export", "/api/admin/snapshot/refresh",
	} {
		rootMux.Handle(plainPath, protectedHandler)
	}

	// ─── 4. Scheduler (Cron) ────────────────────────────────────────────────
	if cfg.SchedulerEnabled && cfg.BillingCronSpec != "" && cfg.IsolationCronSpec != "" {
		isolateWorker := billingUC.NewIsolateWorker(
			subRepo, invRepo, customerRepo, planRepo, accountMgr, notifRepo, settingRepo,
		)
		waWorker := notificationUC.NewWASenderWorker(notifRepo, waSender, settingRepo)
		snapshotJob := func(ctx context.Context) error {
			return reportingRepo.RecomputeDaily(ctx, "tenant-default", timeNowUTC())
		}
		sched := newScheduler(schedulerJobs{
			billing:  runBillingUC,
			isolate:  isolateWorker,
			waSend:   waWorker,
			snapshot: snapshotJob,
		}, schedulerSpecs{
			billing:   cfg.BillingCronSpec,
			isolation: cfg.IsolationCronSpec,
			waSend:    cfg.WaSendCronSpec,
			snapshot:  cfg.SnapshotCronSpec,
		}, "tenant-default")
		defer sched.Stop()
		sched.Start()
		logger.WithComponent("Polyglot").WithFields(map[string]any{
			"billing_cron":   cfg.BillingCronSpec,
			"isolation_cron": cfg.IsolationCronSpec,
			"wa_send_cron":   cfg.WaSendCronSpec,
			"snapshot_cron":  cfg.SnapshotCronSpec,
		}).Info("ISP scheduler started (4 jobs)")
	}

	server := &http.Server{
		Addr:        ":" + cfg.Port,
		Handler:     middleware.Chain(rootMux, middleware.RequestID(), middleware.CORS(cfg.CORSOrigins, cfg.AppEnv), middleware.Recovery()),
		ReadTimeout: 30 * time.Second,
		IdleTimeout: 120 * time.Second,
	}

	return &App{
		cfg:           cfg,
		httpServer:    server,
		registry:      reg,
		waManager:     waManager,
		pgStore:       pgStore,
		redisStore:    redisStore,
		sseHub:        sseHub,
		pingStreamMgr: pingStreamMgr,
	}, nil
}

func (a *App) Run() error {
	logger.WithComponent("Polyglot").WithField("address", a.httpServer.Addr).Info("engine starting")
	if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server failed: %w", err)
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	logger.WithComponent("Polyglot").Info("shutting down server...")
	if a.pingStreamMgr != nil {
		a.pingStreamMgr.Stop()
	}
	if a.sseHub != nil {
		a.sseHub.Close()
	}
	if err := a.httpServer.Shutdown(ctx); err != nil {
		logger.WithComponent("Polyglot").WithError(err).Warn("error shutting down HTTP server")
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

func loadInitialDevices(pgStore *postgres.Store, encKey string) (port.DeviceRepository, port.CredentialVault) {
	return postgres.NewDeviceRepository(pgStore.DB()), postgres.NewCredentialVault(pgStore.DB(), encKey)
}

func timeNowUTC() time.Time {
	n := time.Now().UTC()
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
}

var _ = devicepb.Registration{}
