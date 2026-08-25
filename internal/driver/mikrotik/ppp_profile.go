package mikrotik

import (
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// PPPProfileParams holds the parameters for creating or updating a RouterOS PPP profile.
type PPPProfileParams = port.PPPProfileParams

// PPPProfile represents one row returned by /ppp/profile/print.
type PPPProfile = port.PPPProfile

// IsolirProfileName is the default RouterOS profile name used for
// suspended PPPoE subscribers.
const IsolirProfileName = "isolir"

// IsolirProfileRateLimit gives suspended subscribers just enough bandwidth
// to reach the payment portal. A hard 0/0 would block the redirect entirely.
const IsolirProfileRateLimit = "512k/512k"

// IsolirProfileParams returns a pre-filled PPPProfileParams for the standard
// "isolir" (suspension) profile used in ISP billing. The profile assigns the
// subscriber an address from the dedicated isolation IP pool at a minimal
// rate-limit so firewall NAT rules can redirect web traffic to the payment
// portal instead of dropping everything.
//
// NOTE: RemoteAddress must reference a pool that ALREADY exists on the
// device (RouterOS validates the pool name) — EnsureIsolirInfrastructure
// creates the pool before the profile. SharedUsers is deliberately omitted:
// it is not a valid /ppp/profile attribute on RouterOS 6.x.
//
// Usage: create this profile once on the router on first suspension, then
// reference it by name ("isolir") for subsequent suspensions.
func IsolirProfileParams(isolatedPool string) PPPProfileParams {
	return PPPProfileParams{
		Name:          IsolirProfileName,
		RemoteAddress: isolatedPool,
		RateLimit:     IsolirProfileRateLimit,
		Comment:       "SUSPENDED_PROFILE",
	}
}

// NewPrintPPPProfilesCommand builds the command.Command for /ppp/profile/print.
// Pass a non-empty nameFilter to look up one specific profile by name.
func NewPrintPPPProfilesCommand(nameFilter string) command.Command {
	args := map[string]string{}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/ppp/profile/print",
		Args: args,
	}
}

// NewAddPPPProfileCommand builds the command.Command for /ppp/profile/add.
// Only non-empty fields in params are sent — RouterOS keeps defaults for
// any omitted field.
func NewAddPPPProfileCommand(p PPPProfileParams) command.Command {
	args := map[string]string{
		"name": p.Name,
	}
	setIfNonEmpty(args, "rate-limit", p.RateLimit)
	setIfNonEmpty(args, "local-address", p.LocalAddress)
	setIfNonEmpty(args, "remote-address", p.RemoteAddress)
	setIfNonEmpty(args, "dns-server", p.DNSServer)
	setIfNonEmpty(args, "parent-queue", p.ParentQueue)
	setIfNonEmpty(args, "address-list", p.AddressList)
	setIfNonEmpty(args, "comment", p.Comment)
	setIfNonEmpty(args, "shared-users", p.SharedUsers)
	setIfNonEmpty(args, "only-one", p.OnlyOne)
	setIfNonEmpty(args, "use-mpls", p.UseMPLS)
	setIfNonEmpty(args, "use-compression", p.UseCompression)
	setIfNonEmpty(args, "use-encryption", p.UseEncryption)
	setIfNonEmpty(args, "change-tcp-mss", p.ChangeTCPMSS)
	setIfNonEmpty(args, "bridge-learning", p.BridgeLearning)
	return command.Command{Raw: "/ppp/profile/add", Args: args}
}

// NewSetPPPProfileCommand builds the command.Command for /ppp/profile/set.
// rosID must be the .id from a prior /ppp/profile/print result.
// Only non-empty fields in params are updated.
func NewSetPPPProfileCommand(rosID string, p PPPProfileParams) command.Command {
	args := map[string]string{".id": rosID}
	setIfNonEmpty(args, "name", p.Name)
	setIfNonEmpty(args, "rate-limit", p.RateLimit)
	setIfNonEmpty(args, "local-address", p.LocalAddress)
	setIfNonEmpty(args, "remote-address", p.RemoteAddress)
	setIfNonEmpty(args, "dns-server", p.DNSServer)
	setIfNonEmpty(args, "parent-queue", p.ParentQueue)
	setIfNonEmpty(args, "address-list", p.AddressList)
	setIfNonEmpty(args, "comment", p.Comment)
	setIfNonEmpty(args, "shared-users", p.SharedUsers)
	setIfNonEmpty(args, "only-one", p.OnlyOne)
	setIfNonEmpty(args, "use-mpls", p.UseMPLS)
	setIfNonEmpty(args, "use-compression", p.UseCompression)
	setIfNonEmpty(args, "use-encryption", p.UseEncryption)
	setIfNonEmpty(args, "change-tcp-mss", p.ChangeTCPMSS)
	setIfNonEmpty(args, "bridge-learning", p.BridgeLearning)
	return command.Command{Raw: "/ppp/profile/set", Args: args}
}

// NewRemovePPPProfileCommand builds the command.Command for /ppp/profile/remove.
func NewRemovePPPProfileCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ppp/profile/remove",
		Args: map[string]string{".id": rosID},
	}
}

// ParsePPPProfiles converts command.Result rows from a /ppp/profile/print
// command into typed PPPProfile values. Rows missing ".id" or "name" are skipped.
func ParsePPPProfiles(result command.Result) []PPPProfile {
	profiles := make([]PPPProfile, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		profiles = append(profiles, PPPProfile{
			RosID:          id,
			Name:           name,
			RateLimit:      row["rate-limit"],
			LocalAddress:   row["local-address"],
			RemoteAddress:  row["remote-address"],
			DNSServer:      row["dns-server"],
			ParentQueue:    row["parent-queue"],
			AddressList:    row["address-list"],
			Comment:        row["comment"],
			SharedUsers:    row["shared-users"],
			OnlyOne:        row["only-one"],
			UseMPLS:        row["use-mpls"],
			UseCompression: row["use-compression"],
			UseEncryption:  row["use-encryption"],
			ChangeTCPMSS:   row["change-tcp-mss"],
			BridgeLearning: row["bridge-learning"],
		})
	}
	return profiles
}

// setIfNonEmpty adds key=value to args only when value is not empty.
// Centralised here because every Add/Set builder in this package uses the
// same "send only if provided" pattern from MIKROTIK-COMMAND.md §3.
func setIfNonEmpty(args map[string]string, key, value string) {
	if value != "" {
		args[key] = value
	}
}
