package mikrotik

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// ─── PPPoE Secret builders ────────────────────────────────────────────────

func TestNewAddPPPoESecretCommand(t *testing.T) {
	t.Run("required fields only — defaults applied", func(t *testing.T) {
		cmd := NewAddPPPoESecretCommand(PPPoESecretParams{
			Name:     "budi",
			Password: "pass123",
			Profile:  "10Mbps",
		})
		assert.Equal(t, "/ppp/secret/add", cmd.Raw)
		assert.Equal(t, "budi", cmd.Args["name"])
		assert.Equal(t, "pass123", cmd.Args["password"])
		assert.Equal(t, "10Mbps", cmd.Args["profile"])
		assert.Equal(t, "pppoe", cmd.Args["service"], "service harus default pppoe")
		assert.NotContains(t, cmd.Args, "local-address", "field opsional tidak dikirim jika kosong")
		assert.NotContains(t, cmd.Args, "remote-address")
		assert.NotContains(t, cmd.Args, "disabled")
	})

	t.Run("semua field diisi", func(t *testing.T) {
		cmd := NewAddPPPoESecretCommand(PPPoESecretParams{
			Name:          "andi",
			Password:      "secret",
			Profile:       "20Mbps",
			Service:       "any",
			LocalAddress:  "10.0.0.1",
			RemoteAddress: "pool-dynamic",
			Comment:       "polyglot:sub-001",
			Disabled:      true,
		})
		assert.Equal(t, "any", cmd.Args["service"])
		assert.Equal(t, "10.0.0.1", cmd.Args["local-address"])
		assert.Equal(t, "pool-dynamic", cmd.Args["remote-address"])
		assert.Equal(t, "polyglot:sub-001", cmd.Args["comment"])
		assert.Equal(t, "yes", cmd.Args["disabled"])
	})

	t.Run("profile kosong menggunakan default", func(t *testing.T) {
		cmd := NewAddPPPoESecretCommand(PPPoESecretParams{Name: "x", Password: "y"})
		assert.Equal(t, "default", cmd.Args["profile"])
	})

	t.Run("command tidak dianggap streaming", func(t *testing.T) {
		cmd := NewAddPPPoESecretCommand(PPPoESecretParams{Name: "x", Password: "y"})
		assert.False(t, isStreamingCommand(cmd))
	})
}

func TestNewSetPPPoESecretCommand(t *testing.T) {
	cmd := NewSetPPPoESecretCommand("*3", PPPoESecretParams{
		Password: "newpass",
		Profile:  "5Mbps",
	})
	assert.Equal(t, "/ppp/secret/set", cmd.Raw)
	assert.Equal(t, "*3", cmd.Args["numbers"])
	assert.Equal(t, "newpass", cmd.Args["password"])
	assert.Equal(t, "5Mbps", cmd.Args["profile"])
	assert.NotContains(t, cmd.Args, "service", "field kosong tidak dikirim")
}

func TestNewRemovePPPoESecretCommand(t *testing.T) {
	cmd := NewRemovePPPoESecretCommand("*5")
	assert.Equal(t, "/ppp/secret/remove", cmd.Raw)
	assert.Equal(t, "*5", cmd.Args["numbers"])
}

func TestNewPrintPPPoESecretsCommand(t *testing.T) {
	t.Run("tanpa filter", func(t *testing.T) {
		cmd := NewPrintPPPoESecretsCommand("")
		assert.Equal(t, "/ppp/secret/print", cmd.Raw)
		assert.Empty(t, cmd.Args)
	})
	t.Run("dengan filter nama", func(t *testing.T) {
		cmd := NewPrintPPPoESecretsCommand("budi")
		assert.Equal(t, "budi", cmd.Args["?name"])
	})
	t.Run("tidak streaming", func(t *testing.T) {
		assert.False(t, isStreamingCommand(NewPrintPPPoESecretsCommand("")))
	})
}

// ─── PPPoE Secret parsers ─────────────────────────────────────────────────

func TestParsePPPoESecrets(t *testing.T) {
	t.Run("baris lengkap", func(t *testing.T) {
		result := command.Result{Rows: []map[string]string{
			{
				".id": "*1", "name": "budi", "password": "pass",
				"profile": "10Mbps", "service": "pppoe",
				"local-address": "10.0.0.1", "remote-address": "pool-1",
				"comment": "polyglot:sub-001", "disabled": "false",
				"last-logged-out": "jan/01/2025 10:00:00", "caller-id": "AA:BB:CC:DD:EE:FF",
			},
		}}
		secrets := ParsePPPoESecrets(result)
		require.Len(t, secrets, 1)
		s := secrets[0]
		assert.Equal(t, "*1", s.RosID)
		assert.Equal(t, "budi", s.Name)
		assert.Equal(t, "10Mbps", s.Profile)
		assert.Equal(t, "pppoe", s.Service)
		assert.Equal(t, "10.0.0.1", s.LocalAddress)
		assert.Equal(t, "pool-1", s.RemoteAddress)
		assert.Equal(t, "polyglot:sub-001", s.Comment)
		assert.False(t, s.Disabled)
		assert.Equal(t, "jan/01/2025 10:00:00", s.LastLoggedOut)
		assert.Equal(t, "AA:BB:CC:DD:EE:FF", s.CallerID)
	})

	t.Run("baris tanpa .id dilewati", func(t *testing.T) {
		result := command.Result{Rows: []map[string]string{
			{"name": "budi"},        // tidak ada .id
			{".id": "*2", "name": "andi"}, // valid
		}}
		secrets := ParsePPPoESecrets(result)
		require.Len(t, secrets, 1)
		assert.Equal(t, "andi", secrets[0].Name)
	})

	t.Run("baris tanpa name dilewati", func(t *testing.T) {
		result := command.Result{Rows: []map[string]string{
			{".id": "*1"}, // tidak ada name
		}}
		assert.Empty(t, ParsePPPoESecrets(result))
	})

	t.Run("disabled true di-parse dengan benar", func(t *testing.T) {
		result := command.Result{Rows: []map[string]string{
			{".id": "*1", "name": "x", "disabled": "true"},
		}}
		assert.True(t, ParsePPPoESecrets(result)[0].Disabled)
	})

	t.Run("result kosong tidak panik", func(t *testing.T) {
		assert.Empty(t, ParsePPPoESecrets(command.Result{}))
	})
}

func TestFindPPPoESecretRosID(t *testing.T) {
	result := command.Result{Rows: []map[string]string{
		{".id": "*1", "name": "budi"},
		{".id": "*2", "name": "andi"},
	}}

	t.Run("ditemukan", func(t *testing.T) {
		id, err := FindPPPoESecretRosID(result, "andi")
		require.NoError(t, err)
		assert.Equal(t, "*2", id)
	})

	t.Run("tidak ditemukan", func(t *testing.T) {
		_, err := FindPPPoESecretRosID(result, "siti")
		assert.ErrorIs(t, err, ErrSecretNotFound)
	})
}
