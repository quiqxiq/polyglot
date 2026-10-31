package dns

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// ParseDNSSettings converts command.Result from /ip/dns/print into a DNSSettings struct.
func ParseDNSSettings(result command.Result) DNSSettings {
	if len(result.Rows) == 0 {
		return DNSSettings{}
	}
	row := result.Rows[0]
	return DNSSettings{
		Servers:        row["servers"],
		DynamicServers: row["dynamic-servers"],
		AllowRemote:    strings.EqualFold(row["allow-remote-requests"], "true") || strings.EqualFold(row["allow-remote-requests"], "yes"),
		CacheSize:      row["cache-size"],
		CacheUsed:      row["cache-used"],
	}
}

// ParseDNSStatic converts command.Result rows from /ip/dns/static/print into []DNSStatic.
func ParseDNSStatic(result command.Result) []DNSStatic {
	records := make([]DNSStatic, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		records = append(records, DNSStatic{
			RosID:    id,
			Name:     name,
			Address:  row["address"],
			TTL:      row["ttl"],
			Comment:  row["comment"],
			Disabled: strings.EqualFold(row["disabled"], "true") || strings.EqualFold(row["disabled"], "yes"),
			Dynamic:  strings.EqualFold(row["dynamic"], "true") || strings.EqualFold(row["dynamic"], "yes"),
		})
	}
	return records
}

// ParseDNSCache converts command.Result rows from /ip/dns/cache/print into []DNSCache.
func ParseDNSCache(result command.Result) []DNSCache {
	entries := make([]DNSCache, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		entries = append(entries, DNSCache{
			RosID:  id,
			Name:   name,
			Type:   row["type"],
			Data:   row["data"],
			TTL:    row["ttl"],
			Static: strings.EqualFold(row["static"], "true") || strings.EqualFold(row["static"], "yes"),
		})
	}
	return entries
}
