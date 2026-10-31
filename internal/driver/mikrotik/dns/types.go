package dns

// DNSSettings holds general DNS server configuration.
type DNSSettings struct {
	Servers        string `json:"servers"`
	DynamicServers string `json:"dynamic_servers"`
	AllowRemote    bool   `json:"allow_remote_requests"`
	CacheSize      string `json:"cache_size"`
	CacheUsed      string `json:"cache_used"`
}

// DNSStatic represents a static DNS record (/ip/dns/static).
type DNSStatic struct {
	RosID    string `json:"ros_id"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	TTL      string `json:"ttl"`
	Comment  string `json:"comment"`
	Disabled bool   `json:"disabled"`
	Dynamic  bool   `json:"dynamic"`
}

// DNSCache represents a cached DNS entry (/ip/dns/cache).
type DNSCache struct {
	RosID  string `json:"ros_id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Data   string `json:"data"`
	TTL    string `json:"ttl"`
	Static bool   `json:"static"`
}

// DNSStaticParams holds creation/update parameters for a static DNS record.
type DNSStaticParams struct {
	Name     string
	Address  string
	TTL      string
	Comment  string
	Disabled bool
}
