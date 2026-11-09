package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	authConnect "github.com/quixiq/polyglot/internal/adapter/connect/auth"
	billingConnect "github.com/quixiq/polyglot/internal/adapter/connect/billing"
	cashbookConnect "github.com/quixiq/polyglot/internal/adapter/connect/cashbook"
	customerConnect "github.com/quixiq/polyglot/internal/adapter/connect/customer"
	debtConnect "github.com/quixiq/polyglot/internal/adapter/connect/debt"
	deviceConnect "github.com/quixiq/polyglot/internal/adapter/connect/device"
	hotspotConnect "github.com/quixiq/polyglot/internal/adapter/connect/hotspot"
	ispadminConnect "github.com/quixiq/polyglot/internal/adapter/connect/ispadmin"
	monitorConnect "github.com/quixiq/polyglot/internal/adapter/connect/monitor"
	networkConnect "github.com/quixiq/polyglot/internal/adapter/connect/network"
	notificationConnect "github.com/quixiq/polyglot/internal/adapter/connect/notification"
	oltConnect "github.com/quixiq/polyglot/internal/adapter/connect/olt"
	ontConnect "github.com/quixiq/polyglot/internal/adapter/connect/ont"
	planConnect "github.com/quixiq/polyglot/internal/adapter/connect/plan"
	portalConnect "github.com/quixiq/polyglot/internal/adapter/connect/portal"
	pppConnect "github.com/quixiq/polyglot/internal/adapter/connect/ppp"
	registrationConnect "github.com/quixiq/polyglot/internal/adapter/connect/registration"
	reportConnect "github.com/quixiq/polyglot/internal/adapter/connect/report"
	routerjobConnect "github.com/quixiq/polyglot/internal/adapter/connect/routerjob"
	settingConnect "github.com/quixiq/polyglot/internal/adapter/connect/setting"
	subConnect "github.com/quixiq/polyglot/internal/adapter/connect/subscription"
	adminapi "github.com/quixiq/polyglot/internal/adapter/http/adminapi"
	gatewayHTTP "github.com/quixiq/polyglot/internal/adapter/http/gateway"
	inventoryHTTP "github.com/quixiq/polyglot/internal/adapter/http/inventory"
	"github.com/quixiq/polyglot/internal/adapter/http/middleware"
	portalHTTP "github.com/quixiq/polyglot/internal/adapter/http/portal"
	reportHTTP "github.com/quixiq/polyglot/internal/adapter/http/report"
	uploadHTTP "github.com/quixiq/polyglot/internal/adapter/http/upload"
	webhookHTTP "github.com/quixiq/polyglot/internal/adapter/http/webhook"
	whatsappHTTP "github.com/quixiq/polyglot/internal/adapter/http/whatsapp"
	"github.com/quixiq/polyglot/internal/adapter/mcp"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
	wsAdapter "github.com/quixiq/polyglot/internal/adapter/ws"
	"github.com/quixiq/polyglot/internal/config"
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
	whatsappUC "github.com/quixiq/polyglot/internal/usecase/whatsapp"
)

type routerDeps struct {
	cfg                    config.Config
	jwtService             *auth.JWTService
	casbinEnforcer         *auth.CasbinEnforcer
	reportingRepo          *postgres.ReportingRepository
	userRepo               port.UserRepository
	customerRepo           port.CustomerRepository
	deviceRepo             port.DeviceRepository
	regRepo                port.RegistrationRepository
	cashbookUseCase        *cashbookUC.ManageCashbookUseCase
	notifRepo              port.NotificationRepository
	pgStore                *postgres.Store
	reg                    *registry.Registry
	waManager              *whatsapp.SessionManager
	sseHub                 *wsAdapter.SSEHub
	waSender               port.NotificationSender
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
	paymentSubmissionUC    *billingUC.PaymentSubmissionUseCase
	paymentUC              *billingUC.PaymentUseCase
	chargeUC               *billingUC.GatewayChargeUseCase
	payMethodUC            *billingUC.PaymentMethodUseCase
	massPayUC              *billingUC.MassPaymentUseCase
	debtService            *debtUC.Service
	settlementRepo         port.GatewaySettlementRepository
	gwtxRepo               port.GatewayTransactionRepository
	regManagerUC           *registrationUC.ManageRegistrationUseCase
	regConvertUC           *registrationUC.ConvertUseCase
	portalUCase            *portalUC.UseCase
	upsertImport           *importer.UpsertUseCase
	routerSource           *importer.RouterSource
	reconciler             *importer.Reconciler
	exportUC               *importer.ExportUseCase
	metricsUseCase         *deviceUC.ManageMetricsUseCase
	routerJobRepo          port.RouterJobRepository
	routerJobExporter      *jobUCRouter.Exporter
	manageIsolationUseCase *deviceUC.ManageIsolationUseCase
	oltService             *oltUC.Service
	ontService             *ontUC.Service
	settingReader          port.SettingReader
	cacheStore             port.CacheStore
	notifManagerUC         *notificationUC.ManageNotificationUseCase
	reportManagerUC        *reportUC.ManageReportUseCase
	streamGW               port.MonitorStreamGateway
	connectDriverProvider  func(ctx context.Context, deviceID string) (port.DeviceDriver, error)
	driverResolverConnect  func(ctx context.Context, deviceID string) (port.DeviceDriver, bool)
}

