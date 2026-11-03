package mikrotik

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/quixiq/polyglot/internal/adapter/mcp/mcputil"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/hotspot"
)

type mikhmonGenerateVoucherArgs struct {
	DeviceID string `json:"device_id" jsonschema:"the Mikrotik device ID from inventory"`
	Profile  string `json:"profile" jsonschema:"the Hotspot User Profile name (e.g. 2Jam 5K)"`
	Count    int    `json:"count" jsonschema:"number of vouchers to generate (1-100)"`
	Validity string `json:"validity,omitempty" jsonschema:"optional validity string (e.g. 1d or 2h)"`
}

type mikhmonGenerateVoucherOutput struct {
	DeviceID string   `json:"device_id"`
	Status   string   `json:"status"`
	Summary  string   `json:"summary"`
	Count    int      `json:"count"`
	Vouchers []string `json:"vouchers,omitempty"`
}

// MikhmonGenerateVoucher handles the "mikhmon_generate_voucher" tool.
func (h *Handler) MikhmonGenerateVoucher(ctx context.Context, _ *mcp.CallToolRequest, args mikhmonGenerateVoucherArgs) (*mcp.CallToolResult, mikhmonGenerateVoucherOutput, error) {
	if args.DeviceID == "" || args.Profile == "" || args.Count <= 0 {
		return mcputil.ToolError(mikhmonGenerateVoucherOutput{Status: "error", Summary: "device_id, profile, and positive count are required"})
	}

	if h.Registry == nil {
		return mcputil.ToolError(mikhmonGenerateVoucherOutput{DeviceID: args.DeviceID, Status: "error", Summary: "device registry not configured"})
	}

	driver, err := h.Registry.Get(ctx, args.DeviceID)
	if err != nil {
		return mcputil.ToolError(mikhmonGenerateVoucherOutput{DeviceID: args.DeviceID, Status: "error", Summary: err.Error()})
	}

	uc := h.uc()
	if uc == nil {
		return mcputil.ToolError(mikhmonGenerateVoucherOutput{DeviceID: args.DeviceID, Status: "error", Summary: "mikhmon use case not configured"})
	}

	params := hotspot.VoucherGenerateParams{
		Profile:     args.Profile,
		LimitUptime: args.Validity,
	}

	batch, err := uc.GenerateVouchers(ctx, driver, params, args.Count)
	if err != nil {
		return mcputil.ToolError(mikhmonGenerateVoucherOutput{DeviceID: args.DeviceID, Status: "error", Summary: err.Error()})
	}

	codes := make([]string, len(batch.Vouchers))
	for i, v := range batch.Vouchers {
		codes[i] = fmt.Sprintf("%s / %s", v.Username, v.Password)
	}

	summary := fmt.Sprintf("Successfully generated %d vouchers for profile %q on device %s", len(batch.Vouchers), args.Profile, args.DeviceID)

	return mcputil.ToolOK(mikhmonGenerateVoucherOutput{
		DeviceID: args.DeviceID,
		Status:   "success",
		Summary:  summary,
		Count:    len(batch.Vouchers),
		Vouchers: codes,
	})
}
