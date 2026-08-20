package hotspot

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
)

// parseRowBool converts a RouterOS print row boolean string ("true"/"false")
// into a Go bool. Anything else (empty, "yes", junk) is false.
func parseRowBool(v string) bool {
	return v == "true"
}

// ToProtoHotspotHosts converts raw /ip/hotspot/host/print rows into proto.
// Rows missing ".id" or "mac-address" are skipped.
func ToProtoHotspotHosts(rows []map[string]string) []*devicepb.HotspotHost {
	hosts := make([]*devicepb.HotspotHost, 0, len(rows))
	for _, row := range rows {
		if row[".id"] == "" || row["mac-address"] == "" {
			continue
		}
		hosts = append(hosts, &devicepb.HotspotHost{
			Id:         row[".id"],
			MacAddress: row["mac-address"],
			Address:    row["address"],
			ToAddress:  row["to-address"],
			Server:     row["server"],
			Bypassed:   parseRowBool(row["bypassed"]),
			Authorized: parseRowBool(row["authorized"]),
			Comment:    row["comment"],
		})
	}
	return hosts
}

// ToProtoHotspotServers converts raw /ip/hotspot/print rows into proto.
// Rows missing ".id" or "name" are skipped.
func ToProtoHotspotServers(rows []map[string]string) []*devicepb.HotspotServerInfo {
	servers := make([]*devicepb.HotspotServerInfo, 0, len(rows))
	for _, row := range rows {
		if row[".id"] == "" || row["name"] == "" {
			continue
		}
		servers = append(servers, &devicepb.HotspotServerInfo{
			Id:          row[".id"],
			Name:        row["name"],
			Interface:   row["interface"],
			AddressPool: row["address-pool"],
			Disabled:    parseRowBool(row["disabled"]),
			Comment:     row["comment"],
		})
	}
	return servers
}
