package mikrotik

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/quixiq/polyglot/internal/adapter/mcp/mcputil"
)

// RouterOSListMenu handles the "routeros_list_menu" tool.
// It queries /console/inspect request=child to list submenus, commands, or parameters.
func (h *Handler) RouterOSListMenu(ctx context.Context, _ *mcp.CallToolRequest, args routerosListMenuArgs) (*mcp.CallToolResult, routerosListMenuOutput, error) {
	if args.DeviceID == "" {
		return mcputil.ToolError(routerosListMenuOutput{
			Path:    args.Path,
			Summary: "device_id is required",
			Error:   "device_id is required",
		})
	}

	if h.Registry == nil {
		return mcputil.ToolError(routerosListMenuOutput{
			DeviceID: args.DeviceID,
			Path:     args.Path,
			Summary:  "device registry not configured",
			Error:    "device registry not configured",
		})
	}

	driver, err := h.Registry.Get(ctx, args.DeviceID)
	if err != nil {
		return mcputil.ToolError(routerosListMenuOutput{
			DeviceID: args.DeviceID,
			Path:     args.Path,
			Summary:  err.Error(),
			Error:    err.Error(),
		})
	}

	items, cached, err := h.inspector().ListMenu(ctx, driver, args.DeviceID, args.Path)
	if err != nil {
		return mcputil.ToolError(routerosListMenuOutput{
			DeviceID: args.DeviceID,
			Path:     args.Path,
			Summary:  err.Error(),
			Error:    err.Error(),
		})
	}

	normPath := normalizePath(args.Path)
	summary := fmt.Sprintf("Found %d menu items under %q (cached: %t)", len(items), normPath, cached)
	if normPath == "" {
		summary = fmt.Sprintf("Found %d root menu items (cached: %t)", len(items), cached)
	}

	return mcputil.ToolOK(routerosListMenuOutput{
		DeviceID: args.DeviceID,
		Path:     normPath,
		Items:    items,
		Count:    len(items),
		Cached:   cached,
		Summary:  summary,
	})
}

// RouterOSDescribeCommand handles the "routeros_describe_command" tool.
// It inspects command arguments and their syntax types using request=syntax and request=child.
func (h *Handler) RouterOSDescribeCommand(ctx context.Context, _ *mcp.CallToolRequest, args routerosDescribeCommandArgs) (*mcp.CallToolResult, routerosDescribeCommandOutput, error) {
	if args.DeviceID == "" || args.Path == "" {
		return mcputil.ToolError(routerosDescribeCommandOutput{
			Path:    args.Path,
			Summary: "device_id and path are required",
			Error:   "device_id and path are required",
		})
	}

	if h.Registry == nil {
		return mcputil.ToolError(routerosDescribeCommandOutput{
			DeviceID: args.DeviceID,
			Path:     args.Path,
			Summary:  "device registry not configured",
			Error:    "device registry not configured",
		})
	}

	driver, err := h.Registry.Get(ctx, args.DeviceID)
	if err != nil {
		return mcputil.ToolError(routerosDescribeCommandOutput{
			DeviceID: args.DeviceID,
			Path:     args.Path,
			Summary:  err.Error(),
			Error:    err.Error(),
		})
	}

	cmdArgs, cached, err := h.inspector().DescribeCommand(ctx, driver, args.DeviceID, args.Path)
	if err != nil {
		return mcputil.ToolError(routerosDescribeCommandOutput{
			DeviceID: args.DeviceID,
			Path:     args.Path,
			Summary:  err.Error(),
			Error:    err.Error(),
		})
	}

	normPath := normalizePath(args.Path)
	summary := fmt.Sprintf("Found %d arguments for command %q (cached: %t)", len(cmdArgs), normPath, cached)

	return mcputil.ToolOK(routerosDescribeCommandOutput{
		DeviceID:  args.DeviceID,
		Path:      normPath,
		Arguments: cmdArgs,
		Count:     len(cmdArgs),
		Cached:    cached,
		Summary:   summary,
	})
}
