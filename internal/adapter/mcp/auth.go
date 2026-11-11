package mcp

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/quixiq/polyglot/pkg/logger"
)

// MCP permission scopes. The harness receives only mcp:read so a compromised
// or misconfigured NetOps/CS agent cannot mutate network devices through MCP.
const (
	ScopeMCPRead  = "mcp:read"
	ScopeMCPAdmin = "mcp:admin"
)

const mcpTokenLifetime = time.Hour

// toolScopes maps each registered MCP tool to the scope required to call it.
// Unknown tools fail closed and require mcp:admin.
var toolScopes = map[string]string{
	"customer_lookup":           ScopeMCPRead,
	"get_device_status":         ScopeMCPRead,
	"list_devices":              ScopeMCPRead,
	"mikhmon_get_dashboard":     ScopeMCPRead,
	"routeros_list_menu":        ScopeMCPRead,
	"routeros_describe_command": ScopeMCPRead,
	"routeros_suggest":          ScopeMCPRead,
	"routeros_validate":         ScopeMCPRead,
	"ont_get_status":            ScopeMCPRead,
	"ont_list":                  ScopeMCPRead,

	"run_command":              ScopeMCPAdmin,
	"push_config":              ScopeMCPAdmin,
	"mikhmon_generate_voucher": ScopeMCPAdmin,
	"mikhmon_kick_session":     ScopeMCPAdmin,
	"routeros_run":             ScopeMCPAdmin,
}

// WithBearerAuth enables bearer-token authentication for the HTTP transport and
// enforces per-tool scopes. adminToken grants read and admin scopes; readToken
// grants read-only access and is intended for the DeepSeek harness.
func (s *Server) WithBearerAuth(adminToken, readToken string) *Server {
	s.adminToken = strings.TrimSpace(adminToken)
	s.readToken = strings.TrimSpace(readToken)
	s.mcpServer.AddReceivingMiddleware(s.toolScopeMiddleware())
	return s
}

// AuthConfigured reports whether at least one bearer token is configured.
func (s *Server) AuthConfigured() bool {
	return s.adminToken != "" || s.readToken != ""
}

func (s *Server) verifyToken(_ context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
	now := time.Now()
	switch {
	case s.adminToken != "" && tokensEqual(token, s.adminToken):
		return &auth.TokenInfo{
			Scopes:     []string{ScopeMCPRead, ScopeMCPAdmin},
			Expiration: now.Add(mcpTokenLifetime),
			UserID:     "mcp-admin",
		}, nil
	case s.readToken != "" && tokensEqual(token, s.readToken):
		return &auth.TokenInfo{
			Scopes:     []string{ScopeMCPRead},
			Expiration: now.Add(mcpTokenLifetime),
			UserID:     "mcp-read",
		}, nil
	default:
		return nil, auth.ErrInvalidToken
	}
}

func tokensEqual(got, want string) bool {
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

func (s *Server) toolScopeMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			switch method {
			case "tools/call":
				return s.authorizeToolCall(ctx, next, req)
			case "tools/list":
				res, err := next(ctx, method, req)
				if err != nil || hasScope(tokenScopes(req), ScopeMCPAdmin) {
					return res, err
				}
				list, ok := res.(*mcp.ListToolsResult)
				if !ok || list == nil {
					return res, err
				}
				filtered := make([]*mcp.Tool, 0, len(list.Tools))
				for _, tool := range list.Tools {
					if tool != nil && requiredToolScope(tool.Name) == ScopeMCPRead {
						filtered = append(filtered, tool)
					}
				}
				clone := *list
				clone.Tools = filtered
				clone.NextCursor = ""
				return &clone, nil
			default:
				return next(ctx, method, req)
			}
		}
	}
}

func (s *Server) authorizeToolCall(ctx context.Context, next mcp.MethodHandler, req mcp.Request) (mcp.Result, error) {
	params, ok := req.GetParams().(*mcp.CallToolParamsRaw)
	if !ok || params == nil {
		return nil, &jsonrpc.Error{Code: jsonrpc.CodeInvalidParams, Message: "missing tool parameters"}
	}

	scope := requiredToolScope(params.Name)
	if hasScope(tokenScopes(req), scope) {
		logger.WithComponent("MCP").
			WithField("tool", params.Name).
			WithField("principal", tokenUserID(req)).
			WithField("scope", scope).
			Debug("mcp tool call allowed")
		return next(ctx, "tools/call", req)
	}

	logger.WithComponent("MCP").
		WithField("tool", params.Name).
		WithField("principal", tokenUserID(req)).
		WithField("scope", scope).
		Warn("mcp tool call denied: insufficient scope")
	return nil, &jsonrpc.Error{
		Code:    jsonrpc.CodeInvalidRequest,
		Message: "insufficient scope for tool " + params.Name,
	}
}

func requiredToolScope(tool string) string {
	if scope, ok := toolScopes[tool]; ok {
		return scope
	}
	return ScopeMCPAdmin
}

func tokenScopes(req mcp.Request) []string {
	extra := req.GetExtra()
	if extra == nil || extra.TokenInfo == nil {
		return nil
	}
	return extra.TokenInfo.Scopes
}

func tokenUserID(req mcp.Request) string {
	extra := req.GetExtra()
	if extra == nil || extra.TokenInfo == nil {
		return ""
	}
	return extra.TokenInfo.UserID
}

func hasScope(scopes []string, want string) bool {
	for _, scope := range scopes {
		if scope == want {
			return true
		}
	}
	return false
}
