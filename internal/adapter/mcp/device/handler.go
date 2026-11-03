package device

import (
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/registry"
)

// Handler holds the dependencies needed by generic network device tools.
type Handler struct {
	Registry      *registry.Registry
	GetDeviceRepo func() port.DeviceRepository
	Audit         port.AuditWriter
}

// NewHandler creates a new device tool handler.
func NewHandler(reg *registry.Registry, repoGetter func() port.DeviceRepository, audit port.AuditWriter) *Handler {
	return &Handler{
		Registry:      reg,
		GetDeviceRepo: repoGetter,
		Audit:         audit,
	}
}

func (h *Handler) repo() port.DeviceRepository {
	if h.GetDeviceRepo != nil {
		if r := h.GetDeviceRepo(); r != nil {
			return r
		}
	}
	if h.Registry != nil {
		return h.Registry.DeviceRepo()
	}
	return nil
}
