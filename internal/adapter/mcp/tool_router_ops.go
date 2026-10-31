package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/quixiq/polyglot/internal/port"
)

type deviceOnlyArgs struct {
	DeviceID string `json:"device_id,omitempty" jsonschema:"target device id (defaults to the operator's active router)"`
}

type routerCapabilitiesOutput struct {
	DeviceID         string `json:"device_id"`
	Version          string `json:"routeros_version"`
	AIMode           string `json:"ai_mode"`
	InspectSupported bool   `json:"inspect_supported"`
	Warning          string `json:"warning,omitempty"`
}

func (s *Server) opsFor(ctx context.Context, req *mcp.CallToolRequest, actorID uint, explicit string) (string, port.RouterOps, error) {
	deviceID, err := s.resolveDevice(ctx, req, actorID, explicit)
	if err != nil {
		return "", nil, err
	}
	if s.opsResolver == nil {
		return deviceID, nil, fmt.Errorf("router ops not configured")
	}
	ops, err := s.opsResolver(ctx, deviceID)
	if err != nil {
		return deviceID, nil, err
	}
	return deviceID, ops, nil
}

func (s *Server) getActiveDeviceTool(ctx context.Context, req *mcp.CallToolRequest, args deviceOnlyArgs) (*mcp.CallToolResult, map[string]any, error) {
	actorID, _, ok := s.actorFrom(ctx, req)
	if !ok {
		return toolError(map[string]any{"error": "unauthenticated"})
	}
	if s.activeDevices == nil {
		return toolError(map[string]any{"error": "active device resolver not configured"})
	}
	deviceID, err := s.resolveDevice(ctx, req, actorID, args.DeviceID)
	if err != nil {
		return toolError(map[string]any{"error": err.Error()})
	}
	return toolOK(map[string]any{"device_id": deviceID})
}

func (s *Server) routerCapabilitiesTool(ctx context.Context, req *mcp.CallToolRequest, args deviceOnlyArgs) (*mcp.CallToolResult, routerCapabilitiesOutput, error) {
	actorID, _, ok := s.actorFrom(ctx, req)
	if !ok {
		return toolError(routerCapabilitiesOutput{})
	}
	deviceID, ops, err := s.opsFor(ctx, req, actorID, args.DeviceID)
	if err != nil {
		return toolError(routerCapabilitiesOutput{DeviceID: deviceID})
	}
	caps, err := ops.Capabilities(ctx)
	if err != nil {
		return toolError(routerCapabilitiesOutput{DeviceID: deviceID})
	}
	return toolOK(routerCapabilitiesOutput{
		DeviceID: deviceID, Version: caps.Version, AIMode: caps.AIMode,
		InspectSupported: caps.InspectSupported, Warning: caps.Warning,
	})
}

type routerReadArgs struct {
	DeviceID string            `json:"device_id,omitempty" jsonschema:"target device id (defaults to the operator's active router)"`
	Menu     string            `json:"menu" jsonschema:"RouterOS menu path without leading slash, e.g. ip/firewall/filter"`
	Selector map[string]string `json:"selector,omitempty" jsonschema:"optional unique selector, e.g. {\".id\":\"*A\"}"`
	Limit    int               `json:"limit,omitempty" jsonschema:"max rows (default 100, max 500)"`
}

type routerReadOutput struct {
	DeviceID string              `json:"device_id"`
	Menu     string              `json:"menu"`
	Rows     []map[string]string `json:"rows"`
	RowCount int                 `json:"row_count"`
}

func (s *Server) routerReadTool(ctx context.Context, req *mcp.CallToolRequest, args routerReadArgs) (*mcp.CallToolResult, routerReadOutput, error) {
	actorID, _, ok := s.actorFrom(ctx, req)
	if !ok {
		return toolError(routerReadOutput{})
	}
	deviceID, ops, err := s.opsFor(ctx, req, actorID, args.DeviceID)
	if err != nil {
		return toolError(routerReadOutput{DeviceID: deviceID, Menu: args.Menu})
	}
	rows, err := ops.ReadMenu(ctx, args.Menu, args.Selector)
	if err != nil {
		return toolError(routerReadOutput{DeviceID: deviceID, Menu: args.Menu})
	}
	limit := args.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	if len(rows) > limit {
		rows = rows[:limit]
	}
	return toolOK(routerReadOutput{DeviceID: deviceID, Menu: args.Menu, Rows: rows, RowCount: len(rows)})
}

type inspectPathArgs struct {
	DeviceID string `json:"device_id,omitempty" jsonschema:"target device id"`
	Path     string `json:"path" jsonschema:"RouterOS path without leading slash, e.g. ip,address,add"`
}

