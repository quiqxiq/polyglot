package mikrotik

import (
	"errors"
	"testing"

	"github.com/go-routeros/routeros/v3"
	"github.com/go-routeros/routeros/v3/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/provision"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name string
		cmd  command.Command
		want command.Class
	}{
		{"reboot is destructive", command.Command{Raw: "/system/reboot"}, command.ClassDestructive},
		{"reset-configuration is destructive", command.Command{Raw: "/system/reset-configuration"}, command.ClassDestructive},
		{"ppp secret add is destructive", command.Command{Raw: "/ppp/secret/add"}, command.ClassDestructive},
		{"ppp secret remove is destructive", command.Command{Raw: "/ppp/secret/remove"}, command.ClassDestructive},
		{"ppp profile set is destructive", command.Command{Raw: "/ppp/profile/set"}, command.ClassDestructive},
		{"ppp active remove is destructive", command.Command{Raw: "/ppp/active/remove"}, command.ClassDestructive},
		{"resource print is read-only", command.Command{Raw: "/system/resource/print"}, command.ClassReadOnly},
		{"ppp secret print is read-only", command.Command{Raw: "/ppp/secret/print"}, command.ClassReadOnly},
		{"ping is read-only", command.Command{Raw: "/ping"}, command.ClassReadOnly},
		{"monitor-traffic is read-only", command.Command{Raw: "/interface/monitor-traffic"}, command.ClassReadOnly},
		{"unknown path defaults destructive (fail-safe)", command.Command{Raw: "/some/unknown/path"}, command.ClassDestructive},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Classify(tt.cmd))
		})
	}
}

