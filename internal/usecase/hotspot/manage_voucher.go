package hotspot

import (
	"context"
	"fmt"
	"strings"

	domainHotspot "github.com/quixiq/polyglot/internal/domain/hotspot"
	"github.com/quixiq/polyglot/internal/port"
)

// GenerateVouchers generates a batch of count vouchers and executes their creation commands.
func (u *UseCase) GenerateVouchers(ctx context.Context, driver port.DeviceDriver, p port.VoucherGenerateParams, count int) (port.VoucherBatch, error) {
	return u.gateway.GenerateVouchers(ctx, driver, p, count)
}

// CheckVoucherStatus inspects a voucher username and aggregates all relevant status.
func (u *UseCase) CheckVoucherStatus(ctx context.Context, driver port.DeviceDriver, username string) (*port.VoucherStatusDetails, error) {
	if strings.TrimSpace(username) == "" {
		return nil, domainHotspot.ErrInvalidInput
	}

	users, err := u.gateway.ListUsers(ctx, driver, port.ListUsersFilter{Name: username})
	if err != nil {
		return nil, fmt.Errorf("lookup hotspot user: %w", err)
	}
	if len(users) == 0 {
		return &port.VoucherStatusDetails{
			Found:   false,
			Status:  "not_found",
			Message: fmt.Sprintf("Voucher %q not found on router", username),
		}, nil
	}

	user := users[0]
	details := &port.VoucherStatusDetails{
		Found:   true,
		User:    &user,
		Status:  "unused",
		Message: "Voucher valid",
	}

	if user.Disabled {
		details.Status = "disabled"
		details.Message = "User is currently disabled"
	}

	// 1. Fetch Profile info
	if user.Profile != "" {
		if profiles, err := u.gateway.GetUserProfiles(ctx, driver); err == nil {
			for _, p := range profiles {
				if p.Name == user.Profile {
					profileCopy := p
					details.Profile = &profileCopy
					break
				}
			}
		}
	}

	// 2. Check Active Sessions (Online status)
	if activeSessions, err := u.gateway.ListActiveSessions(ctx, driver); err == nil {
		for _, s := range activeSessions {
			if s.User == username {
				sessionCopy := s
				details.IsOnline = true
				details.ActiveSession = &sessionCopy
				details.Status = "active"
				break
			}
		}
	}

	// 3. Check Cookies
	if cookies, err := u.gateway.ListCookies(ctx, driver); err == nil {
		for _, c := range cookies {
			if c.User == username {
				cookieCopy := c
				details.HasCookie = true
				details.Cookie = &cookieCopy
				break
			}
		}
	}

	// 4. Parse Comment for Validity and Expire info
	if user.Comment != "" {
		if mc, err := u.gateway.ParseUserComment(user.Comment); err == nil {
			if mc.ExpireDate != "" {
				details.ExpireDate = mc.ExpireDate
			}
		}
		if strings.Contains(strings.ToLower(user.Comment), "expired") || user.LimitUptime == "1s" {
			details.Status = "expired"
			details.Message = "Voucher has expired"
		}
	}

	// 5. Calculate remaining uptime & bytes
	if user.LimitUptime != "" {
		details.SisaWaktu = user.LimitUptime
	}
	if user.LimitBytesIn != "" || user.LimitBytesOut != "" {
		details.SisaKuota = user.LimitBytesIn + "/" + user.LimitBytesOut
	}
	if user.MACAddress != "" {
		details.MACLocked = user.MACAddress
	}

	return details, nil
}
