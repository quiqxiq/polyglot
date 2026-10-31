package vpn

import (
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// IPsecIdentity represents an IPsec peer identity (/ip/ipsec/identity).
type IPsecIdentity struct {
	ID         string `json:"id"`
	Peer       string `json:"peer"`
	AuthMethod string `json:"auth_method"`
	Disabled   bool   `json:"disabled"`
	Comment    string `json:"comment"`
}

// IPsecPolicy represents an IPsec traffic selector policy (/ip/ipsec/policy).
type IPsecPolicy struct {
	ID         string `json:"id"`
	Peer       string `json:"peer"`
	Tunnel     bool   `json:"tunnel"`
	SrcAddress string `json:"src_address"`
	DstAddress string `json:"dst_address"`
	Action     string `json:"action"`
	Proposal   string `json:"proposal"`
	Disabled   bool   `json:"disabled"`
	Comment    string `json:"comment"`
}

// IPsecInstalledSA represents an active Security Association (/ip/ipsec/installed-sa).
type IPsecInstalledSA struct {
	ID            string `json:"id"`
	SrcAddress    string `json:"src_address"`
	DstAddress    string `json:"dst_address"`
	SPI           string `json:"spi"`
	AuthAlgorithm string `json:"auth_algorithm"`
	EncrAlgorithm string `json:"encr_algorithm"`
	State         string `json:"state"`
}

// NewPrintIPsecIdentitiesCommand queries IPsec identities with strict proplist.
func NewPrintIPsecIdentitiesCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "peer", "auth-method", "disabled", "comment")
	return command.Command{
		Raw:  "/ip/ipsec/identity/print",
		Args: args,
	}
}

// NewAddIPsecIdentityCommand creates an IPsec identity. Note: secrets are never read back.
func NewAddIPsecIdentityCommand(ident IPsecIdentity, secret string) command.Command {
	args := map[string]string{
		"peer": ident.Peer,
	}
	if ident.AuthMethod != "" {
		args["auth-method"] = ident.AuthMethod
	}
	if secret != "" {
		args["secret"] = secret
	}
	if ident.Comment != "" {
		args["comment"] = ident.Comment
	}
	if ident.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/ip/ipsec/identity/add",
		Args: args,
	}
}

// NewRemoveIPsecIdentityCommand removes an IPsec identity by ID.
func NewRemoveIPsecIdentityCommand(id string) command.Command {
	return command.Command{
		Raw: "/ip/ipsec/identity/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintIPsecPoliciesCommand queries IPsec policies with strict proplist.
func NewPrintIPsecPoliciesCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "peer", "tunnel", "src-address", "dst-address", "action", "proposal", "disabled", "comment")
	return command.Command{
		Raw:  "/ip/ipsec/policy/print",
		Args: args,
	}
}

// NewAddIPsecPolicyCommand creates an IPsec traffic selector policy.
func NewAddIPsecPolicyCommand(p IPsecPolicy) command.Command {
	args := map[string]string{
		"peer":        p.Peer,
		"src-address": p.SrcAddress,
		"dst-address": p.DstAddress,
	}
	if p.Tunnel {
		args["tunnel"] = "yes"
	} else {
		args["tunnel"] = "no"
	}
	if p.Action != "" {
		args["action"] = p.Action
	}
	if p.Proposal != "" {
		args["proposal"] = p.Proposal
	}
	if p.Comment != "" {
		args["comment"] = p.Comment
	}
	if p.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/ip/ipsec/policy/add",
		Args: args,
	}
}

// NewMoveIPsecPolicyCommand reorders an IPsec policy before destination.
func NewMoveIPsecPolicyCommand(id, destination string) command.Command {
	args := map[string]string{
		"numbers": id,
	}
	if destination != "" {
		args["destination"] = destination
	}
	return command.Command{
		Raw:  "/ip/ipsec/policy/move",
		Args: args,
	}
}

// NewRemoveIPsecPolicyCommand removes an IPsec policy by ID.
func NewRemoveIPsecPolicyCommand(id string) command.Command {
	return command.Command{
		Raw: "/ip/ipsec/policy/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintIPsecInstalledSAsCommand queries installed Security Associations.
func NewPrintIPsecInstalledSAsCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "src-address", "dst-address", "spi", "auth-algorithm", "encr-algorithm", "state")
	return command.Command{
		Raw:  "/ip/ipsec/installed-sa/print",
		Args: args,
	}
}

// ParseIPsecIdentities converts reply rows into IPsecIdentity slice.
func ParseIPsecIdentities(rows []map[string]string) []IPsecIdentity {
	out := make([]IPsecIdentity, 0, len(rows))
	for _, r := range rows {
		out = append(out, IPsecIdentity{
			ID:         r[".id"],
			Peer:       r["peer"],
			AuthMethod: r["auth-method"],
			Disabled:   r["disabled"] == "true",
			Comment:    r["comment"],
		})
	}
	return out
}

// ParseIPsecPolicies converts reply rows into IPsecPolicy slice.
func ParseIPsecPolicies(rows []map[string]string) []IPsecPolicy {
	out := make([]IPsecPolicy, 0, len(rows))
	for _, r := range rows {
		out = append(out, IPsecPolicy{
			ID:         r[".id"],
			Peer:       r["peer"],
			Tunnel:     r["tunnel"] == "yes" || r["tunnel"] == "true",
			SrcAddress: r["src-address"],
			DstAddress: r["dst-address"],
			Action:     r["action"],
			Proposal:   r["proposal"],
			Disabled:   r["disabled"] == "true",
			Comment:    r["comment"],
		})
	}
	return out
}

// ParseIPsecInstalledSAs converts reply rows into IPsecInstalledSA slice.
func ParseIPsecInstalledSAs(rows []map[string]string) []IPsecInstalledSA {
	out := make([]IPsecInstalledSA, 0, len(rows))
	for _, r := range rows {
		out = append(out, IPsecInstalledSA{
			ID:            r[".id"],
			SrcAddress:    r["src-address"],
			DstAddress:    r["dst-address"],
			SPI:           r["spi"],
			AuthAlgorithm: r["auth-algorithm"],
			EncrAlgorithm: r["encr-algorithm"],
			State:         r["state"],
		})
	}
	return out
}