func TestTranslate(t *testing.T) {
	tests := []struct {
		name    string
		op      command.Operation
		want    command.Command
		wantErr bool
	}{
		{"get_status", command.OpGetStatus, command.Command{Raw: "/system/resource/print"}, false},
		{"reboot", command.OpReboot, command.Command{Raw: "/system/reboot"}, false},
		{"unsupported op", command.Operation("does_not_exist"), command.Command{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Translate(tt.op)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTranslateProvision(t *testing.T) {
	tests := []struct {
		name         string
		op           provision.Operation
		wantRaw      string
		wantProplist string // "" means: .proplist key must be absent
	}{
		{"list secrets, no projection", provision.ListPPPSecrets{}, "/ppp/secret/print", ""},
		{"list secrets, projection", provision.ListPPPSecrets{Fields: []string{"name", "profile"}}, "/ppp/secret/print", "name,profile"},
		{"list profiles, no projection", provision.ListPPPProfiles{}, "/ppp/profile/print", ""},
		{"list profiles, projection", provision.ListPPPProfiles{Fields: []string{"name", "rate-limit"}}, "/ppp/profile/print", "name,rate-limit"},
		{"list active, no projection", provision.ListActivePPP{}, "/ppp/active/print", ""},
		{"list active, projection", provision.ListActivePPP{Fields: []string{"name", "address"}}, "/ppp/active/print", "name,address"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmds, err := translateProvision(tt.op)
			require.NoError(t, err)
			require.Len(t, cmds, 1)
			assert.Equal(t, tt.wantRaw, cmds[0].Raw)
			got, has := cmds[0].Args[".proplist"]
			if tt.wantProplist == "" {
				assert.False(t, has, "tanpa Fields, .proplist tidak boleh ada (device balikin semua field raw)")
				return
			}
			assert.Equal(t, tt.wantProplist, got)
		})
	}

	t.Run("unsupported operation errors", func(t *testing.T) {
		// A nil Operation matches no case and must fail closed via default.
		// (The sealed interface makes any other unknown op unconstructable
		// from outside the provision package, which is the point.)
		_, err := translateProvision(nil)
		require.Error(t, err)
	})
}

func TestTranslateProvision_CreatePPPSecret(t *testing.T) {
	t.Run("required fields only, optionals omitted", func(t *testing.T) {
		cmds, err := translateProvision(provision.CreatePPPSecret{Name: "tes", Password: "pw"})
		require.NoError(t, err)
		require.Len(t, cmds, 1)
		assert.Equal(t, "/ppp/secret/add", cmds[0].Raw)
		assert.Equal(t, map[string]string{"name": "tes", "password": "pw"}, cmds[0].Args)
	})

	t.Run("optionals included only when non-empty", func(t *testing.T) {
		cmds, err := translateProvision(provision.CreatePPPSecret{
			Name:          "tes",
			Password:      "pw",
			Profile:       "home-10mbps",
			Service:       "pppoe",
			RemoteAddress: "10.0.0.5",
			Comment:       "sub#1",
		})
		require.NoError(t, err)
		require.Len(t, cmds, 1)
		assert.Equal(t, map[string]string{
			"name":           "tes",
			"password":       "pw",
			"profile":        "home-10mbps",
			"service":        "pppoe",
			"remote-address": "10.0.0.5",
			"comment":        "sub#1",
		}, cmds[0].Args)
	})

	t.Run("missing name errors", func(t *testing.T) {
		_, err := translateProvision(provision.CreatePPPSecret{Password: "pw"})
		require.ErrorIs(t, err, errMissingField)
	})

	t.Run("missing password errors", func(t *testing.T) {
		_, err := translateProvision(provision.CreatePPPSecret{Name: "tes"})
		require.ErrorIs(t, err, errMissingField)
	})
}

func TestTranslateProvision_CreatePPPProfile(t *testing.T) {
	t.Run("required name only, optionals omitted", func(t *testing.T) {
		cmds, err := translateProvision(provision.CreatePPPProfile{Name: "home-10mbps"})
		require.NoError(t, err)
		require.Len(t, cmds, 1)
		assert.Equal(t, "/ppp/profile/add", cmds[0].Raw)
		assert.Equal(t, map[string]string{"name": "home-10mbps"}, cmds[0].Args)
	})

	t.Run("optionals included only when non-empty", func(t *testing.T) {
		cmds, err := translateProvision(provision.CreatePPPProfile{
			Name:          "home-10mbps",
			RateLimit:     "10M/10M",
			LocalAddress:  "10.0.0.1",
			RemoteAddress: "pool-pppoe",
			Comment:       "paket 10 mbps",
		})
		require.NoError(t, err)
		require.Len(t, cmds, 1)
		assert.Equal(t, map[string]string{
			"name":           "home-10mbps",
			"rate-limit":     "10M/10M",
			"local-address":  "10.0.0.1",
			"remote-address": "pool-pppoe",
			"comment":        "paket 10 mbps",
		}, cmds[0].Args)
	})

	t.Run("missing name errors", func(t *testing.T) {
		_, err := translateProvision(provision.CreatePPPProfile{RateLimit: "10M/10M"})
		require.ErrorIs(t, err, errMissingField)
	})
}

func TestTranslateProvision_ChangeProfile(t *testing.T) {
	t.Run("emits set-then-kill sequence targeting the subscriber by name", func(t *testing.T) {
		cmds, err := translateProvision(provision.ChangeProfile{Username: "budi", Profile: "home-20mbps"})
		require.NoError(t, err)
		require.Len(t, cmds, 2, "change profile harus sekuens: set profile lalu putus sesi aktif")

		assert.Equal(t, "/ppp/secret/set", cmds[0].Raw)
		assert.Equal(t, map[string]string{"numbers": "budi", "profile": "home-20mbps"}, cmds[0].Args)

		assert.Equal(t, activePPPRemovePath, cmds[1].Raw)
		assert.Equal(t, map[string]string{"numbers": "budi"}, cmds[1].Args)
	})

	t.Run("missing username errors", func(t *testing.T) {
		_, err := translateProvision(provision.ChangeProfile{Profile: "home-20mbps"})
		require.ErrorIs(t, err, errMissingField)
	})

	t.Run("missing profile errors", func(t *testing.T) {
		_, err := translateProvision(provision.ChangeProfile{Username: "budi"})
		require.ErrorIs(t, err, errMissingField)
	})
}

func TestIsNoSuchItem(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error is not no-such-item", nil, false},
		{"routeros no such item trap", errors.New("from RouterOS device: no such item (4)"), true},
		{"unrelated device error", errors.New("from RouterOS device: unknown parameter"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isNoSuchItem(tt.err))
		})
	}
}

func TestIsStreamingCommand(t *testing.T) {
	tests := []struct {
		name string
		cmd  command.Command
		want bool
	}{
		{"ping is streaming", command.Command{Raw: "/ping", Args: map[string]string{"address": "10.0.0.1"}}, true},
		{"monitor-traffic is streaming", command.Command{Raw: "/interface/monitor-traffic", Args: map[string]string{"interface": "ether1"}}, true},
		{"plain print is not streaming", command.Command{Raw: "/interface/print"}, false},
		{"print with follow is streaming", command.Command{Raw: "/interface/print", Args: map[string]string{"follow": ""}}, true},
		{"print with follow-only is streaming", command.Command{Raw: "/log/print", Args: map[string]string{"follow-only": ""}}, true},
		{"print with interval is streaming", command.Command{Raw: "/interface/print", Args: map[string]string{"interval": "1s"}}, true},
		{"resource print is not streaming", command.Command{Raw: "/system/resource/print"}, false},
		{"unrelated args do not trigger streaming", command.Command{Raw: "/interface/print", Args: map[string]string{"disabled": "no"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isStreamingCommand(tt.cmd))
		})
	}
}

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name string
		cmd  command.Command
		want []string
	}{
		{
			name: "no args",
			cmd:  command.Command{Raw: "/system/resource/print"},
			want: []string{"/system/resource/print"},
		},
		{
			name: "attribute with value",
			cmd:  command.Command{Raw: "/ping", Args: map[string]string{"address": "10.0.0.1"}},
			want: []string{"/ping", "=address=10.0.0.1"},
		},
		{
			name: "bare follow flag ignores any value",
			cmd:  command.Command{Raw: "/interface/print", Args: map[string]string{"follow": ""}},
			want: []string{"/interface/print", "follow"},
		},
		{
			name: "bare follow-only flag",
			cmd:  command.Command{Raw: "/log/print", Args: map[string]string{"follow-only": ""}},
			want: []string{"/log/print", "follow-only"},
		},
		{
			name: "attribute with empty value becomes bare",
			cmd:  command.Command{Raw: "/interface/print", Args: map[string]string{"disabled": ""}},
			want: []string{"/interface/print", "=disabled"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildArgs(tt.cmd)
			// Args come from a map, so order beyond the first (Raw) element
			// isn't guaranteed — compare as sets for cmds with >1 arg word.
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestToResult(t *testing.T) {
	t.Run("nil reply", func(t *testing.T) {
		got := toResult(nil)
		assert.Empty(t, got.Rows)
	})

	t.Run("single row", func(t *testing.T) {
		reply := &routeros.Reply{
			Re: []*proto.Sentence{
				{Map: map[string]string{"uptime": "1h2m3s", "version": "7.15"}},
			},
		}
		got := toResult(reply)
		require.Len(t, got.Rows, 1)
		assert.Equal(t, "1h2m3s", got.Rows[0]["uptime"])
	})

	t.Run("multiple rows are all preserved", func(t *testing.T) {
		reply := &routeros.Reply{
			Re: []*proto.Sentence{
				{Map: map[string]string{"name": "ether1"}},
				{Map: map[string]string{"name": "ether2"}},
				{Map: map[string]string{"name": "ether3"}},
			},
		}
		got := toResult(reply)
		require.Len(t, got.Rows, 3)
		assert.Equal(t, "ether2", got.Rows[1]["name"])
	})
}
