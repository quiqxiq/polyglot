package customer

import "github.com/modelcontextprotocol/go-sdk/mcp"

// Register registers Customer tools to the MCP server.
func Register(server *mcp.Server, h *Handler) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "customer_lookup",
		Description: "Lookup ISP customer profile and active subscriptions by phone number, name, or customer ID. Read-only.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, h.CustomerLookup)
}
