package mikrotik

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// ─── PPP Active builders ──────────────────────────────────────────────────

func TestNewPrintPPPActiveCommand(t *testing.T) {
	t.Run("tanpa filter", func(t *testing.T) {
		cmd := NewPrintPPPActiveCommand("")
		assert.Equal(t, "/ppp/active/print", cmd.Raw)
		assert.Empty(t, cmd.Args)
		assert.False(t, isStreamingCommand(cmd))
	})
	t.Run("dengan filter nama", func(t *testing.T) {
		cmd := NewPrintPPPActiveCommand("budi")
		assert.Equal(t, "budi", cmd.Args["?name"])
		assert.False(t, isStreamingCommand(cmd))
	})
}

func TestNewStreamPPPActiveCommand(t *testing.T) {
	t.Run("tanpa filter — streaming", func(t *testing.T) {
		cmd := NewStreamPPPActiveCommand("")
		assert.Equal(t, "/ppp/active/print", cmd.Raw)
		_, hasFollow := cmd.Args["follow"]
		assert.True(t, hasFollow, "harus ada flag follow")
		assert.True(t, isStreamingCommand(cmd), "harus terdeteksi sebagai streaming")
		assert.NotContains(t, cmd.Args, "?name")
	})
	t.Run("dengan filter nama", func(t *testing.T) {
		cmd := NewStreamPPPActiveCommand("andi")
		assert.Equal(t, "andi", cmd.Args["?name"])
		assert.True(t, isStreamingCommand(cmd))
	})
}

func TestNewKickPPPActiveCommand(t *testing.T) {
	cmd := NewKickPPPActiveCommand("*7")
	assert.Equal(t, "/ppp/active/remove", cmd.Raw)
	assert.Equal(t, "*7", cmd.Args[".id"])
}

// ─── PPP Active parsers ───────────────────────────────────────────────────

func TestParsePPPActiveSessions(t *testing.T) {
	t.Run("sesi lengkap", func(t *testing.T) {
		result := command.Result{Rows: []map[string]string{
			{
				".id": "*A1", "name": "budi", "service": "pppoe",
				"caller-id": "AA:BB:CC:DD:EE:FF", "address": "10.10.0.5",
				"uptime": "3h25m10s", "encoding": "", "session-id": "0x8001",
				"limit-bytes-in": "0", "limit-bytes-out": "0", "radius": "false",
			},
		}}
		sessions := ParsePPPActiveSessions(result)
		require.Len(t, sessions, 1)
		s := sessions[0]
		assert.Equal(t, "*A1", s.RosID)
		assert.Equal(t, "budi", s.Name)
		assert.Equal(t, "pppoe", s.Service)
		assert.Equal(t, "AA:BB:CC:DD:EE:FF", s.CallerID)
		assert.Equal(t, "10.10.0.5", s.Address)
		assert.False(t, s.Radius)
	})

	t.Run("radius true di-parse", func(t *testing.T) {
		result := command.Result{Rows: []map[string]string{
			{".id": "*1", "name": "x", "radius": "true"},
		}}
		assert.True(t, ParsePPPActiveSessions(result)[0].Radius)
	})

	t.Run("parse ppp active stats", func(t *testing.T) {
		result := command.Result{Rows: []map[string]string{
			{
				".id": "*A1", "uptime": "3h25m10s",
				"limit-bytes-in": "1000", "limit-bytes-out": "2000",
			},
		}}
		stats := ParsePPPActiveStats(result)
		require.Len(t, stats, 1)
		st := stats[0]
		assert.Equal(t, "*A1", st.RosID)
		assert.Equal(t, "3h25m10s", st.Uptime)
		assert.Equal(t, "1000", st.LimitBytesIn)
		assert.Equal(t, "2000", st.LimitBytesOut)
	})

	t.Run("baris tanpa .id dilewati", func(t *testing.T) {
		result := command.Result{Rows: []map[string]string{
			{"name": "orphan"},
			{".id": "*2", "name": "valid"},
		}}
		sessions := ParsePPPActiveSessions(result)
		require.Len(t, sessions, 1)
		assert.Equal(t, "valid", sessions[0].Name)
	})

	t.Run("result kosong", func(t *testing.T) {
		assert.Empty(t, ParsePPPActiveSessions(command.Result{}))
	})
}

// ─── PPP Profile builders ─────────────────────────────────────────────────

