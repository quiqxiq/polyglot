package mikrotik

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// ─── Queue builders ───────────────────────────────────────────────────────

func TestNewPrintSimpleQueuesCommand(t *testing.T) {
	t.Run("tanpa filter", func(t *testing.T) {
		cmd := NewPrintSimpleQueuesCommand("")
		assert.Equal(t, "/queue/simple/print", cmd.Raw)
		assert.Empty(t, cmd.Args)
		assert.False(t, isStreamingCommand(cmd))
	})
	t.Run("dengan filter nama", func(t *testing.T) {
		cmd := NewPrintSimpleQueuesCommand("sub-budi")
		assert.Equal(t, "sub-budi", cmd.Args["?name"])
	})
}

func TestNewStreamQueueStatsCommand(t *testing.T) {
	t.Run("default interval 1s", func(t *testing.T) {
		cmd := NewStreamQueueStatsCommand(QueueStreamParams{})
		assert.Equal(t, "/queue/simple/print", cmd.Raw)
		assert.Equal(t, "1s", cmd.Args["interval"], "default interval harus 1s")
		_, hasStats := cmd.Args["stats"]
		assert.True(t, hasStats, "harus ada flag stats")
		assert.True(t, isStreamingCommand(cmd), "harus terdeteksi streaming")
	})
	t.Run("interval custom dan filter nama", func(t *testing.T) {
		cmd := NewStreamQueueStatsCommand(QueueStreamParams{
			NameFilter: "sub-budi",
			Interval:   "500ms",
		})
		assert.Equal(t, "500ms", cmd.Args["interval"])
		assert.Equal(t, "sub-budi", cmd.Args["?name"])
		assert.True(t, isStreamingCommand(cmd))
	})
	t.Run("parents only dan parent filter", func(t *testing.T) {
		cmd := NewStreamQueueStatsCommand(QueueStreamParams{
			ParentFilter: "parent-total",
			ParentsOnly:  true,
		})
		assert.Equal(t, "parent-total", cmd.Args["?parent"])
		assert.Equal(t, "false", cmd.Args["?dynamic"])
	})
}

func TestNewAddSimpleQueueCommand(t *testing.T) {
	cmd := NewAddSimpleQueueCommand(SimpleQueueParams{
		Name:     "sub-budi",
		Target:   "10.10.0.5",
		MaxLimit: "10M/10M",
		Comment:  "subscriber budi",
	})
	assert.Equal(t, "/queue/simple/add", cmd.Raw)
	assert.Equal(t, "sub-budi", cmd.Args["name"])
	assert.Equal(t, "10.10.0.5", cmd.Args["target"])
	assert.Equal(t, "10M/10M", cmd.Args["max-limit"])
	assert.Equal(t, "subscriber budi", cmd.Args["comment"])
	assert.Equal(t, "no", cmd.Args["disabled"], "disabled harus selalu dikirim")
}

func TestNewRemoveSimpleQueueCommand(t *testing.T) {
	cmd := NewRemoveSimpleQueueCommand("*9")
	assert.Equal(t, "/queue/simple/remove", cmd.Raw)
	assert.Equal(t, "*9", cmd.Args[".id"])
}

func TestSuspendQueueParams(t *testing.T) {
	t.Run("speed default 1k/1k", func(t *testing.T) {
		p := SuspendQueueParams("10.10.0.5", "unpaid", "")
		assert.Equal(t, "suspended_10_10_0_5", p.Name)
		assert.Equal(t, "10.10.0.5", p.Target)
		assert.Equal(t, "1k/1k", p.MaxLimit)
		assert.Contains(t, p.Comment, "SUSPENDED")
		assert.Contains(t, p.Comment, "unpaid")
	})
	t.Run("speed custom", func(t *testing.T) {
		p := SuspendQueueParams("192.168.1.1", "overdue", "64k/64k")
		assert.Equal(t, "64k/64k", p.MaxLimit)
	})
}

// ─── Queue parsers ────────────────────────────────────────────────────────

