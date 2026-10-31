package routing

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// ParseBGPSessions converts command.Result from /routing/bgp/session/print into []BGPSession.
func ParseBGPSessions(result command.Result) []BGPSession {
	sessions := make([]BGPSession, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		sessions = append(sessions, BGPSession{
			RosID:            id,
			Name:             name,
			RemoteAddress:    row["remote.address"],
			RemoteAS:         row["remote.as"],
			LocalAddress:     row["local.address"],
			LocalAS:          row["local.as"],
			State:            row["state"],
			Uptime:           row["uptime"],
			PrefixesReceived: row["prefix-count"],
			Disabled:         strings.EqualFold(row["disabled"], "true") || strings.EqualFold(row["disabled"], "yes"),
			Comment:          row["comment"],
		})
	}
	return sessions
}

// ParseOSPFNeighbors converts command.Result from /routing/ospf/neighbor/print into []OSPFNeighbor.
func ParseOSPFNeighbors(result command.Result) []OSPFNeighbor {
	neighbors := make([]OSPFNeighbor, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		if id == "" {
			continue
		}
		neighbors = append(neighbors, OSPFNeighbor{
			RosID:     id,
			Instance:  row["instance"],
			Area:      row["area"],
			Address:   row["address"],
			RouterID:  row["router-id"],
			State:     row["state"],
			Priority:  row["priority"],
			Adjacency: row["adjacency"],
		})
	}
	return neighbors
}

// ParseRoutingTables converts command.Result from /routing/table/print into []RoutingTable.
func ParseRoutingTables(result command.Result) []RoutingTable {
	tables := make([]RoutingTable, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		tables = append(tables, RoutingTable{
			RosID:    id,
			Name:     name,
			FIB:      strings.EqualFold(row["fib"], "true") || strings.EqualFold(row["fib"], "yes"),
			Disabled: strings.EqualFold(row["disabled"], "true") || strings.EqualFold(row["disabled"], "yes"),
			Comment:  row["comment"],
		})
	}
	return tables
}
