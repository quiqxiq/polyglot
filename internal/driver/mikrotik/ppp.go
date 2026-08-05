package mikrotik

import (
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// PPPoESecretParams holds the parameters needed to create or update a
// RouterOS PPPoE secret (/ppp/secret/add or /ppp/secret/set). All RouterOS
// attribute name knowledge is confined here — callers (usecase layer) work
// only with this struct, never with raw RouterOS attribute strings.
//
// Field notes (from RouterOS /ppp/secret/add reference):
//   - Name     : PPPoE username — must be unique on the router.
//   - Password : PPPoE password (plaintext in RouterOS, stored as-is).
//   - Profile  : name of an existing /ppp/profile entry (e.g. "10Mbps").
//                Defaults to "default" if empty.
//   - Service  : which PPP service type this secret is valid for.
//                RouterOS accepts: "any", "async", "l2tp", "ovpn", "pppoe",
//                "pptp", "sstp", "pppoe,l2tp" (comma-separated). For ISP
//                PPPoE-only deployments, use "pppoe".
//   - LocalAddress  : IP address assigned to the router end of the PPP link.
//                     Leave empty to inherit from the profile.
//   - RemoteAddress : IP address assigned to the subscriber's CPE.
//                     Leave empty to inherit from the profile (usually a
//                     pool reference).
//   - Comment  : free-text label. Convention: prefix with "polyglot:<subID>"
//                to enable reconciliation queries later.
//   - Disabled : when true, the secret exists but the subscriber cannot log in.
type PPPoESecretParams struct {
	Name          string
	Password      string
	Profile       string
	Service       string
	LocalAddress  string
	RemoteAddress string
	Comment       string
	Disabled      bool
}

// PPPoESecret represents one row returned by /ppp/secret/print. Only fields
// that are stable across RouterOS versions and genuinely useful to the
// usecase layer are mapped — the raw Rows map is still available from
// command.Result if callers need something obscure.
//
// Field notes (from RouterOS /ppp/secret/print output):
//   - RosID         : internal RouterOS ID (e.g. "*1", "*2"). Required for
//                     /ppp/secret/set and /ppp/secret/remove [find .id=<ID>].
//   - Name          : PPPoE username.
//   - Profile       : active profile name.
//   - Service       : service type string.
//   - LocalAddress  : configured local IP (empty = inherited from profile).
//   - RemoteAddress : configured remote IP (empty = pool/inherited).
//   - Comment       : free-text comment.
//   - Disabled      : true when the secret is administratively disabled.
//   - LastLoggedOut : last disconnect timestamp as a RouterOS time string
//                     (e.g. "jan/02/2006 15:04:05"). Empty if never connected.
//   - CallerID      : MAC address or other caller-id of the last connection.
type PPPoESecret struct {
	RosID         string // RouterOS internal .id, needed for set/remove
	Name          string
	Profile       string
	Service       string
	LocalAddress  string
	RemoteAddress string
	Comment       string
	Disabled      bool
	LastLoggedOut string
	CallerID      string
}

// NewAddPPPoESecretCommand builds the command.Command for /ppp/secret/add
// from typed params. The returned Command is ready to pass to Driver.Execute —
// no RouterOS attribute strings need to appear outside this package.
//
// Profile defaults to "default" and Service to "pppoe" when empty, matching
// RouterOS defaults so callers don't have to remember them.
func NewAddPPPoESecretCommand(p PPPoESecretParams) command.Command {
	profile := p.Profile
	if profile == "" {
		profile = "default"
	}
	service := p.Service
	if service == "" {
		service = "pppoe"
	}

	args := map[string]string{
		"name":     p.Name,
		"password": p.Password,
		"profile":  profile,
		"service":  service,
	}
	if p.LocalAddress != "" {
		args["local-address"] = p.LocalAddress
	}
	if p.RemoteAddress != "" {
		args["remote-address"] = p.RemoteAddress
	}
	if p.Comment != "" {
		args["comment"] = p.Comment
	}
	if p.Disabled {
		args["disabled"] = "yes"
	}

	return command.Command{
		Raw:  "/ppp/secret/add",
		Args: args,
	}
}

// NewSetPPPoESecretCommand builds the command.Command for /ppp/secret/set
// (update an existing secret by its RouterOS ID). rosID must be the .id
// value from a previous /ppp/secret/print result (e.g. "*1"). Only non-empty
// fields in params are updated — empty string means "leave unchanged".
//
// Note: RouterOS /ppp/secret/set uses [find .id=<rosID>] syntax to target
// the entry. Callers must have a valid rosID from a prior print — this is
// intentional: usecase layer is expected to read before write for updates.
func NewSetPPPoESecretCommand(rosID string, p PPPoESecretParams) command.Command {
	args := map[string]string{
		"numbers": rosID, // go-routeros maps this to [find .id=<rosID>]
	}
	if p.Password != "" {
		args["password"] = p.Password
	}
	if p.Profile != "" {
		args["profile"] = p.Profile
	}
	if p.Service != "" {
		args["service"] = p.Service
	}
	if p.LocalAddress != "" {
		args["local-address"] = p.LocalAddress
	}
	if p.RemoteAddress != "" {
		args["remote-address"] = p.RemoteAddress
	}
	if p.Comment != "" {
		args["comment"] = p.Comment
	}
	// Disabled is a boolean flag — only set explicitly if caller intends to
	// change it. Use SetPPPoESecretEnabled / SetPPPoESecretDisabled helpers
	// below for clearer intent at the call site.

	return command.Command{
		Raw:  "/ppp/secret/set",
		Args: args,
	}
}

// NewRemovePPPoESecretCommand builds the command.Command for
// /ppp/secret/remove targeting one entry by its RouterOS ID.
// Classified as ClassDestructive (see destructivePaths in commands.go).
func NewRemovePPPoESecretCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ppp/secret/remove",
		Args: map[string]string{"numbers": rosID},
	}
}

