package mikrotik

import (
	"fmt"
	"strings"

	"github.com/go-routeros/routeros/v3"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/monitor"
	"github.com/quixiq/polyglot/internal/domain/provision"
)

// operationMap translates abstract Operations to RouterOS API paths.
// TODO: add more operations as new usecases need them.
var operationMap = map[command.Operation]command.Command{
	command.OpGetStatus: {Raw: "/system/resource/print"},
	command.OpReboot:    {Raw: "/system/reboot"},
}

// Classify reports the risk class of cmd according to RouterOS API conventions,
// FAIL-SAFE: a command is ClassReadOnly only if it matches a known read pattern;
// everything else — including any unrecognized path — is ClassDestructive and
// therefore needs approval. This is a whitelist of safe reads, not a blacklist
// of dangerous writes, so a new write op (e.g. /ppp/secret/add) is destructive
// by default without anyone remembering to list it. See
// docs/adr/0006-klasifikasi-perintah-fail-safe.md.
func Classify(cmd command.Command) command.Class {
	if isReadOnlyCommand(cmd) {
		return command.ClassReadOnly
	}
	return command.ClassDestructive
}

// isReadOnlyCommand reports whether cmd only observes device state. RouterOS
// read commands all end in "/print"; /ping and /interface/monitor-traffic
// observe without mutating (they are the streamingBasePaths). Anything else is
// treated as state-changing by Classify.
func isReadOnlyCommand(cmd command.Command) bool {
	if strings.HasSuffix(cmd.Raw, "/print") {
		return true
	}
	return streamingBasePaths[cmd.Raw]
}

// Translate maps an abstract Operation to a RouterOS-native Command.
func Translate(op command.Operation) (command.Command, error) {
	cmd, ok := operationMap[op]
	if !ok {
		return command.Command{}, fmt.Errorf("mikrotik: unsupported operation %q", op)
	}
	return cmd, nil
}

// translateProvision maps an abstract provisioning operation to the RouterOS
// command sequence that fulfills it. Knowledge of RouterOS paths and field
// names (e.g. /ppp/secret/print, .proplist) lives here per AGENTS.md §1.2,
// never in usecase/. It returns a slice because a single provisioning
// operation may need several ordered commands (see port.ProvisioningDriver);
// read operations like ListPPPSecrets are simply a one-element sequence.
func translateProvision(op provision.Operation) ([]command.Command, error) {
	switch o := op.(type) {
	case provision.ListPPPSecrets:
		return []command.Command{printCmd("/ppp/secret/print", o.Fields)}, nil
	case provision.ListPPPProfiles:
		return []command.Command{printCmd("/ppp/profile/print", o.Fields)}, nil
	case provision.ListActivePPP:
		return []command.Command{printCmd("/ppp/active/print", o.Fields)}, nil
	case provision.CreatePPPSecret:
		return createPPPSecret(o)
	case provision.CreatePPPProfile:
		return createPPPProfile(o)
	case provision.ChangeProfile:
		return changeProfile(o)
	default:
		return nil, fmt.Errorf("mikrotik: unsupported provisioning operation %T", op)
	}
}

// translateStream maps an abstract monitoring operation to the single RouterOS
// streaming command that fulfills it. Knowledge of RouterOS paths, field names,
// and streaming flags (follow/follow-only, stats, the ?name query word) lives
// here per AGENTS.md §1.2, never in usecase/. It returns a single command
// because a streaming observation is always one long-running command (see
// port.StreamingDeviceDriver.TranslateStream). Required fields are validated
// here with a clear error before touching the wire rather than letting the
// device reject the sentence.
func translateStream(op monitor.Operation) (command.Command, error) {
	switch o := op.(type) {
	case monitor.HotspotHosts:
		cmd := printCmd("/ip/hotspot/host/print", o.Fields)
		cmd.Args["follow"] = ""
		return cmd, nil
	case monitor.InterfaceTraffic:
		if o.Interface == "" {
			return command.Command{}, fmt.Errorf("mikrotik: monitor interface traffic: %w: interface", errMissingField)
		}
		return command.Command{
			Raw:  "/interface/monitor-traffic",
			Args: map[string]string{"interface": o.Interface},
		}, nil
	case monitor.DHCPLeases:
		cmd := printCmd("/ip/dhcp-server/lease/print", o.Fields)
		cmd.Args["follow-only"] = ""
		return cmd, nil
	case monitor.QueueStats:
		if o.QueueName == "" {
			return command.Command{}, fmt.Errorf("mikrotik: monitor queue stats: %w: queue name", errMissingField)
		}
		cmd := printCmd("/queue/simple/print", o.Fields)
		cmd.Args["stats"] = ""
		cmd.Args["?name"] = o.QueueName
		cmd.Args["interval=1s"] = ""
		return cmd, nil
	default:
		return command.Command{}, fmt.Errorf("mikrotik: unsupported monitoring operation %T", op)
	}
}