func TestParseSimpleQueues(t *testing.T) {
	result := command.Result{Rows: []map[string]string{
		{
			".id": "*1", "name": "sub-budi",
			"target": "10.10.0.5", "max-limit": "10M/10M",
			"rate": "1234567/2345678", "bytes": "100000/200000",
			"disabled": "false",
		},
		{".id": "", "name": "bad"}, // dilewati
	}}
	queues := ParseSimpleQueues(result)
	require.Len(t, queues, 1)
	q := queues[0]
	assert.Equal(t, "*1", q.RosID)
	assert.Equal(t, "sub-budi", q.Name)
	assert.Equal(t, "10.10.0.5", q.Target)
	assert.Equal(t, "10M/10M", q.MaxLimit)
	assert.Equal(t, "1234567/2345678", q.Rate)
	assert.Equal(t, "100000/200000", q.Bytes)
	assert.False(t, q.Disabled)
}

// ─── Firewall builders ────────────────────────────────────────────────────

func TestNewAddFirewallFilterCommand(t *testing.T) {
	t.Run("BlockIPFilterParams", func(t *testing.T) {
		p := BlockIPFilterParams("192.168.1.100", "unpaid")
		cmd := NewAddFirewallFilterCommand(p)
		assert.Equal(t, "/ip/firewall/filter/add", cmd.Raw)
		assert.Equal(t, "forward", cmd.Args["chain"])
		assert.Equal(t, "drop", cmd.Args["action"])
		assert.Equal(t, "192.168.1.100", cmd.Args["src-address"])
		assert.Contains(t, cmd.Args["comment"], "SUSPENDED")
		assert.Contains(t, cmd.Args["comment"], "192.168.1.100")
	})

	t.Run("BlockAddressListFilterParams forward", func(t *testing.T) {
		p := BlockAddressListFilterParams("forward", "blocked_customers")
		cmd := NewAddFirewallFilterCommand(p)
		assert.Equal(t, "forward", cmd.Args["chain"])
		assert.Equal(t, "blocked_customers", cmd.Args["src-address-list"])
		assert.Equal(t, "0", cmd.Args["place-before"])
	})

	t.Run("BlockAddressListFilterParams input", func(t *testing.T) {
		p := BlockAddressListFilterParams("input", "blocked_customers")
		cmd := NewAddFirewallFilterCommand(p)
		assert.Equal(t, "input", cmd.Args["chain"])
	})
}

func TestParseFirewallFilters(t *testing.T) {
	result := command.Result{Rows: []map[string]string{
		{
			".id": "*1", "chain": "forward", "action": "drop",
			"src-address": "10.0.0.5", "bytes": "1024", "disabled": "false",
		},
		{".id": ""}, // dilewati
	}}
	filters := ParseFirewallFilters(result)
	require.Len(t, filters, 1)
	f := filters[0]
	assert.Equal(t, "*1", f.RosID)
	assert.Equal(t, "forward", f.Chain)
	assert.Equal(t, "drop", f.Action)
	assert.Equal(t, "10.0.0.5", f.SrcAddress)
	assert.False(t, f.Disabled)
}

// ─── Address List builders ────────────────────────────────────────────────

func TestNewAddToAddressListCommand(t *testing.T) {
	cmd := NewAddToAddressListCommand("blocked_customers", "10.0.0.5", "SUSPENDED - unpaid")
	assert.Equal(t, "/ip/firewall/address-list/add", cmd.Raw)
	assert.Equal(t, "blocked_customers", cmd.Args["list"])
	assert.Equal(t, "10.0.0.5", cmd.Args["address"])
	assert.Equal(t, "SUSPENDED - unpaid", cmd.Args["comment"])
}

func TestParseAddressListEntries(t *testing.T) {
	result := command.Result{Rows: []map[string]string{
		{
			".id": "*1", "list": "blocked_customers", "address": "10.0.0.5",
			"dynamic": "false", "disabled": "false",
		},
	}}
	entries := ParseAddressListEntries(result)
	require.Len(t, entries, 1)
	e := entries[0]
	assert.Equal(t, "blocked_customers", e.List)
	assert.Equal(t, "10.0.0.5", e.Address)
	assert.False(t, e.Dynamic)
}

// ─── DHCP builders ────────────────────────────────────────────────────────

