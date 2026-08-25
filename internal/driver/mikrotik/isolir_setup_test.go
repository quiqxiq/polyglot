package mikrotik

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// cannedExec returns preset rows per command path for inspection tests.
func cannedExec(rows map[string][]map[string]string) port.CommandExecutor {
	return func(ctx context.Context, d port.DeviceDriver, cmd command.Command) (command.Result, error) {
		r, ok := rows[cmd.Raw]
		if !ok {
			return command.Result{Rows: nil}, nil
		}
		return command.Result{Rows: r}, nil
	}
}

func TestGateway_InspectIsolirInfrastructure(t *testing.T) {
	cfg := port.IsolirConfig{
		ProfileName: "isolir", PoolName: "pool-isolir",
		PoolRange: "172.16.99.10-172.16.99.254",
		PortalIP:  "192.0.2.10", PortalHTTPPort: "8080",
		RedirectPorts: "80,443",
	}

	t.Run("lengkap semua hadir", func(t *testing.T) {
		g := &Gateway{exec: cannedExec(map[string][]map[string]string{
			"/ip/pool/print": {
				{".id": "*1", "name": "pool-isolir", "ranges": "172.16.99.10-172.16.99.254"},
			},
			"/ppp/profile/print": {
				{".id": "*2", "name": "isolir", "rate-limit": "512k/512k", "remote-address": "pool-isolir"},
			},
			"/ip/firewall/nat/print": {
				{".id": "*A", "chain": "dstnat", "action": "dst-nat", "dst-port": "80", "to-addresses": "192.0.2.10", "to-ports": "8080", "comment": "ISOLIR_REDIRECT 80"},
				{".id": "*B", "chain": "dstnat", "action": "dst-nat", "dst-port": "443", "to-addresses": "192.0.2.10", "to-ports": "8080", "comment": "ISOLIR_REDIRECT 443"},
				{".id": "*C", "chain": "srcnat", "action": "masquerade", "comment": "NAT-LAIN"},
			},
		})}

		ins, err := g.InspectIsolirInfrastructure(context.Background(), nil, cfg)
		require.NoError(t, err)
		assert.True(t, ins.PoolExists)
		assert.Equal(t, "172.16.99.10-172.16.99.254", ins.PoolRanges)
		assert.True(t, ins.ProfileExists)
		assert.Equal(t, "512k/512k", ins.ProfileRateLimit)
		require.Len(t, ins.NATRules, 2)
		assert.True(t, ins.NATRules[0].Exists)
		assert.Equal(t, "*A", ins.NATRules[0].RosID)
		assert.Equal(t, "dst-nat", ins.NATRules[0].Action)
		assert.Equal(t, "192.0.2.10", ins.NATRules[0].ToAddresses)
		assert.True(t, ins.NATRules[1].Exists)
	})

	t.Run("kosong semua", func(t *testing.T) {
		g := &Gateway{exec: cannedExec(map[string][]map[string]string{})}
		ins, err := g.InspectIsolirInfrastructure(context.Background(), nil, cfg)
		require.NoError(t, err)
		assert.False(t, ins.PoolExists)
		assert.False(t, ins.ProfileExists)
		require.Len(t, ins.NATRules, 2)
		assert.False(t, ins.NATRules[0].Exists)
		assert.False(t, ins.NATRules[1].Exists)
		assert.Equal(t, "80", ins.NATRules[0].Port)
		assert.Equal(t, "443", ins.NATRules[1].Port)
	})

	t.Run("rule separuh hadir", func(t *testing.T) {
		g := &Gateway{exec: cannedExec(map[string][]map[string]string{
			"/ip/firewall/nat/print": {
				{".id": "*A", "chain": "dstnat", "action": "redirect", "dst-port": "80", "to-ports": "8080", "comment": "ISOLIR_REDIRECT 80"},
			},
		})}
		ins, err := g.InspectIsolirInfrastructure(context.Background(), nil, cfg)
		require.NoError(t, err)
		assert.True(t, ins.NATRules[0].Exists)
		assert.False(t, ins.NATRules[1].Exists)
		assert.Equal(t, "redirect", ins.NATRules[0].Action)
	})

	t.Run("config tidak valid ditolak", func(t *testing.T) {
		g := &Gateway{exec: cannedExec(map[string][]map[string]string{})}
		_, err := g.InspectIsolirInfrastructure(context.Background(), nil, port.IsolirConfig{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "profile name is required")
	})
}

func TestGateway_RemoveIsolirInfrastructure_ScopedByPorts(t *testing.T) {
	var removed []string
	g := &Gateway{exec: func(ctx context.Context, d port.DeviceDriver, cmd command.Command) (command.Result, error) {
		switch cmd.Raw {
		case "/ip/firewall/nat/print":
			return command.Result{Rows: []map[string]string{
				{".id": "*A", "chain": "dstnat", "action": "dst-nat", "comment": "ISOLIR_REDIRECT 8094"},
				{".id": "*B", "chain": "dstnat", "action": "dst-nat", "comment": "ISOLIR_REDIRECT 80"},
			}}, nil
		case "/ip/firewall/nat/remove":
			removed = append(removed, cmd.Args[".id"])
		}
		return command.Result{}, nil
	}}

	cfg := port.IsolirConfig{
		ProfileName: "probe-iso", PoolName: "probe-pool",
		PoolRange:      "172.16.97.1-172.16.97.9",
		PortalHTTPPort: "8099", RedirectPorts: "8094",
	}
	require.NoError(t, g.RemoveIsolirInfrastructure(context.Background(), nil, cfg))
	assert.Equal(t, []string{"*A"}, removed, "hanya rule milik konfigurasi ini yang dihapus; rule port 80 produksi dibiarkan")
}
