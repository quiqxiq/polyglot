package http

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/internal/adapter/http/middleware"
	"github.com/quixiq/polyglot/internal/adapter/ws"
)

// NewRouter builds and returns the REST API router.
func NewRouter(ctx context.Context) (*gin.Engine, error) {
	r := gin.Default()
	return r, nil
}

// RegisterAuthRoutes registers authentication endpoints.
func RegisterAuthRoutes(r *gin.Engine, authH *AuthHandler, jwtService *auth.JWTService) {
	authGroup := r.Group("/api/v1/auth")
	{
		authGroup.POST("/login", authH.Login)
		authGroup.GET("/me", middleware.AuthenticateJWT(jwtService), authH.Me)
	}
}

// RegisterRBACRoutes registers dynamic Casbin policy management endpoints.
func RegisterRBACRoutes(r *gin.Engine, rbacH *RBACHandler, jwtService *auth.JWTService, enforcer *auth.CasbinEnforcer) {
	rbacGroup := r.Group("/api/v1/rbac")
	rbacGroup.Use(middleware.AuthenticateJWT(jwtService), middleware.AuthorizeCasbin(enforcer))
	{
		rbacGroup.GET("/policies", rbacH.ListPolicies)
		rbacGroup.POST("/policies", rbacH.AddPolicy)
		rbacGroup.DELETE("/policies", rbacH.RemovePolicy)
		rbacGroup.GET("/roles", rbacH.ListRoleAssignments)
		rbacGroup.POST("/roles/assign", rbacH.AssignRole)
		rbacGroup.DELETE("/roles/assign", rbacH.UnassignRole)
	}
}

// RegisterBotRoutes registers REST API endpoints for WA Sessions, Conversations, Knowledge, LLM Configs, and Technicians.
func RegisterBotRoutes(
	r *gin.Engine,
	sessionH *SessionHandler,
	convH *ConversationHandler,
	knowledgeH *KnowledgeHandler,
	llmH *LLMConfigHandler,
	techH *TechnicianHandler,
	sseHub *ws.SSEHub,
	jwtService *auth.JWTService,
	enforcer *auth.CasbinEnforcer,
) {
	api := r.Group("/api/v1")
	api.Use(middleware.CORS(nil))

	// Realtime SSE Stream Endpoint
	api.GET("/events", sseHub.RegisterClient)

	// Protected Admin & Bot Routes
	protected := api.Group("")
	protected.Use(middleware.AuthenticateJWT(jwtService), middleware.AuthorizeCasbin(enforcer))
	{
		// WA Sessions
		protected.GET("/sessions", sessionH.ListSessions)
		protected.POST("/sessions", sessionH.CreateSession)
		protected.GET("/sessions/:id/qr", sessionH.GetQRCode)
		protected.POST("/sessions/:id/pairing", sessionH.GetPairingCode)
		protected.POST("/sessions/:id/pairing-code", sessionH.GetPairingCode)
		protected.PUT("/sessions/:id/toggle-bot", sessionH.ToggleBot)
		protected.PUT("/sessions/:id/webhook", sessionH.UpdateWebhook)
		protected.POST("/sessions/:id/logout", sessionH.LogoutSession)
		protected.POST("/sessions/:id/reconnect", sessionH.ReconnectSession)
		protected.DELETE("/sessions/:id", sessionH.DeleteSession)

		// Conversations
		protected.GET("/conversations", convH.ListConversations)
		protected.GET("/conversations/:id", convH.GetConversation)
		protected.POST("/conversations/:id/takeover", convH.TakeOver)
		protected.POST("/conversations/:id/take-over", convH.TakeOver)
		protected.POST("/conversations/:id/reset-bot", convH.ResetBot)
		protected.POST("/conversations/:id/messages", convH.SendMessage)
		protected.POST("/conversations/:id/close", convH.CloseConversation)

		// Knowledge Base
		protected.GET("/knowledge", knowledgeH.ListKnowledge)
		protected.POST("/knowledge", knowledgeH.CreateKnowledge)
		protected.GET("/knowledge/:id", knowledgeH.GetKnowledge)
		protected.PUT("/knowledge/:id", knowledgeH.UpdateKnowledge)
		protected.DELETE("/knowledge/:id", knowledgeH.DeleteKnowledge)

		// LLM Configs
		protected.GET("/llm-configs", llmH.ListConfigs)
		protected.POST("/llm-configs", llmH.CreateConfig)
		protected.PUT("/llm-configs/:id", llmH.UpdateConfig)
		protected.POST("/llm-configs/:id/activate", llmH.ActivateConfig)
		protected.POST("/llm-configs/:id/test", llmH.TestConfig)
		protected.DELETE("/llm-configs/:id", llmH.DeleteConfig)

		// Technicians
		protected.GET("/technicians", techH.ListTechnicians)
		protected.POST("/technicians", techH.CreateTechnician)
		protected.PUT("/technicians/:id", techH.UpdateTechnician)
		protected.PUT("/technicians/:id/toggle-active", techH.ToggleActive)
		protected.DELETE("/technicians/:id", techH.DeleteTechnician)
	}
}

