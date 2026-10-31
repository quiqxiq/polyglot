package routing

// BGPSession represents a RouterOS BGP connection/session (/routing/bgp/session).
type BGPSession struct {
	RosID            string `json:"ros_id"`
	Name             string `json:"name"`
	RemoteAddress    string `json:"remote_address"`
	RemoteAS         string `json:"remote_as"`
	LocalAddress     string `json:"local_address"`
	LocalAS          string `json:"local_as"`
	State            string `json:"state"`
	Uptime           string `json:"uptime"`
	PrefixesReceived string `json:"prefixes_received"`
	Disabled         bool   `json:"disabled"`
	Comment          string `json:"comment"`
}

// OSPFNeighbor represents a RouterOS OSPF neighbor (/routing/ospf/neighbor).
type OSPFNeighbor struct {
	RosID     string `json:"ros_id"`
	Instance  string `json:"instance"`
	Area      string `json:"area"`
	Address   string `json:"address"`
	RouterID  string `json:"router_id"`
	State     string `json:"state"`
	Priority  string `json:"priority"`
	Adjacency string `json:"adjacency"`
}

// RoutingTable represents a routing table entry (/routing/table).
type RoutingTable struct {
	RosID    string `json:"ros_id"`
	Name     string `json:"name"`
	FIB      bool   `json:"fib"`
	Disabled bool   `json:"disabled"`
	Comment  string `json:"comment"`
}
