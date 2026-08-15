package network

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/mikhmon"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/voucher"
)

// GenerateVouchers generates a batch of count vouchers and executes their creation commands.
func (u *MikhmonUseCase) GenerateVouchers(ctx context.Context, driver port.DeviceDriver, p mikhmon.VoucherGenerateParams, count int) (mikhmon.VoucherBatch, error) {
	batch := mikhmon.NewGenerateVoucherBatchCommands(p, count)
	for _, cmd := range batch.Commands {
		if _, err := ExecuteCommand(ctx, driver, cmd); err != nil {
			return batch, fmt.Errorf("failed to create voucher: %w", err)
		}
	}
	return batch, nil
}

// GetReports fetches sales transaction logs recorded by Mikhmon scripts.
func (u *MikhmonUseCase) GetReports(ctx context.Context, driver port.DeviceDriver) ([]mikhmon.MikhmonTransaction, error) {
	cmd := mikhmon.NewPrintMikhmonReportsCommand()
	res, err := ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return mikhmon.ParseMikhmonTransactions(res), nil
}

// DeleteReport deletes a transaction log script by RouterOS .id.
func (u *MikhmonUseCase) DeleteReport(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := command.Command{Raw: "/system/script/remove", Args: map[string]string{".id": rosID}}
	return ExecuteCommand(ctx, driver, cmd)
}

// GetTodayIncome calculates total sales revenue recorded today by Mikhmon scripts.
func (u *MikhmonUseCase) GetTodayIncome(ctx context.Context, driver port.DeviceDriver) (int64, error) {
	todayStr := time.Now().Format("02.01.06")
	reports, err := u.GetReports(ctx, driver)
	if err != nil {
		return 0, err
	}
	var todayIncome int64
	for _, r := range reports {
		if r.Date == todayStr {
			if val, e := strconv.ParseInt(r.Price, 10, 64); e == nil {
				todayIncome += val
			}
		}
	}
	return todayIncome, nil
}

// GetReportsByFilter fetches sales transaction logs filtered by optional date, month, or year.
func (u *MikhmonUseCase) GetReportsByFilter(ctx context.Context, driver port.DeviceDriver, date, month, year string) ([]mikhmon.MikhmonTransaction, error) {
	all, err := u.GetReports(ctx, driver)
	if err != nil {
		return nil, err
	}
	if date == "" && month == "" && year == "" {
		return all, nil
	}
	filtered := make([]mikhmon.MikhmonTransaction, 0)
	for _, r := range all {
		if date != "" && !strings.HasPrefix(r.Date, date) {
			continue
		}
		if month != "" && !strings.Contains(strings.ToLower(r.Date), strings.ToLower(month)) {
			continue
		}
		if year != "" && !strings.HasSuffix(r.Date, year) {
			continue
		}
		filtered = append(filtered, r)
	}
	return filtered, nil
}

// RenderVoucherHTML converts a generated VoucherBatch into a printable HTML page.
func (u *MikhmonUseCase) RenderVoucherHTML(batch mikhmon.VoucherBatch, layout, hotspotName, dnsName, logo string) (string, error) {
	vouchers := make([]voucher.VoucherData, 0, len(batch.Vouchers))
	for i, v := range batch.Vouchers {
		cardValidity := ""
		if parsed, err := mikhmon.ParseMikhmonComment(v.Comment); err == nil {
			cardValidity = parsed.Tag
		}
		vouchers = append(vouchers, voucher.VoucherData{
			Username:    v.Username,
			Password:    v.Password,
			Comment:     v.Comment,
			Validity:    cardValidity,
			HotspotName: hotspotName,
			DNSName:     dnsName,
			Logo:        logo,
			Number:      i + 1,
		})
	}
	return voucher.Render(vouchers, voucher.Layout(layout), u.TemplateDir)
}
