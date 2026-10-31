package genericcli

import (
	"fmt"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// Catalog holds vendor-specific command knowledge for a device driven
// through the generic scrapligo-based CLI engine (genericssh,
// generictelnet). Constructing a Catalog — plus a scrapligo platform
// definition (a built-in name like "cisco_iosxe", or a custom YAML path
// under internal/platformdef/) — is how a NEW vendor gets supported
// through these generic drivers WITHOUT writing any new Go code, mirroring
// how scrapligo's OWN platform YAML files add prompt/paging support for a
// vendor without new Go code. See
// docs/adr/0004-generic-cli-driver-scrapligo.md.
type Catalog struct {
	// DestructivePrefixes lists literal command prefixes considered
	// destructive for this vendor — checked with strings.HasPrefix against
	// command.Command.Raw. A match classifies the command as MUTATION.
	DestructivePrefixes []string

	// ReadPrefixes lists literal command prefixes considered read-only for
	// this vendor — checked AFTER DestructivePrefixes. A prefix with a
	// trailing space matches that literal prefix ("show "); without a
	// trailing space it is treated as a whole command word ("dir" matches
	// "dir" and "dir flash:", not "directory"). Commands matching neither
	// list are MUTATION (fail-closed), so a zero-value Catalog classifies
	// everything as MUTATION.
	ReadPrefixes []string

	// Operations maps abstract command.Operation values to this vendor's
	// native CLI command string.
	Operations map[command.Operation]command.Command

	// FailedWhenContains lists substrings that, if present in a command's
	// output, mean the device rejected the command even though the
	// session itself completed normally (e.g. "% Invalid input", "%
	// Unknown command", "Error:"). Passed straight through to scrapligo's
	// own failed-when-contains mechanism (see session.go) — this is
	// genuinely vendor-specific error-message knowledge, so it lives here,
	// not in genericcli's shared engine code.
	FailedWhenContains []string

	// ReadDelay, if non-zero, is passed to scrapligo as
	// options.WithReadDelay — the sleep time between channel reads while
	// scrapligo looks for the expected prompt. Almost every vendor is fine
	// with the library default (unset here). It exists because at least
	// one real, documented vendor quirk needs it: RouterOS echoes the
	// prompt/command twice after a command over SSH, which has caused
	// scrapligo to give up waiting before the second echo arrives (see
	// https://github.com/scrapli/scrapligo/issues/95, and
	// internal/platformdef/mikrotik_routeros.yaml's own comments). This is
	// connection-timing behavior, not command knowledge, but it's exactly
	// as vendor-specific as everything else in Catalog, so it lives here
	// rather than becoming a 5th genericcli.NewSession/genericssh.NewDriver
	// parameter (CLAUDE.md §3.4: more than 4 params needs a struct — this
	// keeps the constructor at 4 by folding it into the one already there).
	ReadDelay time.Duration

	// ReturnChar, if non-empty, is passed to scrapligo as
	// options.WithReturnChar — the byte sequence sent as "Enter" after a
	// command. scrapligo's default is "\n". RouterOS is a documented
	// exception: its CLI only executes a line when it receives a carriage
	// return, so the community scrapli driver sets this to "\r\n"
	// (comms_return_char) for every RouterOS platform. Verified in lab:
	// with the default "\n" the G-Net RouterOS v6 firmware (192.168.233.1)
	// never executes commands and SendCommand times out; with "\r\n" both
	// v6 G-Net and v7 vanilla execute correctly. Leave unset for vendors
	// that accept the scrapligo default.
	ReturnChar string
}

// hardDenyTokens are vendor-agnostic CLI command tokens that are
// catastrophic and must never be executed by the AI pipeline, even with
// approval: reloads, configuration wipes, and flash/factory resets. They
// are checked as standalone tokens (start, end, or space-delimited) so
// vendor-phrased commands such as "system reset-configuration" are caught
// too. The result still carries ModeMutation so policy sees deny before
// mode (domain/command/policy.go precedence: hard deny > mutation > read).
var hardDenyTokens = []string{
	"reload",
	"write erase",
	"erase startup-config",
	"erase flash",
	"format flash",
	"factoryreset",
	"factory-reset",
	"reset-configuration",
	"delete startup-config",
}

// Classify reports the AI Router Ops classification of cmd for this
// vendor. Order: hard-deny tokens, then DestructivePrefixes (MUTATION),
// then ReadPrefixes (READ); anything unlisted is MUTATION (fail-closed).
func (c Catalog) Classify(cmd command.Command) command.Classification {
	lower := strings.ToLower(strings.TrimSpace(cmd.Raw))
	for _, token := range hardDenyTokens {
		if containsToken(lower, token) {
			return command.Classification{Mode: command.ModeMutation, DenyReason: "hard deny for AI: " + token}
		}
	}
	for _, prefix := range c.DestructivePrefixes {
		if hasCmdPrefix(lower, prefix) {
			return command.Classification{Mode: command.ModeMutation}
		}
	}
	for _, prefix := range c.ReadPrefixes {
		if hasCmdPrefix(lower, prefix) {
			return command.Classification{Mode: command.ModeRead}
		}
	}
	return command.Classification{Mode: command.ModeMutation}
}

// hasCmdPrefix reports whether raw (already lowercased) begins with prefix.
// A trailing space means a literal prefix ("show "); otherwise the prefix is
// treated as a whole command word ("dir" matches "dir" and "dir flash:",
// but not "directory"). The prefix itself is lowercased so mixed-case
// catalog entries still match.
func hasCmdPrefix(raw, prefix string) bool {
	prefix = strings.ToLower(prefix)
	if strings.HasSuffix(prefix, " ") {
		return strings.HasPrefix(raw, prefix)
	}
	return raw == prefix || strings.HasPrefix(raw, prefix+" ")
}

// containsToken reports whether raw contains token as a standalone command
// token: at the start, at the end, or delimited by spaces anywhere inside.
func containsToken(raw, token string) bool {
	token = strings.ToLower(token)
	if raw == token || strings.HasPrefix(raw, token+" ") || strings.HasSuffix(raw, " "+token) {
		return true
	}
	return strings.Contains(raw, " "+token+" ")
}

// Translate maps an abstract Operation to this vendor's native Command.
func (c Catalog) Translate(op command.Operation) (command.Command, error) {
	cmd, ok := c.Operations[op]
	if !ok {
		return command.Command{}, fmt.Errorf("genericcli: unsupported operation %q", op)
	}
	return cmd, nil
}