// NewPrintPPPoESecretsCommand builds the command.Command for
// /ppp/secret/print. Use nameFilter to limit results to one specific
// username; pass empty string to list all secrets.
//
// The returned result's Rows are parseable via ParsePPPoESecrets below.
//
// RouterOS /ppp/secret/print returns ALL fields by default — no explicit
// proplist is set here so callers always get the full picture.
func NewPrintPPPoESecretsCommand(nameFilter string) command.Command {
	args := map[string]string{}
	if nameFilter != "" {
		// RouterOS where-clause syntax via the go-routeros sentence protocol.
		// ?name=<nameFilter> tells RouterOS to only return entries where
		// the "name" attribute exactly matches — efficient, single-round-trip.
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/ppp/secret/print",
		Args: args,
	}
}

// NewStreamPPPoESecretsCommand builds the command.Command for
// /ppp/secret/print follow, streaming PPPoE subscriber secret updates.
func NewStreamPPPoESecretsCommand(nameFilter string) command.Command {
	args := map[string]string{
		"follow": "",
	}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/ppp/secret/print",
		Args: args,
	}
}

// ParsePPPoESecrets converts command.Result rows (from Driver.Execute on a
// /ppp/secret/print command) into typed PPPoESecret values. Rows that are
// missing the mandatory ".id" or "name" fields are silently skipped — this
// matches how RouterOS sometimes returns incomplete rows for entries that are
// currently being modified or are in an inconsistent state.
func ParsePPPoESecrets(result command.Result) []PPPoESecret {
	secrets := make([]PPPoESecret, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue // skip malformed rows, not an error
		}
		secrets = append(secrets, PPPoESecret{
			RosID:         id,
			Name:          name,
			Profile:       row["profile"],
			Service:       row["service"],
			LocalAddress:  row["local-address"],
			RemoteAddress: row["remote-address"],
			Comment:       row["comment"],
			Disabled:      strings.EqualFold(row["disabled"], "true"),
			LastLoggedOut: row["last-logged-out"],
			CallerID:      row["caller-id"],
		})
	}
	return secrets
}

// FindPPPoESecretRosID looks up the RouterOS internal ID for a PPPoE secret
// by username. Returns ("", ErrSecretNotFound) when the device has no secret
// with that name. Intended as a helper for usecases that need to do a
// read-before-write for /ppp/secret/set or /ppp/secret/remove.
func FindPPPoESecretRosID(result command.Result, name string) (string, error) {
	for _, s := range ParsePPPoESecrets(result) {
		if s.Name == name {
			return s.RosID, nil
		}
	}
	return "", fmt.Errorf("%w: %q", ErrSecretNotFound, name)
}
