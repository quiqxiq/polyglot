package vpn

import (
	"github.com/quixiq/polyglot/internal/domain/command"
)

// NewShowWireGuardClientConfigCommand builds a command to generate client configuration / QR text.
func NewShowWireGuardClientConfigCommand(peerID string, showSensitive bool) command.Command {
	args := map[string]string{
		".id": peerID,
	}
	if showSensitive {
		args["show-sensitive"] = ""
	}
	return command.Command{
		Raw:  "/interface/wireguard/peers/show-client-config",
		Args: args,
	}
}

// ParseWireGuardClientConfig extracts generated client configuration text from replies.
func ParseWireGuardClientConfig(rows []map[string]string) string {
	if len(rows) == 0 {
		return ""
	}
	return rows[0]["ret"]
}
