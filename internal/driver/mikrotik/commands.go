package mikrotik

import (
	"fmt"
	"strings"

	"github.com/quiqxiq/goros/v4"
	"github.com/quiqxiq/goros/v4/transport"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// destructivePaths lists RouterOS API paths considered destructive — these
// always require HITL approval regardless of caller.
// TODO: fill in the full list per Polyglot-Architecture.md §5.3.
var destructivePaths = map[string]bool{
	"/system/reboot":              true,
	"/system/reset-configuration": true,
}

// operationMap translates abstract Operations to RouterOS API paths.
// TODO: add more operations as new usecases need them.
var operationMap = map[command.Operation]command.Command{
	command.OpGetStatus: {Raw: "/system/resource/print"},
	command.OpReboot:    {Raw: "/system/reboot"},
}

// Classify reports the risk class of cmd according to RouterOS API conventions.
func Classify(cmd command.Command) command.Class {
	if destructivePaths[cmd.Raw] {
		return command.ClassDestructive
	}
	return command.ClassReadOnly
}

// Translate maps an abstract Operation to a RouterOS-native Command.
func Translate(op command.Operation) (command.Command, error) {
	cmd, ok := operationMap[op]
	if !ok {
		return command.Command{}, fmt.Errorf("mikrotik: unsupported operation %q", op)
	}
	return cmd, nil
}

// streamingBasePaths lists RouterOS API paths that are ALWAYS streaming
// regardless of arguments — the device keeps sending "!re" sentences
// until explicitly cancelled (or, for /ping with a count, until it
// finishes on its own).
//
// Note: /interface/monitor-traffic is NOT in this map because it supports
// two modes depending on whether the "once" arg is present:
//   - without "once": streaming (use Driver.Stream via NewMonitorTrafficStreamCommand)
//   - with "once": one-shot (use Driver.Execute via NewMonitorTrafficOnceCommand)
//
// The special-case is handled in isStreamingCommand below.
// TODO: extend with other natively-streaming commands as they're actually
// needed (e.g. /tool/torch, /tool/sniffer) — not added now, per agreed scope.
var streamingBasePaths = map[string]bool{
	"/ping": true,
}

// streamingFlagArgs lists Args keys that turn an otherwise one-shot command
// (typically a /.../print) into a streaming one: RouterOS's "print follow",
// "print follow-only", and "print interval=1s" (or any interval) all keep
// the connection open and keep sending rows instead of returning once.
var streamingFlagArgs = []string{"follow", "follow-only", "interval", "stats"}

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

	// /interface/monitor-traffic is streaming ONLY when the "once" arg is
	// absent. With "once", RouterOS sends one snapshot then "!done" — safe
	// for Execute. Without "once", it streams indefinitely until cancelled.
	// See iface.go: NewMonitorTrafficOnceCommand vs NewMonitorTrafficStreamCommand.
	if cmd.Raw == "/interface/monitor-traffic" {
		_, hasOnce := cmd.Args["once"]
		return !hasOnce
	}

	for _, key := range streamingFlagArgs {
		if _, ok := cmd.Args[key]; ok {
			// "interval" is only a streaming flag on /print commands (e.g. /queue/simple/print stats interval=1s).
			// On mutation commands like /system/scheduler/add, "interval" is a configuration property.
			if key == "interval" && !strings.HasSuffix(cmd.Raw, "/print") {
				continue
			}
			return true
		}
	}
	return false
}

// buildArgs converts a vendor-agnostic command.Command into the sentence
// go-routeros expects: Raw first, then one word per Args entry. RouterOS
// API distinguishes bare words (no value, e.g. "follow") from attribute
// words ("=key=value") — follow/follow-only never take a value, so they
// are always written bare regardless of what's stored in cmd.Args[key].
func buildArgs(cmd command.Command) []string {
	args := make([]string, 0, 1+len(cmd.Args))
	args = append(args, cmd.Raw)
	for key, value := range cmd.Args {
		switch key {
		case "follow", "follow-only", "stats", "once", "count-only", "detail", "brief":
			args = append(args, key)
		default:
			if strings.HasPrefix(key, "?") {
				if value == "" {
					args = append(args, key)
				} else {
					args = append(args, key+"="+value)
				}
			} else if value == "" {
				args = append(args, "="+key)
			} else {
				args = append(args, "="+key+"="+value)
			}
		}
	}
	return args
}

// toTransportCommand converts a vendor-agnostic command.Command into the
// canonical goros transport.Command used by Driver.Validate's dry-run. The
// mapping mirrors buildArgs' split of Args into sentence words: keys
// prefixed "?" become RouterOS query words, ".proplist" becomes the field
// projection, everything else is a command attribute. Raw is split at its
// last "/" into the menu path and verb, so Gate 2 can discover the schema
// for (path, verb).
func toTransportCommand(cmd command.Command) *transport.Command {
	tc := &transport.Command{Attributes: map[string]string{}}

	idx := strings.LastIndex(cmd.Raw, "/")
	if idx >= 0 {
		tc.Path = cmd.Raw[:idx]
		tc.Verb = cmd.Raw[idx+1:]
	} else {
		tc.Verb = cmd.Raw
	}

	for key, value := range cmd.Args {
		switch {
		case key == ".proplist":
			tc.Proplist = strings.Split(value, ",")
		case strings.HasPrefix(key, "?"):
			q := strings.TrimPrefix(key, "?")
			if value != "" {
				q += "=" + value
			}
			tc.Queries = append(tc.Queries, q)
		default:
			tc.Attributes[key] = value
		}
	}
	return tc
}

// toResult converts a goros *routeros.Reply (from Driver.Execute) into a
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
