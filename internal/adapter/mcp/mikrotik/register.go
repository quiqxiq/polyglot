package mikrotik

import "github.com/modelcontextprotocol/go-sdk/mcp"

// Register registers MikroTik/Mikhmon tools to the MCP server.
func Register(server *mcp.Server, h *Handler) {
	destructive := true

	mcp.AddTool(server, &mcp.Tool{
		Name:        "mikhmon_get_dashboard",
		Description: "Get Mikrotik Mikhmon hotspot dashboard metrics (CPU, Memory, Uptime, Active users, Today income). Read-only.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, h.MikhmonGetDashboard)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "mikhmon_generate_voucher",
		Description: "Generate a batch of hotspot user vouchers for a given profile. Destructive — approval required.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive},
	}, h.MikhmonGenerateVoucher)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "mikhmon_kick_session",
		Description: "Disconnect an active hotspot user session by username, IP, or session ID. Destructive — approval required.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive},
	}, h.MikhmonKickSession)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "routeros_list_menu",
		Description: "List RouterOS submenus, commands, or parameters at a given hierarchy path using /console/inspect request=child. Read-only and cached.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, h.RouterOSListMenu)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "routeros_describe_command",
		Description: "Describe available arguments and their syntax types for a RouterOS command path (e.g. 'ip,address,add') using /console/inspect request=child and request=syntax. Read-only and cached.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, h.RouterOSDescribeCommand)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "routeros_suggest",
		Description: "Get real-time auto-complete candidates and dynamic parameter values (e.g. interface names, IP pools, profiles) for a console input string using /console/inspect request=completion. Read-only and uncached.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, h.RouterOSSuggest)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "routeros_validate",
		Description: "Dry-run syntax and parameter validation for a RouterOS console command without executing it, using /console/inspect request=error. Returns exact error message and position if invalid. Read-only and uncached.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, h.RouterOSValidate)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "routeros_run",
		Description: "Execute a RouterOS command with pre-flight syntax validation via RouterOS inspector. Destructive — approval required.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive},
	}, h.RouterOSRun)
}