// createPPPSecret builds the RouterOS /ppp/secret/add command for op. Name and
// Password are mandatory on RouterOS, so an empty one is rejected here (a clear
// error before touching the wire) rather than letting the device fail the add.
// Optional attributes are written only when non-empty so RouterOS applies its
// own defaults for the rest.
func createPPPSecret(op provision.CreatePPPSecret) ([]command.Command, error) {
	if op.Name == "" {
		return nil, fmt.Errorf("mikrotik: create ppp secret: %w: name", errMissingField)
	}
	if op.Password == "" {
		return nil, fmt.Errorf("mikrotik: create ppp secret: %w: password", errMissingField)
	}

	args := map[string]string{
		"name":     op.Name,
		"password": op.Password,
	}
	for key, value := range map[string]string{
		"profile":        op.Profile,
		"service":        op.Service,
		"remote-address": op.RemoteAddress,
		"comment":        op.Comment,
	} {
		if value != "" {
			args[key] = value
		}
	}
	return []command.Command{{Raw: "/ppp/secret/add", Args: args}}, nil
}

// createPPPProfile builds the RouterOS /ppp/profile/add command for op. Name is
// mandatory (a profile is referenced by name), so an empty one is rejected here
// before touching the wire. Optional attributes are written only when non-empty
// so RouterOS applies its own defaults for the rest.
func createPPPProfile(op provision.CreatePPPProfile) ([]command.Command, error) {
	if op.Name == "" {
		return nil, fmt.Errorf("mikrotik: create ppp profile: %w: name", errMissingField)
	}

	args := map[string]string{"name": op.Name}
	for key, value := range map[string]string{
		"rate-limit":     op.RateLimit,
		"local-address":  op.LocalAddress,
		"remote-address": op.RemoteAddress,
		"comment":        op.Comment,
	} {
		if value != "" {
			args[key] = value
		}
	}
	return []command.Command{{Raw: "/ppp/profile/add", Args: args}}, nil
}

// activePPPRemovePath drops an online PPPoE session by subscriber name. Removing
// a session that isn't there (subscriber offline) is a no-op for our intent, not
// a failure — Driver.Execute treats RouterOS's "no such item" trap on THIS path
// as success (see isNoSuchItem). It is a package constant so the driver's Execute
// and changeProfile agree on the exact path that carries that idempotent
// semantics.
const activePPPRemovePath = "/ppp/active/remove"

// changeProfile builds the RouterOS sequence that moves subscriber Username onto
// Profile. Username and Profile are both mandatory. The subscriber is targeted
// by name via RouterOS's numbers= selector (empirically works for /ppp/secret/set
// without a prior .id lookup), so no read step is needed to change the stored
// profile. The trailing /ppp/active/remove drops the online session if one
// exists so the new profile takes effect on redial; when the subscriber is
// offline that command's "no such item" trap is swallowed by Execute, leaving
// the profile change as the net effect (see ChangeProfile doc and
// activePPPRemovePath).
func changeProfile(op provision.ChangeProfile) ([]command.Command, error) {
	if op.Username == "" {
		return nil, fmt.Errorf("mikrotik: change profile: %w: username", errMissingField)
	}
	if op.Profile == "" {
		return nil, fmt.Errorf("mikrotik: change profile: %w: profile", errMissingField)
	}
	return []command.Command{
		{Raw: "/ppp/secret/set", Args: map[string]string{"numbers": op.Username, "profile": op.Profile}},
		{Raw: activePPPRemovePath, Args: map[string]string{"numbers": op.Username}},
	}, nil
}