// protectedPlainPaths lists plain-HTTP handlers mounted on protectedMux that
// must also be reachable from the root mux. Keep this in sync with the
// RegisterProtected calls above.
var protectedPlainPaths = []string{
	"/api/devices",
	"/api/v1/devices",
	"/api/cashier/charge",
	"/api/cashier/check-charge",
	"/api/cashier/gateways",
	"/api/reports/daily", "/api/reports/monthly", "/api/reports/yearly",
	"/api/admin/import", "/api/admin/import-router", "/api/admin/reconcile",
	"/api/admin/export", "/api/admin/snapshot/refresh",
	"/api/whatsapp/sessions",
	"/api/whatsapp/session/status",
	"/api/whatsapp/session/qr",
	"/api/whatsapp/session/pair-code",
	"/api/whatsapp/session/reconnect",
	"/api/whatsapp/session/logout",
	"/api/whatsapp/send/message",
	"/api/whatsapp/send/image",
	"/api/whatsapp/send/document",
	"/api/whatsapp/presence",
	"/api/whatsapp/message/read",
	"/api/whatsapp/identity",
}

func buildRouter(d routerDeps) (http.Handler, error) {
	rootMux := http.NewServeMux()

	webhookHTTPHandler := webhookHTTP.NewHandler(d.deviceRepo)
	webhookHTTPHandler.RegisterPublic(rootMux)

	inventoryHTTPHandler := inventoryHTTP.NewHandler(d.deviceRepo)

	uploadHTTPHandler := uploadHTTP.NewHandler(d.cfg.UploadDir)
	uploadHTTPHandler.RegisterPublic(rootMux)

	authPath, authHandler := authConnect.NewAuthServiceHandler(
		d.authUseCase, d.refreshUseCase, d.manageUserUseCase, d.cfg.IsProduction(),
	)
	rootMux.Handle(authPath, authHandler)

	registrationConnect.RegisterPublic(rootMux, d.regManagerUC)
	billingConnect.RegisterPublic(rootMux, d.paymentSubmissionUC)

	portalConnectPath, portalConnectHandler := portalConnect.NewPortalServiceHandler(d.portalUCase)
	rootMux.Handle(portalConnectPath, portalConnectHandler)

	portalHTTPHandler := portalHTTP.NewHandler(d.portalUCase, d.chargeUC)
	portalHTTPHandler.RegisterPublic(rootMux)
	portalHTTPHandler.RegisterAuthenticated(rootMux)

	gatewayHTTPHandler := gatewayHTTP.NewHandler(d.chargeUC)
	gatewayHTTPHandler.RegisterPublic(rootMux)

	mcpServer := mcp.New(d.reg, nil).
		WithMikhmonUseCase(d.hotUC).
		WithOLTService(d.oltService).
		WithCustomerRepository(d.customerRepo).
		WithDeviceRepository(d.deviceRepo)
	rootMux.Handle("/mcp", mcpServer.HTTPHandler())

	wsAdapter.RegisterEventRoutes(rootMux, d.sseHub, d.openTermUC)

	protectedMux := http.NewServeMux()

	var protectedPaths []string
	registerProtected := func(servicePath string, handler http.Handler) {
		protectedMux.Handle(servicePath, handler)
		protectedPaths = append(protectedPaths, servicePath)
	}

	devPath, devHandler := deviceConnect.NewDeviceServiceHandler(d.devUC, d.openTermUC, d.connectDriverProvider, d.metricsUseCase, d.manageIsolationUseCase, d.streamGW)
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

	billingHandler := billingConnect.NewBillingConnectHandler(d.invUC, d.checkoutUC, d.runBillingUC).
		WithPaymentMethods(d.payMethodUC).
		WithMassPayment(d.massPayUC).
		WithGateway(d.chargeUC, d.settlementRepo, d.gwtxRepo)
	paymentHandler := billingConnect.NewPaymentConnectHandler(d.paymentSubmissionUC, d.paymentUC)
	billingPath, billingHTTPHandler := billingConnect.NewBillingServiceHandler(billingHandler, paymentHandler)
	registerProtected(billingPath, billingHTTPHandler)

	debtPath, debtHandler := debtConnect.NewDebtServiceHandler(d.debtService)
	registerProtected(debtPath, debtHandler)

	regPath, regHandler := registrationConnect.NewRegistrationServiceHandler(d.regManagerUC, d.regConvertUC)
	registerProtected(regPath, regHandler)

	rjPath, rjHandler := routerjobConnect.NewRouterJobServiceHandler(d.routerJobRepo, d.routerJobExporter)
	registerProtected(rjPath, rjHandler)

	oltPath, oltHandler := oltConnect.NewOltServiceHandler(d.oltService, d.settingReader)
	registerProtected(oltPath, oltHandler)

	ontPath, ontHandler := ontConnect.NewOntServiceHandler(d.ontService)
	registerProtected(ontPath, ontHandler)

	cashbookPath, cashbookHandler := cashbookConnect.NewCashbookServiceHandler(d.cashbookUseCase)
	registerProtected(cashbookPath, cashbookHandler)

	notifPath, notifHandler := notificationConnect.NewNotificationServiceHandler(d.notifManagerUC)
	registerProtected(notifPath, notifHandler)

	reportPath, reportHandler := reportConnect.NewReportServiceHandler(d.reportManagerUC)
	registerProtected(reportPath, reportHandler)

	adminPath, adminHandler := ispadminConnect.NewISPAdminServiceHandler(
		d.upsertImport, d.routerSource, d.reconciler, d.exportUC, d.driverResolverConnect)
	registerProtected(adminPath, adminHandler)

	mikhmonPath, mikhmonHandler := hotspotConnect.NewHotspotServiceHandler(d.hotUC, d.connectDriverProvider)
	registerProtected(mikhmonPath, mikhmonHandler)

	networkPath, networkHandler := networkConnect.NewNetworkServiceHandler(d.hotUC, d.activeSessionsUC, d.connectDriverProvider)
	registerProtected(networkPath, networkHandler)

	monitorPath, monitorHandler := monitorConnect.NewNetworkMonitorServiceHandler(d.hotUC, d.activeSessionsUC, d.connectDriverProvider, d.streamGW)
	registerProtected(monitorPath, monitorHandler)

	pppPath, pppHandler := pppConnect.NewPPPServiceHandler(d.pppUseCase, d.connectDriverProvider, d.streamGW)
	registerProtected(pppPath, pppHandler)

	probePath, probeHandler := deviceConnect.NewProbeServiceHandler()
	registerProtected(probePath, probeHandler)

	gatewayHTTPHandler.RegisterProtected(protectedMux)
	inventoryHTTPHandler.RegisterProtected(protectedMux)
	reportHTTP.NewHandler(d.reportingRepo).Register(protectedMux)
	adminapi.NewHandler(d.upsertImport, d.routerSource, d.reconciler, d.reportingRepo, d.exportUC, d.driverResolverConnect).
		RegisterProtected(protectedMux)
	whatsappResolver := whatsappUC.NewIdentityResolver(d.userRepo, d.customerRepo, d.cacheStore)
	whatsappHTTP.NewHandler(d.pgStore, d.waManager, whatsappResolver).RegisterProtected(protectedMux)

	protectedHandler := middleware.Chain(
		protectedMux,
		middleware.AuthenticateJWT(d.jwtService),
		middleware.AuthorizeProcedure(d.casbinEnforcer),
	)

	// Plain-HTTP handlers mounted on protectedMux must also be reachable from
	// the root mux; protectedPlainPaths is the single source for that list.
	allProtectedPaths := make([]string, 0, len(protectedPaths)+len(protectedPlainPaths))
	allProtectedPaths = append(allProtectedPaths, protectedPaths...)
	allProtectedPaths = append(allProtectedPaths, protectedPlainPaths...)
	if err := validateProtectedRoutePermissions(allProtectedPaths); err != nil {
		return nil, err
	}

	for _, path := range protectedPaths {
		rootMux.Handle(path, protectedHandler)
	}
	for _, plainPath := range protectedPlainPaths {
		rootMux.Handle(plainPath, protectedHandler)
	}

	return middleware.Chain(
		rootMux,
		middleware.RequestID(),
		middleware.RequestLogger(),
		middleware.CORS(d.cfg.CORSOrigins, d.cfg.AppEnv),
		middleware.Recovery(),
	), nil
}

// validateProtectedRoutePermissions fails startup when a protected route has no
// RBAC mapping. A missing mapping would otherwise fail closed at request time
// (403) and only be discovered by users.
func validateProtectedRoutePermissions(paths []string) error {
	var missing []string
	for _, path := range paths {
		if strings.HasSuffix(path, "/") {
			if !hasPermissionPrefix(path) {
				missing = append(missing, path)
			}
			continue
		}
		if _, ok := auth.PermissionFor(path); !ok {
			missing = append(missing, path)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("protected routes missing RBAC permission mapping: %s", strings.Join(missing, ", "))
}

func hasPermissionPrefix(prefix string) bool {
	for procedure := range auth.ProcedurePermissions {
		if strings.HasPrefix(procedure, prefix) {
			return true
		}
	}
	return false
}
