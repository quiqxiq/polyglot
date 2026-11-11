package ont

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/quixiq/polyglot/internal/adapter/mcp/mcputil"
	domainONT "github.com/quixiq/polyglot/internal/domain/ont"
	"github.com/quixiq/polyglot/internal/port"
)

type ontStatusArgs struct {
	ONTID         string `json:"ont_id,omitempty" jsonschema:"exact ONT ID"`
	CustomerID    string `json:"customer_id,omitempty" jsonschema:"customer ID linked to the ONT"`
	PPPoEUsername string `json:"pppoe_username,omitempty" jsonschema:"PPPoE username linked to the ONT"`
}

type ontSummary struct {
	ID             string     `json:"id"`
	OLTDeviceID    string     `json:"olt_device_id"`
	PONIndex       int        `json:"pon_index"`
	ONUIndex       int        `json:"onu_index"`
	Name           string     `json:"name,omitempty"`
	MACAddress     string     `json:"mac_address,omitempty"`
	SerialNumber   string     `json:"serial_number,omitempty"`
	Model          string     `json:"model,omitempty"`
	State          string     `json:"state"`
	AdminState     string     `json:"admin_state,omitempty"`
	RxPowerDBM     *float64   `json:"rx_power_dbm,omitempty"`
	TxPowerDBM     *float64   `json:"tx_power_dbm,omitempty"`
	DistanceM      *int       `json:"distance_m,omitempty"`
	TemperatureC   *float64   `json:"temperature_c,omitempty"`
	LastSeenAt     *time.Time `json:"last_seen_at,omitempty"`
	LastPolledAt   *time.Time `json:"last_polled_at,omitempty"`
	CustomerID     *string    `json:"customer_id,omitempty"`
	SubscriptionID *string    `json:"subscription_id,omitempty"`
	PPPoEUsername  string     `json:"pppoe_username,omitempty"`
}

type ontStatusOutput struct {
	Status  string       `json:"status"`
	Summary string       `json:"summary"`
	ONTs    []ontSummary `json:"onts,omitempty"`
}

type ontListArgs struct {
	CustomerID  string `json:"customer_id,omitempty" jsonschema:"filter by linked customer ID"`
	OLTDeviceID string `json:"olt_device_id,omitempty" jsonschema:"filter by OLT device ID"`
	State       string `json:"state,omitempty" jsonschema:"filter by state: ONLINE, OFFLINE, UNKNOWN, or REMOVED"`
	Search      string `json:"search,omitempty" jsonschema:"search MAC, serial number, model, or PPPoE username"`
	Limit       int    `json:"limit,omitempty" jsonschema:"maximum number of ONTs to return (default 20, max 100)"`
}

// GetONTStatus handles the "ont_get_status" tool. It only reads inventory and
// the latest persisted optical metrics; it never triggers an ACS refresh.
func (h *Handler) GetONTStatus(ctx context.Context, _ *mcp.CallToolRequest, args ontStatusArgs) (*mcp.CallToolResult, ontStatusOutput, error) {
	if strings.TrimSpace(args.ONTID) == "" && strings.TrimSpace(args.CustomerID) == "" && strings.TrimSpace(args.PPPoEUsername) == "" {
		return mcputil.ToolError(ontStatusOutput{Status: "error", Summary: "at least one criterion (ont_id, customer_id, or pppoe_username) is required"})
	}

	repo := h.repo()
	if repo == nil {
		return mcputil.ToolOK(ontStatusOutput{Status: "success", Summary: "ONT repository not configured in server mode"})
	}

	var matched []domainONT.ONT
	if id := strings.TrimSpace(args.ONTID); id != "" {
		o, err := repo.FindByID(ctx, id)
		if err != nil {
			return mcputil.ToolError(ontStatusOutput{Status: "error", Summary: err.Error()})
		}
		matched = append(matched, o)
	} else {
		filter := port.ONTFilter{
			CustomerID: strings.TrimSpace(args.CustomerID),
			Search:     strings.TrimSpace(args.PPPoEUsername),
			Limit:      20,
		}
		list, err := repo.List(ctx, filter)
		if err != nil {
			return mcputil.ToolError(ontStatusOutput{Status: "error", Summary: err.Error()})
		}
		matched = list
	}

	if len(matched) == 0 {
		return mcputil.ToolOK(ontStatusOutput{Status: "success", Summary: "No ONT matched the search criteria"})
	}

	out := make([]ontSummary, len(matched))
	for i, o := range matched {
		out[i] = summarize(o)
	}
	return mcputil.ToolOK(ontStatusOutput{
		Status:  "success",
		Summary: fmt.Sprintf("Found %d ONT(s)", len(out)),
		ONTs:    out,
	})
}

// ListONTs handles the "ont_list" tool.
func (h *Handler) ListONTs(ctx context.Context, _ *mcp.CallToolRequest, args ontListArgs) (*mcp.CallToolResult, ontStatusOutput, error) {
	repo := h.repo()
	if repo == nil {
		return mcputil.ToolOK(ontStatusOutput{Status: "success", Summary: "ONT repository not configured in server mode"})
	}

	limit := args.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	list, err := repo.List(ctx, port.ONTFilter{
		CustomerID:  strings.TrimSpace(args.CustomerID),
		OLTDeviceID: strings.TrimSpace(args.OLTDeviceID),
		State:       strings.ToUpper(strings.TrimSpace(args.State)),
		Search:      strings.TrimSpace(args.Search),
		Limit:       limit,
	})
	if err != nil {
		return mcputil.ToolError(ontStatusOutput{Status: "error", Summary: err.Error()})
	}
	if len(list) == 0 {
		return mcputil.ToolOK(ontStatusOutput{Status: "success", Summary: "No ONT matched the filters"})
	}

	out := make([]ontSummary, len(list))
	for i, o := range list {
		out[i] = summarize(o)
	}
	return mcputil.ToolOK(ontStatusOutput{
		Status:  "success",
		Summary: fmt.Sprintf("Found %d ONT(s)", len(out)),
		ONTs:    out,
	})
}

func summarize(o domainONT.ONT) ontSummary {
	return ontSummary{
		ID:             o.ID,
		OLTDeviceID:    o.OLTDeviceID,
		PONIndex:       o.PONIndex,
		ONUIndex:       o.ONUIndex,
		Name:           o.Name,
		MACAddress:     o.MACAddress,
		SerialNumber:   o.SerialNumber,
		Model:          o.Model,
		State:          string(o.State),
		AdminState:     string(o.AdminState),
		RxPowerDBM:     o.RxPowerDBM,
		TxPowerDBM:     o.TxPowerDBM,
		DistanceM:      o.DistanceM,
		TemperatureC:   o.TemperatureC,
		LastSeenAt:     o.LastSeenAt,
		LastPolledAt:   o.LastPolledAt,
		CustomerID:     o.CustomerID,
		SubscriptionID: o.SubscriptionID,
		PPPoEUsername:  o.PPPoEUsername,
	}
}
