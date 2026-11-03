package olt

import (
	"github.com/quixiq/polyglot/internal/registry"
	oltUC "github.com/quixiq/polyglot/internal/usecase/olt"
)

// Handler holds dependencies for OLT MCP tools (EPON/GPON).
type Handler struct {
	Registry      *registry.Registry
	GetOLTService func() *oltUC.Service
}

// NewHandler creates a new OLT tool handler.
func NewHandler(reg *registry.Registry, svcGetter func() *oltUC.Service) *Handler {
	return &Handler{
		Registry:      reg,
		GetOLTService: svcGetter,
	}
}

// Service returns the configured OLT service instance, or nil if not configured.
func (h *Handler) Service() *oltUC.Service {
	if h.GetOLTService != nil {
		return h.GetOLTService()
	}
	return nil
}
