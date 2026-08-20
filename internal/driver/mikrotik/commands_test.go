package mikrotik

import (
	"testing"

	"github.com/quiqxiq/goros/v4"
	"github.com/quiqxiq/goros/v4/proto"
	"github.com/quiqxiq/goros/v4/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name string
		cmd  command.Command
		want command.Class
	}{
		{"reboot is destructive", command.Command{Raw: "/system/reboot"}, command.ClassDestructive},
		{"reset-configuration is destructive", command.Command{Raw: "/system/reset-configuration"}, command.ClassDestructive},
		{"resource print is read-only", command.Command{Raw: "/system/resource/print"}, command.ClassReadOnly},
		{"unknown path defaults read-only", command.Command{Raw: "/some/unknown/path"}, command.ClassReadOnly},
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

func TestIsStreamingCommand(t *testing.T) {
	tests := []struct {
		name string
		cmd  command.Command
		want bool
	}{
		{"ping is streaming", command.Command{Raw: "/ping", Args: map[string]string{"address": "10.0.0.1"}}, true},
		{"monitor-traffic without once is streaming", command.Command{Raw: "/interface/monitor-traffic", Args: map[string]string{"interface": "ether1"}}, true},
		{"monitor-traffic with once is NOT streaming", command.Command{Raw: "/interface/monitor-traffic", Args: map[string]string{"interface": "ether1", "once": ""}}, false},
		{"plain print is not streaming", command.Command{Raw: "/interface/print"}, false},
		{"print with follow is streaming", command.Command{Raw: "/interface/print", Args: map[string]string{"follow": ""}}, true},
		{"print with follow-only is streaming", command.Command{Raw: "/log/print", Args: map[string]string{"follow-only": ""}}, true},
		{"print with interval is streaming", command.Command{Raw: "/interface/print", Args: map[string]string{"interval": "1s"}}, true},
		{"resource print is not streaming", command.Command{Raw: "/system/resource/print"}, false},
		{"unrelated args do not trigger streaming", command.Command{Raw: "/interface/print", Args: map[string]string{"disabled": "no"}}, false},
		{"scheduler add with interval is NOT streaming", command.Command{Raw: "/system/scheduler/add", Args: map[string]string{"name": "test", "interval": "00:01:00"}}, false},
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
		{
			name: "query word starting with ?",
			cmd:  command.Command{Raw: "/ppp/secret/print", Args: map[string]string{"?name": "budi"}},
			want: []string{"/ppp/secret/print", "?name=budi"},
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

func TestToTransportCommand(t *testing.T) {
	tests := []struct {
		name string
		cmd  command.Command
		want *transport.Command
	}{
		{
			name: "no args",
			cmd:  command.Command{Raw: "/system/resource/print"},
			want: &transport.Command{Path: "/system/resource", Verb: "print", Attributes: map[string]string{}},
		},
		{
			name: "attribute with value",
			cmd:  command.Command{Raw: "/ip/address/add", Args: map[string]string{"address": "10.0.0.1/24"}},
			want: &transport.Command{Path: "/ip/address", Verb: "add", Attributes: map[string]string{"address": "10.0.0.1/24"}},
		},
		{
			name: "root-level command",
			cmd:  command.Command{Raw: "/ping", Args: map[string]string{"address": "10.0.0.1"}},
			want: &transport.Command{Path: "", Verb: "ping", Attributes: map[string]string{"address": "10.0.0.1"}},
		},
		{
			name: "query word with value",
			cmd:  command.Command{Raw: "/ppp/secret/print", Args: map[string]string{"?name": "budi"}},
			want: &transport.Command{Path: "/ppp/secret", Verb: "print", Queries: []string{"name=budi"}, Attributes: map[string]string{}},
		},
		{
			name: "bare query word",
			cmd:  command.Command{Raw: "/interface/print", Args: map[string]string{"?disabled": ""}},
			want: &transport.Command{Path: "/interface", Verb: "print", Queries: []string{"disabled"}, Attributes: map[string]string{}},
		},
		{
			name: "proplist projection",
			cmd:  command.Command{Raw: "/interface/print", Args: map[string]string{".proplist": "name,rx-byte"}},
			want: &transport.Command{Path: "/interface", Verb: "print", Proplist: []string{"name", "rx-byte"}, Attributes: map[string]string{}},
		},
		{
			name: "mixed query and attribute",
			cmd:  command.Command{Raw: "/ip/hotspot/user/print", Args: map[string]string{"?profile": "vip", "disabled": "no"}},
			want: &transport.Command{Path: "/ip/hotspot/user", Verb: "print", Queries: []string{"profile=vip"}, Attributes: map[string]string{"disabled": "no"}},
		},
		{
			name: "command without leading slash",
			cmd:  command.Command{Raw: "system/resource/print"},
			want: &transport.Command{Path: "system/resource", Verb: "print", Attributes: map[string]string{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toTransportCommand(tt.cmd)
			assert.Equal(t, tt.want, got)
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
