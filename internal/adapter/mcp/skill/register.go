package skill

import "github.com/modelcontextprotocol/go-sdk/mcp"

// Register registers Skill tools to the MCP server.
func Register(server *mcp.Server, h *Handler) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_skills",
		Description: "List active modular skills and standard operational procedures (SOP). Read-only.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, h.ListSkills)
}
