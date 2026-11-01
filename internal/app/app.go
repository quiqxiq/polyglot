package app

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	acsmanager "github.com/quixiq/polyglot/internal/adapter/acsmanager"
	"github.com/quixiq/polyglot/internal/adapter/auth"
	llmadapter "github.com/quixiq/polyglot/internal/adapter/llm"
	"github.com/quixiq/polyglot/internal/adapter/llm/genkit"
	"github.com/quixiq/polyglot/internal/adapter/midtrans"
	oltmanager "github.com/quixiq/polyglot/internal/adapter/oltmanager"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/provisioner"
	redisAdapter "github.com/quixiq/polyglot/internal/adapter/redis"
	storageAdapter "github.com/quixiq/polyglot/internal/adapter/storage"
	tripay "github.com/quixiq/polyglot/internal/adapter/tripay"
	waAdapter "github.com/quixiq/polyglot/internal/adapter/whatsapp"
	wsAdapter "github.com/quixiq/polyglot/internal/adapter/ws"
	"github.com/quixiq/polyglot/internal/adapter/xendit"
	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/device"
	domainllm "github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/driver/genericssh"
	"github.com/quixiq/polyglot/internal/driver/genieacs"
	"github.com/quixiq/polyglot/internal/driver/hioso"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	mikhmon "github.com/quixiq/polyglot/internal/driver/mikrotik/hotspot"
	"github.com/quixiq/polyglot/internal/driver/whatsapp"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/registry"
	authUC "github.com/quixiq/polyglot/internal/usecase/auth"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
	botUC "github.com/quixiq/polyglot/internal/usecase/bot"
	cashbookUC "github.com/quixiq/polyglot/internal/usecase/cashbook"
	chatUC "github.com/quixiq/polyglot/internal/usecase/chat"
	convUC "github.com/quixiq/polyglot/internal/usecase/conversation"
	customerUC "github.com/quixiq/polyglot/internal/usecase/customer"
	debtUC "github.com/quixiq/polyglot/internal/usecase/debt"
	deviceUC "github.com/quixiq/polyglot/internal/usecase/device"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
	"github.com/quixiq/polyglot/internal/usecase/importer"
	llmUC "github.com/quixiq/polyglot/internal/usecase/llm"
	metricsUC "github.com/quixiq/polyglot/internal/usecase/metrics"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
	notificationUC "github.com/quixiq/polyglot/internal/usecase/notification"
	oltUC "github.com/quixiq/polyglot/internal/usecase/olt"
	ontUC "github.com/quixiq/polyglot/internal/usecase/ont"
	planUC "github.com/quixiq/polyglot/internal/usecase/plan"
	portalUC "github.com/quixiq/polyglot/internal/usecase/portal"
	pppUC "github.com/quixiq/polyglot/internal/usecase/ppp"
	registrationUC "github.com/quixiq/polyglot/internal/usecase/registration"
	reportUC "github.com/quixiq/polyglot/internal/usecase/report"
	jobUCRouter "github.com/quixiq/polyglot/internal/usecase/routerjob"
	settingUC "github.com/quixiq/polyglot/internal/usecase/setting"
	skillUC "github.com/quixiq/polyglot/internal/usecase/skill"
	subUC "github.com/quixiq/polyglot/internal/usecase/subscription"
	userUC "github.com/quixiq/polyglot/internal/usecase/user"
	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/response"
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
	scheduler     *Scheduler
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	logger.Init(cfg.LogLevel, cfg.AppEnv)
	logger.WithComponent("App").Info("initializing application container...")

	dbURL := strings.TrimSpace(cfg.DatabaseURL)
	pgStore, err := postgres.NewStore(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	redisURL := strings.TrimSpace(cfg.RedisURL)
	redisStore, err := redisAdapter.NewStore(redisURL)
	if err != nil {
		logger.WithComponent("App").WithError(err).Warn("failed to connect to Redis; falling back to in-memory cache")
	}

	jwtService := auth.NewJWTService(cfg.JWTSecret, cfg.JWTExpiryHours)

	var refreshSvc *auth.RefreshTokenService
	if redisStore != nil {
		refreshSvc = auth.NewRefreshTokenService(redisStore, time.Duration(cfg.RefreshTokenTTLHours)*time.Hour)
	}

	casbinEnforcer, err := auth.NewCasbinEnforcer(ctx, pgStore.DB())
	if err != nil {
		logger.WithComponent("App").WithError(err).Warn("failed to initialize Casbin enforcer")
	} else {
		auth.SeedSystemPolicies(casbinEnforcer)
		if users, err := pgStore.FindAllUsers(ctx); err == nil {
			refs := make([]*auth.UserRef, 0, len(users))
			for _, u := range users {
				refs = append(refs, &auth.UserRef{ID: fmt.Sprintf("%d", u.ID), Role: u.Role})
			}
			auth.EnsureUserRoleAssignments(casbinEnforcer, refs)
		} else {
			logger.WithComponent("App").WithError(err).Warn("failed to load users for role assignment sync")
		}
	}

	sseHub := wsAdapter.NewSSEHub()

	eventHandler := whatsapp.NewEventHandler(pgStore, sseHub)
	chatService := chatUC.NewChatUseCase(pgStore)
	waManager, err := whatsapp.NewSessionManager(cfg.DatabaseURL, pgStore, nil, eventHandler.MakeStatusCallback())
	if err != nil {
		logger.WithComponent("App").WithError(err).Warn("failed to initialize WhatsApp session manager")
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
	settingRepo := postgres.NewSettingRepository(pgStore.DB()).
		WithVault(postgres.NewCredentialVault(pgStore.DB(), cfg.EncryptionKey))
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
		"hioso": func(ctx context.Context, target device.Target) (port.DeviceDriver, error) {
			return hioso.NewDriver(ctx, target)
		},
	}
	reg := registry.New(repo, vault, factories)

	exec := networkUC.ExecuteCommand
	hotGateway := mikhmon.NewGateway(exec)
	sessionGateway := mikrotik.NewGateway(exec)
	queueGateway := mikrotik.NewQueueGateway(exec)

	// ─── 1. Repositories ────────────────────────────────────────────────────
	customerRepo := postgres.NewCustomerRepository(pgStore.DB())
	invRepo := postgres.NewInvoiceRepository(pgStore.DB())
	subRepo := postgres.NewSubscriptionRepository(pgStore.DB(), vault)
	planRepo := postgres.NewServicePlanRepository(pgStore.DB())
	notifRepo := postgres.NewNotificationRepository(pgStore.DB())
	cashRepo := postgres.NewCashbookRepository(pgStore.DB())
	reportingRepo := postgres.NewReportingRepository(pgStore.DB())
	auditLogRepo := postgres.NewAuditLogRepository(pgStore.DB())
	commandAuditWriter := postgres.NewCommandAuditWriter(auditLogRepo)
	regRepo := postgres.NewRegistrationRepository(pgStore.DB())
	portalRepo := postgres.NewPortalRepository(pgStore.DB())
	gwtxRepo := postgres.NewGatewayTransactionRepository(pgStore.DB())
	paymentReader := postgres.NewPaymentReader(pgStore.DB())
	paymentMethodRepo := postgres.NewPaymentMethodRepository(pgStore.DB())
	debtRepo := postgres.NewDebtRepository(pgStore.DB())
	bulkPaymentRepo := postgres.NewBulkPaymentRepository(pgStore.DB())
	settlementRepo := postgres.NewGatewaySettlementRepository(pgStore.DB())
	approvalRepo := postgres.NewApprovalRepository(pgStore.DB())
	restorePointRepo := postgres.NewRestorePointRepository(pgStore.DB())
	commandLogRepo := postgres.NewCommandLogRepository(pgStore.DB())
	activeDeviceRepo := postgres.NewActiveDeviceRepository(pgStore.DB())

	accountMgr := provisioner.New(reg, sessionGateway, hotGateway, sessionGateway, queueGateway)
	paymentProc := postgres.NewPaymentProcessor(pgStore.DB())
	paymentSubmissionRepo := postgres.NewPaymentSubmissionRepository(pgStore.DB())
	paymentReverser := postgres.NewPaymentReverser(pgStore.DB())
	paymentProc.OnPaid = provisioner.BuildOnPaidRestore(accountMgr, subRepo, planRepo, settingRepo, customerRepo)
	waSender := waAdapter.NewSenderAdapter(waManager, pgStore.FindAllSessions)
	tripayAdapter := tripay.NewAdapter(settingRepo)
	midtransAdapter := midtrans.NewAdapter(settingRepo)
	xenditAdapter := xendit.NewAdapter(settingRepo)

	// OLT/ONT EPON (PLAN-ACS-OLT-HIOSO.md Fase B/C).
	ontRepo := postgres.NewONTRepository(pgStore.DB())
	oltManagerProvider := oltmanager.NewProvider(reg.Get)

	// Antrean job router asynchronous (planning v2 Fase 4).
	routerJobRepo := postgres.NewRouterJobRepository(pgStore.DB())
	massPayUC := billingUC.NewMassPaymentUseCase(bulkPaymentRepo, invRepo, paymentProc, routerJobRepo, settingRepo)
	ontAlerts := ontUC.NewAlertEvaluator(ontRepo, customerRepo, notifRepo, settingRepo)
	oltService := oltUC.NewService(oltUC.Deps{
		Devices: repo, Vault: vault, Managers: oltManagerProvider,
		ONTs: ontRepo, Settings: settingRepo, Jobs: routerJobRepo, Evicter: reg,
		Alerts: ontAlerts,
	})
	acsRepo := postgres.NewACSRepository(pgStore.DB())
	acsManager := acsmanager.New(repo, vault, settingRepo)
	ontService := ontUC.NewService(ontUC.Deps{
		ONTs: ontRepo, ACS: acsManager, ACSRepo: acsRepo,
		Settings: settingRepo, Jobs: routerJobRepo,
		Customers: customerRepo, Subs: subRepo,
	})
	routerJobWorker := jobUCRouter.NewWorker(
		routerJobRepo,
		jobUCRouter.NewHandlers(jobUCRouter.HandlerDeps{
			Subs: subRepo, Plans: planRepo, Manager: accountMgr, Settings: settingRepo,
			OLT: oltService, ACS: ontService, MassPay: massPayUC,
		}),
		settingRepo,
	)
	routerJobExporter := jobUCRouter.NewExporter(routerJobRepo, subRepo, planRepo)

	// ─── 2. Use Cases ───────────────────────────────────────────────────────
	authUseCase := authUC.NewAuthUseCase(userRepo, jwtService, refreshSvc, redisStore, casbinEnforcer)
	refreshUseCase := authUC.NewRefreshTokenUseCase(userRepo, jwtService, refreshSvc, casbinEnforcer)
	manageUserUseCase := userUC.NewManageUserUseCase(userRepo, casbinEnforcer)
	manageSettingUseCase := settingUC.NewManageSettingUseCase(settingRepo)
	custUC := customerUC.NewManageCustomerUseCase(customerRepo, subRepo, invRepo, accountMgr).WithSettings(settingRepo).With360(paymentReader, regRepo)
	devUC := deviceUC.NewManageDeviceUseCase(repo, vault, reg, sessionGateway)
	openTermUC := networkUC.NewOpenTerminalUseCase(repo, vault, genericssh.DialSSHPty)
	hotUC := hotspotUC.New(hotGateway)
	activeSessionsUC := networkUC.NewActiveSessionsUseCase(sessionGateway)
	pppUseCase := pppUC.New(sessionGateway)

	invUC := billingUC.NewInvoiceUseCase(invRepo).
		WithCanceller(postgres.NewInvoiceCanceller(pgStore.DB())).
		WithSettings(settingRepo).
		WithNotifications(notifRepo, customerRepo)
	planUCase := planUC.NewManagePlanUseCase(planRepo, subRepo, accountMgr)
	subUCase := subUC.NewManageSubscriptionUseCase(subRepo, planRepo, customerRepo, repo, accountMgr, auditLogRepo, invRepo).WithSettings(settingRepo)
	checkoutUC := billingUC.NewCheckoutUseCase(invRepo, customerRepo, paymentProc)
	lifecycleUC := subUC.NewLifecycleUseCase(subRepo, planRepo, accountMgr, auditLogRepo).
		WithSettings(settingRepo).
		WithProvisionRetrier(routerJobWorker)
	runBillingUC := billingUC.NewRunBillingUseCase(subRepo, planRepo, invRepo).WithSettings(settingRepo).WithNotifications(notifRepo, customerRepo)
	paymentSubmissionUC := billingUC.NewPaymentSubmissionUseCase(paymentSubmissionRepo, invRepo, paymentProc, settingRepo).WithAudit(auditLogRepo)
	paymentUC := billingUC.NewPaymentUseCase(paymentReverser)
	gatewayRegistry := billingUC.NewGatewayRegistry(settingRepo, tripayAdapter, midtransAdapter, xenditAdapter)
	chargeUC := billingUC.NewGatewayChargeUseCaseWithRegistry(invRepo, customerRepo, gwtxRepo, gatewayRegistry, paymentProc, settingRepo).
		WithCashbook(cashRepo)
	payMethodUC := billingUC.NewPaymentMethodUseCase(paymentMethodRepo)
	debtService := debtUC.NewService(debtRepo, customerRepo, notifRepo, settingRepo)

	// AI Router Ops (PLAN-AI-ROUTER-OPS v2 M1): restore point + approval +
	// execution. Router operations are MikroTik-only in v1.
	restoreKey := cfg.RestorePointEncryptionKey
	if restoreKey == "" {
		restoreKey = cfg.EncryptionKey
	}
	rpStorage, rpErr := storageAdapter.NewRestorePointStorage(cfg.RestorePointDir, restoreKey)
	if rpErr != nil && cfg.AIRouterOpsEnabled {
		return nil, fmt.Errorf("restore point storage: %w", rpErr)
	}
	var rpStoragePort port.RestorePointStorage
	if rpStorage != nil {
		rpStoragePort = rpStorage
	}
	aiOpsResolver := func(ctx context.Context, deviceID string) (port.RouterOps, error) {
		drv, derr := reg.Get(ctx, deviceID)
		if derr != nil {
			return nil, derr
		}
		if _, ok := drv.(*mikrotik.Driver); !ok {
			return nil, fmt.Errorf("AI Router Ops v1 supports MikroTik drivers only (device %s)", deviceID)
		}
		return mikrotik.NewAIOps(drv), nil
	}
	aiDriverResolver := func(ctx context.Context, deviceID string) (port.DeviceDriver, error) {
		return reg.Get(ctx, deviceID)
	}
	restoreUC := networkUC.NewRestorePointUseCase(networkUC.RestorePointDeps{
		Repo: restorePointRepo, Storage: rpStoragePort, Devices: repo, Ops: aiOpsResolver, Approvals: approvalRepo,
		Config: networkUC.RestorePointConfig{
			TTL:      time.Duration(cfg.RestorePointRetentionDays) * 24 * time.Hour,
			MaxBytes: cfg.RestorePointMaxBytes, RetentionDays: cfg.RestorePointRetentionDays,
			KeepLast: cfg.RestorePointKeepLast,
		},
	})
	proposalUC := networkUC.NewProposalUseCase(networkUC.ProposalDeps{
		Approvals: approvalRepo, RestorePoints: restoreUC, Devices: repo,
		Driver: aiDriverResolver, Ops: aiOpsResolver, CommandLogs: commandLogRepo,
		Config: networkUC.ProposalConfig{
			TTL:            time.Duration(cfg.ApprovalTTLMinutes) * time.Minute,
			MaxScriptLines: cfg.AIMaxScriptLines, V6Mutations: cfg.AIV6Mutations,
		},
	})
	executionUC := networkUC.NewExecutionUseCase(networkUC.ExecutionDeps{
		Approvals: approvalRepo, RestorePoints: restorePointRepo, Devices: repo,
		Driver: aiDriverResolver, Ops: aiOpsResolver, CommandLogs: commandLogRepo,
		Restore: restoreUC, WatchdogEnabled: cfg.WatchdogEnabled, WatchdogSeconds: cfg.WatchdogRollbackSeconds,
	})
	decisionUC := networkUC.NewDecisionUseCase(networkUC.DecisionDeps{
		Approvals: approvalRepo, RestorePoints: restoreUC, Devices: repo,
		Driver: aiDriverResolver, Ops: aiOpsResolver, Execution: executionUC,
		RequireRejectReason: cfg.ApprovalRequireReasonOnReject,
	})
	deviceAuthorizer := auth.NewDeviceAuthorizer(userRepo)
	activeDeviceUC := deviceUC.NewActiveDeviceUseCase(activeDeviceRepo, repo, deviceAuthorizer, deviceUC.ActiveDeviceConfig{
		TTL:                  time.Duration(cfg.ActiveDeviceTTLHours) * time.Hour,
		SingleDeviceFallback: cfg.AISingleDeviceFallback,
	})

	regEventRepo := postgres.NewRegistrationEventRepository(pgStore.DB())
	regManagerUC := registrationUC.NewManageRegistrationUseCase(regRepo, notifRepo, auditLogRepo).
		WithEvents(regEventRepo).
		WithPlans(planRepo)
	regConvertUC := registrationUC.NewConvertUseCase(registrationUC.ConvertDeps{
		Repo: regRepo, Plans: planRepo, Customers: customerRepo, Events: regEventRepo,
		Subs: subRepo, Audit: auditLogRepo, Manager: accountMgr, Settings: settingRepo,
		Writer:    postgres.NewConversionWriter(pgStore.DB(), vault),
		ONTBinder: ontService,
	})

	portalUCase := portalUC.NewUseCase(portalRepo, customerRepo, subRepo, invRepo, paymentReader, waSender, notifRepo, settingRepo)

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
	routerSource := importer.NewRouterSource(sessionGateway, hotGateway)
	routerSource.SetPlanRepository(planRepo)
	reconciler := importer.NewReconciler(subRepo, sessionGateway)
	exportUC := importer.NewExportUseCase(subRepo, customerRepo, planRepo, repo)

	metricsRepo := postgres.NewMetricsRepository(pgStore.DB())
	metricsUseCase := deviceUC.NewManageMetricsUseCase(repo, metricsRepo, deviceAuthorizer)
	manageIsolationUseCase := deviceUC.NewManageIsolationUseCase(accountMgr, repo).WithSettings(settingRepo)

	pingStreamMgr := metricsUC.NewPingStreamWorker(repo, metricsRepo, func(c context.Context, id string) (port.DeviceDriver, error) {
		return reg.Get(c, id)
	})
	pingStreamMgr.Start(ctx)

	// Pembersih media sementara bot (PDF/gambar hasil generate) — F6-14.
	mediaCleaner := botUC.NewMediaCleanerWorker(botUC.MediaCleanerConfig{})
	go mediaCleaner.Start(ctx)

	connectDriverProvider := func(ctx context.Context, deviceID string) (port.DeviceDriver, error) {
		callerID, callerRoles, hasIdentity := auth.IdentityFromContext(ctx)
		if hasIdentity && deviceAuthorizer != nil {
			canAccess, err := deviceAuthorizer.CanAccessDevice(ctx, callerID, callerRoles, deviceID)
			if err != nil {
				return nil, err
			}
			if !canAccess {
				return nil, response.PermissionDenied(fmt.Sprintf("access to device %s denied", deviceID))
			}
		}
		return reg.Get(ctx, deviceID)
	}

	cashbookUseCase := cashbookUC.NewManageCashbookUseCase(cashRepo)
	llmConfigUseCase := llmUC.NewManageConfigUseCase(pgStore, cfg.EncryptionKey, func(c context.Context, llmCfg *domainllm.Config, apiKey string) error {
		prov, err := genkit.NewProvider(c, llmCfg, apiKey)
		if err != nil {
			return fmt.Errorf("init genkit provider: %w", err)
		}
		if err := prov.TestConnection(c); err != nil {
			return fmt.Errorf("test genkit connection: %w", err)
		}
		return nil
	})

	driverResolverConnect := func(ctx context.Context, deviceID string) (port.DeviceDriver, bool) {
		drv, err := reg.Get(ctx, deviceID)
		if err != nil {
			return nil, false
		}
		return drv, true
	}

	reportManagerUC := reportUC.NewManageReportUseCase(reportingRepo, reportingRepo).WithAdvancedRepository(reportingRepo)
	notifManagerUC := notificationUC.NewManageNotificationUseCase(notifRepo, waSender)
	streamGW := mikrotik.NewMonitorStreamGateway()

	routerHandler := buildRouter(routerDeps{
		cfg:                    cfg,
		jwtService:             jwtService,
		casbinEnforcer:         casbinEnforcer,
		reportingRepo:          reportingRepo,
		reportManagerUC:        reportManagerUC,
		notifManagerUC:         notifManagerUC,
		streamGW:               streamGW,
		userRepo:               userRepo,
		customerRepo:           customerRepo,
		deviceRepo:             repo,
		regRepo:                regRepo,
		cashbookUseCase:        cashbookUseCase,
		notifRepo:              notifRepo,
		pgStore:                pgStore,
		reg:                    reg,
		waManager:              waManager,
		sseHub:                 sseHub,
		waSender:               waSender,
		chatService:            chatService,
		skillUseCase:           skillUseCase,
		convService:            convService,
		botEngine:              botEngine,
		llmConfigUseCase:       llmConfigUseCase,
		authUseCase:            authUseCase,
		refreshUseCase:         refreshUseCase,
		manageUserUseCase:      manageUserUseCase,
		manageSettingUseCase:   manageSettingUseCase,
		custUC:                 custUC,
		devUC:                  devUC,
		openTermUC:             openTermUC,
		hotUC:                  hotUC,
		activeSessionsUC:       activeSessionsUC,
		pppUseCase:             pppUseCase,
		invUC:                  invUC,
		planUCase:              planUCase,
		subUCase:               subUCase,
		checkoutUC:             checkoutUC,
		lifecycleUC:            lifecycleUC,
		runBillingUC:           runBillingUC,
		paymentSubmissionUC:    paymentSubmissionUC,
		paymentUC:              paymentUC,
		chargeUC:               chargeUC,
		payMethodUC:            payMethodUC,
		massPayUC:              massPayUC,
		debtService:            debtService,
		settlementRepo:         settlementRepo,
		gwtxRepo:               gwtxRepo,
		auditWriter:            commandAuditWriter,
		approvalRepo:           approvalRepo,
		restoreUC:              restoreUC,
		proposalUC:             proposalUC,
		decisionUC:             decisionUC,
		activeDevices:          activeDeviceRepo,
		activeDeviceUC:         activeDeviceUC,
		deviceAuthorizer:       deviceAuthorizer,
		aiOpsResolver:          aiOpsResolver,
		regManagerUC:           regManagerUC,
		regConvertUC:           regConvertUC,
		portalUCase:            portalUCase,
		upsertImport:           upsertImport,
		routerSource:           routerSource,
		reconciler:             reconciler,
		exportUC:               exportUC,
		metricsUseCase:         metricsUseCase,
		routerJobRepo:          routerJobRepo,
		routerJobExporter:      routerJobExporter,
		manageIsolationUseCase: manageIsolationUseCase,
		oltService:             oltService,
		ontService:             ontService,
		settingReader:          settingRepo,
		connectDriverProvider:  connectDriverProvider,
		driverResolverConnect:  driverResolverConnect,
	})

	// ─── 4. Scheduler (Cron) ────────────────────────────────────────────────
	var sched *Scheduler
	if cfg.SchedulerEnabled && cfg.BillingCronSpec != "" && cfg.IsolationCronSpec != "" {
		isolateWorker := billingUC.NewIsolateWorker(
			subRepo, invRepo, customerRepo, planRepo, accountMgr, notifRepo, settingRepo,
		)
		reminderWorker := billingUC.NewReminderWorker(subRepo, invRepo, customerRepo, notifRepo).WithSettings(settingRepo)
		waWorker := notificationUC.NewWASenderWorker(notifRepo, waSender, settingRepo)
		snapshotJob := func(ctx context.Context) error {
			return reportingRepo.RecomputeDaily(ctx, "tenant-default", timeNowUTC())
		}
		routerJobsJob := func(ctx context.Context) error {
			res, rerr := routerJobWorker.Run(ctx)
			if res.Processed > 0 {
				logger.WithComponent("App").WithFields(map[string]any{
					"processed": res.Processed, "succeeded": res.Succeeded,
					"retried": res.Retried, "dead": res.Dead,
					"stopped_early": res.StoppedEarly,
				}).Info("router job batch summary")
			}
			return rerr
		}
		driftJob := func(ctx context.Context) error {
			return runDriftReconciliation(ctx, repo, driverResolverConnect, reconciler)
		}
		acsMatchJob := func(ctx context.Context) error {
			stats, aerr := ontService.SyncDevices(ctx)
			if aerr == nil {
				logger.WithComponent("App").WithFields(map[string]any{
					"devices": stats.Devices, "matched": stats.Matched,
					"unmatched": stats.Unmatched, "conflicts": stats.Conflicts,
				}).Info("acs match summary")
			}
			return aerr
		}
		acsTelemetryJob := func(ctx context.Context) error {
			ok, terr := ontService.RefreshTelemetryAll(ctx)
			if terr == nil {
				logger.WithComponent("App").WithField("updated", ok).Info("acs telemetry summary")
			}
			return terr
		}
		oltSyncJob := func(ctx context.Context) error {
			res, serr := oltService.SyncAll(ctx)
			for deviceID, r := range res {
				logger.WithComponent("App").WithFields(map[string]any{
					"device_id": deviceID, "discovered": r.Discovered, "updated": r.Updated,
					"removed": r.Removed, "online": r.Online, "offline": r.Offline,
				}).Info("olt sync summary")
			}
			return serr
		}
		backupScheduler := deviceUC.NewBackupScheduler(deviceUC.BackupScheduler{
			Devices: repo, Restore: restoreUC, Approvals: approvalRepo, TenantID: "tenant-default",
		})
		approvalExpireJob := func(ctx context.Context) error {
			n, err := approvalRepo.ExpirePending(ctx, time.Now().UTC())
			if err == nil && n > 0 {
				logger.WithComponent("Scheduler").WithField("expired", n).Info("approval expiry summary")
			}
			return err
		}
		restoreCleanupJob := func(ctx context.Context) error {
			protected := map[string]bool{}
			for _, status := range []string{"PENDING", "APPROVED", "EXECUTING"} {
				items, _, lerr := approvalRepo.List(ctx, port.ApprovalFilter{TenantID: "tenant-default", Status: status, Limit: 500})
				if lerr != nil {
					return lerr
				}
				for _, a := range items {
					if a.RestorePointID != "" {
						protected[a.RestorePointID] = true
					}
				}
			}
			n, err := restoreUC.Cleanup(ctx, time.Now().UTC(), cfg.RestorePointKeepLast, protected)
			if err == nil && n > 0 {
				logger.WithComponent("Scheduler").WithField("deleted", n).Info("restore point cleanup summary")
			}
			return err
		}
		retentionJob := func(ctx context.Context) error {
			now := time.Now().UTC()
			logCutoff := now.AddDate(0, 0, -cfg.CommandLogRetentionDays)
			logsDeleted, lerr := commandLogRepo.DeleteBefore(ctx, logCutoff)
			if lerr != nil {
				return fmt.Errorf("cleanup command logs: %w", lerr)
			}
			if logsDeleted > 0 {
				logger.WithComponent("Scheduler").WithField("command_logs", logsDeleted).Info("retention cleanup summary")
			}
			return nil
		}
		backupJob := func(ctx context.Context) error {
			created, failed, err := backupScheduler.Run(ctx)
			if err == nil && created+failed > 0 {
				logger.WithComponent("Scheduler").WithFields(map[string]any{"created": created, "failed": failed}).Info("scheduled backup summary")
			}
			return err
		}
		sched = newScheduler(schedulerJobs{
			billing:        runBillingUC,
			isolate:        isolateWorker,
			reminder:       reminderWorker,
			waSend:         waWorker,
			routerJobs:     routerJobsJob,
			drift:          driftJob,
			snapshot:       snapshotJob,
			oltSync:        oltSyncJob,
			acsMatch:       acsMatchJob,
			acsTelemetry:   acsTelemetryJob,
			approvalExpire: approvalExpireJob,
			restoreCleanup: restoreCleanupJob,
			backupDaily:    backupJob,
			backupWeekly:   backupJob,
			retention:      retentionJob,
		}, schedulerSpecs{
			billing:        cfg.BillingCronSpec,
			isolation:      cfg.IsolationCronSpec,
			reminder:       cfg.ReminderCronSpec,
			waSend:         cfg.WaSendCronSpec,
			routerJobs:     cfg.RouterJobCronSpec,
			drift:          cfg.DriftCronSpec,
			snapshot:       cfg.SnapshotCronSpec,
			oltSync:        cfg.OltSyncCronSpec,
			acsMatch:       cfg.ACSMatchCronSpec,
			acsTelemetry:   cfg.ACSTelemetryCron,
			approvalExpire: "@every 5m",
			restoreCleanup: "0 4 * * *",
			backupDaily:    cronFromClock(cfg.BackupDailyAt, ""),
			backupWeekly:   cronFromClock(cfg.BackupWeeklyAt, cfg.BackupWeeklyDay),
			retention:      "30 4 * * *",
		}, "tenant-default", postgres.NewAdvisoryLocker(pgStore.DB()))
		sched.Start()
		logger.WithComponent("App").WithFields(map[string]any{
			"billing_cron":    cfg.BillingCronSpec,
			"isolation_cron":  cfg.IsolationCronSpec,
			"reminder_cron":   cfg.ReminderCronSpec,
			"wa_send_cron":    cfg.WaSendCronSpec,
			"snapshot_cron":   cfg.SnapshotCronSpec,
			"router_job_cron": cfg.RouterJobCronSpec,
			"drift_cron":      cfg.DriftCronSpec,
		}).Info("ISP scheduler started (7 jobs)")
	}

	server := &http.Server{
		Addr:        ":" + cfg.Port,
		Handler:     routerHandler,
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
		scheduler:     sched,
	}, nil
}

