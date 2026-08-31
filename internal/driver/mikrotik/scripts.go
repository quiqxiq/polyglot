package mikrotik

import (
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/device"
)

// BuildPPPOnUpScript generates a RouterOS script for the PPP Profile on-up event.
func BuildPPPOnUpScript(webhookURL, token string) string {
	return fmt.Sprintf(`:local url %q;
:local token %q;
:local mac $"calling-sid";
:local ip $"remote-address";
:local iface $"interface";
:local data ("{\"token\":\"" . $token . "\",\"event\":\"on-up\",\"service\":\"pppoe\",\"user\":\"" . $user . "\",\"ip\":\"" . $ip . "\",\"mac\":\"" . $mac . "\",\"interface\":\"" . $iface . "\"}");
:do {
    /tool fetch url=$url http-method=post http-data=$data http-header-field="Content-Type: application/json" keep-result=no;
} on-error={ :log warning ("Polyglot Webhook: Failed to send on-up for " . $user); };`, webhookURL, token)
}

// BuildPPPOnDownScript generates a RouterOS script for the PPP Profile on-down event.
func BuildPPPOnDownScript(webhookURL, token string) string {
	return fmt.Sprintf(`:local url %q;
:local token %q;
:local iface $"interface";
:local data ("{\"token\":\"" . $token . "\",\"event\":\"on-down\",\"service\":\"pppoe\",\"user\":\"" . $user . "\",\"interface\":\"" . $iface . "\"}");
:do {
    /tool fetch url=$url http-method=post http-data=$data http-header-field="Content-Type: application/json" keep-result=no;
} on-error={ :log warning ("Polyglot Webhook: Failed to send on-down for " . $user); };`, webhookURL, token)
}

// BuildHotspotOnLoginScript generates a RouterOS script for the Hotspot Profile on-login event.
func BuildHotspotOnLoginScript(webhookURL, token string) string {
	return fmt.Sprintf(`:local url %q;
:local token %q;
:local mac $"mac-address";
:local ip $"address";
:local iface $"interface";
:local data ("{\"token\":\"" . $token . "\",\"event\":\"on-login\",\"service\":\"hotspot\",\"user\":\"" . $username . "\",\"ip\":\"" . $ip . "\",\"mac\":\"" . $mac . "\",\"interface\":\"" . $iface . "\"}");
:do {
    /tool fetch url=$url http-method=post http-data=$data http-header-field="Content-Type: application/json" keep-result=no;
} on-error={ :log warning ("Polyglot Webhook: Failed to send on-login for " . $username); };`, webhookURL, token)
}

// BuildHotspotOnLogoutScript generates a RouterOS script for the Hotspot Profile on-logout event.
func BuildHotspotOnLogoutScript(webhookURL, token string) string {
	return fmt.Sprintf(`:local url %q;
:local token %q;
:local mac $"mac-address";
:local ip $"address";
:local data ("{\"token\":\"" . $token . "\",\"event\":\"on-logout\",\"service\":\"hotspot\",\"user\":\"" . $username . "\",\"ip\":\"" . $ip . "\",\"mac\":\"" . $mac . "\"}");
:do {
    /tool fetch url=$url http-method=post http-data=$data http-header-field="Content-Type: application/json" keep-result=no;
} on-error={ :log warning ("Polyglot Webhook: Failed to send on-logout for " . $username); };`, webhookURL, token)
}

// GenerateRouterIntegrationScripts produces all 4 RouterOS integration scripts.
func GenerateRouterIntegrationScripts(webhookURL, token string) device.RouterIntegrationScripts {
	webhookURL = strings.TrimSpace(webhookURL)
	return device.RouterIntegrationScripts{
		PPPOnUpScript:         BuildPPPOnUpScript(webhookURL, token),
		PPPOnDownScript:       BuildPPPOnDownScript(webhookURL, token),
		HotspotOnLoginScript:  BuildHotspotOnLoginScript(webhookURL, token),
		HotspotOnLogoutScript: BuildHotspotOnLogoutScript(webhookURL, token),
		WebhookToken:          token,
	}
}
