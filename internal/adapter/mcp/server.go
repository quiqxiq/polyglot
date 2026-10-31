package mcp

import (
	"context"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	iauth "github.com/quixiq/polyglot/internal/adapter/auth"
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
	mcpServer     *mcp.Server
	registry      *registry.Registry
	audit         port.AuditWriter
	jwt           *iauth.JWTService
	mikhmonUC     *hotspotUC.UseCase
	customerRepo  port.CustomerRepository
	skillRepo     port.SkillRepository
	proposals     ProposalService
	approvals     port.ApprovalRepository
	decider       ApprovalCanceler
	activeDevices ActiveDeviceResolver
	opsResolver   func(ctx context.Context, deviceID string) (port.RouterOps, error)
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

// WithJWTService enables actor resolution for every tool call. /mcp is
// protected by the HTTP auth middleware; this service only re-derives the
// actor (user ID + roles) from the same Bearer token so tools can audit and
// authorize per-user instead of trusting a client-supplied identity.
func (s *Server) WithJWTService(jwt *iauth.JWTService) *Server {
	s.jwt = jwt
	return s
}

// WithProposalService enables mutation proposals (run_command/push_config).
func (s *Server) WithProposalService(p ProposalService) *Server {
	s.proposals = p
	return s
}

// WithApprovalRepository enables get/list/cancel approval tools.
func (s *Server) WithApprovalRepository(r port.ApprovalRepository) *Server {
	s.approvals = r
	return s
}

// WithApprovalCanceler enables cancel_approval to change approval state.
func (s *Server) WithApprovalCanceler(c ApprovalCanceler) *Server {
	s.decider = c
	return s
}

// WithRouterOps enables read/inspect/capability tools (M3).
func (s *Server) WithRouterOps(resolver func(ctx context.Context, deviceID string) (port.RouterOps, error)) *Server {
	s.opsResolver = resolver
	return s
}

// WithActiveDeviceResolver enables tools to target the operator's active
// router when device_id is omitted.
func (s *Server) WithActiveDeviceResolver(r ActiveDeviceResolver) *Server {
	s.activeDevices = r
	return s
}

// actorFrom resolves the authenticated actor for a tool call. It prefers the
// identity the HTTP middleware put in the request context and falls back to
// validating the Authorization header attached to the MCP request (stateless
// streamable HTTP). It NEVER trusts a user ID supplied in tool arguments.
func (s *Server) actorFrom(ctx context.Context, req *mcp.CallToolRequest) (uint, []string, bool) {
	if userID, roles, ok := iauth.IdentityFromContext(ctx); ok && userID > 0 {
		return userID, roles, true
	}
	if s.jwt == nil || req == nil || req.Extra == nil || req.Extra.Header == nil {
		return 0, nil, false
	}
	authz := strings.TrimSpace(req.Extra.Header.Get("Authorization"))
	if authz == "" {
		return 0, nil, false
	}
	parts := strings.SplitN(authz, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return 0, nil, false
	}
	claims, err := s.jwt.ValidateToken(strings.TrimSpace(parts[1]))
	if err != nil || claims.UserID == 0 {
		return 0, nil, false
	}
	roles := claims.Roles
	if len(roles) == 0 && claims.Role != "" {
		roles = []string{claims.Role}
	}
	return claims.UserID, roles, true
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
		Description: "Propose a configuration change for human approval. This tool NEVER applies config; it creates an approval request with a restore point.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: &destructive},
	}, s.pushConfig)

	// AI Router Ops (PLAN-AI-ROUTER-OPS v2). These tools create approval
	// requests only; they never execute a mutation.
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "propose_command",
		Description: "Propose ONE router mutation for human approval. Never executes; returns approval_id + diff + restore point.",
	}, s.proposeCommand)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "propose_config",
		Description: "Propose a multi-line router configuration change for human approval. Never executes.",
	}, s.proposeConfig)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_active_device",
		Description: "Return the operator's active router binding. Read-only.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, s.getActiveDeviceTool)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "router_capabilities",
		Description: "Return RouterOS version, /console/inspect support, and AI mode (full|degraded). Read-only.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, s.routerCapabilitiesTool)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "router_read",
		Description: "Read rows from a router menu (print detail) with an optional unique selector. Read-only.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, s.routerReadTool)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "inspect_children",
		Description: "List child menu nodes via /console/inspect request=child (RouterOS v7). Read-only.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, s.inspectChildrenTool)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "inspect_syntax",
		Description: "Inspect syntax/arguments for a path via /console/inspect request=syntax (RouterOS v7). Read-only.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, s.inspectSyntaxTool)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "inspect_completion",
		Description: "Get completion suggestions via /console/inspect request=completion (RouterOS v7). Read-only.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, s.inspectCompletionTool)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "inspect_validate",
		Description: "Validate a command/script via /console/inspect highlight (RouterOS v7; v6 returns degraded). Read-only.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, s.inspectValidateTool)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_approval",
		Description: "Get an approval request status and result. Read-only.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, s.getApproval)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_approvals",
		Description: "List recent approval requests. Read-only.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, s.listApprovals)
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "cancel_approval",
		Description: "Cancel your own PENDING approval request.",
	}, s.cancelApproval)

	// Mikhmon & Hotspot Tools
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "mikhmon_get_dashboard",
		Description: "Get Mikrotik Mikhmon hotspot dashboard metrics (CPU, Memory, Uptime, Active users, Today income). Read-only.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, s.mikhmonGetDashboard)

	// NOTE (PLAN-AI-ROUTER-OPS v2, K13): mikhmon_generate_voucher and
	// mikhmon_kick_session are deliberately NOT exposed to AI v1. They
	// mutate router state through a usecase rather than a raw command, so
	// the v1 approval/diff/restore model cannot represent them safely yet.
	// They remain available only through the authenticated web UI.

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