// RegisterMikhmonRoutes registers REST API endpoints for Mikhmon & Hotspot administration.
func RegisterMikhmonRoutes(r *gin.Engine, h *MikhmonHandler) {
	api := r.Group("/api/v1/devices/:deviceId/mikhmon")
	{
		api.GET("/dashboard", h.GetDashboard)
		api.GET("/income", h.GetIncome)

		// Hotspot User Profile CRUD
		api.GET("/profiles", h.GetProfiles)
		api.POST("/profiles", h.CreateProfile)
		api.PUT("/profiles/:rosId", h.UpdateProfile)
		api.DELETE("/profiles/:rosId", h.DeleteProfile)

		// Hotspot User CRUD
		api.GET("/users", h.GetUsers)
		api.GET("/users/:rosId", h.GetUser)
		api.POST("/users", h.AddUser)
		api.PUT("/users/:rosId", h.UpdateUser)
		api.DELETE("/users/:rosId", h.RemoveUser)
		api.POST("/users/:rosId/reset-counters", h.ResetUserCounters)

		// Voucher: generate (ke RouterOS saja) + generate+render HTML
		api.GET("/vouchers", h.GetVouchersByTag)
		api.POST("/vouchers/generate", h.GenerateVouchers)
		api.POST("/vouchers/render", h.RenderVoucherHTML)

		// Hotspot Sessions
		api.GET("/active", h.GetActiveSessions)
		api.DELETE("/active/:rosId", h.RemoveActiveSession)
		api.GET("/hotspot/active", h.GetActiveSessions)
		api.GET("/hotspot/inactive", h.GetHotspotInactiveUsers)
		api.DELETE("/hotspot/active/:rosId", h.RemoveActiveSession)

		// PPPoE Sessions
		api.GET("/ppp/active", h.GetPPPActiveSessions)
		api.GET("/ppp/inactive", h.GetPPPInactiveSessions)
		api.DELETE("/ppp/active/:rosId", h.RemovePPPActiveSession)

		// DHCP Leases
		api.GET("/dhcp/leases", h.GetDHCPLeases)
		api.POST("/dhcp/leases/:rosId/block", h.BlockDHCPLease)

		api.GET("/hosts", h.GetHosts)
		api.DELETE("/hosts/:rosId", h.RemoveHost)
		api.GET("/servers", h.GetServers)

		api.GET("/pools", h.GetIPPools)
		api.GET("/parent-queues", h.GetParentQueues)
		api.GET("/nat-rules", h.GetNATRules)

		api.POST("/expire-monitor", h.SetupExpireMonitor)

		// Reports with optional date/month/year filtering
		api.GET("/reports", h.GetReportsByDate)
		api.DELETE("/reports/:rosId", h.DeleteReport)
	}
}

// RegisterDeviceRoutes registers REST API endpoints for Device Inventory CRUD and live testing.
func RegisterDeviceRoutes(r *gin.Engine, h *DeviceHandler) {
	devices := r.Group("/api/v1/devices")
	{
		devices.POST("", h.CreateDevice)
		devices.GET("", h.ListDevices)
		devices.GET("/:deviceId", h.GetDevice)
		devices.PUT("/:deviceId", h.UpdateDevice)
		devices.DELETE("/:deviceId", h.DeleteDevice)
		devices.POST("/:deviceId/test", h.TestConnection)
	}
}
