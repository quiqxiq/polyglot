package dns

import (
	"strconv"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// DNSForwarder represents a DNS forwarder rule in RouterOS v7.
type DNSForwarder struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	DNSServers    []string `json:"dns_servers"`
	DoHServers    string   `json:"doh_servers"`
	VerifyDoHCert bool     `json:"verify_doh_cert"`
	Comment       string   `json:"comment"`
	Disabled      bool     `json:"disabled"`
}

// DNSAdlist represents an ad-blocking list rule (/ip/dns/adlist) in RouterOS v7.
type DNSAdlist struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	File       string `json:"file"`
	SSLVerify  bool   `json:"ssl_verify"`
	Comment    string `json:"comment"`
	Disabled   bool   `json:"disabled"`
	NameCount  int    `json:"name_count"`
	MatchCount int    `json:"match_count"`
}

// NewPrintDNSForwardersCommand queries DNS forwarders with strict proplist.
func NewPrintDNSForwardersCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "dns-servers", "doh-servers", "verify-doh-cert", "comment", "disabled")
	return command.Command{
		Raw:  "/ip/dns/forwarders/print",
		Args: args,
	}
}

// NewAddDNSForwarderCommand creates a DNS forwarder rule.
func NewAddDNSForwarderCommand(f DNSForwarder) command.Command {
	args := map[string]string{
		"name": f.Name,
	}
	if len(f.DNSServers) > 0 {
		args["dns-servers"] = strings.Join(f.DNSServers, ",")
	}
	if f.DoHServers != "" {
		args["doh-servers"] = f.DoHServers
	}
	if f.VerifyDoHCert {
		args["verify-doh-cert"] = "yes"
	}
	if f.Comment != "" {
		args["comment"] = f.Comment
	}
	if f.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/ip/dns/forwarders/add",
		Args: args,
	}
}

// NewRemoveDNSForwarderCommand removes a DNS forwarder by ID.
func NewRemoveDNSForwarderCommand(id string) command.Command {
	return command.Command{
		Raw: "/ip/dns/forwarders/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintDNSAdlistsCommand queries DNS adlists with strict proplist.
func NewPrintDNSAdlistsCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "url", "file", "ssl-verify", "comment", "disabled", "name-count", "match-count")
	return command.Command{
		Raw:  "/ip/dns/adlist/print",
		Args: args,
	}
}

// NewAddDNSAdlistCommand adds a DNS adlist source (URL or local file).
func NewAddDNSAdlistCommand(ad DNSAdlist) command.Command {
	args := make(map[string]string)
	if ad.URL != "" {
		args["url"] = ad.URL
	}
	if ad.File != "" {
		args["file"] = ad.File
	}
	if ad.SSLVerify {
		args["ssl-verify"] = "yes"
	} else {
		args["ssl-verify"] = "no"
	}
	if ad.Comment != "" {
		args["comment"] = ad.Comment
	}
	if ad.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/ip/dns/adlist/add",
		Args: args,
	}
}

// NewRemoveDNSAdlistCommand removes a DNS adlist entry by ID.
func NewRemoveDNSAdlistCommand(id string) command.Command {
	return command.Command{
		Raw: "/ip/dns/adlist/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewReloadDNSAdlistCommand re-downloads/reloads all configured DNS adlists.
func NewReloadDNSAdlistCommand() command.Command {
	return command.Command{
		Raw:  "/ip/dns/adlist/reload",
		Args: map[string]string{},
	}
}

// NewFlushDNSCacheCommand flushes the router's DNS cache.
func NewFlushDNSCacheCommand() command.Command {
	return command.Command{
		Raw:  "/ip/dns/cache/flush",
		Args: map[string]string{},
	}
}

// ParseDNSForwarders converts reply rows into DNSForwarder slice.
func ParseDNSForwarders(rows []map[string]string) []DNSForwarder {
	out := make([]DNSForwarder, 0, len(rows))
	for _, r := range rows {
		var servers []string
		if raw := r["dns-servers"]; raw != "" {
			for _, s := range strings.Split(raw, ",") {
				if s = strings.TrimSpace(s); s != "" {
					servers = append(servers, s)
				}
			}
		}
		out = append(out, DNSForwarder{
			ID:            r[".id"],
			Name:          r["name"],
			DNSServers:    servers,
			DoHServers:    r["doh-servers"],
			VerifyDoHCert: r["verify-doh-cert"] == "yes" || r["verify-doh-cert"] == "true",
			Comment:       r["comment"],
			Disabled:      r["disabled"] == "true",
		})
	}
	return out
}

// ParseDNSAdlists converts reply rows into DNSAdlist slice.
func ParseDNSAdlists(rows []map[string]string) []DNSAdlist {
	out := make([]DNSAdlist, 0, len(rows))
	for _, r := range rows {
		nameCnt, _ := strconv.Atoi(r["name-count"])
		matchCnt, _ := strconv.Atoi(r["match-count"])
		out = append(out, DNSAdlist{
			ID:         r[".id"],
			URL:        r["url"],
			File:       r["file"],
			SSLVerify:  r["ssl-verify"] == "yes" || r["ssl-verify"] == "true",
			Comment:    r["comment"],
			Disabled:   r["disabled"] == "true",
			NameCount:  nameCnt,
			MatchCount: matchCnt,
		})
	}
	return out
}