func TestNewAddPPPProfileCommand(t *testing.T) {
	t.Run("name wajib ada", func(t *testing.T) {
		cmd := NewAddPPPProfileCommand(PPPProfileParams{Name: "10Mbps", RateLimit: "10M/10M"})
		assert.Equal(t, "/ppp/profile/add", cmd.Raw)
		assert.Equal(t, "10Mbps", cmd.Args["name"])
		assert.Equal(t, "10M/10M", cmd.Args["rate-limit"])
		assert.NotContains(t, cmd.Args, "local-address")
	})

	t.Run("IsolirProfileParams menghasilkan param yang benar", func(t *testing.T) {
		p := IsolirProfileParams("pool-isolir")
		assert.Equal(t, "isolir", p.Name)
		assert.Equal(t, "pool-isolir", p.RemoteAddress)
		assert.Equal(t, "512k/512k", p.RateLimit)
		assert.Equal(t, "SUSPENDED_PROFILE", p.Comment)
		assert.Empty(t, p.SharedUsers)

		cmd := NewAddPPPProfileCommand(p)
		assert.Equal(t, "/ppp/profile/add", cmd.Raw)
		assert.Equal(t, "512k/512k", cmd.Args["rate-limit"])
	})
}

func TestNewSetPPPProfileCommand(t *testing.T) {
	cmd := NewSetPPPProfileCommand("*3", PPPProfileParams{RateLimit: "20M/20M"})
	assert.Equal(t, "/ppp/profile/set", cmd.Raw)
	assert.Equal(t, "*3", cmd.Args[".id"])
	assert.Equal(t, "20M/20M", cmd.Args["rate-limit"])
	assert.NotContains(t, cmd.Args, "name")
}

func TestNewRemovePPPProfileCommand(t *testing.T) {
	cmd := NewRemovePPPProfileCommand("*2")
	assert.Equal(t, "/ppp/profile/remove", cmd.Raw)
	assert.Equal(t, "*2", cmd.Args[".id"])
}

func TestParsePPPProfiles(t *testing.T) {
	result := command.Result{Rows: []map[string]string{
		{".id": "*1", "name": "10Mbps", "rate-limit": "10M/10M", "dns-server": "8.8.8.8"},
		{".id": "", "name": "bad"},
	}}
	profiles := ParsePPPProfiles(result)
	require.Len(t, profiles, 1)
	p := profiles[0]
	assert.Equal(t, "*1", p.RosID)
	assert.Equal(t, "10Mbps", p.Name)
	assert.Equal(t, "10M/10M", p.RateLimit)
	assert.Equal(t, "8.8.8.8", p.DNSServer)
}

func TestFilterInactivePPPoESecrets(t *testing.T) {
	secrets := []PPPoESecret{
		{RosID: "*1", Name: "budi"},
		{RosID: "*2", Name: "andi"},
		{RosID: "*3", Name: "siti"},
	}
	active := []PPPActiveSession{
		{RosID: "*A1", Name: "budi"},
	}

	inactive := FilterInactivePPPoESecrets(secrets, active)
	require.Len(t, inactive, 2, "andi dan siti harus terdeteksi non-aktif (offline)")
	names := []string{inactive[0].Name, inactive[1].Name}
	assert.Contains(t, names, "andi")
	assert.Contains(t, names, "siti")
	assert.NotContains(t, names, "budi")
}

func TestEnrichPPPActiveSessionsWithProfiles(t *testing.T) {
	secrets := []PPPoESecret{
		{RosID: "*1", Name: "budi", Profile: "10Mbps_Plan"},
		{RosID: "*2", Name: "andi", Profile: "20Mbps_Plan"},
		{RosID: "*3", Name: "siti", Profile: ""},
	}
	active := []PPPActiveSession{
		{RosID: "*A1", Name: "budi", Profile: ""},
		{RosID: "*A2", Name: "andi", Profile: "Custom_Plan"},
		{RosID: "*A3", Name: "unknown", Profile: ""},
	}

	enriched := EnrichPPPActiveSessionsWithProfiles(active, secrets)
	require.Len(t, enriched, 3)
	assert.Equal(t, "10Mbps_Plan", enriched[0].Profile, "budi harus diisi profil dari secrets")
	assert.Equal(t, "Custom_Plan", enriched[1].Profile, "andi tetap memakai profil yang sudah ada")
	assert.Equal(t, "", enriched[2].Profile, "unknown tanpa secret tetap kosong")
}
