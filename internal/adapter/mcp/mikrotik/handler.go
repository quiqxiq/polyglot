package mikrotik

import (
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/registry"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
)

// Handler holds dependencies for MikroTik and Mikhmon MCP tools.
type Handler struct {
	Registry     *registry.Registry
	GetMikhmonUC func() *hotspotUC.UseCase
	Audit        port.AuditWriter
	Inspector    *InspectClient
}

// HandlerOption configures a Handler.
type HandlerOption func(*Handler)

// WithAudit configures the audit writer for the handler.
func WithAudit(audit port.AuditWriter) HandlerOption {
	return func(h *Handler) {
		h.Audit = audit
	}
}

// WithInspector configures a custom InspectClient for the handler.
func WithInspector(inspector *InspectClient) HandlerOption {
	return func(h *Handler) {
		h.Inspector = inspector
	}
}

// NewHandler creates a new MikroTik tool handler.
func NewHandler(reg *registry.Registry, ucGetter func() *hotspotUC.UseCase, opts ...HandlerOption) *Handler {
	h := &Handler{
		Registry:     reg,
		GetMikhmonUC: ucGetter,
		Inspector:    NewInspectClient(),
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *Handler) uc() *hotspotUC.UseCase {
	if h.GetMikhmonUC != nil {
		return h.GetMikhmonUC()
	}
	return nil
}

func (h *Handler) inspector() *InspectClient {
	if h.Inspector != nil {
		return h.Inspector
	}
	return NewInspectClient()
}
