package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/quixiq/polyglot/internal/adapter/auth"
)

type RBACHandler struct {
	enforcer *auth.CasbinEnforcer
}

func NewRBACHandler(enforcer *auth.CasbinEnforcer) *RBACHandler {
	return &RBACHandler{enforcer: enforcer}
}

type AddPolicyRequest struct {
	Role   string `json:"role" binding:"required"`
	Path   string `json:"path" binding:"required"`
	Method string `json:"method" binding:"required"`
}

type AssignRoleRequest struct {
	User string `json:"user" binding:"required"`
	Role string `json:"role" binding:"required"`
}

func (h *RBACHandler) ListPolicies(c *gin.Context) {
	policies, err := h.enforcer.GetPolicies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"policies": policies})
}

func (h *RBACHandler) AddPolicy(c *gin.Context) {
	var req AddPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role, Path, dan Method wajib diisi"})
		return
	}

	ok, err := h.enforcer.AddPolicy(req.Role, req.Path, req.Method)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !ok {
		c.JSON(http.StatusConflict, gin.H{"error": "Aturan RBAC tersebut sudah ada"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Aturan RBAC berhasil ditambahkan secara dinamis",
		"policy":  []string{req.Role, req.Path, req.Method},
	})
}

func (h *RBACHandler) RemovePolicy(c *gin.Context) {
	var req AddPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role, Path, dan Method wajib diisi"})
		return
	}

	ok, err := h.enforcer.RemovePolicy(req.Role, req.Path, req.Method)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Aturan RBAC tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Aturan RBAC berhasil dihapus"})
}

func (h *RBACHandler) ListRoleAssignments(c *gin.Context) {
	roles, err := h.enforcer.GetGroupingPolicies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

func (h *RBACHandler) AssignRole(c *gin.Context) {
	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User dan Role wajib diisi"})
		return
	}

	ok, err := h.enforcer.AddRoleForUser(req.User, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !ok {
		c.JSON(http.StatusConflict, gin.H{"error": "Penugasan role pengguna tersebut sudah ada"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Role pengguna berhasil ditugaskan",
		"user":    req.User,
		"role":    req.Role,
	})
}

func (h *RBACHandler) UnassignRole(c *gin.Context) {
	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User dan Role wajib diisi"})
		return
	}

	ok, err := h.enforcer.DeleteRoleForUser(req.User, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Penugasan role pengguna tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Penugasan role pengguna berhasil dihapus"})
}
