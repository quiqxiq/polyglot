package vpn

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// IPsecPeerParams holds configuration for creating or updating an IPsec peer.
type IPsecPeerParams struct {
	Name               string `json:"name"`
	Address            string `json:"address"`
	Profile            string `json:"profile"`
	ExchangeMode       string `json:"exchange_mode"`
	SendInitialContact bool   `json:"send_initial_contact"`
	Disabled           bool   `json:"disabled"`
	Comment            string `json:"comment"`
}

// IPsecProfile represents an IKE Phase 1 profile (/ip/ipsec/profile).
type IPsecProfile struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	HashAlgorithm string `json:"hash_algorithm"`
	EncAlgorithm  string `json:"enc_algorithm"`
	DHGroup       string `json:"dh_group"`
	ProposalCheck string `json:"proposal_check"`
	NatTraversal  bool   `json:"nat_traversal"`
	Lifetime      string `json:"lifetime"`
	Disabled      bool   `json:"disabled"`
}

// IPsecProposal represents an IKE Phase 2 proposal (/ip/ipsec/proposal).
type IPsecProposal struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	AuthAlgorithms string `json:"auth_algorithms"`
	EncAlgorithms  string `json:"enc_algorithms"`
	Lifetime       string `json:"lifetime"`
	PFSGroup       string `json:"pfs_group"`
	Disabled       bool   `json:"disabled"`
	Comment        string `json:"comment"`
}

// IPsecModeConfig represents an IKEv2 Mode-Config / address assignment (/ip/ipsec/mode-config).
type IPsecModeConfig struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Responder           bool   `json:"responder"`
	AddressPool         string `json:"address_pool"`
	Address             string `json:"address"`
	AddressPrefixLength string `json:"address_prefix_length"`
	SplitInclude        string `json:"split_include"`
	SystemDNS           bool   `json:"system_dns"`
	Disabled            bool   `json:"disabled"`
}

// NewAddIPsecPeerCommand creates an IPsec peer.
func NewAddIPsecPeerCommand(p IPsecPeerParams) command.Command {
	args := map[string]string{
		"name":    p.Name,
		"address": p.Address,
	}
	rosutil.SetIfNonEmpty(args, "profile", p.Profile)
	rosutil.SetIfNonEmpty(args, "exchange-mode", p.ExchangeMode)
	rosutil.SetIfNonEmpty(args, "comment", p.Comment)
	if p.SendInitialContact {
		args["send-initial-contact"] = "yes"
	}
	if p.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/ip/ipsec/peer/add",
		Args: args,
	}
}

// NewSetIPsecPeerCommand updates an IPsec peer.
func NewSetIPsecPeerCommand(rosID string, p IPsecPeerParams) command.Command {
	args := map[string]string{
		".id": rosID,
	}
	rosutil.SetIfNonEmpty(args, "name", p.Name)
	rosutil.SetIfNonEmpty(args, "address", p.Address)
	rosutil.SetIfNonEmpty(args, "profile", p.Profile)
	rosutil.SetIfNonEmpty(args, "exchange-mode", p.ExchangeMode)
	args["comment"] = p.Comment
	if p.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{
		Raw:  "/ip/ipsec/peer/set",
		Args: args,
	}
}

// NewRemoveIPsecPeerCommand deletes an IPsec peer.
func NewRemoveIPsecPeerCommand(rosID string) command.Command {
	return command.Command{
		Raw: "/ip/ipsec/peer/remove",
		Args: map[string]string{
			".id": rosID,
		},
	}
}

// NewPrintIPsecProfilesCommand queries IPsec Phase 1 profiles with strict proplist.
func NewPrintIPsecProfilesCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "hash-algorithm", "enc-algorithm", "dh-group", "proposal-check", "nat-traversal", "lifetime", "disabled")
	return command.Command{
		Raw:  "/ip/ipsec/profile/print",
		Args: args,
	}
}

// NewAddIPsecProfileCommand creates an IPsec Phase 1 profile.
func NewAddIPsecProfileCommand(p IPsecProfile) command.Command {
	args := map[string]string{
		"name": p.Name,
	}
	rosutil.SetIfNonEmpty(args, "hash-algorithm", p.HashAlgorithm)
	rosutil.SetIfNonEmpty(args, "enc-algorithm", p.EncAlgorithm)
	rosutil.SetIfNonEmpty(args, "dh-group", p.DHGroup)
	rosutil.SetIfNonEmpty(args, "proposal-check", p.ProposalCheck)
	rosutil.SetIfNonEmpty(args, "lifetime", p.Lifetime)
	if p.NatTraversal {
		args["nat-traversal"] = "yes"
	}
	if p.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/ip/ipsec/profile/add",
		Args: args,
	}
}

// NewRemoveIPsecProfileCommand removes an IPsec profile.
func NewRemoveIPsecProfileCommand(id string) command.Command {
	return command.Command{
		Raw: "/ip/ipsec/profile/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintIPsecProposalsCommand queries IPsec Phase 2 proposals with strict proplist.
func NewPrintIPsecProposalsCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "auth-algorithms", "enc-algorithms", "lifetime", "pfs-group", "disabled", "comment")
	return command.Command{
		Raw:  "/ip/ipsec/proposal/print",
		Args: args,
	}
}