func (a *App) Run() error {
	logger.WithComponent("App").WithField("address", a.httpServer.Addr).Info("engine starting")
	if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server failed: %w", err)
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	logger.WithComponent("App").Info("shutting down server...")
	if a.scheduler != nil {
		a.scheduler.Stop()
	}
	if a.pingStreamMgr != nil {
		a.pingStreamMgr.Stop()
	}
	if a.sseHub != nil {
		a.sseHub.Close()
	}
	if err := a.httpServer.Shutdown(ctx); err != nil {
		logger.WithComponent("App").WithError(err).Warn("error shutting down HTTP server")
		_ = a.httpServer.Close()
	}
	if a.waManager != nil {
		a.waManager.DisconnectAll()
	}
	if a.registry != nil {
		_ = a.registry.Close()
	}
	logger.WithComponent("App").Info("shutdown complete")
	return nil
}

func loadInitialDevices(pgStore *postgres.Store, encKey string) (port.DeviceRepository, port.CredentialVault) {
	return postgres.NewDeviceRepository(pgStore.DB()), postgres.NewCredentialVault(pgStore.DB(), encKey)
}

func timeNowUTC() time.Time {
	n := time.Now().UTC()
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
}

// cronFromClock converts "HH:MM" (and optional weekday) into a 5-field cron
// expression. Invalid input returns "" so the scheduler skips the job.
func cronFromClock(hhmm, weekday string) string {
	parts := strings.Split(strings.TrimSpace(hhmm), ":")
	if len(parts) != 2 {
		return ""
	}
	hour, errH := strconv.Atoi(parts[0])
	minute, errM := strconv.Atoi(parts[1])
	if errH != nil || errM != nil || hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return ""
	}
	if strings.TrimSpace(weekday) == "" {
		return fmt.Sprintf("%d %d * * *", minute, hour)
	}
	days := map[string]string{"sun": "0", "mon": "1", "tue": "2", "wed": "3", "thu": "4", "fri": "5", "sat": "6"}
	dow, ok := days[strings.ToLower(strings.TrimSpace(weekday))]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%d %d * * %s", minute, hour, dow)
}
