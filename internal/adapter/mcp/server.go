package mcp

import (
	"context"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/registry"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
)

// Server wraps the MCP SDK server and holds the dependencies every tool
// handler needs: the device driver registry, audit writer, Mikhmon use case,
// Server wraps the MCP SDK server and holds the dependencies every tool
// handler needs: the device driver registry, audit writer, Mikhmon use case,
// customer repository, and skill repository.
type Server struct {
	mcpServer    *mcp.Server
	registry     *registry.Registry
	audit        port.AuditWriter
	mikhmonUC    *hotspotUC.UseCase
	customerRepo port.CustomerRepository
	skillRepo    port.SkillRepository
}

// New builds an MCP Server with registered tools.
func New(reg *registry.Registry, audit port.AuditWriter) *Server {
	s := &Server{
		mcpServer: mcp.NewServer(&mcp.Implementation{Name: "polyglot", Version: "v1.0.0"}, nil),
		registry:  reg,
		audit:     audit,
	}
	s.registerTools()
	return s
}

// WithMikhmonUseCase sets the MikhmonUseCase dependency.
func (s *Server) WithMikhmonUseCase(uc *hotspotUC.UseCase) *Server {
	s.mikhmonUC = uc
	return s
}

// WithCustomerRepository sets the CustomerRepository dependency.
func (s *Server) WithCustomerRepository(repo port.CustomerRepository) *Server {
	s.customerRepo = repo
	return s
}

// WithSkillRepository sets the SkillRepository dependency.
func (s *Server) WithSkillRepository(sr port.SkillRepository) *Server {
	s.skillRepo = sr
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

	// Mikhmon & Hotspot Tools
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mikhmon_get_dashboard",
		Description: "Get Mikrotik Mikhmon hotspot dashboard metrics (CPU, Memory, Uptime, Active users, Today income). Read-only.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, s.mikhmonGetDashboard)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mikhmon_generate_voucher",
		Description: "Generate a batch of hotspot user vouchers for a given profile. Destructive — approval required.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive},
	}, s.mikhmonGenerateVoucher)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mikhmon_kick_session",
		Description: "Disconnect an active hotspot user session by username, IP, or session ID. Destructive — approval required.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive},
	}, s.mikhmonKickSession)

	// Customer & Support Tools
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "customer_lookup",
		Description: "Lookup ISP customer profile and active subscriptions by phone number, name, or customer ID. Read-only.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, s.customerLookup)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_skills",
		Description: "List active modular skills and standard operational procedures (SOP). Read-only.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, s.listSkills)
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
