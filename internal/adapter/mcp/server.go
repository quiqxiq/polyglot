package mcp

import (
	"context"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/registry"
)

// Server wraps the MCP SDK server and holds the dependencies every tool
// handler needs: the device driver registry (for looking up a connected
// driver by device_id) and an optional audit writer (for recording tool
// executions per Polyglot-Architecture.md §2 prinsip 8).
//
// MCP surface is intentionally narrow per Polyglot-Architecture.md §2
// prinsip 3: only get_device_status, run_command, and push_config are
// exposed — operations that genuinely need AI judgment. Admin CRUD
// (device/customer/subscription/plan) stays in internal/adapter/http.
type Server struct {
	mcpServer *mcp.Server
	registry  *registry.Registry
	audit     port.AuditWriter
}

// New builds an MCP Server with all three network tools registered. The
// registry resolves device_id → connected driver; the audit writer records
// executions (pass nil to skip auditing until the audit adapter is wired).
func New(reg *registry.Registry, audit port.AuditWriter) *Server {
	s := &Server{
		mcpServer: mcp.NewServer(&mcp.Implementation{Name: "polyglot", Version: "v1.0.0"}, nil),
		registry:  reg,
		audit:     audit,
	}
	s.registerTools()
	return s
}

func (s *Server) registerTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_device_status",
		Description: "Get the current status of a network device. Read-only — no approval needed.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, s.getDeviceStatus)

	destructive := true
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "run_command",
		Description: "Execute a raw vendor-native command on a device. Potentially destructive — approval required.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive},
	}, s.runCommand)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "push_config",
		Description: "Push a configuration change to a device. Destructive — approval required.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive},
	}, s.pushConfig)
}

// HTTPHandler returns an http.Handler that serves the MCP streamable-HTTP
// transport. Mount it on any HTTP mux (e.g. Gin router at /mcp). The same
// underlying *mcp.Server serves all sessions — it is stateless across
// requests, so this is safe behind a load balancer once the stateless spec
// revision lands (see TECH-STACK-DAN-PERSIAPAN.md §5, 28 Juli 2026).
func (s *Server) HTTPHandler() http.Handler {
	return mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return s.mcpServer
	}, nil)
}

// RunStdio runs the MCP server on the stdio transport (stdin/stdout). Blocks
// until ctx is cancelled or the transport closes. Used when LibreChat (or
// any MCP client) spawns the Go binary as a subprocess.
func (s *Server) RunStdio(ctx context.Context) error {
	return s.mcpServer.Run(ctx, &mcp.StdioTransport{})
}
