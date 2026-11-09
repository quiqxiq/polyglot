package inventory

import (
	"encoding/json"
	"net/http"

	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

// DeviceDTO represents the non-sensitive public representation of an inventory device.
type DeviceDTO struct {
	ID         string   `json:"id"`
	TenantID   string   `json:"tenant_id,omitempty"`
	Name       string   `json:"name"`
	Vendor     string   `json:"vendor"`
	DriverType string   `json:"driver_type"`
	Host       string   `json:"host"`
	Port       int      `json:"port"`
	SSHPort    int      `json:"ssh_port,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Enabled    bool     `json:"enabled"`
	Status     string   `json:"status"`
}

// Handler serves inventory HTTP endpoints for workbench and UI integrations.
type Handler struct {
	devices port.DeviceRepository
}

// NewHandler constructs an inventory HTTP handler.
func NewHandler(devices port.DeviceRepository) *Handler {
	return &Handler{devices: devices}
}

// RegisterProtected mounts device inventory endpoints on the authenticated mux.
func (h *Handler) RegisterProtected(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/devices", h.handleListDevices)
	mux.HandleFunc("GET /api/v1/devices", h.handleListDevices)
}

func (h *Handler) handleListDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if h.devices == nil {
		_ = json.NewEncoder(w).Encode(map[string]any{"devices": []DeviceDTO{}})
		return
	}

	devices, err := h.devices.FindAll(r.Context())
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	dtos := make([]DeviceDTO, 0, len(devices))
	for _, d := range devices {
		status := "disabled"
		if d.Enabled {
			status = "active"
		}
		dtos = append(dtos, DeviceDTO{
			ID:         d.ID,
			TenantID:   d.TenantID,
			Name:       d.Name,
			Vendor:     d.Vendor,
			DriverType: d.DriverType,
			Host:       d.Host,
			Port:       d.Port,
			SSHPort:    d.SSHPort,
			Tags:       d.Tags,
			Enabled:    d.Enabled,
			Status:     status,
		})
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"devices": dtos})
}
