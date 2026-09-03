package app

import (
	"context"
	"net/http"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	authConnect "github.com/quixiq/polyglot/internal/adapter/connect/auth"
	billingConnect "github.com/quixiq/polyglot/internal/adapter/connect/billing"
	botConnect "github.com/quixiq/polyglot/internal/adapter/connect/bot"
	cashbookConnect "github.com/quixiq/polyglot/internal/adapter/connect/cashbook"
	customerConnect "github.com/quixiq/polyglot/internal/adapter/connect/customer"
	deviceConnect "github.com/quixiq/polyglot/internal/adapter/connect/device"
	hotspotConnect "github.com/quixiq/polyglot/internal/adapter/connect/hotspot"
	ispadminConnect "github.com/quixiq/polyglot/internal/adapter/connect/ispadmin"
	llmConnect "github.com/quixiq/polyglot/internal/adapter/connect/llm"
	monitorConnect "github.com/quixiq/polyglot/internal/adapter/connect/monitor"
	networkConnect "github.com/quixiq/polyglot/internal/adapter/connect/network"
	notificationConnect "github.com/quixiq/polyglot/internal/adapter/connect/notification"
	planConnect "github.com/quixiq/polyglot/internal/adapter/connect/plan"
	portalConnect "github.com/quixiq/polyglot/internal/adapter/connect/portal"
	pppConnect "github.com/quixiq/polyglot/internal/adapter/connect/ppp"
	registrationConnect "github.com/quixiq/polyglot/internal/adapter/connect/registration"
	reportConnect "github.com/quixiq/polyglot/internal/adapter/connect/report"
	settingConnect "github.com/quixiq/polyglot/internal/adapter/connect/setting"
	skillConnect "github.com/quixiq/polyglot/internal/adapter/connect/skill"
	subConnect "github.com/quixiq/polyglot/internal/adapter/connect/subscription"
	adminapi "github.com/quixiq/polyglot/internal/adapter/http/adminapi"
	gatewayHTTP "github.com/quixiq/polyglot/internal/adapter/http/gateway"
	"github.com/quixiq/polyglot/internal/adapter/http/middleware"
	portalHTTP "github.com/quixiq/polyglot/internal/adapter/http/portal"
	reportHTTP "github.com/quixiq/polyglot/internal/adapter/http/report"
	webhookHTTP "github.com/quixiq/polyglot/internal/adapter/http/webhook"
	"github.com/quixiq/polyglot/internal/adapter/mcp"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
	wsAdapter "github.com/quixiq/polyglot/internal/adapter/ws"
	"github.com/quixiq/polyglot/internal/config"
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
	deviceUC "github.com/quixiq/polyglot/internal/usecase/device"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
	"github.com/quixiq/polyglot/internal/usecase/importer"
	llmUC "github.com/quixiq/polyglot/internal/usecase/llm"
	networkUC "github.com/quixiq/polyglot/internal/usecase/network"
	planUC "github.com/quixiq/polyglot/internal/usecase/plan"
	portalUC "github.com/quixiq/polyglot/internal/usecase/portal"
	pppUC "github.com/quixiq/polyglot/internal/usecase/ppp"
	registrationUC "github.com/quixiq/polyglot/internal/usecase/registration"
	settingUC "github.com/quixiq/polyglot/internal/usecase/setting"
	skillUC "github.com/quixiq/polyglot/internal/usecase/skill"
	subUC "github.com/quixiq/polyglot/internal/usecase/subscription"
	userUC "github.com/quixiq/polyglot/internal/usecase/user"
)

