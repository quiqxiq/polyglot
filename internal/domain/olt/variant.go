package olt

import (
	"regexp"
	"strings"
)

// Variant adalah varian perangkat OLT HiOSO EPON yang didukung.
type Variant string

const (
	VariantUnknown Variant = "UNKNOWN"
	VariantHA7302  Variant = "HA7302"
	VariantHA7304  Variant = "HA7304"
)

// Capabilities merangkum perilaku per varian (dialek CLI & cara rename).
type Capabilities struct {
	Variant Variant
	// PONCount adalah jumlah port PON default untuk varian ini.
	PONCount int
	// FlatLLID true bila index ONU memakai LLID datar (1..128) — HA7302.
	FlatLLID bool
	// RenameViaSNMP true bila rename ONU harus lewat SNMP SET, bukan CLI.
	RenameViaSNMP bool
}

// CapabilitiesFor mengembalikan capability matrix per varian.
func CapabilitiesFor(v Variant) Capabilities {
	switch v {
	case VariantHA7302:
		return Capabilities{Variant: VariantHA7302, PONCount: 2, FlatLLID: true, RenameViaSNMP: true}
	case VariantHA7304:
		return Capabilities{Variant: VariantHA7304, PONCount: 4, FlatLLID: false, RenameViaSNMP: false}
	default:
		return Capabilities{Variant: VariantUnknown, PONCount: 2}
	}
}

var ha7302Re = regexp.MustCompile(`(?i)ha-?7302`)

// DetectVariant menebak varian dari sysObjectID/sysDescr/firmware signature.
// Urutan: firmware signature HA7302 → sysDescr/sysObjectID → default HA7304
// bila enterprise HiOSO (25355) terdeteksi tanpa petunjuk lain.
func DetectVariant(sysObjectID, sysDescr, firmwareSignature string) Variant {
	hay := strings.Join([]string{sysObjectID, sysDescr, firmwareSignature}, " ")
	switch {
	case ha7302Re.MatchString(hay):
		return VariantHA7302
	case strings.Contains(hay, "25355"):
		return VariantHA7304
	default:
		return VariantUnknown
	}
}

var onuNameRe = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,32}$`)

// ValidateONUName memastikan label ONU aman untuk CLI/SNMP OLT
// (tanpa spasi/tanda kutip; maks 32 karakter).
func ValidateONUName(name string) error {
	if !onuNameRe.MatchString(name) {
		return ErrInvalidInput
	}
	return nil
}

// ValidatePosition memvalidasi PON/ONU index terhadap capability varian.
// HA7302 memakai LLID datar (satu domain index 1..128) sehingga ponIndex
// wajib 1; HA7304 memakai pasangan {PON}.{ONU}.
func ValidatePosition(cap Capabilities, ponIndex, onuIndex int) error {
	if cap.FlatLLID {
		if ponIndex != 1 || onuIndex < 1 || onuIndex > 128 {
			return ErrInvalidInput
		}
		return nil
	}
	if ponIndex < 1 || ponIndex > cap.PONCount {
		return ErrInvalidInput
	}
	if onuIndex < 1 || onuIndex > 128 {
		return ErrInvalidInput
	}
	return nil
}
