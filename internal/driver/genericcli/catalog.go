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
	// command.Command.Raw.
	DestructivePrefixes []string

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
}

// Classify reports the risk class of cmd according to c's destructive
// prefix list. Unlisted commands default to read-only — same fail-safe
// posture as every other vendor's Classify in this project.
func (c Catalog) Classify(cmd command.Command) command.Class {
	for _, prefix := range c.DestructivePrefixes {
		if strings.HasPrefix(cmd.Raw, prefix) {
			return command.ClassDestructive
		}
	}
	return command.ClassReadOnly
}

// Translate maps an abstract Operation to this vendor's native Command.
func (c Catalog) Translate(op command.Operation) (command.Command, error) {
	cmd, ok := c.Operations[op]
	if !ok {
		return command.Command{}, fmt.Errorf("genericcli: unsupported operation %q", op)
	}
	return cmd, nil
}