// isNoSuchItem reports whether err is RouterOS's "no such item" device trap,
// which /ppp/active/remove returns when the targeted subscriber has no active
// session. Only meaningful for activePPPRemovePath, where "no session to drop"
// is a successful no-op rather than an error.
//
// DEVIASI §4: go-routeros surfaces device traps as plain untyped errors with no
// sentinel to match via errors.Is, so substring-matching the device message is
// the only option here; kept scoped to the single idempotent case that needs it.
func isNoSuchItem(err error) bool {
	return err != nil && strings.Contains(err.Error(), "no such item")
}

// printCmd builds a RouterOS "print" command for path, optionally projecting
// only fields via .proplist. Omitting .proplist entirely makes RouterOS return
// every field, which toResult then surfaces raw in Result.Rows — so an empty
// fields slice means "all fields, raw".
func printCmd(path string, fields []string) command.Command {
	args := map[string]string{}
	if len(fields) > 0 {
		args[".proplist"] = strings.Join(fields, ",")
	}
	return command.Command{Raw: path, Args: args}
}

// streamingBasePaths lists RouterOS API paths that are ALWAYS streaming
// regardless of arguments — the device keeps sending "!re" sentences
// until explicitly cancelled (or, for /ping with a count, until it
// finishes on its own).
// TODO: extend with other natively-streaming commands as they're actually
// needed (e.g. /tool/torch, /tool/sniffer) — not added now, per agreed scope.
var streamingBasePaths = map[string]bool{
	"/ping":                      true,
	"/interface/monitor-traffic": true,
}

// streamingFlagArgs lists Args keys that turn an otherwise one-shot command
// (typically a /.../print) into a streaming one: RouterOS's "print follow",
// "print follow-only", and "print interval=1s" (or any interval) all keep
// the connection open and keep sending rows instead of returning once.
var streamingFlagArgs = []string{"follow", "follow-only", "interval"}

// isStreamingCommand reports whether cmd must be run via Driver.Stream
// (Listen — non-blocking, event-driven, no polling) rather than
// Driver.Execute (Run — blocking, returns exactly once). Execute refuses
// streaming commands outright: running a bare /ping through RunArgsContext
// would block forever waiting for a "!done" that a streaming command never
// sends on its own.
func isStreamingCommand(cmd command.Command) bool {
	if streamingBasePaths[cmd.Raw] {
		return true
	}
	for _, key := range streamingFlagArgs {
		if _, ok := cmd.Args[key]; ok {
			return true
		}
	}
	return false
}

// bareWords lists Args keys that RouterOS expects as a standalone word with no
// "=" prefix and no value: the streaming flags follow/follow-only and the
// stats modifier on /queue/simple/print. Whatever is stored in cmd.Args[key]
// for these is ignored — only the key itself is written.
var bareWords = map[string]bool{
	"follow":      true,
	"follow-only": true,
	"stats":       true,
}

// buildArgs converts a vendor-agnostic command.Command into the sentence
// go-routeros expects: Raw first, then one word per Args entry. RouterOS API
// distinguishes three word forms, all produced here from cmd.Args:
//   - bare words (no "=", no value, e.g. "follow"/"stats"): keys in bareWords.
//   - query words ("?key=value", filter the result set): keys prefixed with
//     "?" in cmd.Args (e.g. "?name") — the "?" is already part of the key, so
//     it is written through verbatim, not re-prefixed with "=".
//   - attribute words ("=key=value", or bare "=key" when the value is empty):
//     everything else.
func buildArgs(cmd command.Command) []string {
	args := make([]string, 0, 1+len(cmd.Args))
	args = append(args, cmd.Raw)
	for key, value := range cmd.Args {
		switch {
		case bareWords[key]:
			args = append(args, key)
		case strings.HasPrefix(key, "?"):
			if value == "" {
				args = append(args, key)
			} else {
				args = append(args, key+"="+value)
			}
		case value == "":
			args = append(args, "="+key)
		default:
			args = append(args, "="+key+"="+value)
		}
	}
	return args
}

// toResult converts a go-routeros *Reply (from Driver.Execute) into a
// vendor-agnostic command.Result — one Rows entry per "!re" sentence
// returned, so callers see every row a multi-row /print produced, not
// only the first.
func toResult(reply *routeros.Reply) command.Result {
	if reply == nil {
		return command.Result{}
	}
	rows := make([]map[string]string, 0, len(reply.Re))
	for _, sen := range reply.Re {
		rows = append(rows, sen.Map)
	}
	return command.Result{Rows: rows}
}
