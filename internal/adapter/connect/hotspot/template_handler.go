package hotspot

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/hotspot"
	"github.com/quixiq/polyglot/internal/port"
	templatefs "github.com/quixiq/polyglot/internal/template"
	"github.com/quixiq/polyglot/pkg/response"
	"github.com/quixiq/polyglot/pkg/voucher"
)

// ListTemplates returns the supported voucher template layouts and their
// sections. Templates are embedded read-only (decision #5 — no Save).
func (h *HotspotConnectHandler) ListTemplates(ctx context.Context, req *connect.Request[devicepb.ListTemplatesRequest]) (*connect.Response[devicepb.ListTemplatesResponse], error) {
	infos := voucher.ListTemplates()
	pbTemplates := make([]*devicepb.TemplateInfo, len(infos))
	for i, ti := range infos {
		pbTemplates[i] = &devicepb.TemplateInfo{Name: ti.Name, Sections: ti.Sections}
	}
	return connect.NewResponse(&devicepb.ListTemplatesResponse{Templates: pbTemplates}), nil
}

// GetTemplateSection returns the raw content of one template section
// (header|row|footer) for the given layout, read from the embedded FS.
func (h *HotspotConnectHandler) GetTemplateSection(ctx context.Context, req *connect.Request[devicepb.GetTemplateSectionRequest]) (*connect.Response[devicepb.GetTemplateSectionResponse], error) {
	layout, err := normalizeTemplateLayout(req.Msg.TemplateName)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	section, err := normalizeTemplateSection(req.Msg.Section)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	data, err := templatefs.FS.ReadFile(voucher.TemplateFile(layout, section))
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("template section %s.%s not found: %w", section, layout, err))
	}
	return connect.NewResponse(&devicepb.GetTemplateSectionResponse{Content: string(data)}), nil
}

// RenderVouchers produces the printable HTML for a single voucher (.id), a
// batch (comment tag, unused only), or dummy preview data. Metadata is
// scoped down (decision #4): hotspot name comes from router identity, dns
// name from the first hotspot server, price/validity from the profile's
// on-login script. QR encodes the login URL (decision #6).
func (h *HotspotConnectHandler) RenderVouchers(ctx context.Context, req *connect.Request[devicepb.RenderVouchersRequest]) (*connect.Response[devicepb.RenderVouchersResponse], error) {
	layout, err := normalizeTemplateLayout(req.Msg.TemplateName)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Preview: dummy data, no router access needed.
	if req.Msg.Preview {
		html, err := voucher.RenderWithOptions(previewVoucherData(), voucher.Layout(layout), h.useCase.TemplateDir, voucher.Options{QRMode: voucher.QRModeLoginURL})
		if err != nil {
			return nil, response.MapDomainError(err)
		}
		return connect.NewResponse(&devicepb.RenderVouchersResponse{Html: html, TotalVouchers: 3}), nil
	}

	if req.Msg.UserId == "" && req.Msg.Comment == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user_id or comment (batch tag) required"))
	}
	if req.Msg.UserId != "" && req.Msg.Comment != "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user_id and comment are mutually exclusive"))
	}

	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	// Gather voucher users: single by .id or batch by comment tag (unused only).
	var users []port.HotspotUser
	if req.Msg.UserId != "" {
		u, err := h.useCase.GetUser(ctx, driver, req.Msg.UserId)
		if err != nil {
			return nil, response.MapDomainError(err)
		}
		users = []port.HotspotUser{u}
	} else {
		users, err = h.useCase.GetUsers(ctx, driver, port.ListUsersFilter{
			Comment:    sanitizeBatchTag(req.Msg.Comment),
			OnlyUnused: true,
		})
		if err != nil {
			return nil, response.MapDomainError(err)
		}
	}
	if len(users) == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("no vouchers found for the given filter"))
	}

	// Scope-down metadata: identity + first hotspot server + profile on-login.
	hotspotName := ""
	if identity, err := h.useCase.GetSystemIdentity(ctx, driver); err == nil {
		hotspotName = identity
	}
	dnsName := hotspotName
	if servers, err := h.useCase.GetHotspotServers(ctx, driver); err == nil && len(servers) > 0 {
		if addr := servers[0]["address"]; addr != "" {
			dnsName = addr
		} else if name := servers[0]["name"]; name != "" {
			dnsName = name
		}
		if hotspotName == "" {
			hotspotName = servers[0]["name"]
		}
	}

	profileMeta := map[string]hotspot.ProfileMeta{}
	if profiles, err := h.useCase.GetProfiles(ctx, driver); err == nil {
		for _, p := range profiles {
			if meta, perr := hotspot.ParseOnLoginScript(p.OnLogin); perr == nil {
				profileMeta[p.Name] = meta
			}
		}
	}

	data := make([]voucher.VoucherData, 0, len(users))
	for i, u := range users {
		meta := profileMeta[u.Profile]
		data = append(data, voucher.VoucherData{
			Username:        u.Name,
			Password:        u.Password,
			Price:           formatPrice(meta.Price),
			Validity:        meta.Validity,
			LimitUptime:     u.LimitUptime,
			LimitBytesTotal: u.LimitBytesIn,
			HotspotName:     hotspotName,
			DNSName:         dnsName,
			Comment:         u.Comment,
			Number:          i + 1,
		})
	}

	html, err := voucher.RenderWithOptions(data, voucher.Layout(layout), h.useCase.TemplateDir, voucher.Options{QRMode: voucher.QRModeLoginURL})
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.RenderVouchersResponse{
		Html:          html,
		TotalVouchers: int32(len(data)),
	}), nil
}

// previewVoucherData returns dummy voucher rows matching the legacy admin
// preview (mikhmon / 1234).
func previewVoucherData() []voucher.VoucherData {
	base := voucher.VoucherData{
		HotspotName: "Mikhmon",
		DNSName:     "192.168.88.1",
		TimeStamp:   "",
	}
	return []voucher.VoucherData{
		{Username: "mikhmon", Password: "1234", Price: "5000", Validity: "1d", Comment: "vc-PREV", HotspotName: base.HotspotName, DNSName: base.DNSName, Number: 1},
		{Username: "mikhmon1", Password: "5678", Price: "10000", Validity: "7d", Comment: "vc-PREV", HotspotName: base.HotspotName, DNSName: base.DNSName, Number: 2},
		{Username: "mikhmon2", Password: "9012", Price: "20000", Validity: "30d", Comment: "vc-PREV", HotspotName: base.HotspotName, DNSName: base.DNSName, Number: 3},
	}
}

// normalizeTemplateLayout validates and defaults the template name.
func normalizeTemplateLayout(layout string) (string, error) {
	if layout == "" {
		return string(voucher.LayoutDefault), nil
	}
	switch layout {
	case string(voucher.LayoutDefault), string(voucher.LayoutSmall), string(voucher.LayoutThermal):
		return layout, nil
	}
	return "", fmt.Errorf("unknown template %q (default|small|thermal)", layout)
}

// normalizeTemplateSection validates the template section name.
func normalizeTemplateSection(section string) (string, error) {
	switch section {
	case voucher.SectionHeader, voucher.SectionRow, voucher.SectionFooter:
		return section, nil
	}
	return "", fmt.Errorf("unknown section %q (header|row|footer)", section)
}

// formatPrice renders a float price the way the legacy UI does: integers
// without decimals, otherwise two decimals.
func formatPrice(p float64) string {
	if p == 0 {
		return ""
	}
	if p == math.Trunc(p) {
		return strconv.FormatFloat(p, 'f', 0, 64)
	}
	return strconv.FormatFloat(p, 'f', 2, 64)
}
