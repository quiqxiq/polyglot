package device

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/quixiq/polyglot/internal/adapter/mcp/mcputil"
)

// ListDevicesArgs is the input schema for list_devices.
type ListDevicesArgs struct {
	Vendor     string `json:"vendor,omitempty" jsonschema:"optional filter by vendor (e.g. mikrotik, zte)"`
	DriverType string `json:"driver_type,omitempty" jsonschema:"optional filter by driver type (e.g. mikrotik, zteolt)"`
}

// Item represents a single device in list_devices.
type Item struct {
	ID         string            `json:"id"`
	TenantID   string            `json:"tenant_id,omitempty"`
	Name       string            `json:"name"`
	Vendor     string            `json:"vendor"`
	DriverType string            `json:"driver_type"`
	Host       string            `json:"host"`
	Port       int               `json:"port"`
	SSHPort    int               `json:"ssh_port,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
	Enabled    bool              `json:"enabled"`
	Extra      map[string]string `json:"extra,omitempty"`
	Status     string            `json:"status"`
}

// ListDevicesOutput is the response payload for list_devices.
type ListDevicesOutput struct {
	Status  string `json:"status"`
	Summary string `json:"summary"`
	Devices []Item `json:"devices,omitempty"`
}

// ListDevices handles the "list_devices" tool.
func (h *Handler) ListDevices(ctx context.Context, _ *mcp.CallToolRequest, args ListDevicesArgs) (*mcp.CallToolResult, ListDevicesOutput, error) {
	repo := h.repo()
	if repo == nil {
		return mcputil.ToolOK(ListDevicesOutput{
			Status:  "success",
			Summary: "Device repository not configured",
		})
	}

	devices, err := repo.FindAll(ctx)
	if err != nil {
		return mcputil.ToolError(ListDevicesOutput{Status: "error", Summary: err.Error()})
	}

	var items []Item
	vendorFilter := strings.ToLower(strings.TrimSpace(args.Vendor))
	driverFilter := strings.ToLower(strings.TrimSpace(args.DriverType))

	for _, d := range devices {
		if vendorFilter != "" && !strings.EqualFold(d.Vendor, vendorFilter) {
			continue
		}
		if driverFilter != "" && !strings.EqualFold(d.DriverType, driverFilter) {
			continue
		}

		status := "disabled"
		if d.Enabled {
			status = "active"
		}

		items = append(items, Item{
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
			Extra:      d.Extra,
			Status:     status,
		})
	}

	if len(items) == 0 {
		return mcputil.ToolOK(ListDevicesOutput{
			Status:  "success",
			Summary: "No devices found matching criteria",
		})
	}

	return mcputil.ToolOK(ListDevicesOutput{
		Status:  "success",
		Summary: fmt.Sprintf("Found %d devices", len(items)),
		Devices: items,
	})
}