func TestNewPrintDHCPLeasesCommand(t *testing.T) {
	t.Run("tanpa filter", func(t *testing.T) {
		cmd := NewPrintDHCPLeasesCommand("")
		assert.Equal(t, "/ip/dhcp-server/lease/print", cmd.Raw)
		assert.Empty(t, cmd.Args)
		assert.False(t, isStreamingCommand(cmd))
	})
	t.Run("dengan MAC filter", func(t *testing.T) {
		cmd := NewPrintDHCPLeasesCommand("AA:BB:CC:DD:EE:FF")
		assert.Equal(t, "AA:BB:CC:DD:EE:FF", cmd.Args["?mac-address"])
	})
}

func TestNewStreamDHCPLeasesCommand(t *testing.T) {
	cmd := NewStreamDHCPLeasesCommand("")
	assert.Equal(t, "/ip/dhcp-server/lease/print", cmd.Raw)
	_, hasFollow := cmd.Args["follow"]
	assert.True(t, hasFollow)
	assert.True(t, isStreamingCommand(cmd))
}

func TestNewSetDHCPLeaseBlockCommand(t *testing.T) {
	t.Run("block", func(t *testing.T) {
		cmd := NewSetDHCPLeaseBlockCommand("*3", DHCPLeaseBlockParams{
			Blocked: true, Comment: "SUSPENDED - unpaid",
		})
		assert.Equal(t, "/ip/dhcp-server/lease/set", cmd.Raw)
		assert.Equal(t, "*3", cmd.Args[".id"])
		assert.Equal(t, "yes", cmd.Args["blocked"])
		assert.Equal(t, "SUSPENDED - unpaid", cmd.Args["comment"])
	})
	t.Run("unblock", func(t *testing.T) {
		cmd := NewSetDHCPLeaseBlockCommand("*3", DHCPLeaseBlockParams{Blocked: false})
		assert.Equal(t, "no", cmd.Args["blocked"])
	})
}

func TestParseDHCPLeases(t *testing.T) {
	result := command.Result{Rows: []map[string]string{
		{
			".id": "*1", "address": "192.168.1.10", "mac-address": "AA:BB:CC:DD:EE:FF",
			"status": "bound", "blocked": "false", "host-name": "mypc",
		},
	}}
	leases := ParseDHCPLeases(result)
	require.Len(t, leases, 1)
	l := leases[0]
	assert.Equal(t, "192.168.1.10", l.Address)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", l.MACAddress)
	assert.Equal(t, "bound", l.Status)
	assert.False(t, l.Blocked)
	assert.Equal(t, "mypc", l.HostName)
}

// ─── Interface builders ───────────────────────────────────────────────────

func TestNewMonitorTrafficCommands(t *testing.T) {
	t.Run("once command — tidak streaming", func(t *testing.T) {
		cmd := NewMonitorTrafficOnceCommand("ether1")
		assert.Equal(t, "/interface/monitor-traffic", cmd.Raw)
		assert.Equal(t, "ether1", cmd.Args["interface"])
		_, hasOnce := cmd.Args["once"]
		assert.True(t, hasOnce, "once harus ada")
		assert.False(t, isStreamingCommand(cmd), "dengan once tidak boleh streaming")
	})

	t.Run("stream command — streaming", func(t *testing.T) {
		cmd := NewMonitorTrafficStreamCommand("ether1")
		assert.Equal(t, "/interface/monitor-traffic", cmd.Raw)
		assert.Equal(t, "ether1", cmd.Args["interface"])
		_, hasOnce := cmd.Args["once"]
		assert.False(t, hasOnce, "tanpa once")
		assert.True(t, isStreamingCommand(cmd), "tanpa once harus streaming")
	})
}

func TestParseInterfaceTrafficStats(t *testing.T) {
	t.Run("baris normal", func(t *testing.T) {
		result := command.Result{Rows: []map[string]string{
			{
				"rx-bits-per-second": "10000000", "tx-bits-per-second": "5000000",
				"rx-packets-per-second": "100", "tx-packets-per-second": "50",
			},
		}}
		stats := ParseInterfaceTrafficStats(result)
		assert.Equal(t, "10000000", stats.RxBitsPerSecond)
		assert.Equal(t, "5000000", stats.TxBitsPerSecond)
		assert.Equal(t, "100", stats.RxPacketsPerSecond)
	})

	t.Run("result kosong — zero value", func(t *testing.T) {
		stats := ParseInterfaceTrafficStats(command.Result{})
		assert.Equal(t, InterfaceTrafficStats{}, stats)
	})
}

