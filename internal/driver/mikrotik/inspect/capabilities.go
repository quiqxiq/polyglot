package inspect

import (
	"strings"
)

// AIMode describes how much RouterOS AI support is available.
type AIMode string

const (
	AIModeFull     AIMode = "full"
	AIModeDegraded AIMode = "degraded"
)

// Capabilities is the cached AI-support profile for one RouterOS device.
type Capabilities struct {
	Version          string `json:"version"`
	Major            int    `json:"major"`
	Minor            int    `json:"minor"`
	Patch            int    `json:"patch"`
	InspectSupported bool   `json:"inspect_supported"`
	AIMode           AIMode `json:"ai_mode"`
	// ScriptInspectSafe is false for RouterOS versions where request=syntax/
	// completion on scripting paths can deadlock the REST/API server
	// (<= 7.20.8; fixed in 7.21.4). request=child remains safe.
	ScriptInspectSafe bool   `json:"script_inspect_safe"`
	Warning           string `json:"warning,omitempty"`
}

// BuildCapabilities derives capabilities from a version string.
func BuildCapabilities(version string, major, minor, patch int) Capabilities {
	c := Capabilities{
		Version:           version,
		Major:             major,
		Minor:             minor,
		Patch:             patch,
		InspectSupported:  major >= 7,
		ScriptInspectSafe: versionAtLeast(major, minor, patch, 7, 21, 4),
	}
	if c.InspectSupported {
		c.AIMode = AIModeFull
	} else {
		c.AIMode = AIModeDegraded
		c.Warning = "RouterOS AI features are optimized for RouterOS v7. /console/inspect is v7-only; on v6 syntax validation is disabled (degraded mode). Every mutation still requires human approval and a restore point."
	}
	return c
}

// ScriptPathsUnsafe reports whether syntax/completion inspection on a script
// path must be skipped for these capabilities. Conservative: any scripting
// keyword in the path disables syntax/completion on unsafe versions.
func (c Capabilities) ScriptPathsUnsafe(path string) bool {
	if c.ScriptInspectSafe {
		return false
	}
	lower := strings.ToLower(path)
	for _, marker := range []string{"where", "do", "else", "rule", "command", "on-error"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func versionAtLeast(maj, min, patch, wantMaj, wantMin, wantPatch int) bool {
	if maj != wantMaj {
		return maj > wantMaj
	}
	if min != wantMin {
		return min > wantMin
	}
	return patch >= wantPatch
}
