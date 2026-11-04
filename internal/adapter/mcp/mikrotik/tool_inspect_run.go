package mikrotik

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/quixiq/polyglot/internal/adapter/mcp/mcputil"
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/usecase/network"
)

// RouterOSRun handles the "routeros_run" tool.
// It executes a RouterOS command with optional pre-flight syntax validation via /console/inspect.
func (h *Handler) RouterOSRun(ctx context.Context, _ *mcp.CallToolRequest, args routerosRunArgs) (*mcp.CallToolResult, routerosRunOutput, error) {
	if args.DeviceID == "" || strings.TrimSpace(args.Command) == "" {
		return mcputil.ToolError(routerosRunOutput{
			DeviceID: args.DeviceID,
			Command:  args.Command,
			Error:    "device_id and command are required",
		})
	}

	if h.Registry == nil {
		return mcputil.ToolError(routerosRunOutput{
			DeviceID: args.DeviceID,
			Command:  args.Command,
			Error:    "device registry not configured",
		})
	}

	driver, err := h.Registry.Get(ctx, args.DeviceID)
	if err != nil {
		return mcputil.ToolError(routerosRunOutput{
			DeviceID: args.DeviceID,
			Command:  args.Command,
			Error:    err.Error(),
		})
	}

	var raw string
	var cmdArgs map[string]string
	var consoleLine string

	if len(args.Args) == 0 && strings.Contains(strings.TrimSpace(args.Command), " ") {
		raw, cmdArgs = parseConsoleCommand(args.Command)
		consoleLine = normalizeConsoleInput(args.Command)
	} else {
		raw = args.Command
		cmdArgs = args.Args
		consoleLine = buildConsoleLine(raw, cmdArgs)
	}

	autoValidate := true
	if args.AutoValidate != nil {
		autoValidate = *args.AutoValidate
	}

	validated := false
	if autoValidate {
		valRes, valErr := h.inspector().Validate(ctx, driver, consoleLine)
		if valErr != nil {
			return mcputil.ToolError(routerosRunOutput{
				DeviceID: args.DeviceID,
				Command:  args.Command,
				Error:    fmt.Sprintf("pre-flight validation error: %s", valErr.Error()),
			})
		}
		if !valRes.Valid {
			errMsg := fmt.Sprintf("validation failed: %s (position %d)", valRes.ErrorMessage, valRes.ErrorPosition)
			remediation := formatRouterOSRemediation(args.Command, errMsg, valRes.ErrorPosition)
			return mcputil.ToolError(routerosRunOutput{
				DeviceID:        args.DeviceID,
				Command:         args.Command,
				Validated:       false,
				Executed:        false,
				ValidationError: valRes.ErrorMessage,
				Remediation:     remediation,
				Error:           remediation,
			})
		}
		validated = true
	}

	cmd := command.Command{
		Raw:  raw,
		Args: cmdArgs,
	}

	result, err := network.ExecuteCommandPreApproved(ctx, driver, cmd)
	if err != nil {
		remediation := formatRouterOSRemediation(args.Command, err.Error(), -1)
		return mcputil.ToolError(routerosRunOutput{
			DeviceID:    args.DeviceID,
			Command:     args.Command,
			Validated:   validated,
			Executed:    false,
			Remediation: remediation,
			Error:       remediation,
		})
	}

	mcputil.AuditBestEffort(ctx, h.Audit, args.DeviceID, cmd.Raw, result)

	out := routerosRunOutput{
		DeviceID:  args.DeviceID,
		Command:   args.Command,
		Validated: validated,
		Executed:  true,
		Output:    result.Output,
		Rows:      result.Rows,
		RowCount:  len(result.Rows),
	}
	if out.Output == "" && out.RowCount == 0 {
		out.Output = "ok"
	}

	return mcputil.ToolOK(out)
}

// formatRouterOSRemediation formats an error with actionable guidance directing the model to inspect/suggest tools.
func formatRouterOSRemediation(commandStr, errMsg string, pos int) string {
	cleanCmd := strings.TrimSpace(commandStr)
	suggestPrefix := cleanCmd
	parts := strings.Fields(cleanCmd)
	if len(parts) > 1 {
		suggestPrefix = strings.Join(parts[:len(parts)-1], " ") + " "
	}

	posStr := ""
	if pos >= 0 {
		posStr = fmt.Sprintf(" [at position %d]", pos)
	}

	return fmt.Sprintf("%s%s\n"+
		"STOP: Do NOT retry 'routeros_run' or 'run_command' blindly.\n"+
		"REQUIRED REMEDIATION STEPS:\n"+
		"1. Call 'routeros_suggest' with device_id and input=%q to inspect real valid commands, submenus, or arguments on the router.\n"+
		"2. Call 'routeros_describe_command' or 'routeros_list_menu' if you need the full parameter schema.\n"+
		"3. Call 'routeros_validate' with the corrected command to dry-run syntax before executing.",
		errMsg, posStr, suggestPrefix)
}