type routerDeps struct {
	cfg                    config.Config
	jwtService             *auth.JWTService
	casbinEnforcer         *auth.CasbinEnforcer
	reportingRepo          *postgres.ReportingRepository
	userRepo               port.UserRepository
	customerRepo           port.CustomerRepository
	regRepo                port.RegistrationRepository
	cashbookUseCase        *cashbookUC.ManageCashbookUseCase
	notifRepo              port.NotificationRepository
	pgStore                *postgres.Store
	reg                    *registry.Registry
	waManager              *whatsapp.SessionManager
	sseHub                 *wsAdapter.SSEHub
	waSender               port.NotificationSender
	chatService            *chatUC.ChatUseCase
	skillUseCase           *skillUC.ManageSkillUseCase
	convService            *convUC.ConversationUseCase
	botEngine              *botUC.Engine
	llmConfigUseCase       *llmUC.ManageConfigUseCase
	authUseCase            *authUC.AuthUseCase
	refreshUseCase         *authUC.RefreshTokenUseCase
	manageUserUseCase      *userUC.ManageUserUseCase
	manageSettingUseCase   *settingUC.ManageSettingUseCase
	custUC                 *customerUC.ManageCustomerUseCase
	devUC                  *deviceUC.ManageDeviceUseCase
	openTermUC             *networkUC.OpenTerminalUseCase
	hotUC                  *hotspotUC.UseCase
	activeSessionsUC       *networkUC.ActiveSessionsUseCase
	pppUseCase             *pppUC.UseCase
	invUC                  *billingUC.InvoiceUseCase
	planUCase              *planUC.ManagePlanUseCase
	subUCase               *subUC.ManageSubscriptionUseCase
	checkoutUC             *billingUC.CheckoutUseCase
	lifecycleUC            *subUC.LifecycleUseCase
	runBillingUC           *billingUC.RunBillingUseCase
	chargeUC               *billingUC.GatewayChargeUseCase
	regManagerUC           *registrationUC.ManageRegistrationUseCase
	regConvertUC           *registrationUC.ConvertUseCase
	portalUCase            *portalUC.UseCase
	upsertImport           *importer.UpsertUseCase
	routerSource           *importer.RouterSource
	reconciler             *importer.Reconciler
	exportUC               *importer.ExportUseCase
	metricsUseCase         *deviceUC.ManageMetricsUseCase
	manageIsolationUseCase *deviceUC.ManageIsolationUseCase
	connectDriverProvider  func(ctx context.Context, deviceID string) (port.DeviceDriver, error)
	driverResolverConnect  func(ctx context.Context, deviceID string) (port.DeviceDriver, bool)
}