type inspectRowsOutput struct {
	DeviceID string              `json:"device_id"`
	Rows     []map[string]string `json:"rows"`
	RowCount int                 `json:"row_count"`
}

func (s *Server) inspectChildrenTool(ctx context.Context, req *mcp.CallToolRequest, args inspectPathArgs) (*mcp.CallToolResult, inspectRowsOutput, error) {
	return s.inspectRows(ctx, req, args, func(ins port.RouterInspector) ([]map[string]string, error) {
		return ins.InspectChildren(ctx, args.Path)
	})
}

func (s *Server) inspectSyntaxTool(ctx context.Context, req *mcp.CallToolRequest, args inspectPathArgs) (*mcp.CallToolResult, inspectRowsOutput, error) {
	return s.inspectRows(ctx, req, args, func(ins port.RouterInspector) ([]map[string]string, error) {
		return ins.InspectSyntax(ctx, args.Path)
	})
}

type inspectCompletionArgs struct {
	DeviceID string `json:"device_id,omitempty" jsonschema:"target device id"`
	Input    string `json:"input" jsonschema:"partial command input, e.g. /ip/addr/set "`
	Path     string `json:"path,omitempty" jsonschema:"optional path context"`
}

func (s *Server) inspectCompletionTool(ctx context.Context, req *mcp.CallToolRequest, args inspectCompletionArgs) (*mcp.CallToolResult, inspectRowsOutput, error) {
	actorID, _, ok := s.actorFrom(ctx, req)
	if !ok {
		return toolError(inspectRowsOutput{})
	}
	deviceID, ops, err := s.opsFor(ctx, req, actorID, args.DeviceID)
	if err != nil {
		return toolError(inspectRowsOutput{DeviceID: deviceID})
	}
	ins, ok := ops.(port.RouterInspector)
	if !ok {
		return toolError(inspectRowsOutput{DeviceID: deviceID})
	}
	rows, err := ins.InspectCompletion(ctx, args.Input, args.Path)
	if err != nil {
		return toolError(inspectRowsOutput{DeviceID: deviceID})
	}
	return toolOK(inspectRowsOutput{DeviceID: deviceID, Rows: rows, RowCount: len(rows)})
}

type inspectValidateArgs struct {
	DeviceID string `json:"device_id,omitempty" jsonschema:"target device id"`
	Script   string `json:"script" jsonschema:"command/script to validate"`
}

type inspectValidateOutput struct {
	DeviceID  string   `json:"device_id"`
	Supported bool     `json:"supported"`
	Valid     bool     `json:"valid"`
	Errors    []string `json:"errors,omitempty"`
	Warnings  []string `json:"warnings,omitempty"`
}

func (s *Server) inspectValidateTool(ctx context.Context, req *mcp.CallToolRequest, args inspectValidateArgs) (*mcp.CallToolResult, inspectValidateOutput, error) {
	actorID, _, ok := s.actorFrom(ctx, req)
	if !ok {
		return toolError(inspectValidateOutput{})
	}
	deviceID, ops, err := s.opsFor(ctx, req, actorID, args.DeviceID)
	if err != nil {
		return toolError(inspectValidateOutput{DeviceID: deviceID})
	}
	ins, ok := ops.(port.RouterInspector)
	if !ok {
		return toolError(inspectValidateOutput{DeviceID: deviceID})
	}
	res, err := ins.InspectValidate(ctx, args.Script)
	if err != nil {
		return toolError(inspectValidateOutput{DeviceID: deviceID})
	}
	return toolOK(inspectValidateOutput{DeviceID: deviceID, Supported: res.Supported, Valid: res.Valid, Errors: res.Errors, Warnings: res.Warnings})
}

func (s *Server) inspectRows(ctx context.Context, req *mcp.CallToolRequest, args inspectPathArgs, call func(port.RouterInspector) ([]map[string]string, error)) (*mcp.CallToolResult, inspectRowsOutput, error) {
	actorID, _, ok := s.actorFrom(ctx, req)
	if !ok {
		return toolError(inspectRowsOutput{})
	}
	deviceID, ops, err := s.opsFor(ctx, req, actorID, args.DeviceID)
	if err != nil {
		return toolError(inspectRowsOutput{DeviceID: deviceID})
	}
	ins, ok := ops.(port.RouterInspector)
	if !ok {
		return toolError(inspectRowsOutput{DeviceID: deviceID})
	}
	rows, err := call(ins)
	if err != nil {
		return toolError(inspectRowsOutput{DeviceID: deviceID})
	}
	return toolOK(inspectRowsOutput{DeviceID: deviceID, Rows: rows, RowCount: len(rows)})
}
