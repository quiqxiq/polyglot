package wifi

import (
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// WirelessSecurityProfile represents a legacy wireless security profile (/interface/wireless/security-profiles).
type WirelessSecurityProfile struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Mode                string `json:"mode"`
	AuthenticationTypes string `json:"authentication_types"`
	UnicastCiphers      string `json:"unicast_ciphers"`
	GroupCiphers        string `json:"group_ciphers"`
	Comment             string `json:"comment"`
}

// NewPrintWirelessSecurityProfilesCommand queries legacy wireless security profiles.
// PSK secrets are deliberately not in proplist to avoid credential exposure.
func NewPrintWirelessSecurityProfilesCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "mode", "authentication-types", "unicast-ciphers", "group-ciphers", "comment")
	return command.Command{
		Raw:  "/interface/wireless/security-profiles/print",
		Args: args,
	}
}

// NewAddWirelessSecurityProfileCommand creates a legacy wireless security profile.
func NewAddWirelessSecurityProfileCommand(p WirelessSecurityProfile, wpaKey, wpa2Key string) command.Command {
	args := map[string]string{
		"name": p.Name,
	}
	rosutil.SetIfNonEmpty(args, "mode", p.Mode)
	rosutil.SetIfNonEmpty(args, "authentication-types", p.AuthenticationTypes)
	rosutil.SetIfNonEmpty(args, "unicast-ciphers", p.UnicastCiphers)
	rosutil.SetIfNonEmpty(args, "group-ciphers", p.GroupCiphers)
	rosutil.SetIfNonEmpty(args, "wpa-pre-shared-key", wpaKey)
	rosutil.SetIfNonEmpty(args, "wpa2-pre-shared-key", wpa2Key)
	rosutil.SetIfNonEmpty(args, "comment", p.Comment)
	return command.Command{
		Raw:  "/interface/wireless/security-profiles/add",
		Args: args,
	}
}

// NewSetWirelessSecurityProfileCommand updates a legacy wireless security profile.
func NewSetWirelessSecurityProfileCommand(id string, p WirelessSecurityProfile, wpaKey, wpa2Key string) command.Command {
	args := map[string]string{
		".id": id,
	}
	rosutil.SetIfNonEmpty(args, "name", p.Name)
	rosutil.SetIfNonEmpty(args, "mode", p.Mode)
	rosutil.SetIfNonEmpty(args, "authentication-types", p.AuthenticationTypes)
	rosutil.SetIfNonEmpty(args, "unicast-ciphers", p.UnicastCiphers)
	rosutil.SetIfNonEmpty(args, "group-ciphers", p.GroupCiphers)
	rosutil.SetIfNonEmpty(args, "wpa-pre-shared-key", wpaKey)
	rosutil.SetIfNonEmpty(args, "wpa2-pre-shared-key", wpa2Key)
	args["comment"] = p.Comment
	return command.Command{
		Raw:  "/interface/wireless/security-profiles/set",
		Args: args,
	}
}

// NewRemoveWirelessSecurityProfileCommand removes a legacy wireless security profile.
func NewRemoveWirelessSecurityProfileCommand(id string) command.Command {
	return command.Command{
		Raw: "/interface/wireless/security-profiles/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// ParseWirelessSecurityProfiles converts reply rows into WirelessSecurityProfile slice.
func ParseWirelessSecurityProfiles(rows []map[string]string) []WirelessSecurityProfile {
	out := make([]WirelessSecurityProfile, 0, len(rows))
	for _, r := range rows {
		out = append(out, WirelessSecurityProfile{
			ID:                  r[".id"],
			Name:                r["name"],
			Mode:                r["mode"],
			AuthenticationTypes: r["authentication-types"],
			UnicastCiphers:      r["unicast-ciphers"],
			GroupCiphers:        r["group-ciphers"],
			Comment:             r["comment"],
		})
	}
	return out
}
