package hotspot

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/hotspot"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

// GenerateVouchers generates a batch of count vouchers. All legacy fields are
// mapped: server, time_limit, data_limit and comment (batch tag).
func (h *HotspotConnectHandler) GenerateVouchers(ctx context.Context, req *connect.Request[devicepb.GenerateVouchersRequest]) (*connect.Response[devicepb.GenerateVouchersResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	count := int(req.Msg.Count)
	if count <= 0 {
		count = 1
	}

	params := hotspot.VoucherGenerateParams{
		Server:      req.Msg.Server,
		Profile:     req.Msg.Profile,
		Prefix:      sanitizeBatchTag(req.Msg.Prefix),
		UserLength:  int(req.Msg.UserLength),
		CharSet:     hotspot.CharSet(req.Msg.CharacterSet),
		LimitUptime: req.Msg.TimeLimit,
		LimitBytes:  req.Msg.DataLimit,
		CommentTag:  sanitizeBatchTag(req.Msg.Comment),
	}

	batch, err := h.useCase.GenerateVouchers(ctx, driver, params, count)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	limitBytes := dataLimitBytes(req.Msg.DataLimit)
	pbVouchers := make([]*devicepb.HotspotUser, len(batch.Vouchers))
	for i, u := range batch.Vouchers {
		pbVouchers[i] = &devicepb.HotspotUser{
			Name:        u.Username,
			Password:    u.Password,
			Profile:     req.Msg.Profile,
			LimitUptime: req.Msg.TimeLimit,
			LimitBytes:  limitBytes,
			Comment:     u.Comment,
		}
	}

	return connect.NewResponse(&devicepb.GenerateVouchersResponse{
		Vouchers: pbVouchers,
		Message:  fmt.Sprintf("successfully generated %d vouchers", len(pbVouchers)),
	}), nil
}

// GetVoucherBatch returns all vouchers of a batch tag that have never logged
// in (uptime=0s), the replacement for legacy post_cache_voucher.php.
func (h *HotspotConnectHandler) GetVoucherBatch(ctx context.Context, req *connect.Request[devicepb.GetVoucherBatchRequest]) (*connect.Response[devicepb.GetVoucherBatchResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	comment := req.Msg.Comment
	users, err := h.useCase.GetUsers(ctx, driver, port.ListUsersFilter{
		Comment: comment,
	})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	if len(users) == 0 && sanitizeBatchTag(comment) != comment {
		users, err = h.useCase.GetUsers(ctx, driver, port.ListUsersFilter{
			Comment: sanitizeBatchTag(comment),
		})
		if err != nil {
			return nil, response.MapDomainError(err)
		}
	}
	if len(users) == 0 {
		allUsers, err := h.useCase.GetUsers(ctx, driver, port.ListUsersFilter{})
		if err == nil {
			tag := strings.TrimSpace(comment)
			for _, u := range allUsers {
				if u.Comment == tag || strings.Contains(u.Comment, tag) {
					users = append(users, u)
				}
			}
		}
	}

	return connect.NewResponse(&devicepb.GetVoucherBatchResponse{
		Vouchers: toProtoHotspotUsers(users),
		Count:    int32(len(users)),
	}), nil
}

// sanitizeBatchTag strips whitespace from batch tags/prefixes so they cannot
// break the vc-<code>-<date>-<tag> comment format (legacy notes §6.3).
func sanitizeBatchTag(s string) string {
	return strings.Join(strings.Fields(s), "")
}
