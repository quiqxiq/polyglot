package device

import "github.com/modelcontextprotocol/go-sdk/mcp"

// Register registers generic network device tools to the MCP server.
func Register(server *mcp.Server, h *Handler) {
	destructive := true

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_device_status",
		Description: "Check live connectivity and operational status of a network device by directly probing it via API/SSH. Returns live status (online or error/timeout with diagnostic summary). Read-only — no approval needed.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, h.GetDeviceStatus)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "run_command",
		Description: "Execute a raw vendor-native command on non-MikroTik devices (Cisco, OLT, Switch). NOTE: For MikroTik RouterOS devices, DO NOT use this tool — use 'routeros_run' with 'routeros_validate' and 'routeros_suggest' instead for pre-flight syntax safety. Potentially destructive — approval required.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive},
	}, h.RunCommand)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "push_config",
		Description: "Push a configuration change to a device. Destructive — approval required.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive},
	}, h.PushConfig)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_devices",
		Description: "List registered network devices (MikroTik, OLT, Switch) from inventory database. NOTE: Device status in the list is only the inventory registration state (active/inactive in DB). To verify live physical connectivity and operational health, you must call 'get_device_status'. Read-only.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, h.ListDevices)
}
