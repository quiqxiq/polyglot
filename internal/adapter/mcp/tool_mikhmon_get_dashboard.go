package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mikhmonGetDashboardArgs struct {
	DeviceID string `json:"device_id" jsonschema:"the Mikrotik device ID from inventory"`
}

type mikhmonDashboardOutput struct {
	DeviceID    string `json:"device_id"`
	Status      string `json:"status"`
	Summary     string `json:"summary"`
	Uptime      string `json:"uptime,omitempty"`
	Version     string `json:"version,omitempty"`
	CPULoad     int    `json:"cpu_load"`
	BoardName   string `json:"board_name,omitempty"`
	Identity    string `json:"identity,omitempty"`
	TotalUsers  int    `json:"total_users"`
	ActiveUsers int    `json:"active_users"`
	TodayIncome int64  `json:"today_income"`
}

func (s *Server) mikhmonGetDashboard(ctx context.Context, _ *mcp.CallToolRequest, args mikhmonGetDashboardArgs) (*mcp.CallToolResult, mikhmonDashboardOutput, error) {
	if args.DeviceID == "" {
		return toolError(mikhmonDashboardOutput{Status: "error", Summary: "device_id is required"})
	}

	driver, err := s.registry.Get(ctx, args.DeviceID)
	if err != nil {
		return toolError(mikhmonDashboardOutput{DeviceID: args.DeviceID, Status: "error", Summary: err.Error()})
	}

	if s.mikhmonUC == nil {
		return toolError(mikhmonDashboardOutput{DeviceID: args.DeviceID, Status: "error", Summary: "mikhmon use case not configured"})
	}

	summary, err := s.mikhmonUC.GetDashboardSummary(ctx, driver)
	if err != nil {
		return toolError(mikhmonDashboardOutput{DeviceID: args.DeviceID, Status: "error", Summary: err.Error()})
	}

	humanSummary := fmt.Sprintf("Identity: %s, Board: %s, CPU Load: %d%%, Active Users: %d/%d, Today Income: Rp %d",
		summary.Identity, summary.BoardName, summary.CPULoad, summary.ActiveUsers, summary.TotalUsers, summary.TodayIncome)

	out := mikhmonDashboardOutput{
		DeviceID:    args.DeviceID,
		Status:      "success",
		Summary:     humanSummary,
		Uptime:      summary.Uptime,
		Version:     summary.Version,
		CPULoad:     summary.CPULoad,
		BoardName:   summary.BoardName,
		Identity:    summary.Identity,
		TotalUsers:  summary.TotalUsers,
		ActiveUsers: summary.ActiveUsers,
		TodayIncome: summary.TodayIncome,
	}

	return toolOK(out)
}