func buildRouter(d routerDeps) http.Handler {
	rootMux := http.NewServeMux()

	webhookHTTPHandler := webhookHTTP.NewHandler()
	webhookHTTPHandler.RegisterPublic(rootMux)

	authPath, authHandler := authConnect.NewAuthServiceHandler(
		d.authUseCase, d.refreshUseCase, d.manageUserUseCase, d.cfg.AppEnv == "production",
	)
	rootMux.Handle(authPath, authHandler)

	pubRegPath, pubRegHandler := registrationConnect.NewPublicSubmitHandler(d.regManagerUC)
	rootMux.Handle(pubRegPath, pubRegHandler)

	portalConnectPath, portalConnectHandler := portalConnect.NewPortalServiceHandler(d.portalUCase)
	rootMux.Handle(portalConnectPath, portalConnectHandler)

	portalHTTPHandler := portalHTTP.NewHandler(d.portalUCase)
	portalHTTPHandler.RegisterPublic(rootMux)
	portalHTTPHandler.RegisterAuthenticated(rootMux)

	gatewayHTTPHandler := gatewayHTTP.NewHandler(d.chargeUC)
	gatewayHTTPHandler.RegisterPublic(rootMux)

	mcpServer := mcp.New(d.reg, nil).
		WithMikhmonUseCase(d.hotUC).
		WithCustomerRepository(d.customerRepo)
	rootMux.Handle("/mcp", mcpServer.HTTPHandler())

	wsAdapter.RegisterEventRoutes(rootMux, d.sseHub, d.openTermUC)

	protectedMux := http.NewServeMux()

	var protectedPaths []string
	registerProtected := func(servicePath string, handler http.Handler) {
		protectedMux.Handle(servicePath, handler)
		protectedPaths = append(protectedPaths, servicePath)
	}

	devPath, devHandler := deviceConnect.NewDeviceServiceHandler(d.devUC, d.openTermUC, d.connectDriverProvider, d.metricsUseCase, d.manageIsolationUseCase)
	registerProtected(devPath, devHandler)

	custPath, custHandler := customerConnect.NewCustomerServiceHandler(d.custUC)
	registerProtected(custPath, custHandler)

	rbacPath, rbacHandler := authConnect.NewRBACServiceHandler(d.casbinEnforcer)
	registerProtected(rbacPath, rbacHandler)

	userPath, userHandler := authConnect.NewUserServiceHandler(d.manageUserUseCase)
	registerProtected(userPath, userHandler)

	settingPath, settingHandler := settingConnect.NewSettingServiceHandler(d.manageSettingUseCase)
	registerProtected(settingPath, settingHandler)

	planPath, planHandler := planConnect.NewPlanServiceHandler(d.planUCase)
	registerProtected(planPath, planHandler)

	subPath, subHandler := subConnect.NewSubscriptionServiceHandler(d.subUCase, d.lifecycleUC)
	registerProtected(subPath, subHandler)

	billingPath, billingHandler := billingConnect.NewBillingServiceHandler(d.invUC, d.checkoutUC, d.runBillingUC)
	registerProtected(billingPath, billingHandler)

	regPath, regHandler := registrationConnect.NewRegistrationServiceHandler(d.regManagerUC, d.regConvertUC, d.regRepo)
	registerProtected(regPath, regHandler)

	cashbookPath, cashbookHandler := cashbookConnect.NewCashbookServiceHandler(d.cashbookUseCase)
	registerProtected(cashbookPath, cashbookHandler)

	notifPath, notifHandler := notificationConnect.NewNotificationServiceHandler(d.notifRepo, d.waSender)
	registerProtected(notifPath, notifHandler)

	reportPath, reportHandler := reportConnect.NewReportServiceHandler(d.reportingRepo, d.reportingRepo)
	registerProtected(reportPath, reportHandler)

	adminPath, adminHandler := ispadminConnect.NewISPAdminServiceHandler(
		d.upsertImport, d.routerSource, d.reconciler, d.exportUC, d.driverResolverConnect)
	registerProtected(adminPath, adminHandler)

	mikhmonPath, mikhmonHandler := hotspotConnect.NewHotspotServiceHandler(d.hotUC, d.connectDriverProvider)
	registerProtected(mikhmonPath, mikhmonHandler)

	networkPath, networkHandler := networkConnect.NewNetworkServiceHandler(d.hotUC, d.activeSessionsUC, d.connectDriverProvider)
	registerProtected(networkPath, networkHandler)

	monitorPath, monitorHandler := monitorConnect.NewNetworkMonitorServiceHandler(d.hotUC, d.activeSessionsUC, d.connectDriverProvider)
	registerProtected(monitorPath, monitorHandler)

	pppPath, pppHandler := pppConnect.NewPPPServiceHandler(d.pppUseCase, d.connectDriverProvider)
	registerProtected(pppPath, pppHandler)

	waPath, waHandler := botConnect.NewWhatsAppServiceHandler(d.pgStore, d.waManager, d.chatService)
	registerProtected(waPath, waHandler)

	skillPath, skillHandler := skillConnect.NewSkillServiceHandler(d.skillUseCase)
	registerProtected(skillPath, skillHandler)

	llmPath, llmHandler := llmConnect.NewLLMConfigServiceHandler(d.llmConfigUseCase)
	registerProtected(llmPath, llmHandler)

	botPath, botHandler := botConnect.NewBotServiceHandler(d.convService, d.botEngine)
	registerProtected(botPath, botHandler)

	probePath, probeHandler := deviceConnect.NewProbeServiceHandler()
	registerProtected(probePath, probeHandler)

	gatewayHTTPHandler.RegisterProtected(protectedMux)
	reportHTTP.NewHandler(d.reportingRepo).Register(protectedMux)
	adminapi.NewHandler(d.upsertImport, d.routerSource, d.reconciler, d.reportingRepo, d.exportUC, d.driverResolverConnect).
		RegisterProtected(protectedMux)

	protectedHandler := middleware.Chain(
		protectedMux,
		middleware.AuthenticateJWT(d.jwtService),
		middleware.AuthorizeProcedure(d.casbinEnforcer),
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

	return middleware.Chain(
		rootMux,
		middleware.RequestID(),
		middleware.RequestLogger(),
		middleware.CORS(d.cfg.CORSOrigins, d.cfg.AppEnv),
		middleware.Recovery(),
	)
}
