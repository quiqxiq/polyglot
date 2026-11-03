package olt

import "github.com/modelcontextprotocol/go-sdk/mcp"

// Register registers OLT management tools to the MCP server.
// Tools can be added here as new capabilities are implemented (e.g. ONU status, optical power).
func Register(_ *mcp.Server, _ *Handler) {
	// Future tools to register:
	// - olt_get_onu_status
	// - olt_get_optical_power
	// - olt_reboot_onu
}
