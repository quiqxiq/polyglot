package mcp

import (
	"context"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/quixiq/polyglot/internal/adapter/mcp/customer"
	"github.com/quixiq/polyglot/internal/adapter/mcp/device"
	"github.com/quixiq/polyglot/internal/adapter/mcp/mikrotik"
	"github.com/quixiq/polyglot/internal/adapter/mcp/olt"
	ontmcp "github.com/quixiq/polyglot/internal/adapter/mcp/ont"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/registry"
	hotspotUC "github.com/quixiq/polyglot/internal/usecase/hotspot"
	oltUC "github.com/quixiq/polyglot/internal/usecase/olt"
)

// Server wraps the MCP SDK server and holds the dependencies every tool
// handler needs: the device driver registry, audit writer, Mikhmon use case,
// customer repository, and skill repository.
type Server struct {
	mcpServer    *mcp.Server
	registry     *registry.Registry
	audit        port.AuditWriter
	mikhmonUC    *hotspotUC.UseCase
	oltSvc       *oltUC.Service
	ontRepo      port.ONTRepository
	customerRepo port.CustomerRepository
	deviceRepo   port.DeviceRepository
	adminToken   string
	readToken    string
}

// New builds an MCP Server with registered tools.
func New(reg *registry.Registry, audit port.AuditWriter) *Server {
	s := &Server{
		mcpServer: mcp.NewServer(&mcp.Implementation{Name: "polyglot", Version: "v1.0.0"}, nil),
		registry:  reg,
		audit:     audit,
	}
	if reg != nil {
		s.deviceRepo = reg.DeviceRepo()
	}
	s.registerTools()
	return s
}

// WithDeviceRepository sets the DeviceRepository dependency.
func (s *Server) WithDeviceRepository(repo port.DeviceRepository) *Server {
	s.deviceRepo = repo
	return s
}

// WithMikhmonUseCase sets the MikhmonUseCase dependency.
func (s *Server) WithMikhmonUseCase(uc *hotspotUC.UseCase) *Server {
	s.mikhmonUC = uc
	return s
}

// WithOLTService sets the OLT Service dependency.
func (s *Server) WithOLTService(svc *oltUC.Service) *Server {
	s.oltSvc = svc
	return s
}

// WithCustomerRepository sets the CustomerRepository dependency.
func (s *Server) WithCustomerRepository(repo port.CustomerRepository) *Server {
	s.customerRepo = repo
	return s
}

// WithONTRepository sets the ONTRepository dependency.
func (s *Server) WithONTRepository(repo port.ONTRepository) *Server {
	s.ontRepo = repo
	return s
}

func (s *Server) registerTools() {
	deviceH := device.NewHandler(s.registry, func() port.DeviceRepository { return s.deviceRepo }, s.audit)
	device.Register(s.mcpServer, deviceH)

	mikrotikH := mikrotik.NewHandler(s.registry, func() *hotspotUC.UseCase { return s.mikhmonUC }, mikrotik.WithAudit(s.audit))
	mikrotik.Register(s.mcpServer, mikrotikH)

	oltH := olt.NewHandler(s.registry, func() *oltUC.Service { return s.oltSvc })
	olt.Register(s.mcpServer, oltH)

	customerH := customer.NewHandler(func() port.CustomerRepository { return s.customerRepo })
	customer.Register(s.mcpServer, customerH)

	ontH := ontmcp.NewHandler(func() port.ONTRepository { return s.ontRepo })
	ontmcp.Register(s.mcpServer, ontH)
}

// HTTPHandler returns an http.Handler that serves the MCP streamable-HTTP
// transport. Mount it on any HTTP mux (e.g. at /mcp). The same
// underlying *mcp.Server serves all sessions.
//
// When no bearer token is configured the endpoint fails closed with 503: MCP
// exposes device commands and must never be anonymous.
func (s *Server) HTTPHandler() http.Handler {
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return s.mcpServer
	}, nil)
	if !s.AuthConfigured() {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "MCP authentication is not configured", http.StatusServiceUnavailable)
		})
	}
	return auth.RequireBearerToken(s.verifyToken, nil)(handler)
}

// RunStdio runs the MCP server on the stdio transport (stdin/stdout). Blocks
// until ctx is cancelled or the transport closes. Used when LibreChat (or
// any MCP client) spawns the Go binary as a subprocess.
func (s *Server) RunStdio(ctx context.Context) error {
	return s.mcpServer.Run(ctx, &mcp.StdioTransport{})
}