// NewAddIPsecProposalCommand creates an IPsec Phase 2 proposal.
func NewAddIPsecProposalCommand(p IPsecProposal) command.Command {
	args := map[string]string{
		"name": p.Name,
	}
	rosutil.SetIfNonEmpty(args, "auth-algorithms", p.AuthAlgorithms)
	rosutil.SetIfNonEmpty(args, "enc-algorithms", p.EncAlgorithms)
	rosutil.SetIfNonEmpty(args, "lifetime", p.Lifetime)
	rosutil.SetIfNonEmpty(args, "pfs-group", p.PFSGroup)
	rosutil.SetIfNonEmpty(args, "comment", p.Comment)
	if p.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/ip/ipsec/proposal/add",
		Args: args,
	}
}

// NewRemoveIPsecProposalCommand removes an IPsec proposal.
func NewRemoveIPsecProposalCommand(id string) command.Command {
	return command.Command{
		Raw: "/ip/ipsec/proposal/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintIPsecModeConfigsCommand queries IPsec mode configs with strict proplist.
func NewPrintIPsecModeConfigsCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "responder", "address-pool", "address", "address-prefix-length", "split-include", "system-dns", "disabled")
	return command.Command{
		Raw:  "/ip/ipsec/mode-config/print",
		Args: args,
	}
}

// NewAddIPsecModeConfigCommand creates an IPsec mode config.
func NewAddIPsecModeConfigCommand(mc IPsecModeConfig) command.Command {
	args := map[string]string{
		"name": mc.Name,
	}
	if mc.Responder {
		args["responder"] = "yes"
	}
	rosutil.SetIfNonEmpty(args, "address-pool", mc.AddressPool)
	rosutil.SetIfNonEmpty(args, "address", mc.Address)
	rosutil.SetIfNonEmpty(args, "address-prefix-length", mc.AddressPrefixLength)
	rosutil.SetIfNonEmpty(args, "split-include", mc.SplitInclude)
	if mc.SystemDNS {
		args["system-dns"] = "yes"
	}
	if mc.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/ip/ipsec/mode-config/add",
		Args: args,
	}
}

// NewRemoveIPsecModeConfigCommand removes an IPsec mode config.
func NewRemoveIPsecModeConfigCommand(id string) command.Command {
	return command.Command{
		Raw: "/ip/ipsec/mode-config/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// ParseIPsecProfiles converts reply rows into IPsecProfile slice.
func ParseIPsecProfiles(rows []map[string]string) []IPsecProfile {
	out := make([]IPsecProfile, 0, len(rows))
	for _, r := range rows {
		out = append(out, IPsecProfile{
			ID:            r[".id"],
			Name:          r["name"],
			HashAlgorithm: r["hash-algorithm"],
			EncAlgorithm:  r["enc-algorithm"],
			DHGroup:       r["dh-group"],
			ProposalCheck: r["proposal-check"],
			NatTraversal:  strings.EqualFold(r["nat-traversal"], "yes"),
			Lifetime:      r["lifetime"],
			Disabled:      strings.EqualFold(r["disabled"], "true") || strings.EqualFold(r["disabled"], "yes"),
		})
	}
	return out
}

// ParseIPsecProposals converts reply rows into IPsecProposal slice.
func ParseIPsecProposals(rows []map[string]string) []IPsecProposal {
	out := make([]IPsecProposal, 0, len(rows))
	for _, r := range rows {
		out = append(out, IPsecProposal{
			ID:             r[".id"],
			Name:           r["name"],
			AuthAlgorithms: r["auth-algorithms"],
			EncAlgorithms:  r["enc-algorithms"],
			Lifetime:       r["lifetime"],
			PFSGroup:       r["pfs-group"],
			Disabled:       strings.EqualFold(r["disabled"], "true") || strings.EqualFold(r["disabled"], "yes"),
			Comment:        r["comment"],
		})
	}
	return out
}

// ParseIPsecModeConfigs converts reply rows into IPsecModeConfig slice.
func ParseIPsecModeConfigs(rows []map[string]string) []IPsecModeConfig {
	out := make([]IPsecModeConfig, 0, len(rows))
	for _, r := range rows {
		out = append(out, IPsecModeConfig{
			ID:                  r[".id"],
			Name:                r["name"],
			Responder:           strings.EqualFold(r["responder"], "yes"),
			AddressPool:         r["address-pool"],
			Address:             r["address"],
			AddressPrefixLength: r["address-prefix-length"],
			SplitInclude:        r["split-include"],
			SystemDNS:           strings.EqualFold(r["system-dns"], "yes"),
			Disabled:            strings.EqualFold(r["disabled"], "true") || strings.EqualFold(r["disabled"], "yes"),
		})
	}
	return out
}
