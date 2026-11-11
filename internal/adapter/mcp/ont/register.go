package ont

import "github.com/modelcontextprotocol/go-sdk/mcp"

// Register registers read-only ONT/ONU tools to the MCP server.
func Register(server *mcp.Server, h *Handler) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ont_get_status",
		Description: "Get ONT/ONU status, optical power, and last-seen data by ONT ID, customer ID, or PPPoE username. Read-only and never triggers a refresh.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, h.GetONTStatus)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ont_list",
		Description: "List ONT/ONU inventory filtered by customer, OLT, state, or search term. Read-only.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, h.ListONTs)
}
