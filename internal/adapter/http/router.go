package http

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/internal/adapter/http/middleware"
	"github.com/quixiq/polyglot/internal/adapter/ws"
)

// NewRouter builds and returns the REST API router for remaining un-migrated REST routes (Auth & RBAC) and SSE Events.
func NewRouter(ctx context.Context) (*gin.Engine, error) {
	r := gin.Default()
	return r, nil
}

// RegisterAuthRoutes registers authentication endpoints (POST /api/v1/auth/login, GET /api/v1/auth/me).
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

// RegisterEventRoutes registers the realtime SSE event streaming endpoint.
func RegisterEventRoutes(r *gin.Engine, sseHub *ws.SSEHub) {
	api := r.Group("/api/v1")
	api.Use(middleware.CORS(nil))
	api.GET("/events", sseHub.RegisterClient)
}
