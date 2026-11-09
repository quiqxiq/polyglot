package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	acsmanager "github.com/quixiq/polyglot/internal/adapter/acsmanager"
	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/internal/adapter/midtrans"
	oltmanager "github.com/quixiq/polyglot/internal/adapter/oltmanager"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/provisioner"
	redisAdapter "github.com/quixiq/polyglot/internal/adapter/redis"
	tripay "github.com/quixiq/polyglot/internal/adapter/tripay"
	waAdapter "github.com/quixiq/polyglot/internal/adapter/whatsapp"
	wsAdapter "github.com/quixiq/polyglot/internal/adapter/ws"
	"github.com/quixiq/polyglot/internal/adapter/xendit"
	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/device"
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
	cashbookUC "github.com/quixiq/polyglot/internal/usecase/cashbook"
	customerUC "github.com/quixiq/polyglot/internal/usecase/customer"
	debtUC "github.com/quixiq/polyglot/internal/usecase/debt"
	deviceUC "github.com/quixiq/polyglot/internal/usecase/device"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
	"github.com/quixiq/polyglot/internal/usecase/importer"
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
	waManager, err := whatsapp.NewSessionManager(cfg.DatabaseURL, pgStore, nil, eventHandler.MakeStatusCallback())
	if err != nil {
		logger.WithComponent("App").WithError(err).Warn("failed to initialize WhatsApp session manager")
	}

	userRepo := postgres.NewUserRepository(pgStore.DB())
	settingRepo := postgres.NewSettingRepository(pgStore.DB()).
		WithVault(postgres.NewCredentialVault(pgStore.DB(), cfg.EncryptionKey))

	if waManager != nil {
		waManager.SetMessageCallback(eventHandler.MakeMessageCallback(nil))
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
	regRepo := postgres.NewRegistrationRepository(pgStore.DB())
	portalRepo := postgres.NewPortalRepository(pgStore.DB())
	gwtxRepo := postgres.NewGatewayTransactionRepository(pgStore.DB())
	paymentReader := postgres.NewPaymentReader(pgStore.DB())
	paymentMethodRepo := postgres.NewPaymentMethodRepository(pgStore.DB())
	debtRepo := postgres.NewDebtRepository(pgStore.DB())
	bulkPaymentRepo := postgres.NewBulkPaymentRepository(pgStore.DB())
	settlementRepo := postgres.NewGatewaySettlementRepository(pgStore.DB())

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

	deviceAuthorizer := auth.NewDeviceAuthorizer(userRepo)

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

	routerHandler, err := buildRouter(routerDeps{
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
		cacheStore:             redisStore,
		connectDriverProvider:  connectDriverProvider,
		driverResolverConnect:  driverResolverConnect,
	})
	if err != nil {
		return nil, fmt.Errorf("build router: %w", err)
	}

	// ─── 4. Scheduler (Cron) ────────────────────────────────────────────────
	sched := buildScheduler(ctx, schedulerDeps{
		cfg:            cfg,
		pgStore:        pgStore,
		subscriptions:  subRepo,
		invoices:       invRepo,
		customers:      customerRepo,
		plans:          planRepo,
		accountManager: accountMgr,
		notifications:  notifRepo,
		settings:       settingRepo,
		waSender:       waSender,
		reporting:      reportingRepo,
		routerJobs:     routerJobWorker,
		devices:        repo,
		driverResolver: driverResolverConnect,
		reconciler:     reconciler,
		ontService:     ontService,
		oltService:     oltService,
		billing:        runBillingUC,
	})

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