// ─── System builders ──────────────────────────────────────────────────────

func TestNewStreamSystemResourceCommand(t *testing.T) {
	cmd := NewStreamSystemResourceCommand("")
	assert.Equal(t, "/system/resource/print", cmd.Raw)
	assert.Equal(t, "1s", cmd.Args["interval"])
	assert.True(t, isStreamingCommand(cmd), "harus terdeteksi sebagai streaming")
}

func TestParseSystemResource(t *testing.T) {
	t.Run("baris normal", func(t *testing.T) {
		result := command.Result{Rows: []map[string]string{
			{
				"cpu-load": "25", "cpu-count": "4", "uptime": "5d2h",
				"version": "7.10.2 (stable)", "board-name": "RB750Gr3",
				"free-memory": "52428800", "total-memory": "67108864",
				"voltage": "24000",
			},
		}}
		r := ParseSystemResource(result)
		assert.Equal(t, 25, r.CPULoad)
		assert.Equal(t, 4, r.CPUCount)
		assert.Equal(t, "5d2h", r.Uptime)
		assert.Equal(t, "7.10.2 (stable)", r.Version)
		assert.Equal(t, "RB750Gr3", r.BoardName)
		assert.Equal(t, "24000", r.Voltage)
	})

	t.Run("voltage fallback ke board-voltage", func(t *testing.T) {
		result := command.Result{Rows: []map[string]string{
			{"cpu-load": "0", "board-voltage": "12000"},
		}}
		r := ParseSystemResource(result)
		assert.Equal(t, "12000", r.Voltage)
	})

	t.Run("result kosong — zero value", func(t *testing.T) {
		r := ParseSystemResource(command.Result{})
		assert.Equal(t, SystemResource{}, r)
	})
}

func TestParseLogEntries(t *testing.T) {
	result := command.Result{Rows: []map[string]string{
		{".id": "*1", "time": "10:00:00", "topics": "ppp,info", "message": "budi logged in"},
		{".id": ""}, // dilewati
	}}
	entries := ParseLogEntries(result)
	require.Len(t, entries, 1)
	assert.Equal(t, "ppp,info", entries[0].Topics)
	assert.Equal(t, "budi logged in", entries[0].Message)
}

func TestNewPingCommand(t *testing.T) {
	t.Run("count default 4", func(t *testing.T) {
		cmd := NewPingCommand("10.0.0.1", "")
		assert.Equal(t, "/ping", cmd.Raw)
		assert.Equal(t, "10.0.0.1", cmd.Args["address"])
		assert.Equal(t, "4", cmd.Args["count"])
		assert.True(t, isStreamingCommand(cmd))
	})
	t.Run("count custom", func(t *testing.T) {
		cmd := NewPingCommand("8.8.8.8", "10")
		assert.Equal(t, "10", cmd.Args["count"])
	})
}

func TestFilterInactiveHotspotUsers(t *testing.T) {
	users := []HotspotUser{
		{RosID: "*1", Name: "voucher10"},
		{RosID: "*2", Name: "voucher20"},
	}
	active := []HotspotActiveSession{
		{RosID: "*A1", User: "voucher10"},
	}

	inactive := FilterInactiveHotspotUsers(users, active)
	require.Len(t, inactive, 1)
	assert.Equal(t, "voucher20", inactive[0].Name)
}

func TestIPPool(t *testing.T) {
	cmd := NewPrintIPPoolsCommand("hs-pool")
	assert.Equal(t, "/ip/pool/print", cmd.Raw)
	assert.Equal(t, "hs-pool", cmd.Args["?name"])

	result := command.Result{Rows: []map[string]string{
		{".id": "*1", "name": "hs-pool", "ranges": "192.168.1.10-192.168.1.200"},
	}}
	pools := ParseIPPools(result)
	require.Len(t, pools, 1)
	assert.Equal(t, "hs-pool", pools[0].Name)
	assert.Equal(t, "192.168.1.10-192.168.1.200", pools[0].Ranges)
}

func TestNewPrintParentQueuesCommand(t *testing.T) {
	cmd := NewPrintParentQueuesCommand()
	assert.Equal(t, "/queue/simple/print", cmd.Raw)
	assert.Equal(t, "false", cmd.Args["?dynamic"])
}


