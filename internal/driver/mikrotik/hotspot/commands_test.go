// Driver command-builder tests (Fase 0 — PLAN-ISP-CORE-HARDENING.md).
// Test RED mengenkapsulasi bug yang ditemukan pada analisis: /ip/hotspot/user/set
// belum menulis argumen disabled sehingga suspend tidak pernah benar-benar
// menonaktifkan user di router (F1-6).
package hotspot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAddUserCommand_WritesDisabledFlag(t *testing.T) {
	on := NewAddUserCommand(HotspotUserParams{Name: "u1", Profile: "PLAN", Disabled: true})
	assert.Equal(t, "/ip/hotspot/user/add", on.Raw)
	assert.Equal(t, "u1", on.Args["name"])
	assert.Equal(t, "yes", on.Args["disabled"])

	off := NewAddUserCommand(HotspotUserParams{Name: "u1", Profile: "PLAN"})
	assert.NotContains(t, off.Args, "disabled")
}

func TestNewSetUserCommand_WritesProfileAndComment(t *testing.T) {
	cmd := NewSetUserCommand("*1", HotspotUserParams{
		Name: "u1", Profile: "PLAN", Comment: "polyglot:sub-1",
	})
	assert.Equal(t, "/ip/hotspot/user/set", cmd.Raw)
	assert.Equal(t, "*1", cmd.Args[".id"])
	assert.Equal(t, "PLAN", cmd.Args["profile"])
	assert.Equal(t, "polyglot:sub-1", cmd.Args["comment"])
}

func TestNewSetUserCommand_WritesDisabledState(t *testing.T) {
	on := NewSetUserCommand("*1", HotspotUserParams{Name: "u1", Profile: "PLAN", Disabled: true})
	assert.Equal(t, "yes", on.Args["disabled"])

	off := NewSetUserCommand("*1", HotspotUserParams{Name: "u1", Profile: "PLAN", Disabled: false})
	assert.Equal(t, "no", off.Args["disabled"])
}

func TestNewAddUserCommand_WritesMACAndAddress(t *testing.T) {
	cmd := NewAddUserCommand(HotspotUserParams{
		Name: "u1", MACAddress: "AA:BB:CC:DD:EE:FF", Address: "10.0.0.5",
	})
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", cmd.Args["mac-address"])
	assert.Equal(t, "10.0.0.5", cmd.Args["address"])
}

func TestNewSetUserCommand_WritesMACAndAddress(t *testing.T) {
	cmd := NewSetUserCommand("*1", HotspotUserParams{
		Name: "u1", MACAddress: "AA:BB:CC:DD:EE:FF", Address: "10.0.0.5",
	})
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", cmd.Args["mac-address"])
	assert.Equal(t, "10.0.0.5", cmd.Args["address"])
}
