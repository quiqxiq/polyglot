package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/mikhmon"
	"github.com/quixiq/polyglot/internal/usecase/network"
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

func (s *Server) mikhmonGenerateVoucher(ctx context.Context, _ *mcp.CallToolRequest, args mikhmonGenerateVoucherArgs) (*mcp.CallToolResult, mikhmonGenerateVoucherOutput, error) {
	if args.DeviceID == "" || args.Profile == "" || args.Count <= 0 {
		return toolError(mikhmonGenerateVoucherOutput{Status: "error", Summary: "device_id, profile, and positive count are required"})
	}

	driver, err := s.registry.Get(ctx, args.DeviceID)
	if err != nil {
		return toolError(mikhmonGenerateVoucherOutput{DeviceID: args.DeviceID, Status: "error", Summary: err.Error()})
	}

	uc := s.mikhmonUC
	if uc == nil {
		uc = network.NewMikhmonUseCase("")
	}

	params := mikhmon.VoucherGenerateParams{
		Profile:     args.Profile,
		LimitUptime: args.Validity,
	}

	batch, err := uc.GenerateVouchers(ctx, driver, params, args.Count)
	if err != nil {
		return toolError(mikhmonGenerateVoucherOutput{DeviceID: args.DeviceID, Status: "error", Summary: err.Error()})
	}

	codes := make([]string, len(batch.Vouchers))
	for i, v := range batch.Vouchers {
		codes[i] = fmt.Sprintf("%s / %s", v.Username, v.Password)
	}

	summary := fmt.Sprintf("Successfully generated %d vouchers for profile %q on device %s", len(batch.Vouchers), args.Profile, args.DeviceID)

	return toolOK(mikhmonGenerateVoucherOutput{
		DeviceID: args.DeviceID,
		Status:   "success",
		Summary:  summary,
		Count:    len(batch.Vouchers),
		Vouchers: codes,
	})
}
