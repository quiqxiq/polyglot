package mikrotik

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// IsolationRedirectComment menandai rule dst-nat redirect halaman bayar
// milik aplikasi — dipakai untuk lookup idempoten & pembersihan.
const IsolationRedirectComment = "POLYGLOT_ISOLIR_REDIRECT"

// FirewallNATParams holds parameters for /ip/firewall/nat/add|set.
//
// Field notes (RouterOS /ip/firewall/nat reference):
//   - Chain          : "dstnat" untuk redirect trafik masuk pelanggan.
//   - Action         : "redirect" (ke router) atau "dst-nat" (ke host lain).
//   - SrcAddressList : address-list sumber (mis. "ISOLIR_USERS").
//   - Protocol       : "tcp" (umumnya cukup; UDP opsional).
//   - DstPort        : "80,443" dsb.
//   - ToAddresses    : target dst-nat (host halaman bayar); kosongkan bila
//     action=redirect ke router itu sendiri.
//   - ToPorts        : port tujuan setelah NAT.
//   - Comment        : penanda kepemilikan rule (lihat IsolationRedirectComment).
//   - Disabled       : true = rule tersimpan nonaktif (aman untuk E2E/dry-run).
type FirewallNATParams struct {
	Chain          string
	Action         string
	SrcAddressList string
	Protocol       string
	DstPort        string
	ToAddresses    string
	ToPorts        string
	Comment        string
	Disabled       bool
}

// IsolationRedirectNATParams returns pre-filled params for the global
// payment-page redirect rule applied to isolated subscribers.
// disabled=true digunakan saat pembuatan awal/E2E agar trafik tak tersentuh.
func IsolationRedirectNATParams(srcList, protocol, dstPorts, toAddresses, toPorts string, disabled bool) FirewallNATParams {
	return FirewallNATParams{
		Chain:          "dstnat",
		Action:         "dst-nat",
		SrcAddressList: srcList,
		Protocol:       protocol,
		DstPort:        dstPorts,
		ToAddresses:    toAddresses,
		ToPorts:        toPorts,
		Comment:        IsolationRedirectComment,
		Disabled:       disabled,
	}
}

// NewAddFirewallNATCommand builds the command.Command for
// /ip/firewall/nat/add. Only non-empty fields are sent.
func NewAddFirewallNATCommand(p FirewallNATParams) command.Command {
	args := map[string]string{
		"chain":  p.Chain,
		"action": p.Action,
	}
	setIfNonEmpty(args, "src-address-list", p.SrcAddressList)
	setIfNonEmpty(args, "protocol", p.Protocol)
	setIfNonEmpty(args, "dst-port", p.DstPort)
	setIfNonEmpty(args, "to-addresses", p.ToAddresses)
	setIfNonEmpty(args, "to-ports", p.ToPorts)
	setIfNonEmpty(args, "comment", p.Comment)
	if p.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{Raw: "/ip/firewall/nat/add", Args: args}
}

// NewSetFirewallNATCommand updates an existing NAT rule by RouterOS .id;
// only non-empty fields change. Set Disabled via this command to toggle.
func NewSetFirewallNATCommand(rosID string, p FirewallNATParams) command.Command {
	args := map[string]string{}
	if p.Chain != "" {
		args["chain"] = p.Chain
	}
	if p.Action != "" {
		args["action"] = p.Action
	}
	setIfNonEmpty(args, "src-address-list", p.SrcAddressList)
	setIfNonEmpty(args, "protocol", p.Protocol)
	setIfNonEmpty(args, "dst-port", p.DstPort)
	setIfNonEmpty(args, "to-addresses", p.ToAddresses)
	setIfNonEmpty(args, "to-ports", p.ToPorts)
	if p.Comment != "" {
		args["comment"] = p.Comment
	}
	if p.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{Raw: "/ip/firewall/nat/set", Args: withNumbers(rosID, args)}
}

// withNumbers merges the "numbers" targeting key into args (RouterOS set/
// remove convention: [find .id=<rosID>]). Existing keys win.
func withNumbers(rosID string, args map[string]string) map[string]string {
	out := map[string]string{"numbers": rosID}
	for k, v := range args {
		if _, exists := out[k]; !exists {
			out[k] = v
		}
	}
	return out
}

// NewRemoveFirewallNATCommand deletes a NAT rule by RouterOS .id.
func NewRemoveFirewallNATCommand(rosID string) command.Command {
	return command.Command{Raw: "/ip/firewall/nat/remove", Args: map[string]string{"numbers": rosID}}
}

// NewPrintFirewallNATCommand lists NAT rules optionally filtered by
// chain/comment/src-address-list.
func NewPrintFirewallNATCommand(chain, comment, srcAddressList string) command.Command {
	args := map[string]string{}
	setIfNonEmpty(args, "?chain", chain)
	setIfNonEmpty(args, "?comment", comment)
	setIfNonEmpty(args, "?src-address-list", srcAddressList)
	return command.Command{Raw: "/ip/firewall/nat/print", Args: args}
}

// ParseFirewallNATRules parses /ip/firewall/nat/print results.
func ParseFirewallNATRules(result command.Result) []FirewallNATRule {
	out := make([]FirewallNATRule, 0, len(result.Rows))
	for _, row := range result.Rows {
		out = append(out, FirewallNATRule{
			RosID:          row[".id"],
			Chain:          row["chain"],
			Action:         row["action"],
			SrcAddressList: row["src-address-list"],
			Protocol:       row["protocol"],
			DstPort:        row["dst-port"],
			ToAddresses:    row["to-addresses"],
			ToPorts:        row["to-ports"],
			Comment:        row["comment"],
			Disabled:       strings.EqualFold(row["disabled"], "true"),
		})
	}
	return out
}

// FirewallNATRule is the parsed /ip/firewall/nat row.
type FirewallNATRule struct {
	RosID          string
	Chain          string
	Action         string
	SrcAddressList string
	Protocol       string
	DstPort        string
	ToAddresses    string
	ToPorts        string
	Comment        string
	Disabled       bool
}

// FindIsolationRedirectRules returns app-owned redirect rules from a list.
func FindIsolationRedirectRules(rules []FirewallNATRule) []FirewallNATRule {
	var out []FirewallNATRule
	for _, r := range rules {
		if strings.Contains(r.Comment, IsolationRedirectComment) {
			out = append(out, r)
		}
	}
	return out
}
