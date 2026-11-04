package mikrotik

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/quixiq/polyglot/internal/adapter/mcp/mcputil"
)

// RouterOSSuggest handles the "routeros_suggest" tool.
// It queries /console/inspect request=completion to fetch dynamic candidate values or arguments.
func (h *Handler) RouterOSSuggest(ctx context.Context, _ *mcp.CallToolRequest, args routerosSuggestArgs) (*mcp.CallToolResult, routerosSuggestOutput, error) {
	if args.DeviceID == "" || args.Input == "" {
		return mcputil.ToolError(routerosSuggestOutput{
			Input:   args.Input,
			Summary: "device_id and input are required",
			Error:   "device_id and input are required",
		})
	}

	if h.Registry == nil {
		return mcputil.ToolError(routerosSuggestOutput{
			DeviceID: args.DeviceID,
			Input:    args.Input,
			Summary:  "device registry not configured",
			Error:    "device registry not configured",
		})
	}

	driver, err := h.Registry.Get(ctx, args.DeviceID)
	if err != nil {
		return mcputil.ToolError(routerosSuggestOutput{
			DeviceID: args.DeviceID,
			Input:    args.Input,
			Summary:  err.Error(),
			Error:    err.Error(),
		})
	}

	suggestions, err := h.inspector().Suggest(ctx, driver, args.Input)
	if err != nil {
		return mcputil.ToolError(routerosSuggestOutput{
			DeviceID: args.DeviceID,
			Input:    args.Input,
			Summary:  err.Error(),
			Error:    err.Error(),
		})
	}

	summary := fmt.Sprintf("Found %d suggestion candidates for %q", len(suggestions), args.Input)

	return mcputil.ToolOK(routerosSuggestOutput{
		DeviceID:    args.DeviceID,
		Input:       args.Input,
		Suggestions: suggestions,
		Count:       len(suggestions),
		Summary:     summary,
	})
}

// RouterOSValidate handles the "routeros_validate" tool.
// It dry-runs syntax validation without executing using /console/inspect request=error.
func (h *Handler) RouterOSValidate(ctx context.Context, _ *mcp.CallToolRequest, args routerosValidateArgs) (*mcp.CallToolResult, routerosValidateOutput, error) {
	if args.DeviceID == "" || args.Input == "" {
		return mcputil.ToolError(routerosValidateOutput{
			Input:   args.Input,
			Summary: "device_id and input are required",
			Error:   "device_id and input are required",
		})
	}

	if h.Registry == nil {
		return mcputil.ToolError(routerosValidateOutput{
			DeviceID: args.DeviceID,
			Input:    args.Input,
			Summary:  "device registry not configured",
			Error:    "device registry not configured",
		})
	}

	driver, err := h.Registry.Get(ctx, args.DeviceID)
	if err != nil {
		return mcputil.ToolError(routerosValidateOutput{
			DeviceID: args.DeviceID,
			Input:    args.Input,
			Summary:  err.Error(),
			Error:    err.Error(),
		})
	}

	result, err := h.inspector().Validate(ctx, driver, args.Input)
	if err != nil {
		return mcputil.ToolError(routerosValidateOutput{
			DeviceID: args.DeviceID,
			Input:    args.Input,
			Summary:  err.Error(),
			Error:    err.Error(),
		})
	}

	if !result.Valid {
		summary := fmt.Sprintf("Syntax error: %s (position %d)", result.ErrorMessage, result.ErrorPosition)
		remediation := formatRouterOSRemediation(args.Input, summary, result.ErrorPosition)
		return mcputil.ToolOK(routerosValidateOutput{
			DeviceID:      args.DeviceID,
			Input:         args.Input,
			Valid:         false,
			ErrorMessage:  result.ErrorMessage,
			ErrorPosition: result.ErrorPosition,
			Summary:       summary,
			Remediation:   remediation,
		})
	}

	return mcputil.ToolOK(routerosValidateOutput{
		DeviceID: args.DeviceID,
		Input:    args.Input,
		Valid:    true,
		Summary:  "Command syntax is valid",
	})
}
