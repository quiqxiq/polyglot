package ipservice

import (
	"strconv"
)

// ParseIPServices converts reply rows into IPService slice.
func ParseIPServices(rows []map[string]string) []IPService {
	out := make([]IPService, 0, len(rows))
	for _, r := range rows {
		p, _ := strconv.Atoi(r["port"])
		out = append(out, IPService{
			ID:          r[".id"],
			Name:        r["name"],
			Port:        p,
			Disabled:    r["disabled"] == "true",
			Address:     r["address"],
			Certificate: r["certificate"],
		})
	}
	return out
}

// ParseIPCloud converts reply rows into IPCloud struct.
func ParseIPCloud(rows []map[string]string) *IPCloud {
	if len(rows) == 0 {
		return nil
	}
	r := rows[0]
	return &IPCloud{
		DDNSEnabled:   r["ddns-enabled"] == "yes" || r["ddns-enabled"] == "true",
		UpdateTime:    r["update-time"] == "yes" || r["update-time"] == "true",
		PublicAddress: r["public-address"],
		DNSName:       r["dns-name"],
		Status:        r["status"],
	}
}

// ParseNeighbors converts reply rows into NeighborEntry slice.
func ParseNeighbors(rows []map[string]string) []NeighborEntry {
	out := make([]NeighborEntry, 0, len(rows))
	for _, r := range rows {
		out = append(out, NeighborEntry{
			ID:                r[".id"],
			Interface:         r["interface"],
			MacAddress:        r["mac-address"],
			Identity:          r["identity"],
			Platform:          r["platform"],
			Version:           r["version"],
			Unpack:            r["unpack"],
			Age:               r["age"],
			SystemDescription: r["system-description"],
			SystemCaps:        r["system-caps"],
		})
	}
	return out
}

// ParseDiscoverySettings converts reply rows into DiscoverySettings.
func ParseDiscoverySettings(rows []map[string]string) *DiscoverySettings {
	if len(rows) == 0 {
		return nil
	}
	return &DiscoverySettings{
		DiscoverInterfaceList: rows[0]["discover-interface-list"],
	}
}

// ParseARPEntries converts reply rows into ARPEntry slice.
func ParseARPEntries(rows []map[string]string) []ARPEntry {
	out := make([]ARPEntry, 0, len(rows))
	for _, r := range rows {
		out = append(out, ARPEntry{
			ID:         r[".id"],
			Address:    r["address"],
			MacAddress: r["mac-address"],
			Interface:  r["interface"],
			Status:     r["status"],
			Dynamic:    r["dynamic"] == "true",
			DHCP:       r["DHCP"] == "true" || r["dhcp"] == "true",
			Disabled:   r["disabled"] == "true",
			Comment:    r["comment"],
		})
	}
	return out
}

// ParseSNMPSettings converts reply rows into SNMPSettings.
func ParseSNMPSettings(rows []map[string]string) *SNMPSettings {
	if len(rows) == 0 {
		return nil
	}
	r := rows[0]
	return &SNMPSettings{
		Enabled:  r["enabled"] == "yes" || r["enabled"] == "true",
		Contact:  r["contact"],
		Location: r["location"],
	}
}

// ParseSNMPCommunities converts reply rows into SNMPCommunity slice.
func ParseSNMPCommunities(rows []map[string]string) []SNMPCommunity {
	out := make([]SNMPCommunity, 0, len(rows))
	for _, r := range rows {
		out = append(out, SNMPCommunity{
			ID:          r[".id"],
			Name:        r["name"],
			Addresses:   r["addresses"],
			Security:    r["security"],
			ReadAccess:  r["read-access"] == "yes" || r["read-access"] == "true",
			WriteAccess: r["write-access"] == "yes" || r["write-access"] == "true",
		})
	}
	return out
}
