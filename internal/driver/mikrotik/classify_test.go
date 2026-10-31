package mikrotik

import (
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestClassify_ReadAllowlist(t *testing.T) {
	read := []string{
		"/system/resource/print",
		"/interface/print",
		"/ip/address/print",
		"/ip/firewall/filter/print",
		"/ip/route/print",
		"/system/identity/get",
		"/export",
		"/ip/firewall/export",
		"/interface/monitor",
		"/interface/monitor-traffic",
		"/ping",
		"/tool/ping",
		"/tool/traceroute",
		"/interface/check",
		"/console/inspect",
		"/tool/sniff",
		"/tool/torch",
		"/ip/address/find",
		"/ip/neighbor/discover",
		"/tool/scan",
		"/tool/whois",
	}
	for _, raw := range read {
		t.Run(raw, func(t *testing.T) {
			got := Classify(command.Command{Raw: raw})
			if got.Mode != command.ModeRead || got.IsDenied() {
				t.Fatalf("%s: got %+v, want READ", raw, got)
			}
		})
	}
}

func TestClassify_MutationVerbs(t *testing.T) {
	mutation := []string{
		"/ip/firewall/filter/add",
		"/ip/firewall/filter/set",
		"/ip/firewall/filter/remove",
		"/ip/address/add",
		"/ip/address/remove",
		"/interface/enable",
		"/interface/disable",
		"/ip/route/set",
		"/system/identity/set",
		"/system/reboot",
		"/system/shutdown",
		"/system/backup/save",
		"/file/add",
		"/file/set",
		"/tool/fetch",
		"/system/script/run",
		"/system/scheduler/add",
		"/user/add",
		"/ppp/secret/add",
		"/queue/simple/add",
		"/ip/service/set",
		"/system/upgrade",
		"/system/downgrade",
		"/import",
		"/system/backup/load",
		"/system/reset-configuration",
		"/file/remove",
		"/system/script/add",
		"/system/scheduler/set",
	}
	for _, raw := range mutation {
		t.Run(raw, func(t *testing.T) {
			got := Classify(command.Command{Raw: raw})
			if got.Mode != command.ModeMutation {
				t.Fatalf("%s: got %+v, want MUTATION", raw, got)
			}
		})
	}
}

func TestClassify_HardDeny(t *testing.T) {
	denied := map[string]string{
		"/system/reset-configuration": "reset-configuration",
		"/system/backup/load":         "backup/load",
		"/file/remove":                "file/remove",
		"/tool/fetch":                 "tool/fetch",
		"/system/script/add":          "script/add",
		"/system/scheduler/add":       "scheduler/add",
	}
	for raw, reason := range denied {
		t.Run(raw, func(t *testing.T) {
			got := Classify(command.Command{Raw: raw})
			if !got.IsDenied() {
				t.Fatalf("%s: expected deny, got %+v", raw, got)
			}
			if got.DenyReason == "" || !contains(got.DenyReason, reason) {
				t.Fatalf("%s: deny reason %q missing %q", raw, got.DenyReason, reason)
			}
		})
	}
}

func TestClassify_FailClosed(t *testing.T) {
	cases := []command.Command{
		{Raw: ""},
		{Raw: "   "},
		{Raw: "/"},
		{Raw: "/some/unknown/path"},
		{Raw: "/ip/address/add", Args: map[string]string{"address": "1.1.1.1/24"}},
		{Raw: "/export file=foo"}, // raw with inline args is not a known verb
	}
	for _, cmd := range cases {
		got := Classify(cmd)
		if got.Mode != command.ModeMutation {
			t.Fatalf("%q: expected default MUTATION, got %+v", cmd.Raw, got)
		}
	}
}

func TestClassify_ReadVerbWithContentsIsMutation(t *testing.T) {
	got := Classify(command.Command{Raw: "/file/print", Args: map[string]string{"contents": "x"}})
	if got.Mode != command.ModeMutation {
		t.Fatalf("expected mutation when contents= is present, got %+v", got)
	}
}

func TestClassify_ExportWithFileArgIsStillRead(t *testing.T) {
	got := Classify(command.Command{Raw: "/export", Args: map[string]string{"file": "backup"}})
	if got.Mode != command.ModeRead {
		t.Fatalf("export file= should stay READ, got %+v", got)
	}
}

func TestClassify_DecideIntegration(t *testing.T) {
	if command.Decide(Classify(command.Command{Raw: "/system/resource/print"})) != command.DecisionAutoApprove {
		t.Fatal("read should auto-approve")
	}
	if command.Decide(Classify(command.Command{Raw: "/ip/firewall/filter/add"})) != command.DecisionRequireApproval {
		t.Fatal("mutation should require approval")
	}
	if command.Decide(Classify(command.Command{Raw: "/system/reset-configuration"})) != command.DecisionDeny {
		t.Fatal("hard-deny should deny")
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
